package sensor_integration_test

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"

	"backend/internal/gateway"
	"backend/internal/infra/transport/http/dto"
	"backend/internal/sensor"
	sensorProfile "backend/internal/sensor/profile"
	sharedCrypto "backend/internal/shared/crypto"
	"backend/internal/shared/identity"
	"backend/internal/tenant"
	"backend/tests/helper"

	"github.com/nats-io/nats.go"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	tenant1IdStr = "11111111-1111-1111-1111-111111111111"
	tenant2IdStr = "22222222-2222-2222-2222-222222222222"
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

func preSetupCreateGateway(
	gatewayID string,
	name string,
) helper.IntegrationTestPreSetup {
	return preSetupCreateGatewayWithTenant(gatewayID, name, nil)
}

func preSetupCreateGatewayWithTenant(
	gatewayID string,
	name string,
	tenantID *string,
) helper.IntegrationTestPreSetup {
	return func(deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)
		entity := gateway.GatewayEntity{
			ID:       gatewayID,
			Name:     name,
			TenantId: tenantID,
			Status:   string(gateway.GATEWAY_STATUS_ACTIVE),
		}
		return db.Create(&entity).Error == nil
	}
}

/*
Popola il DB con 2 tenant.

NOTA: è una "monkey patch", però funziona
*/
func populateTenantDefaultData(db *gorm.DB) error {
	tenants := []tenant.TenantEntity{
		{ID: tenant1IdStr, Name: "Tenant 1", CanImpersonate: true},
		{ID: tenant2IdStr, Name: "Tenant 2", CanImpersonate: false},
	}

	// Tenant 1 e 2
	for _, tenant := range tenants {
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&tenant).Error; err != nil {
			return fmt.Errorf("failed to create tenant %v: %v", tenant.ID, err)
		}
	}
	return nil
}

func preSetupCreateSensor(
	sensorID string,
	gatewayID string,
	name string,
	intervalMs int64,
	profile sensorProfile.SensorProfile,
	status sensor.SensorStatus,
) helper.IntegrationTestPreSetup {
	return func(deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)
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
) helper.IntegrationTestPostSetup {
	return func(deps helper.IntegrationTestDeps) {
		db := (*gorm.DB)(deps.CloudDB)
		_ = db.Where("gateway_id = ?", gatewayID).Delete(&sensor.SensorEntity{}).Error
		_ = db.Where("id = ?", gatewayID).Delete(&gateway.GatewayEntity{}).Error
	}
}

func postSetupDeleteSensor(
	sensorID string,
) helper.IntegrationTestPostSetup {
	return func(deps helper.IntegrationTestDeps) {
		db := (*gorm.DB)(deps.CloudDB)
		_ = db.Where("id = ?", sensorID).Delete(&sensor.SensorEntity{}).Error
	}
}

func postSetupUnsubscribe(
	subscription **nats.Subscription,
) helper.IntegrationTestPostSetup {
	return func(deps helper.IntegrationTestDeps) {
		if *subscription == nil {
			return
		}
		_ = (*subscription).Unsubscribe()
		_ = (*nats.Conn)(deps.NatsTestConn).Flush()
		*subscription = nil
	}
}

func checkNoSensorForGateway(
	gatewayID string,
) helper.IntegrationTestCheck {
	return func(
		r *httptest.ResponseRecorder,
		deps helper.IntegrationTestDeps,
	) bool {
		db := (*gorm.DB)(deps.CloudDB)
		var count int64
		err := db.Model(&sensor.SensorEntity{}).Where("gateway_id = ?", gatewayID).Count(&count).Error
		return err == nil && count == 0
	}
}

func checkSensorExists(
	sensorID string,
) helper.IntegrationTestCheck {
	return func(
		r *httptest.ResponseRecorder,
		deps helper.IntegrationTestDeps,
	) bool {
		db := (*gorm.DB)(deps.CloudDB)
		var count int64
		err := db.Model(&sensor.SensorEntity{}).Where("id = ?", sensorID).Count(&count).Error
		return err == nil && count == 1
	}
}

func checkSensorNotExists(
	sensorID string,
) helper.IntegrationTestCheck {
	return func(
		r *httptest.ResponseRecorder,
		deps helper.IntegrationTestDeps,
	) bool {
		db := (*gorm.DB)(deps.CloudDB)
		var count int64
		err := db.Model(&sensor.SensorEntity{}).Where("id = ?", sensorID).Count(&count).Error
		return err == nil && count == 0
	}
}

func checkSensorStatus(
	sensorID string,
	status sensor.SensorStatus,
) helper.IntegrationTestCheck {
	return func(
		r *httptest.ResponseRecorder,
		deps helper.IntegrationTestDeps,
	) bool {
		db := (*gorm.DB)(deps.CloudDB)
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
	reply dto.CommandResponse,
	subject string,
	onMessage ...func(*nats.Msg),
) helper.IntegrationTestPreSetup {
	return func(deps helper.IntegrationTestDeps) bool {
		nc := (*nats.Conn)(deps.NatsTestConn)
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
