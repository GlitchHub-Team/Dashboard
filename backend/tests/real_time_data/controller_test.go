package real_time_data_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"

	"backend/internal/infra/transport/http/dto"
	"backend/internal/real_time_data"
	"backend/internal/sensor"
	sensorProfile "backend/internal/sensor/profile"
	"backend/internal/shared/identity"
	"backend/tests/real_time_data/mocks"
)

func setupTestServer(t *testing.T, controller *real_time_data.Controller, requester *identity.Requester) *httptest.Server {
	t.Helper()
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	engine.Use(func(ctx *gin.Context) {
		if requester != nil {
			ctx.Set("requester", *requester)
		}
		ctx.Next()
	})

	engine.GET("/api/v1/tenant_id/:tenant_id/sensor/:sensor_id/real_time_data", controller.GetRealTimeData)
	return httptest.NewServer(engine)
}

func parseHTTPResponseError(t *testing.T, resp *http.Response) string {
	t.Helper()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}

	errStr, ok := response["error"].(string)
	if !ok {
		t.Fatalf("response does not contain 'error' string field: %s", string(bodyBytes))
	}
	return errStr
}

func buildURL(baseURL string, tenantId, sensorId string) string {
	return baseURL + "/api/v1/tenant_id/" + tenantId + "/sensor/" + sensorId + "/real_time_data"
}

func TestController_GetRealTimeData(t *testing.T) {
	targetSensorId := uuid.New()
	targetTenantId := uuid.New()

	requester := identity.Requester{
		RequesterUserId:   1,
		RequesterTenantId: &targetTenantId,
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	expectedCmd := real_time_data.RetrieveRealTimeDataCommand{
		Requester: requester,
		SensorId:  targetSensorId,
		TenantId:  targetTenantId,
	}

	mockTimestamp := time.Now()
	mockSample := &real_time_data.HeartRateSample{
		BaseSample: real_time_data.BaseSample{
			Profile:   sensorProfile.HEART_RATE,
			Timestamp: mockTimestamp,
		},
		Data: real_time_data.HeartRateSampleData{
			BpmValue: 75,
		},
	}

	type testCase struct {
		name               string
		setupFunc          func(*mocks.MockGetRealTimeDataUseCase) *gomock.Call
		targetTenantId     string
		targetSensorId     string
		requester          *identity.Requester
		expectedStatusCode int
		expectedErrString  string
		checkWS            bool
	}

	genericErr := errors.New("database connection timeout")
	cases := []testCase{
		{
			name: "Success: Upgrades to WS, receives mapped data, and closes on error",
			setupFunc: func(mockUC *mocks.MockGetRealTimeDataUseCase) *gomock.Call {
				return mockUC.EXPECT().
					RetrieveRealTimeData(gomock.Eq(expectedCmd)).
					DoAndReturn(func(cmd real_time_data.RetrieveRealTimeDataCommand) (chan real_time_data.RealTimeSample, chan real_time_data.RealTimeError, error) {
						dataChan := make(chan real_time_data.RealTimeSample, 1)
						errChan := make(chan real_time_data.RealTimeError, 1)

						go func() {
							dataChan <- mockSample
							time.Sleep(50 * time.Millisecond)
							errChan <- real_time_data.RealTimeError{
								Err:       errors.New("sensor connection lost"),
								Timestamp: time.Now(),
							}
						}()
						return dataChan, errChan, nil
					}).
					Times(1)
			},
			targetTenantId:     targetTenantId.String(),
			targetSensorId:     targetSensorId.String(),
			requester:          &requester,
			expectedStatusCode: http.StatusSwitchingProtocols,
			checkWS:            true,
		},
		{
			name:               "Fail: Requester not found in context yields 401",
			setupFunc:          nil,
			targetTenantId:     targetTenantId.String(),
			targetSensorId:     targetSensorId.String(),
			requester:          nil,
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "Fail: URI binding fails yields 400",
			setupFunc:          nil,
			targetTenantId:     "invalid-uuid",
			targetSensorId:     "invalid-uuid",
			requester:          &requester,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "Fail: UseCase returns ErrSensorNotFound yields 404",
			setupFunc: func(mockUC *mocks.MockGetRealTimeDataUseCase) *gomock.Call {
				return mockUC.EXPECT().
					RetrieveRealTimeData(gomock.Eq(expectedCmd)).
					Return(nil, nil, sensor.ErrSensorNotFound).
					Times(1)
			},
			targetTenantId:     targetTenantId.String(),
			targetSensorId:     targetSensorId.String(),
			requester:          &requester,
			expectedStatusCode: http.StatusNotFound,
			expectedErrString:  sensor.ErrSensorNotFound.Error(),
		},
		{
			name: "Fail: UseCase returns generic error yields 500",
			setupFunc: func(mockUC *mocks.MockGetRealTimeDataUseCase) *gomock.Call {
				return mockUC.EXPECT().
					RetrieveRealTimeData(gomock.Eq(expectedCmd)).
					Return(nil, nil, genericErr).
					Times(1)
			},
			targetTenantId:     targetTenantId.String(),
			targetSensorId:     targetSensorId.String(),
			requester:          &requester,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrString:  genericErr.Error(),
		},
		{
			name: "Fail: WS upgrade fails yields 400",
			setupFunc: func(mockUC *mocks.MockGetRealTimeDataUseCase) *gomock.Call {
				return mockUC.EXPECT().
					RetrieveRealTimeData(gomock.Eq(expectedCmd)).
					Return(make(chan real_time_data.RealTimeSample), make(chan real_time_data.RealTimeError), nil).
					Times(1)
			},
			targetTenantId:     targetTenantId.String(),
			targetSensorId:     targetSensorId.String(),
			requester:          &requester,
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockController := gomock.NewController(t)

			mockUC := mocks.NewMockGetRealTimeDataUseCase(mockController)
			if tc.setupFunc != nil {
				tc.setupFunc(mockUC)
			}

			logger := zaptest.NewLogger(t)
			controller := real_time_data.NewController(logger, mockUC)
			server := setupTestServer(t, controller, tc.requester)
			defer server.Close()

			httpEndpoint := buildURL(server.URL, tc.targetTenantId, tc.targetSensorId)

			if tc.checkWS {
				wsURL := "ws" + strings.TrimPrefix(httpEndpoint, "http")

				// a. Dial
				ws, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
				if err != nil {
					var body []byte
					if resp != nil {
						body, _ = io.ReadAll(resp.Body)
					}
					t.Fatalf("expected no error dialing websocket, got %v. Response body: %s", err, string(body))
				}
				defer resp.Body.Close() //nolint:errcheck
				defer ws.Close()        //nolint:errcheck

				// b. Check status code
				if resp.StatusCode != http.StatusSwitchingProtocols {
					t.Fatalf("expected status %d, got %d", http.StatusSwitchingProtocols, resp.StatusCode)
				}

				// c. Read and check message
				_, message, err := ws.ReadMessage()
				if err != nil {
					t.Fatalf("expected no error reading sample message, got %v", err)
				}

				var receivedData real_time_data.RealTimeSampleOutDTO
				if err := json.Unmarshal(message, &receivedData); err != nil {
					t.Fatalf("failed to unmarshal sample data: %v", err)
				}

				if receivedData.Profile != "HeartRate" {
					t.Errorf("expected Profile 'HeartRate', got '%s'", receivedData.Profile)
				}

				dataBytes, err := json.Marshal(receivedData.Data)
				if err != nil {
					t.Fatalf("failed to marshal inner payload: %v", err)
				}

				var heartRateData dto.HeartRateData
				if err := json.Unmarshal(dataBytes, &heartRateData); err != nil {
					t.Fatalf("failed to unmarshal internal payload: %v", err)
				}

				if heartRateData.BpmValue != 75 {
					t.Errorf("expected BPM 75, got %d", heartRateData.BpmValue)
				}

				// d. Read and check error message
				_, errMsg, err := ws.ReadMessage()
				if err != nil {
					t.Fatalf("expected no error reading error message, got %v", err)
				}

				var receivedErr real_time_data.RealTimeErrorOutDTO
				if err := json.Unmarshal(errMsg, &receivedErr); err != nil {
					t.Fatalf("failed to unmarshal error payload: %v", err)
				}

				if receivedErr.Error != "sensor connection lost" {
					t.Errorf("expected error text 'sensor connection lost', got '%s'", receivedErr.Error)
				}

			} else {
				resp, err := http.Get(httpEndpoint) //nolint:noctx
				if err != nil {
					t.Fatalf("expected no error making request, got %v", err)
				}
				defer resp.Body.Close() //nolint:errcheck

				if resp.StatusCode != tc.expectedStatusCode {
					t.Errorf("expected status %d, got %d", tc.expectedStatusCode, resp.StatusCode)
				}

				if tc.expectedErrString != "" {
					errStr := parseHTTPResponseError(t, resp)
					if errStr != tc.expectedErrString {
						t.Errorf("expected error %q, got %q", tc.expectedErrString, errStr)
					}
				}
			}
		})
	}
}
