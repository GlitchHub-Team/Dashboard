package sensor_test

import (
	"errors"
	"net/http"
	"testing"

	transportHttp "backend/internal/infra/transport/http"
	"backend/internal/sensor"
	"backend/internal/shared/identity"
	helper "backend/tests/helper"
	"backend/tests/sensor/mocks"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestSensorController_InterruptSensor(t *testing.T) {
	targetTenantId := uuid.New()
	targetSensorId := uuid.New()

	requester := identity.Requester{
		RequesterUserId:   uint(1),
		RequesterTenantId: &targetTenantId,
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	expectedCommand := sensor.InterruptSensorCommand{
		Requester: requester,
		SensorId:  targetSensorId,
	}

	useCaseOk := func(mockUC *mocks.MockInterruptSensorUseCase) *gomock.Call {
		return mockUC.EXPECT().
			InterruptSensor(gomock.Eq(expectedCommand)).
			Return(nil).
			Times(1)
	}

	useCaseNeverCalled := func(mockUC *mocks.MockInterruptSensorUseCase) *gomock.Call {
		return mockUC.EXPECT().InterruptSensor(gomock.Any()).Times(0)
	}

	useCaseSensorNotFound := func(mockUC *mocks.MockInterruptSensorUseCase) *gomock.Call {
		return mockUC.EXPECT().
			InterruptSensor(gomock.Eq(expectedCommand)).
			Return(sensor.ErrSensorNotFound).
			Times(1)
	}

	useCaseUnauthorizedAccess := func(mockUC *mocks.MockInterruptSensorUseCase) *gomock.Call {
		return mockUC.EXPECT().
			InterruptSensor(gomock.Eq(expectedCommand)).
			Return(identity.ErrUnauthorizedAccess).
			Times(1)
	}

	errMock := errors.New("unexpected server error")
	useCaseUnexpectedErr := func(mockUC *mocks.MockInterruptSensorUseCase) *gomock.Call {
		return mockUC.EXPECT().
			InterruptSensor(gomock.Eq(expectedCommand)).
			Return(errMock).
			Times(1)
	}

	validURL := "/sensor/" + targetSensorId.String() + "/interrupt"
	invalidURL := "/sensor/not-a-uuid/interrupt"

	cases := []helper.GenericControllerTestCase[any, mocks.MockInterruptSensorUseCase]{
		{
			Name:         "401 Unauthorized: credenziali mancanti",
			Method:       "POST",
			Url:          validURL,
			InputDto:     nil,
			Requester:    requester,
			OmitIdentity: true,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockInterruptSensorUseCase]{
				useCaseNeverCalled,
			},
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedResponse: gin.H{
				"error": transportHttp.ErrMissingIdentity.Error(),
			},
		},
		{
			Name:      "400 Bad Request: sensorId nell'url non valido",
			Method:    "POST",
			Url:       invalidURL,
			InputDto:  nil,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockInterruptSensorUseCase]{
				useCaseNeverCalled,
			},
			ExpectedStatus: http.StatusBadRequest,
			ExpectedResponse: gin.H{
				"error": sensor.ErrInvalidSensorID.Error(),
			},
		},
		{
			Name:      "200 OK: sensorId valido",
			Method:    "POST",
			Url:       validURL,
			InputDto:  nil,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockInterruptSensorUseCase]{
				useCaseOk,
			},
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: nil,
		},
		{
			Name:      "404 Not Found: sensor non trovato",
			Method:    "POST",
			Url:       validURL,
			InputDto:  nil,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockInterruptSensorUseCase]{
				useCaseSensorNotFound,
			},
			ExpectedStatus: http.StatusNotFound,
			ExpectedResponse: gin.H{
				"error": sensor.ErrSensorNotFound.Error(),
			},
		},
		{
			Name:      "404 Not Found: accesso non autorizzato",
			Method:    "POST",
			Url:       validURL,
			InputDto:  nil,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockInterruptSensorUseCase]{
				useCaseUnauthorizedAccess,
			},
			ExpectedStatus: http.StatusNotFound,
			ExpectedResponse: gin.H{
				"error": sensor.ErrSensorNotFound.Error(),
			},
		},
		{
			Name:      "500 Server Error: errore interno",
			Method:    "POST",
			Url:       validURL,
			InputDto:  nil,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockInterruptSensorUseCase]{
				useCaseUnexpectedErr,
			},
			ExpectedStatus: http.StatusInternalServerError,
			ExpectedResponse: gin.H{
				"error": errMock.Error(),
			},
		},
		{
			Name:      "200 OK: caso valido",
			Method:    "POST",
			Url:       validURL,
			InputDto:  nil,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockInterruptSensorUseCase]{
				useCaseOk,
			},
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: nil,
		},
	}

	mountMethod := "POST"
	mountURL := "/sensor/:sensor_id/interrupt"

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			mockUseCase := helper.SetupMockUseCase(
				mocks.NewMockInterruptSensorUseCase,
				tc.SetupSteps,
				t,
			)

			sensorController := sensor.NewSensorController(
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				mockUseCase,
				nil,
			)

			helper.ExecuteControllerTest(
				t,
				tc,
				mountMethod,
				mountURL,
				sensorController.InterruptSensor,
			)
		})
	}
}

func TestSensorController_ResumeSensor(t *testing.T) {
	targetTenantId := uuid.New()
	targetSensorId := uuid.New()

	requester := identity.Requester{
		RequesterUserId:   uint(1),
		RequesterTenantId: &targetTenantId,
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	expectedCommand := sensor.ResumeSensorCommand{
		Requester: requester,
		SensorId:  targetSensorId,
	}

	useCaseOk := func(mockUC *mocks.MockResumeSensorUseCase) *gomock.Call {
		return mockUC.EXPECT().
			ResumeSensor(gomock.Eq(expectedCommand)).
			Return(nil).
			Times(1)
	}

	useCaseNeverCalled := func(mockUC *mocks.MockResumeSensorUseCase) *gomock.Call {
		return mockUC.EXPECT().ResumeSensor(gomock.Any()).Times(0)
	}

	useCaseSensorNotFound := func(mockUC *mocks.MockResumeSensorUseCase) *gomock.Call {
		return mockUC.EXPECT().
			ResumeSensor(gomock.Eq(expectedCommand)).
			Return(sensor.ErrSensorNotFound).
			Times(1)
	}

	useCaseUnauthorizedAccess := func(mockUC *mocks.MockResumeSensorUseCase) *gomock.Call {
		return mockUC.EXPECT().
			ResumeSensor(gomock.Eq(expectedCommand)).
			Return(identity.ErrUnauthorizedAccess).
			Times(1)
	}

	errMock := errors.New("unexpected server error")
	useCaseUnexpectedErr := func(mockUC *mocks.MockResumeSensorUseCase) *gomock.Call {
		return mockUC.EXPECT().
			ResumeSensor(gomock.Eq(expectedCommand)).
			Return(errMock).
			Times(1)
	}

	validURL := "/sensor/" + targetSensorId.String() + "/resume"
	invalidURL := "/sensor/not-a-uuid/resume"

	cases := []helper.GenericControllerTestCase[any, mocks.MockResumeSensorUseCase]{
		{
			Name:         "401 Unauthorized: credenziali mancanti",
			Method:       "POST",
			Url:          validURL,
			InputDto:     nil,
			Requester:    requester,
			OmitIdentity: true,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockResumeSensorUseCase]{
				useCaseNeverCalled,
			},
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedResponse: gin.H{
				"error": transportHttp.ErrMissingIdentity.Error(),
			},
		},
		{
			Name:      "400 Bad Request: sensorId nell'url non valido",
			Method:    "POST",
			Url:       invalidURL,
			InputDto:  nil,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockResumeSensorUseCase]{
				useCaseNeverCalled,
			},
			ExpectedStatus: http.StatusBadRequest,
			ExpectedResponse: gin.H{
				"error": sensor.ErrInvalidSensorID.Error(),
			},
		},
		{
			Name:      "200 OK: sensorId valido",
			Method:    "POST",
			Url:       validURL,
			InputDto:  nil,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockResumeSensorUseCase]{
				useCaseOk,
			},
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: nil,
		},
		{
			Name:      "404 Not Found: sensor non trovato",
			Method:    "POST",
			Url:       validURL,
			InputDto:  nil,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockResumeSensorUseCase]{
				useCaseSensorNotFound,
			},
			ExpectedStatus: http.StatusNotFound,
			ExpectedResponse: gin.H{
				"error": sensor.ErrSensorNotFound.Error(),
			},
		},
		{
			Name:      "404 Not Found: accesso non autorizzato",
			Method:    "POST",
			Url:       validURL,
			InputDto:  nil,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockResumeSensorUseCase]{
				useCaseUnauthorizedAccess,
			},
			ExpectedStatus: http.StatusNotFound,
			ExpectedResponse: gin.H{
				"error": sensor.ErrSensorNotFound.Error(),
			},
		},
		{
			Name:      "500 Server Error: errore interno",
			Method:    "POST",
			Url:       validURL,
			InputDto:  nil,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockResumeSensorUseCase]{
				useCaseUnexpectedErr,
			},
			ExpectedStatus: http.StatusInternalServerError,
			ExpectedResponse: gin.H{
				"error": errMock.Error(),
			},
		},
		{
			Name:      "200 OK: caso valido",
			Method:    "POST",
			Url:       validURL,
			InputDto:  nil,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockResumeSensorUseCase]{
				useCaseOk,
			},
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: nil,
		},
	}

	mountMethod := "POST"
	mountURL := "/sensor/:sensor_id/resume"

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			mockUseCase := helper.SetupMockUseCase(
				mocks.NewMockResumeSensorUseCase,
				tc.SetupSteps,
				t,
			)

			sensorController := sensor.NewSensorController(
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				mockUseCase,
			)

			helper.ExecuteControllerTest(
				t,
				tc,
				mountMethod,
				mountURL,
				sensorController.ResumeSensor,
			)
		})
	}
}
