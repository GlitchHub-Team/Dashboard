package real_time_data_integration_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	httpDto "backend/internal/infra/transport/http/dto"
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

	pathWithJwt := func(path, jwtToken string) string {
		return path + "?jwt=" + jwtToken
	}

	tests := []*helper.IntegrationTestCase{
		{
			PreSetups: nil,
			Name:      "Fallimento: Token JWT mancante",
			Method:    http.MethodGet,
			Path:      realTimeDataPath(targetTenantID, targetSensorID),
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
			Path:      pathWithJwt("/api/v1/tenant/invalid-uuid/sensor/invalid-uuid/real_time_data", tenantAdminJWT),
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
			Path:      pathWithJwt(realTimeDataPath(targetTenantID, uuid.New()), tenantAdminJWT),
			Body:      nil,

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: helper.ErrJsonString(sensor.ErrSensorNotFound),
			ResponseChecks:   nil,
			PostSetups:       nil,
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupSensorChain(
					t, targetTenantID, targetGatewayID, targetSensorID, string(sensorProfile.HEART_RATE), string(sensor.Active),
				),
			},
			Name:   "Fallimento: Tentativo di accesso da un tenant non autorizzato",
			Method: http.MethodGet,
			Path:   pathWithJwt(realTimeDataPath(targetTenantID, targetSensorID), unauthorizedAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: helper.ErrJsonString(sensor.ErrSensorNotFound),
			ResponseChecks:   nil,
			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteSensorChain(t, targetTenantID, targetGatewayID, targetSensorID),
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

	t.Run("Super Admin case", func(t *testing.T) {
		targetTenantID := uuid.New()
		targetGatewayID := uuid.New()
		targetSensorID := uuid.New()

		superAdminJWT, err := helper.NewSuperAdminJWT(deps, 1)
		if err != nil {
			t.Fatalf("failed to generate super admin jwt: %v", err)
		}
		executeWSTest(t, deps, superAdminJWT, targetTenantID, targetGatewayID, targetSensorID)
	})

	t.Run("Tenant Admin case", func(t *testing.T) {
		targetTenantID := uuid.New()
		targetGatewayID := uuid.New()
		targetSensorID := uuid.New()

		tenantAdminJWT, err := helper.NewTenantAdminJWT(deps, targetTenantID, 1)
		if err != nil {
			t.Fatalf("failed to generate tenant admin jwt: %v", err)
		}
		executeWSTest(t, deps, tenantAdminJWT, targetTenantID, targetGatewayID, targetSensorID)
	})

	t.Run("Tenant User case", func(t *testing.T) {
		targetTenantID := uuid.New()
		targetGatewayID := uuid.New()
		targetSensorID := uuid.New()

		tenantUserJWT, err := helper.NewTenantUserJWT(deps, targetTenantID, 1)
		if err != nil {
			t.Fatalf("failed to generate tenant user jwt: %v", err)
		}
		executeWSTest(t, deps, tenantUserJWT, targetTenantID, targetGatewayID, targetSensorID)
	})
}

func executeWSTest(t *testing.T, deps helper.IntegrationTestDeps, jwt string, targetTenantID, targetGatewayID, targetSensorID uuid.UUID) {
	t.Helper()

	// 1. Inserimento record DB
	setupFunc := preSetupSensorChain(t, targetTenantID, targetGatewayID, targetSensorID, string(sensorProfile.HEART_RATE), string(sensor.Active))
	if ok := setupFunc(deps); !ok {
		t.Fatalf("database presetup failed")
	}
	t.Cleanup(func() {
		cleanupFunc := postSetupDeleteSensorChain(t, targetTenantID, targetGatewayID, targetSensorID)
		cleanupFunc(deps)
	})

	// 2. Avvio del server HTTP reale per consentire il protocol switch
	server := httptest.NewServer(deps.Router)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + realTimeDataPath(targetTenantID, targetSensorID) + "?jwt=" + jwt

	// 3. Negoziazione WebSocket
	ws, resp, err := websocket.DefaultDialer.Dial(wsURL, http.Header{})
	if err != nil {
		bodyDump := "nessun body"
		if resp != nil && resp.Body != nil {
			bytes, _ := io.ReadAll(resp.Body)
			bodyDump = string(bytes)
		}
		// Se fallisce mostrerà "404 page not found" (problema di routing)
		// o "{"error": "sensor not found"}" (problema di setup DB / foreign key)
		t.Fatalf("failed to dial websocket: %v (status: %d) | body: %s | URL: %s", err, resp.StatusCode, bodyDump, wsURL)
	}
	defer resp.Body.Close() //nolint:errcheck
	defer ws.Close()        //nolint:errcheck

	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("expected HTTP 101 Switching Protocols, got %d", resp.StatusCode)
	}

	// 4. Simulazione deterministica del flusso NATS
	natsSubject := fmt.Sprintf("sensor.%s.%s.%s", targetTenantID.String(), targetGatewayID.String(), targetSensorID.String())
	targetTimestamp := time.Now().UTC().Format(time.RFC3339Nano)
	rawPayload := fmt.Sprintf(`{"timestamp":"%s","data":{"BpmValue":82}}`, targetTimestamp)

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
	err = ws.SetReadDeadline(time.Now().Add(3 * time.Second))
	if err != nil {
		t.Fatalf("cannot set deadline for mock NATS publisher: %v", err)
	}
	_, message, err := ws.ReadMessage()

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
	var data httpDto.HeartRateData
	err = json.Unmarshal(dataBytes, &data)
	if err != nil {
		t.Fatalf("expected nil error when unmarshaling data, got %v", err)
	}

	if data.BpmValue != 82 {
		t.Errorf("expected bpmValue 82, got %v", data.BpmValue)
	}
}
