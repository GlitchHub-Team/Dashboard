package sensor_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"backend/internal/gateway"
	transportHttp "backend/internal/infra/transport/http"
	transportHttpDto "backend/internal/infra/transport/http/dto"
	"backend/internal/sensor"
	"backend/internal/shared/identity"
	"backend/internal/tenant"
	helper "backend/tests/helper"
	"backend/tests/sensor/mocks"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestSensorController_GetSensor(t *testing.T) {
	targetTenantId := uuid.New()
	targetGatewayId := uuid.New()
	targetSensorId := uuid.New()

	requester := identity.Requester{
		RequesterUserId:   uint(1),
		RequesterTenantId: &targetTenantId,
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	expectedSensor := sensor.Sensor{
		Id:        targetSensorId,
		GatewayId: targetGatewayId,
		Name:      "Heart monitor",
		Interval:  1200 * time.Millisecond,
		Status:    sensor.Active,
		Profile:   sensor.HEART_RATE,
	}

	expectedCommand := sensor.GetSensorCommand{
		Requester: requester,
		SensorId:  targetSensorId,
	}

	useCaseOk := func(mockUC *mocks.MockGetSensorUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetSensorById(gomock.Eq(expectedCommand)).
			Return(expectedSensor, nil).
			Times(1)
	}

	useCaseNeverCalled := func(mockUC *mocks.MockGetSensorUseCase) *gomock.Call {
		return mockUC.EXPECT().GetSensorById(gomock.Any()).Times(0)
	}

	useCaseSensorNotFound := func(mockUC *mocks.MockGetSensorUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetSensorById(gomock.Eq(expectedCommand)).
			Return(sensor.Sensor{}, sensor.ErrSensorNotFound).
			Times(1)
	}

	useCaseUnauthorized := func(mockUC *mocks.MockGetSensorUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetSensorById(gomock.Eq(expectedCommand)).
			Return(sensor.Sensor{}, identity.ErrUnauthorizedAccess).
			Times(1)
	}

	errMock := errors.New("unexpected server error")
	useCaseUnexpectedErr := func(mockUC *mocks.MockGetSensorUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetSensorById(gomock.Eq(expectedCommand)).
			Return(sensor.Sensor{}, errMock).
			Times(1)
	}

	validURL := "/sensor/" + targetSensorId.String()
	invalidURL := "/sensor/not-a-uuid"

	cases := []helper.GenericControllerTestCase[any, mocks.MockGetSensorUseCase]{
		{
			Name:         "401 Unauthorized: Utente senza credenziali",
			Method:       "GET",
			Url:          validURL,
			InputDto:     nil,
			Requester:    requester,
			OmitIdentity: true,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetSensorUseCase]{
				useCaseNeverCalled,
			},
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedResponse: gin.H{
				"error": transportHttp.ErrMissingIdentity.Error(),
			},
		},
		{
			Name:      "400 Bad Request: sensorId invalido",
			Method:    "GET",
			Url:       invalidURL,
			InputDto:  nil,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetSensorUseCase]{
				useCaseNeverCalled,
			},
			ExpectedStatus: http.StatusBadRequest,
			ExpectedResponse: gin.H{
				"error": sensor.ErrInvalidSensorID.Error(),
			},
		},
		{
			Name:      "200 OK: sensorId valido",
			Method:    "GET",
			Url:       validURL,
			InputDto:  nil,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetSensorUseCase]{
				useCaseOk,
			},
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: sensor.NewSensorResponseDTO(expectedSensor),
		},
		{
			Name:      "404 Not Found: sensore non trovato",
			Method:    "GET",
			Url:       validURL,
			InputDto:  nil,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetSensorUseCase]{
				useCaseSensorNotFound,
			},
			ExpectedStatus: http.StatusNotFound,
			ExpectedResponse: gin.H{
				"error": sensor.ErrSensorNotFound.Error(),
			},
		},
		{
			Name:      "404 Not Found: utente non autorizzato",
			Method:    "GET",
			Url:       validURL,
			InputDto:  nil,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetSensorUseCase]{
				useCaseUnauthorized,
			},
			ExpectedStatus: http.StatusNotFound,
			ExpectedResponse: gin.H{
				"error": sensor.ErrSensorNotFound.Error(),
			},
		},
		{
			Name:      "500 Server Error: errore generico",
			Method:    "GET",
			Url:       validURL,
			InputDto:  nil,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetSensorUseCase]{
				useCaseUnexpectedErr,
			},
			ExpectedStatus: http.StatusInternalServerError,
			ExpectedResponse: gin.H{
				"error": errMock.Error(),
			},
		},
	}

	mountMethod := "GET"
	mountURL := "/sensor/:sensorId"

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			mockUseCase := helper.SetupMockUseCase(
				mocks.NewMockGetSensorUseCase,
				tc.SetupSteps,
				t,
			)

			sensorController := sensor.NewSensorController(
				nil,
				nil,
				nil,
				mockUseCase,
				nil,
				nil,
				nil,
				nil,
			)

			helper.ExecuteControllerTest(
				t,
				tc,
				mountMethod,
				mountURL,
				sensorController.GetSensor,
			)
		})
	}
}

func TestSensorController_GetSensorsByGateway(t *testing.T) {
	targetTenantId := uuid.New()
	targetGatewayId := uuid.New()
	sensorOneId := uuid.New()
	sensorTwoId := uuid.New()

	requester := identity.Requester{
		RequesterUserId:   uint(1),
		RequesterTenantId: &targetTenantId,
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	sensors := []sensor.Sensor{
		{
			Id:        sensorOneId,
			GatewayId: targetGatewayId,
			Name:      "A",
			Interval:  1500 * time.Millisecond,
			Status:    sensor.Active,
			Profile:   sensor.HEART_RATE,
		},
		{
			Id:        sensorTwoId,
			GatewayId: targetGatewayId,
			Name:      "B",
			Interval:  2 * time.Second,
			Status:    sensor.Inactive,
			Profile:   sensor.ENVIRONMENTAL_SENSING,
		},
	}

	defaultCmd := sensor.GetSensorsByGatewayCommand{
		Requester: requester,
		Page:      transportHttpDto.DEFAULT_PAGINATION.Page,
		Limit:     transportHttpDto.DEFAULT_PAGINATION.Limit,
		GatewayId: targetGatewayId,
	}

	paginatedCmd := sensor.GetSensorsByGatewayCommand{
		Requester: requester,
		Page:      2,
		Limit:     10,
		GatewayId: targetGatewayId,
	}

	useCaseOkDefault := func(mockUC *mocks.MockGetSensorsByGatewayUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetSensorsByGateway(gomock.Eq(defaultCmd)).
			Return(sensors, uint(2), nil).
			Times(1)
	}

	useCaseOkPagination := func(mockUC *mocks.MockGetSensorsByGatewayUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetSensorsByGateway(gomock.Eq(paginatedCmd)).
			Return(sensors, uint(22), nil).
			Times(1)
	}

	useCaseNeverCalled := func(mockUC *mocks.MockGetSensorsByGatewayUseCase) *gomock.Call {
		return mockUC.EXPECT().GetSensorsByGateway(gomock.Any()).Times(0)
	}

	useCaseGatewayNotFound := func(mockUC *mocks.MockGetSensorsByGatewayUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetSensorsByGateway(gomock.Eq(defaultCmd)).
			Return(nil, uint(0), gateway.ErrGatewayNotFound).
			Times(1)
	}

	useCaseUnauthorized := func(mockUC *mocks.MockGetSensorsByGatewayUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetSensorsByGateway(gomock.Eq(defaultCmd)).
			Return(nil, uint(0), identity.ErrUnauthorizedAccess).
			Times(1)
	}

	errMock := errors.New("unexpected server error")
	useCaseUnexpectedErr := func(mockUC *mocks.MockGetSensorsByGatewayUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetSensorsByGateway(gomock.Eq(defaultCmd)).
			Return(nil, uint(0), errMock).
			Times(1)
	}

	validURL := "/gateway/" + targetGatewayId.String() + "/sensors"
	validURLWithCorrectPagination := validURL + "?page=2&limit=10"
	invalidGatewayURL := "/gateway/not-a-uuid/sensors"
	invalidPaginationURL := validURL + "?page=abc&limit=def"

	cases := []helper.GenericControllerTestCase[any, mocks.MockGetSensorsByGatewayUseCase]{
		{
			Name:         "401 Unauthorized: utente senza credenziali",
			Method:       "GET",
			Url:          validURL,
			InputDto:     nil,
			Requester:    requester,
			OmitIdentity: true,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetSensorsByGatewayUseCase]{
				useCaseNeverCalled,
			},
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedResponse: gin.H{
				"error": transportHttp.ErrMissingIdentity.Error(),
			},
		},
		{
			Name:      "400 Bad Request: gatewayId non valido",
			Method:    "GET",
			Url:       invalidGatewayURL,
			InputDto:  nil,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetSensorsByGatewayUseCase]{
				useCaseNeverCalled,
			},
			ExpectedStatus: http.StatusBadRequest,
			ExpectedResponse: gin.H{
				"error": gateway.ErrInvalidGatewayID.Error(),
			},
		},
		{
			Name:      "200 OK: gatewayId valido",
			Method:    "GET",
			Url:       validURL,
			InputDto:  nil,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetSensorsByGatewayUseCase]{
				useCaseOkDefault,
			},
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: sensor.NewSensorsResponseDTO(sensors, uint(2)),
		},
		{
			Name:      "200 OK: pagination di default",
			Method:    "GET",
			Url:       validURL,
			InputDto:  nil,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetSensorsByGatewayUseCase]{
				useCaseOkDefault,
			},
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: sensor.NewSensorsResponseDTO(sensors, uint(2)),
		},
		{
			Name:      "400 Bad Request: pagination errata",
			Method:    "GET",
			Url:       invalidPaginationURL,
			InputDto:  nil,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetSensorsByGatewayUseCase]{
				useCaseNeverCalled,
			},
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedResponse: helper.HasError{},
		},
		{
			Name:      "200 OK: paginato corretta",
			Method:    "GET",
			Url:       validURLWithCorrectPagination,
			InputDto:  nil,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetSensorsByGatewayUseCase]{
				useCaseOkPagination,
			},
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: sensor.NewSensorsResponseDTO(sensors, uint(22)),
		},
		{
			Name:      "404 Not Found: gateway non trovato",
			Method:    "GET",
			Url:       validURL,
			InputDto:  nil,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetSensorsByGatewayUseCase]{
				useCaseGatewayNotFound,
			},
			ExpectedStatus: http.StatusNotFound,
			ExpectedResponse: gin.H{
				"error": gateway.ErrGatewayNotFound.Error(),
			},
		},
		{
			Name:      "404 Not Found: gateway non accessibile dall'utente",
			Method:    "GET",
			Url:       validURL,
			InputDto:  nil,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetSensorsByGatewayUseCase]{
				useCaseUnauthorized,
			},
			ExpectedStatus: http.StatusNotFound,
			ExpectedResponse: gin.H{
				"error": gateway.ErrGatewayNotFound.Error(),
			},
		},
		{
			Name:      "500 Server Error: server error generico",
			Method:    "GET",
			Url:       validURL,
			InputDto:  nil,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetSensorsByGatewayUseCase]{
				useCaseUnexpectedErr,
			},
			ExpectedStatus: http.StatusInternalServerError,
			ExpectedResponse: gin.H{
				"error": errMock.Error(),
			},
		},
	}

	mountMethod := "GET"
	mountURL := "/gateway/:gatewayId/sensors"

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			mockUseCase := helper.SetupMockUseCase(
				mocks.NewMockGetSensorsByGatewayUseCase,
				tc.SetupSteps,
				t,
			)

			sensorController := sensor.NewSensorController(
				nil,
				nil,
				nil,
				nil,
				mockUseCase,
				nil,
				nil,
				nil,
			)

			helper.ExecuteControllerTest(
				t,
				tc,
				mountMethod,
				mountURL,
				sensorController.GetSensorsByGateway,
			)
		})
	}
}

func TestSensorController_GetSensorsByTenant(t *testing.T) {
	targetTenantId := uuid.New()
	targetGatewayId := uuid.New()
	targetSensorId := uuid.New()

	requester := identity.Requester{
		RequesterUserId:   uint(1),
		RequesterTenantId: &targetTenantId,
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	sensors := []sensor.Sensor{
		{
			Id:        targetSensorId,
			GatewayId: targetGatewayId,
			Name:      "Tenant sensor",
			Interval:  2200 * time.Millisecond,
			Status:    sensor.Active,
			Profile:   sensor.HEART_RATE,
		},
	}

	defaultCmd := sensor.GetSensorsByTenantCommand{
		Requester: requester,
		Page:      transportHttpDto.DEFAULT_PAGINATION.Page,
		Limit:     transportHttpDto.DEFAULT_PAGINATION.Limit,
		TenantId:  targetTenantId,
	}

	paginatedCmd := sensor.GetSensorsByTenantCommand{
		Requester: requester,
		Page:      2,
		Limit:     10,
		TenantId:  targetTenantId,
	}

	useCaseOkDefault := func(mockUC *mocks.MockGetSensorsByTenantUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetSensorsByTenant(gomock.Eq(defaultCmd)).
			Return(sensors, uint(1), nil).
			Times(1)
	}

	useCaseOkPagination := func(mockUC *mocks.MockGetSensorsByTenantUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetSensorsByTenant(gomock.Eq(paginatedCmd)).
			Return(sensors, uint(11), nil).
			Times(1)
	}

	useCaseNeverCalled := func(mockUC *mocks.MockGetSensorsByTenantUseCase) *gomock.Call {
		return mockUC.EXPECT().GetSensorsByTenant(gomock.Any()).Times(0)
	}

	useCaseTenantNotFound := func(mockUC *mocks.MockGetSensorsByTenantUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetSensorsByTenant(gomock.Eq(defaultCmd)).
			Return(nil, uint(0), tenant.ErrTenantNotFound).
			Times(1)
	}

	useCaseUnauthorized := func(mockUC *mocks.MockGetSensorsByTenantUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetSensorsByTenant(gomock.Eq(defaultCmd)).
			Return(nil, uint(0), identity.ErrUnauthorizedAccess).
			Times(1)
	}

	errMock := errors.New("unexpected server error")
	useCaseUnexpectedErr := func(mockUC *mocks.MockGetSensorsByTenantUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetSensorsByTenant(gomock.Eq(defaultCmd)).
			Return(nil, uint(0), errMock).
			Times(1)
	}

	baseURL := "/tenant/" + targetTenantId.String() + "/sensors"
	invalidTenantIdURL := "/tenant/not-a-uuid/sensors"
	invalidPaginationURL := baseURL + "?page=abc&limit=def"
	validPaginationURL := baseURL + "?page=2&limit=10"

	cases := []helper.GenericControllerTestCase[any, mocks.MockGetSensorsByTenantUseCase]{
		{
			Name:         "401 Unauthorized: utente senza credenziali",
			Method:       "GET",
			Url:          baseURL,
			InputDto:     nil,
			Requester:    requester,
			OmitIdentity: true,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetSensorsByTenantUseCase]{
				useCaseNeverCalled,
			},
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedResponse: gin.H{
				"error": transportHttp.ErrMissingIdentity.Error(),
			},
		},
		{
			Name:      "400 Bad Request: tenantId non valido",
			Method:    "GET",
			Url:       invalidTenantIdURL,
			InputDto:  nil,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetSensorsByTenantUseCase]{
				useCaseNeverCalled,
			},
			ExpectedStatus: http.StatusBadRequest,
			ExpectedResponse: gin.H{
				"error": tenant.ErrInvalidTenantID.Error(),
			},
		},
		{
			Name:      "401 Unauthorized: tenantId non trovato",
			Method:    "GET",
			Url:       baseURL,
			InputDto:  nil,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetSensorsByTenantUseCase]{
				useCaseTenantNotFound,
			},
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedResponse: gin.H{
				"error": tenant.ErrTenantNotFound.Error(),
			},
		},
		{
			Name:      "200 OK: pagination di default",
			Method:    "GET",
			Url:       baseURL,
			InputDto:  nil,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetSensorsByTenantUseCase]{
				useCaseOkDefault,
			},
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: sensor.NewSensorsResponseDTO(sensors, uint(1)),
		},
		{
			Name:      "400 Bad Request: pagination errata",
			Method:    "GET",
			Url:       invalidPaginationURL,
			InputDto:  nil,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetSensorsByTenantUseCase]{
				useCaseNeverCalled,
			},
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedResponse: helper.HasError{},
		},
		{
			Name:      "200 OK: paginato corretta",
			Method:    "GET",
			Url:       validPaginationURL,
			InputDto:  nil,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetSensorsByTenantUseCase]{
				useCaseOkPagination,
			},
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: sensor.NewSensorsResponseDTO(sensors, uint(11)),
		},
		{
			Name:      "401 Unauthorized: unauthorized access",
			Method:    "GET",
			Url:       baseURL,
			InputDto:  nil,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetSensorsByTenantUseCase]{
				useCaseUnauthorized,
			},
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedResponse: gin.H{
				"error": identity.ErrUnauthorizedAccess.Error(),
			},
		},
		{
			Name:      "200 OK: caso valido",
			Method:    "GET",
			Url:       baseURL,
			InputDto:  nil,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetSensorsByTenantUseCase]{
				useCaseOkDefault,
			},
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: sensor.NewSensorsResponseDTO(sensors, uint(1)),
		},
		{
			Name:      "500 Server Error: errore generico",
			Method:    "GET",
			Url:       baseURL,
			InputDto:  nil,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetSensorsByTenantUseCase]{
				useCaseUnexpectedErr,
			},
			ExpectedStatus: http.StatusInternalServerError,
			ExpectedResponse: gin.H{
				"error": errMock.Error(),
			},
		},
	}

	mountMethod := "GET"
	mountURL := "/tenant/:tenantId/sensors"

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			mockUseCase := helper.SetupMockUseCase(
				mocks.NewMockGetSensorsByTenantUseCase,
				tc.SetupSteps,
				t,
			)

			sensorController := sensor.NewSensorController(
				nil,
				nil,
				nil,
				nil,
				nil,
				mockUseCase,
				nil,
				nil,
			)

			helper.ExecuteControllerTest(
				t,
				tc,
				mountMethod,
				mountURL,
				sensorController.GetSensorsByTenant,
			)
		})
	}
}
