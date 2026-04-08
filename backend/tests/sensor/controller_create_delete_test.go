package sensor_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"backend/internal/gateway"
	transportHttp "backend/internal/infra/transport/http"
	"backend/internal/sensor"
	"backend/internal/shared/identity"
	helper "backend/tests/helper"
	"backend/tests/sensor/mocks"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

// CREATE ==============================================================================================================

func TestSensorController_CreateSensor(t *testing.T) {
	targetTenantId := uuid.New()
	targetGatewayId := uuid.New()
	targetSensorId := uuid.New()

	requester := identity.Requester{
		RequesterUserId:   uint(1),
		RequesterTenantId: &targetTenantId,
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	validPayload := sensor.CreateSensorBodyDTO{
		Name:      "Heart monitor",
		Interval:  1500,
		Profile:   sensor.HEART_RATE,
		GatewayId: targetGatewayId,
	}

	invalidPayload := map[string]any{
		"data_interval": 0,
		"profile":       "invalid_profile",
		"gateway_id":    "not-uuid",
	}

	expectedSensor := sensor.Sensor{
		Id:        targetSensorId,
		GatewayId: targetGatewayId,
		Name:      "Heart monitor",
		Interval:  1500 * time.Millisecond,
		Status:    sensor.Active,
		Profile:   sensor.HEART_RATE,
	}

	expectedCommand := sensor.CreateSensorCommand{
		Requester: requester,
		Name:      validPayload.Name,
		Interval:  1500 * time.Millisecond,
		Profile:   validPayload.Profile,
		GatewayId: validPayload.GatewayId,
	}

	useCaseOk := func(mockUC *mocks.MockCreateSensorUseCase) *gomock.Call {
		return mockUC.EXPECT().
			CreateSensor(gomock.Eq(expectedCommand)).
			Return(expectedSensor, nil).
			Times(1)
	}

	useCaseNeverCalled := func(mockUC *mocks.MockCreateSensorUseCase) *gomock.Call {
		return mockUC.EXPECT().CreateSensor(gomock.Any()).Times(0)
	}

	useCaseGatewayNotFound := func(mockUC *mocks.MockCreateSensorUseCase) *gomock.Call {
		return mockUC.EXPECT().
			CreateSensor(gomock.Eq(expectedCommand)).
			Return(sensor.Sensor{}, gateway.ErrGatewayNotFound).
			Times(1)
	}

	useCaseUnauthorizedAccess := func(mockUC *mocks.MockCreateSensorUseCase) *gomock.Call {
		return mockUC.EXPECT().
			CreateSensor(gomock.Eq(expectedCommand)).
			Return(sensor.Sensor{}, identity.ErrUnauthorizedAccess).
			Times(1)
	}

	errMock := errors.New("unexpected error")
	useCaseUnexpectedErr := func(mockUC *mocks.MockCreateSensorUseCase) *gomock.Call {
		return mockUC.EXPECT().
			CreateSensor(gomock.Eq(expectedCommand)).
			Return(sensor.Sensor{}, errMock).
			Times(1)
	}

	cases := []helper.GenericControllerTestCase[any, mocks.MockCreateSensorUseCase]{
		{
			Name:      "200 OK",
			Method:    "POST",
			Url:       "/sensor",
			InputDto:  validPayload,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockCreateSensorUseCase]{
				useCaseOk,
			},
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: sensor.NewSensorResponseDTO(expectedSensor),
		},
		{
			Name:      "400 Bad Request: Invalid body",
			Method:    "POST",
			Url:       "/sensor",
			InputDto:  invalidPayload,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockCreateSensorUseCase]{
				useCaseNeverCalled,
			},
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedResponse: helper.HasError{},
		},
		{
			Name:         "401 Unauthorized: No identity",
			Method:       "POST",
			Url:          "/sensor",
			InputDto:     validPayload,
			Requester:    requester,
			OmitIdentity: true,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockCreateSensorUseCase]{
				useCaseNeverCalled,
			},
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedResponse: gin.H{
				"error": transportHttp.ErrMissingIdentity.Error(),
			},
		},
		{
			Name:      "404 Not Found: Gateway not found (obfuscated)",
			Method:    "POST",
			Url:       "/sensor",
			InputDto:  validPayload,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockCreateSensorUseCase]{
				useCaseGatewayNotFound,
			},
			ExpectedStatus: http.StatusNotFound,
			ExpectedResponse: gin.H{
				"error": gateway.ErrGatewayNotFound.Error(),
			},
		},
		{
			Name:      "404 Not Found: Unauthorized access (obfuscated)",
			Method:    "POST",
			Url:       "/sensor",
			InputDto:  validPayload,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockCreateSensorUseCase]{
				useCaseUnauthorizedAccess,
			},
			ExpectedStatus: http.StatusNotFound,
			ExpectedResponse: gin.H{
				"error": gateway.ErrGatewayNotFound.Error(),
			},
		},
		{
			Name:      "500 Server Error: Unexpected error",
			Method:    "POST",
			Url:       "/sensor",
			InputDto:  validPayload,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockCreateSensorUseCase]{
				useCaseUnexpectedErr,
			},
			ExpectedStatus: http.StatusInternalServerError,
			ExpectedResponse: gin.H{
				"error": errMock.Error(),
			},
		},
	}

	mountMethod := "POST"
	mountURL := "/sensor"

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			mockUseCase := helper.SetupMockUseCase(
				mocks.NewMockCreateSensorUseCase,
				tc.SetupSteps,
				t,
			)

			sensorController := sensor.NewSensorController(
				nil,
				mockUseCase,
				nil,
				nil,
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
				sensorController.CreateSensor,
			)
		})
	}
}

// DELETE ==============================================================================================================

func TestSensorController_DeleteSensor(t *testing.T) {
	targetGatewayId := uuid.New()
	targetSensorId := uuid.New()

	requester := identity.Requester{
		RequesterUserId: uint(1),
		RequesterRole:   identity.ROLE_TENANT_ADMIN,
	}

	type Empty struct{}
	validPayload := Empty{}

	expectedSensor := sensor.Sensor{
		Id:        targetSensorId,
		GatewayId: targetGatewayId,
		Name:      "Heart monitor",
		Interval:  2 * time.Second,
		Status:    sensor.Inactive,
		Profile:   sensor.HEART_RATE,
	}

	expectedCommand := sensor.DeleteSensorCommand{
		Requester: requester,
		SensorId:  targetSensorId,
	}

	useCaseOk := func(mockUC *mocks.MockDeleteSensorUseCase) *gomock.Call {
		return mockUC.EXPECT().
			DeleteSensor(gomock.Eq(expectedCommand)).
			Return(expectedSensor, nil).
			Times(1)
	}

	useCaseNeverCalled := func(mockUC *mocks.MockDeleteSensorUseCase) *gomock.Call {
		return mockUC.EXPECT().DeleteSensor(gomock.Any()).Times(0)
	}

	useCaseNotFound := func(mockUC *mocks.MockDeleteSensorUseCase) *gomock.Call {
		return mockUC.EXPECT().
			DeleteSensor(gomock.Eq(expectedCommand)).
			Return(sensor.Sensor{}, sensor.ErrSensorNotFound).
			Times(1)
	}

	useCaseUnauthorizedAccess := func(mockUC *mocks.MockDeleteSensorUseCase) *gomock.Call {
		return mockUC.EXPECT().
			DeleteSensor(gomock.Eq(expectedCommand)).
			Return(sensor.Sensor{}, identity.ErrUnauthorizedAccess).
			Times(1)
	}

	errMock := errors.New("unexpected error")
	useCaseUnexpectedErr := func(mockUC *mocks.MockDeleteSensorUseCase) *gomock.Call {
		return mockUC.EXPECT().
			DeleteSensor(gomock.Eq(expectedCommand)).
			Return(sensor.Sensor{}, errMock).
			Times(1)
	}

	validURL := "/sensor/" + targetSensorId.String()
	invalidURL := "/sensor/not-a-uuid"

	cases := []helper.GenericControllerTestCase[Empty, mocks.MockDeleteSensorUseCase]{
		{
			Name:      "200 OK",
			Method:    "DELETE",
			Url:       validURL,
			InputDto:  validPayload,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockDeleteSensorUseCase]{
				useCaseOk,
			},
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: sensor.NewSensorResponseDTO(expectedSensor),
		},
		{
			Name:      "400 Bad Request: Invalid sensor ID",
			Method:    "DELETE",
			Url:       invalidURL,
			InputDto:  validPayload,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockDeleteSensorUseCase]{
				useCaseNeverCalled,
			},
			ExpectedStatus: http.StatusBadRequest,
			ExpectedResponse: gin.H{
				"error": "invalid sensor ID",
			},
		},
		{
			Name:         "401 Unauthorized: No identity",
			Method:       "DELETE",
			Url:          validURL,
			InputDto:     validPayload,
			Requester:    requester,
			OmitIdentity: true,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockDeleteSensorUseCase]{
				useCaseNeverCalled,
			},
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedResponse: gin.H{
				"error": transportHttp.ErrMissingIdentity.Error(),
			},
		},
		{
			Name:      "404 Not Found: Sensor not found",
			Method:    "DELETE",
			Url:       validURL,
			InputDto:  validPayload,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockDeleteSensorUseCase]{
				useCaseNotFound,
			},
			ExpectedStatus: http.StatusNotFound,
			ExpectedResponse: gin.H{
				"error": sensor.ErrSensorNotFound.Error(),
			},
		},
		{
			Name:      "404 Not Found: Unauthorized access (obfuscated)",
			Method:    "DELETE",
			Url:       validURL,
			InputDto:  validPayload,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockDeleteSensorUseCase]{
				useCaseUnauthorizedAccess,
			},
			ExpectedStatus: http.StatusNotFound,
			ExpectedResponse: gin.H{
				"error": sensor.ErrSensorNotFound.Error(),
			},
		},
		{
			Name:      "500 Server Error: Unexpected error",
			Method:    "DELETE",
			Url:       validURL,
			InputDto:  validPayload,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockDeleteSensorUseCase]{
				useCaseUnexpectedErr,
			},
			ExpectedStatus: http.StatusInternalServerError,
			ExpectedResponse: gin.H{
				"error": errMock.Error(),
			},
		},
	}

	mountMethod := "DELETE"
	mountURL := "/sensor/:sensor_id"

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			mockUseCase := helper.SetupMockUseCase(
				mocks.NewMockDeleteSensorUseCase,
				tc.SetupSteps,
				t,
			)

			sensorController := sensor.NewSensorController(
				nil,
				nil,
				mockUseCase,
				nil,
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
				sensorController.DeleteSensor,
			)
		})
	}
}
