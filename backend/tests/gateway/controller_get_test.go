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

func mustUUIDParseError(t *testing.T, raw string) error {
	t.Helper()

	_, err := uuid.Parse(raw)
	if err == nil {
		t.Fatalf("expected invalid uuid error for %q", raw)
	}

	return err
}

func TestGatewayController_GetGateway(t *testing.T) {
	gatewayID := uuid.New()
	tenantID := uuid.New()
	requester := identity.Requester{RequesterUserId: 1, RequesterRole: identity.ROLE_SUPER_ADMIN}

	publicID := "pub-gw"
	resultGateway := gateway.Gateway{
		Id:               gatewayID,
		Name:             "Gateway A",
		TenantId:         &tenantID,
		Status:           gateway.GATEWAY_STATUS_ACTIVE,
		IntervalLimit:    5 * time.Second,
		PublicIdentifier: &publicID,
	}

	expectedCommand := gateway.GetGatewayByIdCommand{
		Requester: requester,
		GatewayId: gatewayID,
	}

	useCaseOk := func(mockUC *mocks.MockGetGatewayUseCase) *gomock.Call {
		return mockUC.EXPECT().GetGateway(gomock.Eq(expectedCommand)).Return(resultGateway, nil).Times(1)
	}
	useCaseNeverCalled := func(mockUC *mocks.MockGetGatewayUseCase) *gomock.Call {
		return mockUC.EXPECT().GetGateway(gomock.Any()).Times(0)
	}
	useCaseNotFound := func(mockUC *mocks.MockGetGatewayUseCase) *gomock.Call {
		return mockUC.EXPECT().GetGateway(gomock.Eq(expectedCommand)).Return(gateway.Gateway{}, gateway.ErrGatewayNotFound).Times(1)
	}
	useCaseUnauthorized := func(mockUC *mocks.MockGetGatewayUseCase) *gomock.Call {
		return mockUC.EXPECT().GetGateway(gomock.Eq(expectedCommand)).Return(gateway.Gateway{}, identity.ErrUnauthorizedAccess).Times(1)
	}
	errMock := errors.New("get gateway failed")
	useCaseUnexpectedErr := func(mockUC *mocks.MockGetGatewayUseCase) *gomock.Call {
		return mockUC.EXPECT().GetGateway(gomock.Eq(expectedCommand)).Return(gateway.Gateway{}, errMock).Times(1)
	}

	invalidGatewayIDErr := mustUUIDParseError(t, "not-a-uuid")

	cases := []helper.GenericControllerTestCase[any, mocks.MockGetGatewayUseCase]{
		{
			Name:      "200 OK",
			Method:    "GET",
			Url:       "/gateway/" + gatewayID.String(),
			InputDto:  struct{}{},
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetGatewayUseCase]{
				useCaseOk,
			},
			ExpectedStatus: http.StatusOK,
			ExpectedResponse: gin.H{
				"gateway_id":        gatewayID.String(),
				"name":              resultGateway.Name,
				"tenant_id":         tenantID.String(),
				"status":            string(resultGateway.Status),
				"interval":          resultGateway.IntervalLimit.Milliseconds(),
				"public_identifier": publicID,
			},
		},
		{
			Name:      "400 Bad Request: gatewayId non valido",
			Method:    "GET",
			Url:       "/gateway/not-a-uuid",
			InputDto:  struct{}{},
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetGatewayUseCase]{
				useCaseNeverCalled,
			},
			ExpectedStatus: http.StatusBadRequest,
			ExpectedResponse: gin.H{
				"error": invalidGatewayIDErr.Error(),
			},
		},
		{
			Name:         "401 Unauthorized: identity mancante",
			Url:          "/gateway/" + gatewayID.String(),
			InputDto:     struct{}{},
			Requester:    requester,
			OmitIdentity: true,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetGatewayUseCase]{
				useCaseNeverCalled,
			},
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedResponse: gin.H{
				"error": transportHttp.ErrMissingIdentity.Error(),
			},
		},
		{
			Name:      "400 Bad Request: use case returns not found",
			Method:    "GET",
			Url:       "/gateway/" + gatewayID.String(),
			InputDto:  struct{}{},
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetGatewayUseCase]{
				useCaseNotFound,
			},
			ExpectedStatus: http.StatusBadRequest,
			ExpectedResponse: gin.H{
				"error": gateway.ErrGatewayNotFound.Error(),
			},
		},
		{
			Name:      "500 Server Error",
			Method:    "GET",
			Url:       "/gateway/" + gatewayID.String(),
			InputDto:  struct{}{},
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetGatewayUseCase]{
				useCaseUnexpectedErr,
			},
			ExpectedStatus: http.StatusInternalServerError,
			ExpectedResponse: gin.H{
				"error": errMock.Error(),
			},
		},
		{
			Name:      "401 Unauthorized: use case reports unauthorized access",
			Method:    "GET",
			Url:       "/gateway/" + gatewayID.String(),
			InputDto:  struct{}{},
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetGatewayUseCase]{
				useCaseUnauthorized,
			},
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedResponse: gin.H{
				"error": identity.ErrUnauthorizedAccess.Error(),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			mockUseCase := helper.SetupMockUseCase(mocks.NewMockGetGatewayUseCase, tc.SetupSteps, t)

			controller := gateway.NewGatewayController(
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
				nil,
				mockUseCase,
				nil,
			)

			helper.ExecuteControllerTest(t, tc, "GET", "/gateway/:gateway_id", controller.GetGateway)
		})
	}
}

func TestGatewayController_GetAllGateways(t *testing.T) {
	gatewayID1 := uuid.New()
	gatewayID2 := uuid.New()
	tenantID1 := uuid.New()
	tenantID2 := uuid.New()
	requester := identity.Requester{RequesterUserId: 1, RequesterRole: identity.ROLE_SUPER_ADMIN}

	publicID := "pub-list"
	gateways := []gateway.Gateway{
		{
			Id:               gatewayID1,
			Name:             "Gateway A",
			TenantId:         &tenantID1,
			Status:           gateway.GATEWAY_STATUS_ACTIVE,
			IntervalLimit:    3 * time.Second,
			PublicIdentifier: &publicID,
		},
		{
			Id:            gatewayID2,
			Name:          "Gateway B",
			TenantId:      &tenantID2,
			Status:        gateway.GATEWAY_STATUS_DECOMMISSIONED,
			IntervalLimit: 6 * time.Second,
		},
	}

	expectedCommand := gateway.GetAllGatewaysCommand{
		Requester: requester,
		Page:      2,
		Limit:     10,
	}

	useCaseOk := func(mockUC *mocks.MockGetAllGatewaysUseCase) *gomock.Call {
		return mockUC.EXPECT().GetAllGateways(gomock.Eq(expectedCommand)).Return(gateways, uint(7), nil).Times(1)
	}
	useCaseNeverCalled := func(mockUC *mocks.MockGetAllGatewaysUseCase) *gomock.Call {
		return mockUC.EXPECT().GetAllGateways(gomock.Any()).Times(0)
	}
	errMock := errors.New("get all failed")
	useCaseUnexpectedErr := func(mockUC *mocks.MockGetAllGatewaysUseCase) *gomock.Call {
		return mockUC.EXPECT().GetAllGateways(gomock.Eq(expectedCommand)).Return(nil, uint(0), errMock).Times(1)
	}

	cases := []helper.GenericControllerTestCase[any, mocks.MockGetAllGatewaysUseCase]{
		{
			Name:      "200 OK",
			Method:    "GET",
			Url:       "/gateway?page=2&limit=10",
			InputDto:  struct{}{},
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetAllGatewaysUseCase]{
				useCaseOk,
			},
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: gateway.NewGatewayListResponseDTO(gateways, uint(7)),
		},
		{
			Name:         "401 Unauthorized: identity mancante",
			Method:       "GET",
			Url:          "/gateway?page=2&limit=10",
			InputDto:     struct{}{},
			Requester:    requester,
			OmitIdentity: true,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetAllGatewaysUseCase]{
				useCaseNeverCalled,
			},
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedResponse: gin.H{
				"error": transportHttp.ErrMissingIdentity.Error(),
			},
		},
		{
			Name:      "400 Bad Request: query validation error",
			Method:    "GET",
			Url:       "/gateway?page=0&limit=10",
			InputDto:  struct{}{},
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetAllGatewaysUseCase]{
				useCaseNeverCalled,
			},
			ExpectedStatus: http.StatusBadRequest,
			ExpectedResponse: gin.H{
				"error": "invalid format",
				"fields": gin.H{
					"page": "min",
				},
			},
		},
		{
			Name:      "400 Bad Request: query parse error",
			Method:    "GET",
			Url:       "/gateway?page=abc&limit=10",
			InputDto:  struct{}{},
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetAllGatewaysUseCase]{
				useCaseNeverCalled,
			},
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedResponse: helper.HasError{},
		},
		{
			Name:      "500 Server Error",
			Method:    "GET",
			Url:       "/gateway?page=2&limit=10",
			InputDto:  struct{}{},
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetAllGatewaysUseCase]{
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
			mockUseCase := helper.SetupMockUseCase(mocks.NewMockGetAllGatewaysUseCase, tc.SetupSteps, t)

			controller := gateway.NewGatewayController(
				nil,
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

			helper.ExecuteControllerTest(t, tc, "GET", "/gateway", controller.GetAllGateways)
		})
	}
}

func TestGatewayController_GetGatewaysByTenant(t *testing.T) {
	gatewayID := uuid.New()
	tenantID := uuid.New()
	requester := identity.Requester{RequesterUserId: 1, RequesterRole: identity.ROLE_SUPER_ADMIN}

	publicID := "pub-tenant"
	gateways := []gateway.Gateway{
		{
			Id:               gatewayID,
			Name:             "Gateway Tenant",
			TenantId:         &tenantID,
			Status:           gateway.GATEWAY_STATUS_ACTIVE,
			IntervalLimit:    4 * time.Second,
			PublicIdentifier: &publicID,
		},
	}

	expectedCommand := gateway.GetGatewaysByTenantCommand{
		Requester: requester,
		TenantId:  tenantID,
		Page:      3,
		Limit:     8,
	}

	useCaseOk := func(mockUC *mocks.MockGetGatewaysByTenantUseCase) *gomock.Call {
		return mockUC.EXPECT().GetGatewaysByTenant(gomock.Eq(expectedCommand)).Return(gateways, uint(1), nil).Times(1)
	}
	useCaseNeverCalled := func(mockUC *mocks.MockGetGatewaysByTenantUseCase) *gomock.Call {
		return mockUC.EXPECT().GetGatewaysByTenant(gomock.Any()).Times(0)
	}
	useCaseNotFound := func(mockUC *mocks.MockGetGatewaysByTenantUseCase) *gomock.Call {
		return mockUC.EXPECT().GetGatewaysByTenant(gomock.Eq(expectedCommand)).Return(nil, uint(0), gateway.ErrGatewayNotFound).Times(1)
	}
	useCaseUnauthorized := func(mockUC *mocks.MockGetGatewaysByTenantUseCase) *gomock.Call {
		return mockUC.EXPECT().GetGatewaysByTenant(gomock.Eq(expectedCommand)).Return(nil, uint(0), identity.ErrUnauthorizedAccess).Times(1)
	}
	errMock := errors.New("get gateways by tenant failed")
	useCaseUnexpectedErr := func(mockUC *mocks.MockGetGatewaysByTenantUseCase) *gomock.Call {
		return mockUC.EXPECT().GetGatewaysByTenant(gomock.Eq(expectedCommand)).Return(nil, uint(0), errMock).Times(1)
	}

	invalidTenantIDErr := mustUUIDParseError(t, "not-a-uuid")

	cases := []helper.GenericControllerTestCase[any, mocks.MockGetGatewaysByTenantUseCase]{
		{
			Name:      "200 OK",
			Method:    "GET",
			Url:       "/tenant/" + tenantID.String() + "/gateway?page=3&limit=8",
			InputDto:  struct{}{},
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetGatewaysByTenantUseCase]{
				useCaseOk,
			},
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: gateway.NewGatewayListResponseDTO(gateways, uint(1)),
		},
		{
			Name:      "400 Bad Request: tenantId non valido",
			Method:    "GET",
			Url:       "/tenant/not-a-uuid/gateway?page=3&limit=8",
			InputDto:  struct{}{},
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetGatewaysByTenantUseCase]{
				useCaseNeverCalled,
			},
			ExpectedStatus: http.StatusBadRequest,
			ExpectedResponse: gin.H{
				"error": invalidTenantIDErr.Error(),
			},
		},
		{
			Name:         "401 Unauthorized: identity mancante",
			Method:       "GET",
			Url:          "/tenant/" + tenantID.String() + "/gateway?page=3&limit=8",
			InputDto:     struct{}{},
			Requester:    requester,
			OmitIdentity: true,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetGatewaysByTenantUseCase]{
				useCaseNeverCalled,
			},
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedResponse: gin.H{
				"error": transportHttp.ErrMissingIdentity.Error(),
			},
		},
		{
			Name:      "400 Bad Request: query validation error",
			Method:    "GET",
			Url:       "/tenant/" + tenantID.String() + "/gateway?page=0&limit=8",
			InputDto:  struct{}{},
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetGatewaysByTenantUseCase]{
				useCaseNeverCalled,
			},
			ExpectedStatus: http.StatusBadRequest,
			ExpectedResponse: gin.H{
				"error": "invalid format",
				"fields": gin.H{
					"page": "min",
				},
			},
		},
		{
			Name:      "400 Bad Request: query parse error",
			Method:    "GET",
			Url:       "/tenant/" + tenantID.String() + "/gateway?page=abc&limit=8",
			InputDto:  struct{}{},
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetGatewaysByTenantUseCase]{
				useCaseNeverCalled,
			},
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedResponse: helper.HasError{},
		},
		{
			Name:      "400 Bad Request: use case returns not found",
			Method:    "GET",
			Url:       "/tenant/" + tenantID.String() + "/gateway?page=3&limit=8",
			InputDto:  struct{}{},
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetGatewaysByTenantUseCase]{
				useCaseNotFound,
			},
			ExpectedStatus: http.StatusBadRequest,
			ExpectedResponse: gin.H{
				"error": gateway.ErrGatewayNotFound.Error(),
			},
		},
		{
			Name:      "401 Unauthorized: use case reports unauthorized access",
			Method:    "GET",
			Url:       "/tenant/" + tenantID.String() + "/gateway?page=3&limit=8",
			InputDto:  struct{}{},
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetGatewaysByTenantUseCase]{
				useCaseUnauthorized,
			},
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedResponse: gin.H{
				"error": identity.ErrUnauthorizedAccess.Error(),
			},
		},
		{
			Name:      "500 Server Error",
			Method:    "GET",
			Url:       "/tenant/" + tenantID.String() + "/gateway?page=3&limit=8",
			InputDto:  struct{}{},
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetGatewaysByTenantUseCase]{
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
			mockUseCase := helper.SetupMockUseCase(mocks.NewMockGetGatewaysByTenantUseCase, tc.SetupSteps, t)

			controller := gateway.NewGatewayController(
				nil,
				nil,
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
			)

			helper.ExecuteControllerTest(t, tc, "GET", "/tenant/:tenant_id/gateway", controller.GetGatewaysByTenant)
		})
	}
}

func TestGatewayController_GetGatewayByTenantID(t *testing.T) {
	gatewayID := uuid.New()
	tenantID := uuid.New()
	requester := identity.Requester{RequesterUserId: 1, RequesterRole: identity.ROLE_SUPER_ADMIN}

	publicID := "pub-tenant-gateway"
	resultGateway := gateway.Gateway{
		Id:               gatewayID,
		Name:             "Gateway Tenant Detail",
		TenantId:         &tenantID,
		Status:           gateway.GATEWAY_STATUS_ACTIVE,
		IntervalLimit:    7 * time.Second,
		PublicIdentifier: &publicID,
	}

	expectedCommand := gateway.GetGatewayByTenantIDCommand{
		Requester: requester,
		TenantId:  tenantID,
		GatewayId: gatewayID,
	}

	useCaseOk := func(mockUC *mocks.MockGetGatewayByTenantIDUseCase) *gomock.Call {
		return mockUC.EXPECT().GetGatewayByTenantID(gomock.Eq(expectedCommand)).Return(resultGateway, nil).Times(1)
	}
	useCaseNeverCalled := func(mockUC *mocks.MockGetGatewayByTenantIDUseCase) *gomock.Call {
		return mockUC.EXPECT().GetGatewayByTenantID(gomock.Any()).Times(0)
	}
	useCaseNotFound := func(mockUC *mocks.MockGetGatewayByTenantIDUseCase) *gomock.Call {
		return mockUC.EXPECT().GetGatewayByTenantID(gomock.Eq(expectedCommand)).Return(gateway.Gateway{}, gateway.ErrGatewayNotFound).Times(1)
	}
	useCaseUnauthorized := func(mockUC *mocks.MockGetGatewayByTenantIDUseCase) *gomock.Call {
		return mockUC.EXPECT().GetGatewayByTenantID(gomock.Eq(expectedCommand)).Return(gateway.Gateway{}, identity.ErrUnauthorizedAccess).Times(1)
	}
	errMock := errors.New("get gateway by tenant failed")
	useCaseUnexpectedErr := func(mockUC *mocks.MockGetGatewayByTenantIDUseCase) *gomock.Call {
		return mockUC.EXPECT().GetGatewayByTenantID(gomock.Eq(expectedCommand)).Return(gateway.Gateway{}, errMock).Times(1)
	}

	invalidTenantIDErr := mustUUIDParseError(t, "not-a-uuid")
	invalidGatewayIDErr := mustUUIDParseError(t, "also-not-a-uuid")

	cases := []helper.GenericControllerTestCase[any, mocks.MockGetGatewayByTenantIDUseCase]{
		{
			Name:      "200 OK",
			Method:    "GET",
			Url:       "/tenant/" + tenantID.String() + "/gateway/" + gatewayID.String(),
			InputDto:  struct{}{},
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetGatewayByTenantIDUseCase]{
				useCaseOk,
			},
			ExpectedStatus: http.StatusOK,
			ExpectedResponse: gin.H{
				"gateway_id":        gatewayID.String(),
				"name":              resultGateway.Name,
				"tenant_id":         tenantID.String(),
				"status":            string(resultGateway.Status),
				"interval":          resultGateway.IntervalLimit.Milliseconds(),
				"public_identifier": publicID,
			},
		},
		{
			Name:      "400 Bad Request: tenantId non valido",
			Method:    "GET",
			Url:       "/tenant/not-a-uuid/gateway/" + gatewayID.String(),
			InputDto:  struct{}{},
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetGatewayByTenantIDUseCase]{
				useCaseNeverCalled,
			},
			ExpectedStatus: http.StatusBadRequest,
			ExpectedResponse: gin.H{
				"error": invalidTenantIDErr.Error(),
			},
		},
		{
			Name:      "400 Bad Request: gatewayId non valido",
			Method:    "GET",
			Url:       "/tenant/" + tenantID.String() + "/gateway/also-not-a-uuid",
			InputDto:  struct{}{},
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetGatewayByTenantIDUseCase]{
				useCaseNeverCalled,
			},
			ExpectedStatus: http.StatusBadRequest,
			ExpectedResponse: gin.H{
				"error": invalidGatewayIDErr.Error(),
			},
		},
		{
			Name:         "401 Unauthorized: identity mancante",
			Method:       "GET",
			Url:          "/tenant/" + tenantID.String() + "/gateway/" + gatewayID.String(),
			InputDto:     struct{}{},
			Requester:    requester,
			OmitIdentity: true,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetGatewayByTenantIDUseCase]{
				useCaseNeverCalled,
			},
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedResponse: gin.H{
				"error": transportHttp.ErrMissingIdentity.Error(),
			},
		},
		{
			Name:      "400 Bad Request: use case returns not found",
			Method:    "GET",
			Url:       "/tenant/" + tenantID.String() + "/gateway/" + gatewayID.String(),
			InputDto:  struct{}{},
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetGatewayByTenantIDUseCase]{
				useCaseNotFound,
			},
			ExpectedStatus: http.StatusBadRequest,
			ExpectedResponse: gin.H{
				"error": gateway.ErrGatewayNotFound.Error(),
			},
		},
		{
			Name:      "401 Unauthorized: use case reports unauthorized access",
			Method:    "GET",
			Url:       "/tenant/" + tenantID.String() + "/gateway/" + gatewayID.String(),
			InputDto:  struct{}{},
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetGatewayByTenantIDUseCase]{
				useCaseUnauthorized,
			},
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedResponse: gin.H{
				"error": identity.ErrUnauthorizedAccess.Error(),
			},
		},
		{
			Name:      "500 Server Error",
			Method:    "GET",
			Url:       "/tenant/" + tenantID.String() + "/gateway/" + gatewayID.String(),
			InputDto:  struct{}{},
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockGetGatewayByTenantIDUseCase]{
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
			mockUseCase := helper.SetupMockUseCase(mocks.NewMockGetGatewayByTenantIDUseCase, tc.SetupSteps, t)

			controller := gateway.NewGatewayController(
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
				nil,
				nil,
				mockUseCase,
			)

			helper.ExecuteControllerTest(t, tc, "GET", "/tenant/:tenant_id/gateway/:gateway_id", controller.GetGatewayByTenantID)
		})
	}
}
