package real_time_data_integration_test

import (
	"fmt"
	"net/http"

	"backend/tests/helper"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func authHeader(jwt string) http.Header {
	header := http.Header{}
	header.Set("Authorization", "Bearer "+jwt)
	return header
}

func realTimeDataPath(tenantID uuid.UUID, sensorID uuid.UUID) string {
	return fmt.Sprintf("/api/v1/tenant_id/%s/sensor/%s/real_time_data", tenantID.String(), sensorID.String())
}

func preSetupSensorChain(
	tenantID uuid.UUID,
	gatewayID uuid.UUID,
	sensorID uuid.UUID,
	profile string,
	status string,
) helper.IntegrationTestPreSetup {
	return func(deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)

		err := db.Exec(`INSERT INTO tenants (id, name, can_impersonate) VALUES (?, 'RealTime Data Tenant', false)`, tenantID).Error
		if err != nil {
			return false
		}

		var tID *uuid.UUID
		if tenantID != uuid.Nil {
			tID = &tenantID
		}

		err = db.Exec(`INSERT INTO gateways (id, tenant_id, name, mac_address) VALUES (?, ?, 'RealTime Gateway', 'AA:BB:CC:DD:EE:FF')`, gatewayID, tID).Error
		if err != nil {
			return false
		}

		err = db.Exec(`INSERT INTO sensors (id, gateway_id, profile, status, name) VALUES (?, ?, ?, ?, 'RealTime Sensor')`, sensorID, gatewayID, profile, status).Error

		return err == nil
	}
}

func postSetupDeleteSensorChain(
	tenantID uuid.UUID,
	gatewayID uuid.UUID,
	sensorID uuid.UUID,
) helper.IntegrationTestPostSetup {
	return func(deps helper.IntegrationTestDeps) {
		db := (*gorm.DB)(deps.CloudDB)

		db.Exec(`DELETE FROM sensors WHERE id = ?`, sensorID)
		db.Exec(`DELETE FROM gateways WHERE id = ?`, gatewayID)
		db.Exec(`DELETE FROM tenants WHERE id = ?`, tenantID)
	}
}
