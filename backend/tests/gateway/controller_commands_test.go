package gateway_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"backend/internal/gateway"
	transportHttp "backend/internal/infra/transport/http"
	"backend/internal/shared/identity"
	mocks "backend/tests/gateway/mocks"
	helper "backend/tests/helper"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestGatewayController_CommissionGateway(t *testing.T) {
	gatewayID := uuid.New()
	tenantID := uuid.New()
	requester := identity.Requester{RequesterUserId: 1, RequesterRole: identity.ROLE_SUPER_ADMIN}

	input := map[string]any{"tenant_id": tenantID.String(), "commission_token": "tok"}
	expectedCommand := gateway.CommissionGatewayCommand{
		Requester:       requester,
		GatewayId:       gatewayID,
		TenantId:        tenantID,
		CommissionToken: "tok",
	}
	resultGateway := gateway.Gateway{Id: gatewayID, Name: "GW-C", TenantId: &tenantID, Status: gateway.GATEWAY_STATUS_ACTIVE, IntervalLimit: 2 * time.Second}

	useCaseOk := func(mockUC *mocks.MockCommissionGatewayUseCase) *gomock.Call {
		return mockUC.EXPECT().CommissionGateway(gomock.Eq(expectedCommand)).Return(resultGateway, nil).Times(1)
	}
	useCaseNever := func(mockUC *mocks.MockCommissionGatewayUseCase) *gomock.Call {
		return mockUC.EXPECT().CommissionGateway(gomock.Any()).Times(0)
	}
	useCaseNotFound := func(mockUC *mocks.MockCommissionGatewayUseCase) *gomock.Call {
		return mockUC.EXPECT().CommissionGateway(gomock.Eq(expectedCommand)).Return(gateway.Gateway{}, gateway.ErrGatewayNotFound).Times(1)
	}

	validURL := "/gateway/" + gatewayID.String() + "/commission"

	cases := []helper.GenericControllerTestCase[any, mocks.MockCommissionGatewayUseCase]{
		{
			Name:      "200 OK",
			Method:    "POST",
			Url:       validURL,
			InputDto:  input,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockCommissionGatewayUseCase]{
				useCaseOk,
			},
			ExpectedStatus: http.StatusOK,
			ExpectedResponse: gin.H{
				"gateway_id":        gatewayID.String(),
				"name":              "GW-C",
				"tenant_id":         tenantID.String(),
				"status":            string(gateway.GATEWAY_STATUS_ACTIVE),
				"interval":          int64(2000),
				"public_identifier": nil,
			},
		},
		{
			Name:      "400 Bad Request: gatewayId non valido",
			Method:    "POST",
			Url:       "/gateway/not-a-uuid/commission",
			InputDto:  input,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockCommissionGatewayUseCase]{
				useCaseNever,
			},
			ExpectedStatus: http.StatusBadRequest,
			ExpectedResponse: gin.H{
				"error": gateway.ErrInvalidGatewayID.Error(),
			},
		},
		{
			Name:      "404 Not Found",
			Method:    "POST",
			Url:       validURL,
			InputDto:  input,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockCommissionGatewayUseCase]{
				useCaseNotFound,
			},
			ExpectedStatus: http.StatusNotFound,
			ExpectedResponse: gin.H{
				"error": gateway.ErrGatewayNotFound.Error(),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			mockUseCase := helper.SetupMockUseCase(mocks.NewMockCommissionGatewayUseCase, tc.SetupSteps, t)
			controller := gateway.NewGatewayController(nil, nil, nil, nil, nil, mockUseCase, nil, nil, nil, nil, nil, nil, nil)
			helper.ExecuteControllerTest(t, tc, "POST", "/gateway/:gateway_id/commission", controller.CommissionGateway)
		})
	}
}

func TestGatewayController_DecommissionGateway(t *testing.T) {
	gatewayID := uuid.New()
	requester := identity.Requester{RequesterUserId: 1, RequesterRole: identity.ROLE_SUPER_ADMIN}

	expectedCommand := gateway.DecommissionGatewayCommand{Requester: requester, GatewayId: gatewayID}
	resultGateway := gateway.Gateway{Id: gatewayID, Name: "GW-D", Status: gateway.GATEWAY_STATUS_DECOMMISSIONED, IntervalLimit: 2 * time.Second}

	useCaseOk := func(mockUC *mocks.MockDecommissionGatewayUseCase) *gomock.Call {
		return mockUC.EXPECT().DecommissionGateway(gomock.Eq(expectedCommand)).Return(resultGateway, nil).Times(1)
	}
	useCaseNever := func(mockUC *mocks.MockDecommissionGatewayUseCase) *gomock.Call {
		return mockUC.EXPECT().DecommissionGateway(gomock.Any()).Times(0)
	}

	cases := []helper.GenericControllerTestCase[any, mocks.MockDecommissionGatewayUseCase]{
		{
			Name:      "200 OK",
			Method:    "POST",
			Url:       "/gateway/" + gatewayID.String() + "/decommission",
			InputDto:  struct{}{},
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockDecommissionGatewayUseCase]{
				useCaseOk,
			},
			ExpectedStatus: http.StatusOK,
			ExpectedResponse: gin.H{
				"gateway_id":        gatewayID.String(),
				"name":              "GW-D",
				"tenant_id":         "",
				"status":            string(gateway.GATEWAY_STATUS_DECOMMISSIONED),
				"interval":          int64(2000),
				"public_identifier": nil,
			},
		},
		{
			Name:         "401 Unauthorized: identity mancante",
			Method:       "POST",
			Url:          "/gateway/" + gatewayID.String() + "/decommission",
			InputDto:     struct{}{},
			Requester:    requester,
			OmitIdentity: true,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockDecommissionGatewayUseCase]{
				useCaseNever,
			},
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedResponse: gin.H{
				"error": transportHttp.ErrMissingIdentity.Error(),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			mockUseCase := helper.SetupMockUseCase(mocks.NewMockDecommissionGatewayUseCase, tc.SetupSteps, t)
			controller := gateway.NewGatewayController(nil, nil, nil, nil, nil, nil, mockUseCase, nil, nil, nil, nil, nil, nil)
			helper.ExecuteControllerTest(t, tc, "POST", "/gateway/:gateway_id/decommission", controller.DecommissionGateway)
		})
	}
}

func TestGatewayController_InterruptResumeResetReboot(t *testing.T) {
	gatewayID := uuid.New()
	requester := identity.Requester{RequesterUserId: 1, RequesterRole: identity.ROLE_SUPER_ADMIN}

	t.Run("Interrupt 200 OK", func(t *testing.T) {
		expected := gateway.InterruptGatewayCommand{Requester: requester, GatewayId: gatewayID}
		steps := []helper.MockUseCaseSetupFunc[mocks.MockInterruptGatewayUseCase]{
			func(mockUC *mocks.MockInterruptGatewayUseCase) *gomock.Call {
				return mockUC.EXPECT().InterruptGateway(gomock.Eq(expected)).Return(gateway.Gateway{}, nil).Times(1)
			},
		}
		mockUseCase := helper.SetupMockUseCase(mocks.NewMockInterruptGatewayUseCase, steps, t)
		controller := gateway.NewGatewayController(nil, nil, nil, nil, nil, nil, nil, mockUseCase, nil, nil, nil, nil, nil)

		tc := helper.GenericControllerTestCase[any, mocks.MockInterruptGatewayUseCase]{
			Method:         "POST",
			Url:            "/gateway/" + gatewayID.String() + "/interrupt",
			InputDto:       struct{}{},
			Requester:      requester,
			ExpectedStatus: http.StatusOK,
			ExpectedResponse: gin.H{
				"result": "Invio dei dati da parte del gateway interrotto correttamente",
			},
		}
		helper.ExecuteControllerTest(t, tc, "POST", "/gateway/:gateway_id/interrupt", controller.InterruptGateway)
	})

	t.Run("Resume 200 OK", func(t *testing.T) {
		expected := gateway.ResumeGatewayCommand{Requester: requester, GatewayId: gatewayID}
		steps := []helper.MockUseCaseSetupFunc[mocks.MockResumeGatewayUseCase]{
			func(mockUC *mocks.MockResumeGatewayUseCase) *gomock.Call {
				return mockUC.EXPECT().ResumeGateway(gomock.Eq(expected)).Return(gateway.Gateway{}, nil).Times(1)
			},
		}
		mockUseCase := helper.SetupMockUseCase(mocks.NewMockResumeGatewayUseCase, steps, t)
		controller := gateway.NewGatewayController(nil, nil, nil, nil, nil, nil, nil, nil, mockUseCase, nil, nil, nil, nil)

		tc := helper.GenericControllerTestCase[any, mocks.MockResumeGatewayUseCase]{
			Method:         "POST",
			Url:            "/gateway/" + gatewayID.String() + "/resume",
			InputDto:       struct{}{},
			Requester:      requester,
			ExpectedStatus: http.StatusOK,
			ExpectedResponse: gin.H{
				"result": "Invio dei dati da parte del gateway ripreso correttamente",
			},
		}
		helper.ExecuteControllerTest(t, tc, "POST", "/gateway/:gateway_id/resume", controller.ResumeGateway)
	})

	t.Run("Reset 200 OK", func(t *testing.T) {
		expected := gateway.ResetGatewayCommand{Requester: requester, GatewayId: gatewayID}
		steps := []helper.MockUseCaseSetupFunc[mocks.MockResetGatewayUseCase]{
			func(mockUC *mocks.MockResetGatewayUseCase) *gomock.Call {
				return mockUC.EXPECT().ResetGateway(gomock.Eq(expected)).Return(gateway.Gateway{}, nil).Times(1)
			},
		}
		mockUseCase := helper.SetupMockUseCase(mocks.NewMockResetGatewayUseCase, steps, t)
		controller := gateway.NewGatewayController(nil, nil, nil, nil, nil, nil, nil, nil, nil, mockUseCase, nil, nil, nil)

		tc := helper.GenericControllerTestCase[any, mocks.MockResetGatewayUseCase]{
			Method:         "POST",
			Url:            "/gateway/" + gatewayID.String() + "/reset",
			InputDto:       struct{}{},
			Requester:      requester,
			ExpectedStatus: http.StatusOK,
			ExpectedResponse: gin.H{
				"result": "Reset del gateway eseguito correttamente",
			},
		}
		helper.ExecuteControllerTest(t, tc, "POST", "/gateway/:gateway_id/reset", controller.ResetGateway)
	})

	t.Run("Reboot 200 OK", func(t *testing.T) {
		expected := gateway.RebootGatewayCommand{Requester: requester, GatewayId: gatewayID}
		steps := []helper.MockUseCaseSetupFunc[mocks.MockRebootGatewayUseCase]{
			func(mockUC *mocks.MockRebootGatewayUseCase) *gomock.Call {
				return mockUC.EXPECT().RebootGateway(gomock.Eq(expected)).Return(gateway.Gateway{}, nil).Times(1)
			},
		}
		mockUseCase := helper.SetupMockUseCase(mocks.NewMockRebootGatewayUseCase, steps, t)
		controller := gateway.NewGatewayController(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, mockUseCase, nil, nil)

		tc := helper.GenericControllerTestCase[any, mocks.MockRebootGatewayUseCase]{
			Method:         "POST",
			Url:            "/gateway/" + gatewayID.String() + "/reboot",
			InputDto:       struct{}{},
			Requester:      requester,
			ExpectedStatus: http.StatusOK,
			ExpectedResponse: gin.H{
				"result": "Reboot del gateway eseguito correttamente",
			},
		}
		helper.ExecuteControllerTest(t, tc, "POST", "/gateway/:gateway_id/reboot", controller.RebootGateway)
	})

	t.Run("Reboot: errore use case non trovato viene restituito 404", func(t *testing.T) {
		expected := gateway.RebootGatewayCommand{Requester: requester, GatewayId: gatewayID}
		steps := []helper.MockUseCaseSetupFunc[mocks.MockRebootGatewayUseCase]{
			func(mockUC *mocks.MockRebootGatewayUseCase) *gomock.Call {
				return mockUC.EXPECT().RebootGateway(gomock.Eq(expected)).Return(gateway.Gateway{}, gateway.ErrGatewayNotFound).Times(1)
			},
		}
		mockUseCase := helper.SetupMockUseCase(mocks.NewMockRebootGatewayUseCase, steps, t)
		controller := gateway.NewGatewayController(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, mockUseCase, nil, nil)

		tc := helper.GenericControllerTestCase[any, mocks.MockRebootGatewayUseCase]{
			Method:         "POST",
			Url:            "/gateway/" + gatewayID.String() + "/reboot",
			InputDto:       struct{}{},
			Requester:      requester,
			ExpectedStatus: http.StatusNotFound,
			ExpectedResponse: gin.H{
				"error": gateway.ErrGatewayNotFound.Error(),
			},
		}
		helper.ExecuteControllerTest(t, tc, "POST", "/gateway/:gateway_id/reboot", controller.RebootGateway)
	})

	t.Run("Interrupt 500 Server Error", func(t *testing.T) {
		expected := gateway.InterruptGatewayCommand{Requester: requester, GatewayId: gatewayID}
		errMock := errors.New("interrupt failed")
		steps := []helper.MockUseCaseSetupFunc[mocks.MockInterruptGatewayUseCase]{
			func(mockUC *mocks.MockInterruptGatewayUseCase) *gomock.Call {
				return mockUC.EXPECT().InterruptGateway(gomock.Eq(expected)).Return(gateway.Gateway{}, errMock).Times(1)
			},
		}
		mockUseCase := helper.SetupMockUseCase(mocks.NewMockInterruptGatewayUseCase, steps, t)
		controller := gateway.NewGatewayController(nil, nil, nil, nil, nil, nil, nil, mockUseCase, nil, nil, nil, nil, nil)

		tc := helper.GenericControllerTestCase[any, mocks.MockInterruptGatewayUseCase]{
			Method:         "POST",
			Url:            "/gateway/" + gatewayID.String() + "/interrupt",
			InputDto:       struct{}{},
			Requester:      requester,
			ExpectedStatus: http.StatusInternalServerError,
			ExpectedResponse: gin.H{
				"error": errMock.Error(),
			},
		}
		helper.ExecuteControllerTest(t, tc, "POST", "/gateway/:gateway_id/interrupt", controller.InterruptGateway)
	})
}
