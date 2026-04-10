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

func TestGatewayController_CreateGateway(t *testing.T) {
	requester := identity.Requester{
		RequesterUserId: 1,
		RequesterRole:   identity.ROLE_SUPER_ADMIN,
	}

	input := map[string]any{
		"name":     "Gateway A",
		"interval": 3000,
	}

	expectedCommand := gateway.CreateGatewayCommand{
		Requester: requester,
		Name:      "Gateway A",
		Interval:  3 * time.Second,
	}

	createdGateway := gateway.Gateway{
		Id:            uuid.New(),
		Name:          "Gateway A",
		Status:        gateway.GATEWAY_STATUS_DECOMMISSIONED,
		IntervalLimit: 3 * time.Second,
	}

	useCaseOk := func(mockUC *mocks.MockCreateGatewayUseCase) *gomock.Call {
		return mockUC.EXPECT().CreateGateway(gomock.Eq(expectedCommand)).Return(createdGateway, nil).Times(1)
	}
	useCaseNeverCalled := func(mockUC *mocks.MockCreateGatewayUseCase) *gomock.Call {
		return mockUC.EXPECT().CreateGateway(gomock.Any()).Times(0)
	}
	errMock := errors.New("unexpected create error")
	useCaseUnexpectedErr := func(mockUC *mocks.MockCreateGatewayUseCase) *gomock.Call {
		return mockUC.EXPECT().CreateGateway(gomock.Eq(expectedCommand)).Return(gateway.Gateway{}, errMock).Times(1)
	}

	cases := []helper.GenericControllerTestCase[any, mocks.MockCreateGatewayUseCase]{
		{
			Name:      "200 OK",
			Method:    "POST",
			Url:       "/gateway",
			InputDto:  input,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockCreateGatewayUseCase]{
				useCaseOk,
			},
			ExpectedStatus: http.StatusOK,
			ExpectedResponse: gin.H{
				"gateway_id":        createdGateway.Id.String(),
				"name":              createdGateway.Name,
				"tenant_id":         "",
				"status":            string(createdGateway.Status),
				"interval":          createdGateway.IntervalLimit.Milliseconds(),
				"public_identifier": nil,
			},
		},
		{
			Name:      "400 Bad Request: body non valido",
			Method:    "POST",
			Url:       "/gateway",
			InputDto:  map[string]any{"name": "Gateway A", "interval": 0},
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockCreateGatewayUseCase]{
				useCaseNeverCalled,
			},
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedResponse: helper.HasError{},
		},
		{
			Name:         "401 Unauthorized: identity mancante",
			Method:       "POST",
			Url:          "/gateway",
			InputDto:     input,
			Requester:    requester,
			OmitIdentity: true,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockCreateGatewayUseCase]{
				useCaseNeverCalled,
			},
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedResponse: gin.H{
				"error": transportHttp.ErrMissingIdentity.Error(),
			},
		},
		{
			Name:      "500 Server Error",
			Method:    "POST",
			Url:       "/gateway",
			InputDto:  input,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockCreateGatewayUseCase]{
				useCaseUnexpectedErr,
			},
			ExpectedStatus: http.StatusInternalServerError,
			ExpectedResponse: gin.H{
				"error": errMock.Error(),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			mockUseCase := helper.SetupMockUseCase(mocks.NewMockCreateGatewayUseCase, tc.SetupSteps, t)

			controller := gateway.NewGatewayController(
				nil,
				mockUseCase,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
			)

			helper.ExecuteControllerTest(t, tc, "POST", "/gateway", controller.CreateGateway)
		})
	}
}

func TestGatewayController_DeleteGateway(t *testing.T) {
	gatewayID := uuid.New()
	requester := identity.Requester{
		RequesterUserId: 1,
		RequesterRole:   identity.ROLE_SUPER_ADMIN,
	}

	expectedCommand := gateway.DeleteGatewayCommand{
		Requester: requester,
		GatewayId: gatewayID,
	}

	deletedGateway := gateway.Gateway{
		Id:            gatewayID,
		Name:          "Gateway A",
		Status:        gateway.GATEWAY_STATUS_DECOMMISSIONED,
		IntervalLimit: 3 * time.Second,
	}

	useCaseOk := func(mockUC *mocks.MockDeleteGatewayUseCase) *gomock.Call {
		return mockUC.EXPECT().DeleteGateway(gomock.Eq(expectedCommand)).Return(deletedGateway, nil).Times(1)
	}
	useCaseNeverCalled := func(mockUC *mocks.MockDeleteGatewayUseCase) *gomock.Call {
		return mockUC.EXPECT().DeleteGateway(gomock.Any()).Times(0)
	}
	useCaseNotFound := func(mockUC *mocks.MockDeleteGatewayUseCase) *gomock.Call {
		return mockUC.EXPECT().DeleteGateway(gomock.Eq(expectedCommand)).Return(gateway.Gateway{}, gateway.ErrGatewayNotFound).Times(1)
	}
	errMock := errors.New("delete failed")
	useCaseUnexpectedErr := func(mockUC *mocks.MockDeleteGatewayUseCase) *gomock.Call {
		return mockUC.EXPECT().DeleteGateway(gomock.Eq(expectedCommand)).Return(gateway.Gateway{}, errMock).Times(1)
	}

	validURL := "/gateway/" + gatewayID.String()
	invalidURL := "/gateway/not-a-uuid"

	cases := []helper.GenericControllerTestCase[any, mocks.MockDeleteGatewayUseCase]{
		{
			Name:      "200 OK",
			Method:    "DELETE",
			Url:       validURL,
			InputDto:  struct{}{},
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockDeleteGatewayUseCase]{
				useCaseOk,
			},
			ExpectedStatus: http.StatusOK,
			ExpectedResponse: gin.H{
				"gateway_id":        deletedGateway.Id.String(),
				"name":              deletedGateway.Name,
				"tenant_id":         "",
				"status":            string(deletedGateway.Status),
				"interval":          deletedGateway.IntervalLimit.Milliseconds(),
				"public_identifier": nil,
			},
		},
		{
			Name:      "400 Bad Request: gatewayId non valido",
			Method:    "DELETE",
			Url:       invalidURL,
			InputDto:  struct{}{},
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockDeleteGatewayUseCase]{
				useCaseNeverCalled,
			},
			ExpectedStatus: http.StatusBadRequest,
			ExpectedResponse: gin.H{
				"error": gateway.ErrInvalidGatewayID.Error(),
			},
		},
		{
			Name:         "401 Unauthorized: identity mancante",
			Method:       "DELETE",
			Url:          validURL,
			InputDto:     struct{}{},
			Requester:    requester,
			OmitIdentity: true,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockDeleteGatewayUseCase]{
				useCaseNeverCalled,
			},
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedResponse: gin.H{
				"error": transportHttp.ErrMissingIdentity.Error(),
			},
		},
		{
			Name:      "404 Not Found",
			Method:    "DELETE",
			Url:       validURL,
			InputDto:  struct{}{},
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockDeleteGatewayUseCase]{
				useCaseNotFound,
			},
			ExpectedStatus: http.StatusNotFound,
			ExpectedResponse: gin.H{
				"error": gateway.ErrGatewayNotFound.Error(),
			},
		},
		{
			Name:      "500 Server Error",
			Method:    "DELETE",
			Url:       validURL,
			InputDto:  struct{}{},
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockDeleteGatewayUseCase]{
				useCaseUnexpectedErr,
			},
			ExpectedStatus: http.StatusInternalServerError,
			ExpectedResponse: gin.H{
				"error": errMock.Error(),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			mockUseCase := helper.SetupMockUseCase(mocks.NewMockDeleteGatewayUseCase, tc.SetupSteps, t)

			controller := gateway.NewGatewayController(
				nil,
				nil,
				mockUseCase,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
			)

			helper.ExecuteControllerTest(t, tc, "DELETE", "/gateway/:gateway_id", controller.DeleteGateway)
		})
	}
}
