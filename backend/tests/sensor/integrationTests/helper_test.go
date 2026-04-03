package sensor_integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/gateway"
	clouddb "backend/internal/infra/database/cloud_db/connection"
	sensordb "backend/internal/infra/database/sensor_db"
	natsutils "backend/internal/infra/nats"
	"backend/internal/sensor"
	sharedCrypto "backend/internal/shared/crypto"
	"backend/internal/shared/identity"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"gorm.io/gorm"
)

func mustGenerateJWTForRequester(t *testing.T, jwtManager sharedCrypto.AuthTokenManager, requester identity.Requester) string {
	t.Helper()

	jwt, err := jwtManager.GenerateForRequester(requester)
	if err != nil {
		t.Fatalf("failed to generate JWT: %v", err)
	}
	if jwt == "" {
		t.Fatalf("generated JWT is empty")
	}

	return jwt
}

func mustJSONBody(t *testing.T, payload any) *bytes.Reader {
	t.Helper()

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal json payload: %v", err)
	}

	return bytes.NewReader(jsonBody)
}

func authHeader(jwt string) http.Header {
	header := http.Header{}
	header.Set("Authorization", "Bearer "+jwt)
	return header
}

func preSetupCreateGateway(
	gatewayID string,
	name string,
) func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool {
	return preSetupCreateGatewayWithTenant(gatewayID, name, nil)
}

func preSetupCreateGatewayWithTenant(
	gatewayID string,
	name string,
	tenantID *string,
) func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool {
	return func(
		cloudDB clouddb.CloudDBConnection,
		_ sensordb.SensorDBConnection,
		_ *nats.Conn,
		_ natsutils.NatsTestConnection,
		_ jetstream.JetStream,
		_ jetstream.JetStream,
	) bool {
		db := (*gorm.DB)(cloudDB)
		entity := gateway.GatewayEntity{
			ID:       gatewayID,
			Name:     name,
			TenantId: tenantID,
			Status:   string(gateway.GATEWAY_STATUS_ACTIVE),
		}
		return db.Create(&entity).Error == nil
	}
}

func preSetupCreateSensor(
	sensorID string,
	gatewayID string,
	name string,
	intervalMs int64,
	profile sensor.SensorProfile,
	status sensor.SensorStatus,
) func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool {
	return func(
		cloudDB clouddb.CloudDBConnection,
		_ sensordb.SensorDBConnection,
		_ *nats.Conn,
		_ natsutils.NatsTestConnection,
		_ jetstream.JetStream,
		_ jetstream.JetStream,
	) bool {
		db := (*gorm.DB)(cloudDB)
		entity := sensor.SensorEntity{
			ID:        sensorID,
			GatewayID: gatewayID,
			Name:      name,
			Interval:  intervalMs,
			Profile:   string(profile),
			Status:    string(status),
		}
		return db.Create(&entity).Error == nil
	}
}

func postSetupDeleteByGateway(
	gatewayID string,
) func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) {
	return func(
		cloudDB clouddb.CloudDBConnection,
		_ sensordb.SensorDBConnection,
		_ *nats.Conn,
		_ natsutils.NatsTestConnection,
		_ jetstream.JetStream,
		_ jetstream.JetStream,
	) {
		db := (*gorm.DB)(cloudDB)
		_ = db.Where("gateway_id = ?", gatewayID).Delete(&sensor.SensorEntity{}).Error
		_ = db.Where("id = ?", gatewayID).Delete(&gateway.GatewayEntity{}).Error
	}
}

func postSetupDeleteSensor(
	sensorID string,
) func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) {
	return func(
		cloudDB clouddb.CloudDBConnection,
		_ sensordb.SensorDBConnection,
		_ *nats.Conn,
		_ natsutils.NatsTestConnection,
		_ jetstream.JetStream,
		_ jetstream.JetStream,
	) {
		db := (*gorm.DB)(cloudDB)
		_ = db.Where("id = ?", sensorID).Delete(&sensor.SensorEntity{}).Error
	}
}

func postSetupUnsubscribe(
	subscription **nats.Subscription,
) func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) {
	return func(
		_ clouddb.CloudDBConnection,
		_ sensordb.SensorDBConnection,
		_ *nats.Conn,
		natsTestConn natsutils.NatsTestConnection,
		_ jetstream.JetStream,
		_ jetstream.JetStream,
	) {
		if *subscription == nil {
			return
		}
		_ = (*subscription).Unsubscribe()
		_ = (*nats.Conn)(natsTestConn).Flush()
		*subscription = nil
	}
}

func checkNoSensorForGateway(
	gatewayID string,
) func(*httptest.ResponseRecorder, clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool {
	return func(
		_ *httptest.ResponseRecorder,
		cloudDB clouddb.CloudDBConnection,
		_ sensordb.SensorDBConnection,
		_ *nats.Conn,
		_ natsutils.NatsTestConnection,
		_ jetstream.JetStream,
		_ jetstream.JetStream,
	) bool {
		db := (*gorm.DB)(cloudDB)
		var count int64
		err := db.Model(&sensor.SensorEntity{}).Where("gateway_id = ?", gatewayID).Count(&count).Error
		return err == nil && count == 0
	}
}

func checkSensorExists(
	sensorID string,
) func(*httptest.ResponseRecorder, clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool {
	return func(
		_ *httptest.ResponseRecorder,
		cloudDB clouddb.CloudDBConnection,
		_ sensordb.SensorDBConnection,
		_ *nats.Conn,
		_ natsutils.NatsTestConnection,
		_ jetstream.JetStream,
		_ jetstream.JetStream,
	) bool {
		db := (*gorm.DB)(cloudDB)
		var count int64
		err := db.Model(&sensor.SensorEntity{}).Where("id = ?", sensorID).Count(&count).Error
		return err == nil && count == 1
	}
}

func checkSensorNotExists(
	sensorID string,
) func(*httptest.ResponseRecorder, clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool {
	return func(
		_ *httptest.ResponseRecorder,
		cloudDB clouddb.CloudDBConnection,
		_ sensordb.SensorDBConnection,
		_ *nats.Conn,
		_ natsutils.NatsTestConnection,
		_ jetstream.JetStream,
		_ jetstream.JetStream,
	) bool {
		db := (*gorm.DB)(cloudDB)
		var count int64
		err := db.Model(&sensor.SensorEntity{}).Where("id = ?", sensorID).Count(&count).Error
		return err == nil && count == 0
	}
}

func checkSensorStatus(
	sensorID string,
	status sensor.SensorStatus,
) func(*httptest.ResponseRecorder, clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool {
	return func(
		_ *httptest.ResponseRecorder,
		cloudDB clouddb.CloudDBConnection,
		_ sensordb.SensorDBConnection,
		_ *nats.Conn,
		_ natsutils.NatsTestConnection,
		_ jetstream.JetStream,
		_ jetstream.JetStream,
	) bool {
		db := (*gorm.DB)(cloudDB)
		var entity sensor.SensorEntity
		if err := db.Where("id = ?", sensorID).First(&entity).Error; err != nil {
			return false
		}

		return entity.Status == string(status)
	}
}

func preSetupCommandResponseListener(
	subscription **nats.Subscription,
	shouldReply bool,
	reply sensor.CommandResponse,
	subject string,
	onMessage ...func(*nats.Msg),
) func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool {
	return func(
		_ clouddb.CloudDBConnection,
		_ sensordb.SensorDBConnection,
		_ *nats.Conn,
		natsTestConn natsutils.NatsTestConnection,
		_ jetstream.JetStream,
		_ jetstream.JetStream,
	) bool {
		nc := (*nats.Conn)(natsTestConn)
		sub, err := nc.Subscribe(subject, func(msg *nats.Msg) {
			if len(onMessage) > 0 && onMessage[0] != nil {
				onMessage[0](msg)
			}

			if !shouldReply {
				return
			}

			payload, err := json.Marshal(reply)
			if err != nil {
				return
			}
			_ = msg.Respond(payload)
		})
		if err != nil {
			return false
		}

		if err := nc.Flush(); err != nil {
			_ = sub.Unsubscribe()
			return false
		}

		*subscription = sub
		return true
	}
}
