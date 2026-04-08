package historical_data_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	historical_data "backend/internal/historical_data"
	"backend/internal/shared/identity"
	"backend/internal/tenant"
	"backend/tests/historical_data/mocks"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type hasError struct{}

type mockUseCaseSetupFunc[T any] func(*T) *gomock.Call

type genericControllerTestCase[MockT any] struct {
	name             string
	method           string
	url              string
	requester        identity.Requester
	omitIdentity     bool
	setupSteps       []mockUseCaseSetupFunc[MockT]
	expectedStatus   int
	expectedResponse any
}

func setupMockUseCase[MockT any](
	constructor func(*gomock.Controller) *MockT,
	setupSteps []mockUseCaseSetupFunc[MockT],
	t *testing.T,
) *MockT {
	ctrl := gomock.NewController(t)
	mockUseCase := constructor(ctrl)

	var expectedCalls []any
	for _, step := range setupSteps {
		if call := step(mockUseCase); call != nil {
			expectedCalls = append(expectedCalls, call)
		}
	}
	if len(expectedCalls) > 0 {
		gomock.InOrder(expectedCalls...)
	}
	return mockUseCase
}

func executeControllerTest[MockT any](
	t *testing.T,
	tc genericControllerTestCase[MockT],
	mountMethod string,
	mountURL string,
	controllerFunc func(*gin.Context),
) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	if validator, ok := binding.Validator.Engine().(*validator.Validate); ok {
		validator.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})
	}

	router.Use(func(ctx *gin.Context) {
		if !tc.omitIdentity {
			ctx.Set("requester", tc.requester)
		}
		ctx.Next()
	})

	router.Handle(mountMethod, mountURL, controllerFunc)

	req, err := http.NewRequest(tc.method, tc.url, nil) //nolint:noctx
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != tc.expectedStatus {
		t.Fatalf("expected status %d, got %d. Response: %s", tc.expectedStatus, w.Code, w.Body.String())
	}

	checkHTTPResponse(w, tc.expectedResponse, t)
}

func checkHTTPResponse[T any](w *httptest.ResponseRecorder, expectedResponse T, t *testing.T) {
	if reflect.ValueOf(expectedResponse) == reflect.Zero(reflect.TypeFor[T]()) {
		return
	}

	actualBytes := w.Body.Bytes()
	expectedBytes, err := json.Marshal(expectedResponse)
	if err != nil {
		t.Fatalf("failed to marshal expected response: %v", err)
	}

	var actualObj, expectedObj any
	if err := json.Unmarshal(actualBytes, &actualObj); err != nil {
		t.Fatalf("failed to unmarshal actual response: %v. body was %s", err, string(actualBytes))
	}
	if err := json.Unmarshal(expectedBytes, &expectedObj); err != nil {
		t.Fatalf("failed to unmarshal expected response: %v", err)
	}

	if _, onlyCheckError := any(expectedResponse).(hasError); onlyCheckError {
		actualMap, ok := actualObj.(map[string]any)
		if !ok {
			t.Fatalf("expected json object, got %v", actualObj)
		}
		if _, ok := actualMap["error"]; !ok {
			t.Fatalf("expected object with error key, got %#v", actualObj)
		}
		return
	}

	if !reflect.DeepEqual(expectedObj, actualObj) {
		t.Fatalf("response mismatch. expected %#v got %#v", expectedObj, actualObj)
	}
}

func TestController_GetSensorHistoricalData(t *testing.T) {
	targetTenantId := uuid.New()
	targetSensorId := uuid.New()

	expectedSamples := []historical_data.HistoricalSample{
		{
			SensorId:  targetSensorId,
			GatewayId: uuid.New(),
			TenantId:  targetTenantId,
			Profile:   string(sensorProfile.HEART_RATE),
			Timestamp: time.Date(2026, 3, 29, 12, 0, 0, 0, time.UTC),
			Data:      json.RawMessage(`{"BpmValue":72}`),
		},
	}
	expectedResponse, err := historical_data.NewHistoricalDataResponseDTO(expectedSamples)
	if err != nil {
		t.Fatalf("failed to build expected response: %v", err)
	}

	useCaseOK := func(mockUC *mocks.MockGetSensorHistoricalDataUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetSensorHistoricalData(gomock.Any()).
			Return(expectedSamples, nil).
			Times(1)
	}
	useCaseNeverCalled := func(mockUC *mocks.MockGetSensorHistoricalDataUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetSensorHistoricalData(gomock.Any()).
			Times(0)
	}
	useCaseTenantNotFound := func(mockUC *mocks.MockGetSensorHistoricalDataUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetSensorHistoricalData(gomock.Any()).
			Return(nil, tenant.ErrTenantNotFound).
			Times(1)
	}
	useCaseUnauthorized := func(mockUC *mocks.MockGetSensorHistoricalDataUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetSensorHistoricalData(gomock.Any()).
			Return(nil, identity.ErrUnauthorizedAccess).
			Times(1)
	}
	useCaseInvalidDateRange := func(mockUC *mocks.MockGetSensorHistoricalDataUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetSensorHistoricalData(gomock.Any()).
			Return(nil, historical_data.ErrInvalidDateRange).
			Times(1)
	}
	errMock := errors.New("unexpected error")
	useCaseUnexpectedError := func(mockUC *mocks.MockGetSensorHistoricalDataUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetSensorHistoricalData(gomock.Any()).
			Return(nil, errMock).
			Times(1)
	}

	requester := identity.Requester{
		RequesterUserId:   1,
		RequesterTenantId: &targetTenantId,
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	baseURL := "/tenant/" + targetTenantId.String() + "/sensor/" + targetSensorId.String() + "/historical_data"

	cases := []genericControllerTestCase[mocks.MockGetSensorHistoricalDataUseCase]{
		{
			name:             "Success: returns historical data",
			method:           http.MethodGet,
			url:              baseURL,
			requester:        requester,
			setupSteps:       []mockUseCaseSetupFunc[mocks.MockGetSensorHistoricalDataUseCase]{useCaseOK},
			expectedStatus:   http.StatusOK,
			expectedResponse: expectedResponse,
		},
		{
			name:             "Success: supports query params",
			method:           http.MethodGet,
			url:              baseURL + "?from=2026-03-29T12:00:00Z&to=2026-03-29T13:00:00Z&limit=10",
			requester:        requester,
			setupSteps:       []mockUseCaseSetupFunc[mocks.MockGetSensorHistoricalDataUseCase]{useCaseOK},
			expectedStatus:   http.StatusOK,
			expectedResponse: expectedResponse,
		},
		{
			name:             "Fail: missing identity",
			method:           http.MethodGet,
			url:              baseURL,
			omitIdentity:     true,
			setupSteps:       []mockUseCaseSetupFunc[mocks.MockGetSensorHistoricalDataUseCase]{useCaseNeverCalled},
			expectedStatus:   http.StatusUnauthorized,
			expectedResponse: hasError{},
		},
		{
			name:             "Fail: invalid tenant id in URI",
			method:           http.MethodGet,
			url:              "/tenant/not-a-uuid/sensor/" + targetSensorId.String() + "/historical_data",
			requester:        requester,
			setupSteps:       []mockUseCaseSetupFunc[mocks.MockGetSensorHistoricalDataUseCase]{useCaseNeverCalled},
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: hasError{},
		},
		{
			name:             "Fail: invalid timestamp query",
			method:           http.MethodGet,
			url:              baseURL + "?from=not-a-date",
			requester:        requester,
			setupSteps:       []mockUseCaseSetupFunc[mocks.MockGetSensorHistoricalDataUseCase]{useCaseNeverCalled},
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: hasError{},
		},
		{
			name:             "Fail: invalid limit query validation",
			method:           http.MethodGet,
			url:              baseURL + "?limit=-1",
			requester:        requester,
			setupSteps:       []mockUseCaseSetupFunc[mocks.MockGetSensorHistoricalDataUseCase]{useCaseNeverCalled},
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: hasError{},
		},
		{
			name:             "Fail: tenant not found",
			method:           http.MethodGet,
			url:              baseURL,
			requester:        requester,
			setupSteps:       []mockUseCaseSetupFunc[mocks.MockGetSensorHistoricalDataUseCase]{useCaseTenantNotFound},
			expectedStatus:   http.StatusNotFound,
			expectedResponse: hasError{},
		},
		{
			name:             "Fail: unauthorized access",
			method:           http.MethodGet,
			url:              baseURL,
			requester:        requester,
			setupSteps:       []mockUseCaseSetupFunc[mocks.MockGetSensorHistoricalDataUseCase]{useCaseUnauthorized},
			expectedStatus:   http.StatusNotFound,
			expectedResponse: hasError{},
		},
		{
			name:             "Fail: invalid date range returned by use case",
			method:           http.MethodGet,
			url:              baseURL,
			requester:        requester,
			setupSteps:       []mockUseCaseSetupFunc[mocks.MockGetSensorHistoricalDataUseCase]{useCaseInvalidDateRange},
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: hasError{},
		},
		{
			name:             "Fail: unexpected use case error",
			method:           http.MethodGet,
			url:              baseURL,
			requester:        requester,
			setupSteps:       []mockUseCaseSetupFunc[mocks.MockGetSensorHistoricalDataUseCase]{useCaseUnexpectedError},
			expectedStatus:   http.StatusInternalServerError,
			expectedResponse: hasError{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockUseCase := setupMockUseCase(
				mocks.NewMockGetSensorHistoricalDataUseCase,
				tc.setupSteps,
				t,
			)
			controller := historical_data.NewHistoricalDataController(zap.NewNop(), mockUseCase)

			executeControllerTest(
				t,
				tc,
				http.MethodGet,
				"/tenant/:tenant_id/sensor/:sensor_id/historical_data",
				controller.GetSensorHistoricalData,
			)
		})
	}
}
