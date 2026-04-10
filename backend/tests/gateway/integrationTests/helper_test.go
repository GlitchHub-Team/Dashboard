package gateway_integrationtests

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"backend/internal/gateway"
	"backend/internal/infra/transport/http/dto"
	"backend/tests/helper"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"gorm.io/gorm"
)

type gatewayHTTPResponse struct {
	GatewayID        string  `json:"gateway_id"`
	GatewayName      string  `json:"name"`
	TenantID         string  `json:"tenant_id"`
	Status           string  `json:"status"`
	Interval         int64   `json:"interval"`
	PublicIdentifier *string `json:"public_identifier"`
}

type gatewayCommandPayload struct {
	GatewayID string `json:"gatewayId"`
}

type commissionGatewayCommandPayload struct {
	GatewayID         string `json:"gatewayId"`
	TenantID          string `json:"tenantId"`
	CommissionedToken string `json:"commissionedToken"`
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

func preSetupRawCommandResponseListener(
	subscription **nats.Subscription,
	subject string,
	rawPayload []byte,
	onMessage ...func(*nats.Msg),
) helper.IntegrationTestPreSetup {
	return func(deps helper.IntegrationTestDeps) bool {
		nc := (*nats.Conn)(deps.NatsTestConn)
		sub, err := nc.Subscribe(subject, func(msg *nats.Msg) {
			if len(onMessage) > 0 && onMessage[0] != nil {
				onMessage[0](msg)
			}

			_ = msg.Respond(rawPayload)
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

func postSetupUnsubscribe(subscription **nats.Subscription) helper.IntegrationTestPostSetup {
	return func(deps helper.IntegrationTestDeps) {
		if *subscription == nil {
			return
		}
		_ = (*subscription).Unsubscribe()
		_ = (*nats.Conn)(deps.NatsTestConn).Flush()
		*subscription = nil
	}
}

func postSetupDeleteGatewayByName(name string) helper.IntegrationTestPostSetup {
	return func(deps helper.IntegrationTestDeps) {
		db := (*gorm.DB)(deps.CloudDB)
		_ = db.Where("name = ?", name).Delete(&gateway.GatewayEntity{}).Error
	}
}

func postSetupDeleteGatewayByID(id string) helper.IntegrationTestPostSetup {
	return func(deps helper.IntegrationTestDeps) {
		db := (*gorm.DB)(deps.CloudDB)
		_ = db.Where("id = ?", id).Delete(&gateway.GatewayEntity{}).Error
	}
}

func postSetupComposite(setups ...helper.IntegrationTestPostSetup) helper.IntegrationTestPostSetup {
	return func(deps helper.IntegrationTestDeps) {
		for _, setup := range setups {
			if setup != nil {
				setup(deps)
			}
		}
	}
}

func postSetupsWithFinal(preCount int, final helper.IntegrationTestPostSetup) []helper.IntegrationTestPostSetup {
	if preCount <= 0 {
		return []helper.IntegrationTestPostSetup{}
	}

	postSetups := make([]helper.IntegrationTestPostSetup, preCount)
	postSetups[preCount-1] = final
	return postSetups
}

func checkGatewayNotExistsByName(name string) helper.IntegrationTestCheck {
	return func(
		r *httptest.ResponseRecorder,
		deps helper.IntegrationTestDeps,
	) bool {
		db := (*gorm.DB)(deps.CloudDB)
		var count int64
		err := db.Model(&gateway.GatewayEntity{}).Where("name = ?", name).Count(&count).Error
		return err == nil && count == 0
	}
}

func checkGatewayExistsByID(id string) helper.IntegrationTestCheck {
	return func(
		r *httptest.ResponseRecorder,
		deps helper.IntegrationTestDeps,
	) bool {
		db := (*gorm.DB)(deps.CloudDB)
		var count int64
		err := db.Model(&gateway.GatewayEntity{}).Where("id = ?", id).Count(&count).Error
		return err == nil && count == 1
	}
}

func checkGatewayNotExistsByID(id string) helper.IntegrationTestCheck {
	return func(
		r *httptest.ResponseRecorder,
		deps helper.IntegrationTestDeps,
	) bool {
		db := (*gorm.DB)(deps.CloudDB)
		var count int64
		err := db.Model(&gateway.GatewayEntity{}).Where("id = ?", id).Count(&count).Error
		return err == nil && count == 0
	}
}

func preSetupCreateGatewayWithState(
	gatewayID string,
	name string,
	interval int64,
	status gateway.GatewayStatus,
	tenantID *string,
	publicIdentifier *string,
) helper.IntegrationTestPreSetup {
	return func(deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)
		entity := gateway.GatewayEntity{
			ID:               gatewayID,
			Name:             name,
			Interval:         interval,
			Status:           string(status),
			TenantId:         tenantID,
			PublicIdentifier: publicIdentifier,
		}
		return db.Create(&entity).Error == nil
	}
}

func checkGatewayState(
	gatewayID string,
	status gateway.GatewayStatus,
	tenantID *string,
	interval int64,
) helper.IntegrationTestCheck {
	return func(
		r *httptest.ResponseRecorder,
		deps helper.IntegrationTestDeps,
	) bool {
		db := (*gorm.DB)(deps.CloudDB)
		var entity gateway.GatewayEntity
		if err := db.Where("id = ?", gatewayID).First(&entity).Error; err != nil {
			return false
		}

		if entity.Status != string(status) {
			return false
		}

		if entity.Interval != interval {
			return false
		}

		if tenantID == nil {
			return entity.TenantId == nil
		}

		if entity.TenantId == nil {
			return false
		}

		return *entity.TenantId == *tenantID
	}
}

func checkDeleteGatewayResponseAndCommand(
	cmd *gatewayCommandPayload,
	expectedGatewayID string,
	expectedGatewayName string,
) helper.IntegrationTestCheck {
	return func(
		w *httptest.ResponseRecorder,
		deps helper.IntegrationTestDeps,
	) bool {
		var resp gatewayHTTPResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			return false
		}

		if resp.GatewayID != expectedGatewayID || resp.GatewayName != expectedGatewayName {
			return false
		}

		if cmd.GatewayID != expectedGatewayID {
			return false
		}

		db := (*gorm.DB)(deps.CloudDB)
		var count int64
		err := db.Model(&gateway.GatewayEntity{}).Where("id = ?", expectedGatewayID).Count(&count).Error
		return err == nil && count == 0
	}
}

func checkCommissionResponseAndCommand(
	t *testing.T,
	cmd *commissionGatewayCommandPayload,
	expectedGatewayID string,
	expectedGatewayName string,
	expectedTenantID string,
	expectedCommissionToken string,
	expectedInterval int64,
) helper.IntegrationTestCheck {
	t.Helper()

	return func(
		w *httptest.ResponseRecorder,
		deps helper.IntegrationTestDeps,
	) bool {
		var resp gatewayHTTPResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Errorf("failed to unmarshal response: %v", err)
			return false
		}

		if resp.GatewayID != expectedGatewayID || resp.GatewayName != expectedGatewayName {
			return false
		}

		if resp.TenantID != expectedTenantID {
			return false
		}

		if resp.Status != string(gateway.GATEWAY_STATUS_ACTIVE) {
			return false
		}

		if resp.Interval != expectedInterval {
			return false
		}

		if cmd.GatewayID != expectedGatewayID || cmd.TenantID != expectedTenantID || cmd.CommissionedToken != expectedCommissionToken {
			return false
		}

		db := (*gorm.DB)(deps.CloudDB)
		var entity gateway.GatewayEntity
		if err := db.Where("id = ?", expectedGatewayID).First(&entity).Error; err != nil {
			return false
		}

		if entity.Status != string(gateway.GATEWAY_STATUS_ACTIVE) || entity.TenantId == nil || *entity.TenantId != expectedTenantID {
			return false
		}

		return true
	}
}

func checkGatewayCommandAndState(
	cmd *gatewayCommandPayload,
	expectedGatewayID string,
	expectedStatus gateway.GatewayStatus,
	expectedTenantID *string,
	expectedInterval int64,
) helper.IntegrationTestCheck {
	return func(
		w *httptest.ResponseRecorder,
		deps helper.IntegrationTestDeps,
	) bool {
		if cmd != nil && cmd.GatewayID != expectedGatewayID {
			return false
		}

		db := (*gorm.DB)(deps.CloudDB)
		var entity gateway.GatewayEntity
		if err := db.Where("id = ?", expectedGatewayID).First(&entity).Error; err != nil {
			return false
		}

		if entity.Status != string(expectedStatus) || entity.Interval != expectedInterval {
			return false
		}

		if expectedTenantID == nil {
			return entity.TenantId == nil
		}

		if entity.TenantId == nil {
			return false
		}

		return *entity.TenantId == *expectedTenantID
	}
}

func checkCreateGatewayResponseAndDB(
	t *testing.T,
	cmd *createGatewayCommandPayload,
	expectedName string,
	expectedInterval int64,
) helper.IntegrationTestCheck {
	t.Helper()

	return func(w *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		var resp createGatewayResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Errorf("errore unmarshaling response: %v", err)
			return false
		}

		gatewayID, err := uuid.Parse(resp.GatewayID)
		if err != nil {
			t.Errorf("gateway_id non valido in response: %v", err)
			return false
		}

		if resp.GatewayName != expectedName {
			t.Errorf("name errato: got=%s want=%s", resp.GatewayName, expectedName)
			return false
		}

		if resp.Interval != expectedInterval {
			t.Errorf("interval errato: got=%d want=%d", resp.Interval, expectedInterval)
			return false
		}

		if resp.Status != string(gateway.GATEWAY_STATUS_DECOMMISSIONED) {
			t.Errorf("status errato: got=%s want=%s", resp.Status, string(gateway.GATEWAY_STATUS_DECOMMISSIONED))
			return false
		}

		if resp.TenantID != "" {
			t.Errorf("tenant_id deve essere vuoto, got=%s", resp.TenantID)
			return false
		}

		if resp.PublicIdentifier != nil {
			t.Errorf("publicIdentifier deve essere nil")
			return false
		}

		db := (*gorm.DB)(deps.CloudDB)
		var entity gateway.GatewayEntity
		if err := db.Where("id = ?", gatewayID.String()).First(&entity).Error; err != nil {
			t.Errorf("errore lettura db: %v", err)
			return false
		}

		if entity.Name != expectedName || entity.Interval != expectedInterval {
			t.Errorf("dati db non coerenti: got name=%s interval=%d", entity.Name, entity.Interval)
			return false
		}

		if entity.Status != string(gateway.GATEWAY_STATUS_DECOMMISSIONED) {
			t.Errorf("status db errato: got=%s", entity.Status)
			return false
		}

		if entity.TenantId != nil {
			t.Errorf("tenant_id db deve essere nil")
			return false
		}

		if entity.PublicIdentifier != nil {
			t.Errorf("publicIdentifier db deve essere nil")
			return false
		}

		if cmd.GatewayID == "" || cmd.GatewayID != entity.ID {
			t.Errorf("gateway id nel comando non coerente: cmd=%s db=%s", cmd.GatewayID, entity.ID)
			return false
		}

		if cmd.Interval != expectedInterval {
			t.Errorf("interval nel comando errato: got=%d want=%d", cmd.Interval, expectedInterval)
			return false
		}

		return true
	}
}
