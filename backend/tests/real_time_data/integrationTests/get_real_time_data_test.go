package real_time_data_integration_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"backend/internal/real_time_data"
	"backend/internal/sensor"
	sensorProfile "backend/internal/sensor/profile"
	"backend/tests/helper"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
)

func TestGetRealTimeDataIntegrationErrors(t *testing.T) {
	t.Setenv("DASHBOARD_CREDS_PATH", "dashboard.creds")
	t.Setenv("TEST_CREDS_PATH", "admin_test.creds")
	t.Setenv("CA_PEM_PATH", "ca.pem")

	deps := helper.SetupIntegrationTest(t)

	targetTenantID := uuid.New()
	targetGatewayID := uuid.New()
	targetSensorID := uuid.New()

	unauthorizedTenantID := uuid.New()

	tenantAdminJWT, err := helper.NewTenantAdminJWT(deps, targetTenantID, 1)
	if err != nil {
		t.Fatalf("failed to generate tenant admin jwt: %v", err)
	}

	unauthorizedAdminJWT, err := helper.NewTenantAdminJWT(deps, unauthorizedTenantID, 2)
	if err != nil {
		t.Fatalf("failed to generate unauthorized tenant admin jwt: %v", err)
	}

	tests := []*helper.IntegrationTestCase{
		{
			PreSetups: nil,
			Name:      "Fallimento: Token JWT mancante",
			Method:    http.MethodGet,
			Path:      realTimeDataPath(targetTenantID, targetSensorID),
			Header:    nil,
			Body:      nil,

			WantStatusCode:   http.StatusUnauthorized,
			WantResponseBody: "",
			ResponseChecks:   nil,
			PostSetups:       nil,
		},
		{
			PreSetups: nil,
			Name:      "Fallimento: UUID uri params non validi",
			Method:    http.MethodGet,
			Path:      "/api/v1/tenant_id/invalid-uuid/sensor/invalid-uuid/real_time_data",
			Header:    authHeader(tenantAdminJWT),
			Body:      nil,

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: "",
			ResponseChecks:   nil,
			PostSetups:       nil,
		},
		{
			PreSetups: nil,
			Name:      "Fallimento: Sensore non trovato nel database",
			Method:    http.MethodGet,
			Path:      realTimeDataPath(targetTenantID, targetSensorID),
			Header:    authHeader(tenantAdminJWT),
			Body:      nil,

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: helper.ErrJsonString(sensor.ErrSensorNotFound),
			ResponseChecks:   nil,
			PostSetups:       nil,
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupSensorChain(targetTenantID, targetGatewayID, targetSensorID, string(sensorProfile.HEART_RATE), string(sensor.Active)),
			},
			Name:   "Fallimento: Tentativo di accesso da un tenant non autorizzato",
			Method: http.MethodGet,
			Path:   realTimeDataPath(targetTenantID, targetSensorID),
			Header: authHeader(unauthorizedAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: helper.ErrJsonString(sensor.ErrSensorNotFound),
			ResponseChecks:   nil,
			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteSensorChain(targetTenantID, targetGatewayID, targetSensorID),
			},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupSensorChain(targetTenantID, targetGatewayID, targetSensorID, string(sensorProfile.HEART_RATE), string(sensor.Inactive)),
			},
			Name:   "Fallimento: Tentativo di recupero stream per sensore disattivato",
			Method: http.MethodGet,
			Path:   realTimeDataPath(targetTenantID, targetSensorID),
			Header: authHeader(tenantAdminJWT),
			Body:   nil,

			// A inactive sensor (via its status field or a nil gateway tenant mapping) returns ErrSensorNotActive
			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: helper.ErrJsonString(sensor.ErrSensorNotActive),
			ResponseChecks:   nil,
			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteSensorChain(targetTenantID, targetGatewayID, targetSensorID),
			},
		},
	}

	helper.RunIntegrationTests(t, tests, deps)
}

func TestGetRealTimeDataIntegrationSuccess(t *testing.T) {
	t.Setenv("DASHBOARD_CREDS_PATH", "dashboard.creds")
	t.Setenv("TEST_CREDS_PATH", "admin_test.creds")
	t.Setenv("CA_PEM_PATH", "ca.pem")

	deps := helper.SetupIntegrationTest(t)

	targetTenantID := uuid.New()
	targetGatewayID := uuid.New()
	targetSensorID := uuid.New()

	tenantAdminJWT, err := helper.NewTenantAdminJWT(deps, targetTenantID, 1)
	if err != nil {
		t.Fatalf("failed to generate tenant admin jwt: %v", err)
	}

	// 1. Inserimento record DB
	setupFunc := preSetupSensorChain(targetTenantID, targetGatewayID, targetSensorID, string(sensorProfile.HEART_RATE), string(sensor.Active))
	if ok := setupFunc(deps); !ok {
		t.Fatalf("database presetup failed")
	}
	t.Cleanup(func() {
		cleanupFunc := postSetupDeleteSensorChain(targetTenantID, targetGatewayID, targetSensorID)
		cleanupFunc(deps)
	})

	// 2. Avvio del server HTTP reale per consentire il protocol switch
	server := httptest.NewServer(deps.Router)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + realTimeDataPath(targetTenantID, targetSensorID)

	// 3. Negoziazione WebSocket
	header := http.Header{}
	header.Set("Authorization", "Bearer "+tenantAdminJWT)

	ws, resp, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		t.Fatalf("failed to dial websocket: %v (status: %d)", err, resp.StatusCode)
	}
	defer ws.Close()

	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("expected HTTP 101 Switching Protocols, got %d", resp.StatusCode)
	}

	// 4. Simulazione deterministica del flusso NATS
	natsSubject := fmt.Sprintf("sensor.%s.%s.%s", targetTenantID.String(), targetGatewayID.String(), targetSensorID.String())
	targetTimestamp := time.Now().UTC().Format(time.RFC3339Nano)
	rawPayload := fmt.Sprintf(`{"timestamp":"%s","data":{"bpmValue":82}}`, targetTimestamp)

	// Spawns a background worker that publishes the payload every 10ms.
	// This completely eliminates the race condition: the exact millisecond
	// the subscriber is ready, it will catch the next tick's payload.
	publishDone := make(chan struct{})
	go func() {
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-publishDone:
				return
			case <-ticker.C:
				_ = (*nats.Conn)(deps.NatsTestConn).Publish(natsSubject, []byte(rawPayload))
			}
		}
	}()

	// 5. Lettura bloccante con timeout
	ws.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, message, err := ws.ReadMessage()

	// Ferma il publisher in background immediatamente dopo la lettura (o il timeout)
	close(publishDone)

	if err != nil {
		t.Fatalf("failed to read from websocket: %v", err)
	}

	// 6. Validazione dei frame in uscita dal socket
	var receivedData real_time_data.RealTimeSampleOutDTO
	if err := json.Unmarshal(message, &receivedData); err != nil {
		t.Fatalf("failed to unmarshal output DTO: %v", err)
	}

	if receivedData.Profile != "HeartRate" {
		t.Errorf("expected Profile HeartRate, got %s", receivedData.Profile)
	}

	dataBytes, _ := json.Marshal(receivedData.Data)
	var heartRateMap map[string]interface{}
	json.Unmarshal(dataBytes, &heartRateMap)

	bpmValue, ok := heartRateMap["bpmValue"].(float64)
	if !ok || bpmValue != 82 {
		t.Errorf("expected bpmValue 82, got %v", bpmValue)
	}
}
