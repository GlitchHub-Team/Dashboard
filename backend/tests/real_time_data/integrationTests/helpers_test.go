package real_time_data_integration_test

import (
	"fmt"
	"testing"

	"backend/internal/gateway"
	"backend/internal/sensor"
	"backend/internal/tenant"
	"backend/tests/helper"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func realTimeDataPath(tenantID, sensorID uuid.UUID) string {
	return fmt.Sprintf("/api/v1/tenant/%s/sensor/%s/real_time_data", tenantID.String(), sensorID.String())
}

func preSetupSensorChain(
	t *testing.T,
	tenantId uuid.UUID,
	gatewayId uuid.UUID,
	sensorId uuid.UUID,

	sensorProfile string,
	sensorStatus string,
) helper.IntegrationTestPreSetup {
	t.Helper()

	return func(deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)

		// Crea tenant
		if tenantId == uuid.Nil {
			t.Errorf("Cannot setup sensor with nil tenant id")
			return false
		}
		tenantIdStr := tenantId.String()
		err := db.Save(&tenant.TenantEntity{
			ID: tenantId.String(),
			Name: "Test tenant",
			CanImpersonate: true,
		}).Error
		if err != nil {
			t.Errorf("error creating tenant: %v", err)
			return false
		}

		// Crea gateway
		err = db.Save(&gateway.GatewayEntity{
			ID: gatewayId.String(),
			Name: "Test gateway",
			TenantId: &tenantIdStr,
			Status: string(gateway.GATEWAY_STATUS_ACTIVE),
			PublicIdentifier: "test-public-identifier",
		}).Error
		if err != nil {
			t.Errorf("error creating tenant: %v", err)
			return false
		}

		// Crea sensore
		err = db.Save(&sensor.SensorEntity{
			ID: sensorId.String(),
			GatewayID: gatewayId.String(),
			Name: "Test Sensor",
			Interval: 1000,
			Profile: sensorProfile,
			Status: sensorStatus,
		}).Error
		if err != nil {
			t.Errorf("error creating tenant: %v", err)
			return false
		}

		return true
	}
}

func postSetupDeleteSensorChain(
	t *testing.T,
	tenantId uuid.UUID,
	gatewayId uuid.UUID,
	sensorId uuid.UUID,
) helper.IntegrationTestPostSetup {
	t.Helper()
	return func(deps helper.IntegrationTestDeps) {
		db := (*gorm.DB)(deps.CloudDB)

		err := db.Where("id = ?", sensorId.String()).Delete(&sensor.SensorEntity{}).Error
		if err != nil {
			t.Errorf("cannot delete sensor: %v", err)
		}

		err = db.Where("id = ?", gatewayId.String()).Delete(&gateway.GatewayEntity{}).Error
		if err != nil {
			t.Errorf("cannot delete sensor: %v", err)
		}

		err = db.Where("id = ?", tenantId.String()).Delete(&tenant.TenantEntity{}).Error
		if err != nil {
			t.Errorf("cannot delete sensor: %v", err)
		}
	}
}
