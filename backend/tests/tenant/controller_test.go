package tenant_test

import (
	"errors"
	"net/http"
	"testing"

	transportHttp "backend/internal/infra/transport/http"
	httpdto "backend/internal/infra/transport/http/dto"
	"backend/internal/shared/identity"
	"backend/internal/tenant"
	"backend/tests/helper"
	tenantMocks "backend/tests/tenant/mocks"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestController_CreateTenant(t *testing.T) {
	tenantName := "Tenant A"
	requester := identity.Requester{RequesterUserId: 1, RequesterRole: identity.ROLE_SUPER_ADMIN}
	inputDTO := tenant.CreateTenantDTO{
		TenantNameField: httpdto.TenantNameField{TenantName: tenantName},
		CanImpersonate:  true,
	}

	tenantID := uuid.New()
	expectedCommand := tenant.CreateTenantCommand{
		Requester:      requester,
		Name:           tenantName,
		CanImpersonate: true,
	}
	expectedCreated := tenant.Tenant{Id: tenantID, Name: tenantName, CanImpersonate: true}

	useCaseOK := func(mockUC *tenantMocks.MockCreateTenantUseCase) *gomock.Call {
		return mockUC.EXPECT().
			CreateTenant(gomock.Eq(expectedCommand)).
			Return(expectedCreated, nil).
			Times(1)
	}
	useCaseNeverCalled := func(mockUC *tenantMocks.MockCreateTenantUseCase) *gomock.Call {
		return mockUC.EXPECT().
			CreateTenant(gomock.Any()).
			Times(0)
	}
	useCaseUnauthorized := func(mockUC *tenantMocks.MockCreateTenantUseCase) *gomock.Call {
		return mockUC.EXPECT().
			CreateTenant(gomock.Eq(expectedCommand)).
			Return(tenant.Tenant{}, identity.ErrUnauthorizedAccess).
			Times(1)
	}
	useCaseAlreadyExists := func(mockUC *tenantMocks.MockCreateTenantUseCase) *gomock.Call {
		return mockUC.EXPECT().
			CreateTenant(gomock.Eq(expectedCommand)).
			Return(tenant.Tenant{}, tenant.ErrTenantAlreadyExists).
			Times(1)
	}
	errMock := errors.New("unexpected error")
	useCaseUnexpectedErr := func(mockUC *tenantMocks.MockCreateTenantUseCase) *gomock.Call {
		return mockUC.EXPECT().
			CreateTenant(gomock.Eq(expectedCommand)).
			Return(tenant.Tenant{}, errMock).
			Times(1)
	}

	baseURL := "/tenant"
	cases := []helper.GenericControllerTestCase[any, tenantMocks.MockCreateTenantUseCase]{
		{
			Name:             "200 OK",
			Method:           http.MethodPost,
			Url:              baseURL,
			InputDto:         inputDTO,
			Requester:        requester,
			SetupSteps:       []helper.MockUseCaseSetupFunc[tenantMocks.MockCreateTenantUseCase]{useCaseOK},
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: tenant.NewTenantResponseDTO(expectedCreated),
		},
		{
			Name:             "401 Unauthorized: missing identity",
			Method:           http.MethodPost,
			Url:              baseURL,
			InputDto:         inputDTO,
			OmitIdentity:     true,
			SetupSteps:       []helper.MockUseCaseSetupFunc[tenantMocks.MockCreateTenantUseCase]{useCaseNeverCalled},
			ExpectedStatus:   http.StatusUnauthorized,
			ExpectedResponse: gin.H{"error": transportHttp.ErrMissingIdentity.Error()},
		},
		{
			Name:             "400 Bad Request: invalid body",
			Method:           http.MethodPost,
			Url:              baseURL,
			InputDto:         tenant.CreateTenantDTO{},
			Requester:        requester,
			SetupSteps:       []helper.MockUseCaseSetupFunc[tenantMocks.MockCreateTenantUseCase]{useCaseNeverCalled},
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedResponse: helper.HasError{},
		},
		{
			Name:             "401 Unauthorized: use case unauthorized",
			Method:           http.MethodPost,
			Url:              baseURL,
			InputDto:         inputDTO,
			Requester:        requester,
			SetupSteps:       []helper.MockUseCaseSetupFunc[tenantMocks.MockCreateTenantUseCase]{useCaseUnauthorized},
			ExpectedStatus:   http.StatusUnauthorized,
			ExpectedResponse: gin.H{"error": identity.ErrUnauthorizedAccess.Error()},
		},
		{
			Name:             "400 Bad Request: tenant already exists",
			Method:           http.MethodPost,
			Url:              baseURL,
			InputDto:         inputDTO,
			Requester:        requester,
			SetupSteps:       []helper.MockUseCaseSetupFunc[tenantMocks.MockCreateTenantUseCase]{useCaseAlreadyExists},
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedResponse: gin.H{"error": tenant.ErrTenantAlreadyExists.Error()},
		},
		{
			Name:             "500 Internal Server Error",
			Method:           http.MethodPost,
			Url:              baseURL,
			InputDto:         inputDTO,
			Requester:        requester,
			SetupSteps:       []helper.MockUseCaseSetupFunc[tenantMocks.MockCreateTenantUseCase]{useCaseUnexpectedErr},
			ExpectedStatus:   http.StatusInternalServerError,
			ExpectedResponse: gin.H{"error": errMock.Error()},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			mockUseCase := helper.SetupMockUseCase(tenantMocks.NewMockCreateTenantUseCase, tc.SetupSteps, t)
			controller := tenant.NewTenantController(zap.NewNop(), mockUseCase, nil, nil, nil, nil)
			helper.ExecuteControllerTest(t, tc, http.MethodPost, "/tenant", controller.CreateTenant)
		})
	}
}

func TestController_DeleteTenant(t *testing.T) {
	tenantID := uuid.New()
	requester := identity.Requester{RequesterUserId: 1, RequesterRole: identity.ROLE_SUPER_ADMIN}

	expectedCmd := tenant.DeleteTenantCommand{Requester: requester, TenantId: tenantID}
	expectedDeleted := tenant.Tenant{Id: tenantID, Name: "Tenant A", CanImpersonate: true}

	useCaseOK := func(mockUC *tenantMocks.MockDeleteTenantUseCase) *gomock.Call {
		return mockUC.EXPECT().
			DeleteTenant(gomock.Eq(expectedCmd)).
			Return(expectedDeleted, nil).
			Times(1)
	}
	useCaseNeverCalled := func(mockUC *tenantMocks.MockDeleteTenantUseCase) *gomock.Call {
		return mockUC.EXPECT().
			DeleteTenant(gomock.Any()).
			Times(0)
	}
	useCaseUnauthorized := func(mockUC *tenantMocks.MockDeleteTenantUseCase) *gomock.Call {
		return mockUC.EXPECT().
			DeleteTenant(gomock.Eq(expectedCmd)).
			Return(tenant.Tenant{}, identity.ErrUnauthorizedAccess).
			Times(1)
	}
	useCaseNotFound := func(mockUC *tenantMocks.MockDeleteTenantUseCase) *gomock.Call {
		return mockUC.EXPECT().
			DeleteTenant(gomock.Eq(expectedCmd)).
			Return(tenant.Tenant{}, tenant.ErrTenantNotFound).
			Times(1)
	}
	errMock := errors.New("unexpected delete error")
	useCaseUnexpectedErr := func(mockUC *tenantMocks.MockDeleteTenantUseCase) *gomock.Call {
		return mockUC.EXPECT().
			DeleteTenant(gomock.Eq(expectedCmd)).
			Return(tenant.Tenant{}, errMock).
			Times(1)
	}

	mountPath := "/tenant/:tenant_id"
	baseURL := "/tenant/" + tenantID.String()

	cases := []helper.GenericControllerTestCase[any, tenantMocks.MockDeleteTenantUseCase]{
		{
			Name:             "200 OK",
			Method:           http.MethodDelete,
			Url:              baseURL,
			Requester:        requester,
			SetupSteps:       []helper.MockUseCaseSetupFunc[tenantMocks.MockDeleteTenantUseCase]{useCaseOK},
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: tenant.NewTenantResponseDTO(expectedDeleted),
		},
		{
			Name:             "401 Unauthorized: missing identity",
			Method:           http.MethodDelete,
			Url:              baseURL,
			OmitIdentity:     true,
			SetupSteps:       []helper.MockUseCaseSetupFunc[tenantMocks.MockDeleteTenantUseCase]{useCaseNeverCalled},
			ExpectedStatus:   http.StatusUnauthorized,
			ExpectedResponse: gin.H{"error": transportHttp.ErrMissingIdentity.Error()},
		},
		{
			Name:             "400 Bad Request: invalid tenant id in URI",
			Method:           http.MethodDelete,
			Url:              "/tenant/not-a-uuid",
			Requester:        requester,
			SetupSteps:       []helper.MockUseCaseSetupFunc[tenantMocks.MockDeleteTenantUseCase]{useCaseNeverCalled},
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedResponse: helper.HasError{},
		},
		{
			Name:             "401 Unauthorized: use case unauthorized",
			Method:           http.MethodDelete,
			Url:              baseURL,
			Requester:        requester,
			SetupSteps:       []helper.MockUseCaseSetupFunc[tenantMocks.MockDeleteTenantUseCase]{useCaseUnauthorized},
			ExpectedStatus:   http.StatusUnauthorized,
			ExpectedResponse: gin.H{"error": identity.ErrUnauthorizedAccess.Error()},
		},
		{
			Name:             "404 Not Found",
			Method:           http.MethodDelete,
			Url:              baseURL,
			Requester:        requester,
			SetupSteps:       []helper.MockUseCaseSetupFunc[tenantMocks.MockDeleteTenantUseCase]{useCaseNotFound},
			ExpectedStatus:   http.StatusNotFound,
			ExpectedResponse: gin.H{"error": tenant.ErrTenantNotFound.Error()},
		},
		{
			Name:             "500 Internal Server Error",
			Method:           http.MethodDelete,
			Url:              baseURL,
			Requester:        requester,
			SetupSteps:       []helper.MockUseCaseSetupFunc[tenantMocks.MockDeleteTenantUseCase]{useCaseUnexpectedErr},
			ExpectedStatus:   http.StatusInternalServerError,
			ExpectedResponse: gin.H{"error": errMock.Error()},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			mockUseCase := helper.SetupMockUseCase(tenantMocks.NewMockDeleteTenantUseCase, tc.SetupSteps, t)
			controller := tenant.NewTenantController(zap.NewNop(), nil, mockUseCase, nil, nil, nil)
			helper.ExecuteControllerTest(t, tc, http.MethodDelete, mountPath, controller.DeleteTenant)
		})
	}
}

func TestController_GetTenant(t *testing.T) {
	targetTenantID := uuid.New()
	requester := identity.Requester{RequesterUserId: 1, RequesterRole: identity.ROLE_SUPER_ADMIN}
	inputDTO := tenant.GetTenantDTO{TenantIdField: httpdto.TenantIdField{TenantId: targetTenantID.String()}}

	expectedCmd := tenant.GetTenantCommand{Requester: requester, TenantId: targetTenantID}
	expectedTenant := tenant.Tenant{Id: targetTenantID, Name: "Tenant A", CanImpersonate: true}

	useCaseOK := func(mockUC *tenantMocks.MockGetTenantUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetTenant(gomock.Eq(expectedCmd)).
			Return(expectedTenant, nil).
			Times(1)
	}
	useCaseNeverCalled := func(mockUC *tenantMocks.MockGetTenantUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetTenant(gomock.Any()).
			Times(0)
	}
	useCaseUnauthorized := func(mockUC *tenantMocks.MockGetTenantUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetTenant(gomock.Eq(expectedCmd)).
			Return(tenant.Tenant{}, identity.ErrUnauthorizedAccess).
			Times(1)
	}
	useCaseNotFound := func(mockUC *tenantMocks.MockGetTenantUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetTenant(gomock.Eq(expectedCmd)).
			Return(tenant.Tenant{}, tenant.ErrTenantNotFound).
			Times(1)
	}
	errMock := errors.New("unexpected get error")
	useCaseUnexpectedErr := func(mockUC *tenantMocks.MockGetTenantUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetTenant(gomock.Eq(expectedCmd)).
			Return(tenant.Tenant{}, errMock).
			Times(1)
	}

	baseURL := "/tenant/get"
	cases := []helper.GenericControllerTestCase[any, tenantMocks.MockGetTenantUseCase]{
		{
			Name:             "200 OK",
			Method:           http.MethodPost,
			Url:              baseURL,
			InputDto:         inputDTO,
			Requester:        requester,
			SetupSteps:       []helper.MockUseCaseSetupFunc[tenantMocks.MockGetTenantUseCase]{useCaseOK},
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: tenant.NewTenantResponseDTO(expectedTenant),
		},
		{
			Name:             "401 Unauthorized: missing identity",
			Method:           http.MethodPost,
			Url:              baseURL,
			InputDto:         inputDTO,
			OmitIdentity:     true,
			SetupSteps:       []helper.MockUseCaseSetupFunc[tenantMocks.MockGetTenantUseCase]{useCaseNeverCalled},
			ExpectedStatus:   http.StatusUnauthorized,
			ExpectedResponse: gin.H{"error": transportHttp.ErrMissingIdentity.Error()},
		},
		{
			Name:             "400 Bad Request: invalid body",
			Method:           http.MethodPost,
			Url:              baseURL,
			InputDto:         tenant.GetTenantDTO{},
			Requester:        requester,
			SetupSteps:       []helper.MockUseCaseSetupFunc[tenantMocks.MockGetTenantUseCase]{useCaseNeverCalled},
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedResponse: helper.HasError{},
		},
		{
			Name:             "401 Unauthorized: use case unauthorized",
			Method:           http.MethodPost,
			Url:              baseURL,
			InputDto:         inputDTO,
			Requester:        requester,
			SetupSteps:       []helper.MockUseCaseSetupFunc[tenantMocks.MockGetTenantUseCase]{useCaseUnauthorized},
			ExpectedStatus:   http.StatusUnauthorized,
			ExpectedResponse: gin.H{"error": identity.ErrUnauthorizedAccess.Error()},
		},
		{
			Name:             "404 Not Found: tenant not found",
			Method:           http.MethodPost,
			Url:              baseURL,
			InputDto:         inputDTO,
			Requester:        requester,
			SetupSteps:       []helper.MockUseCaseSetupFunc[tenantMocks.MockGetTenantUseCase]{useCaseNotFound},
			ExpectedStatus:   http.StatusNotFound,
			ExpectedResponse: gin.H{"error": tenant.ErrTenantNotFound.Error()},
		},
		{
			Name:             "500 Internal Server Error",
			Method:           http.MethodPost,
			Url:              baseURL,
			InputDto:         inputDTO,
			Requester:        requester,
			SetupSteps:       []helper.MockUseCaseSetupFunc[tenantMocks.MockGetTenantUseCase]{useCaseUnexpectedErr},
			ExpectedStatus:   http.StatusInternalServerError,
			ExpectedResponse: gin.H{"error": errMock.Error()},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			mockUseCase := helper.SetupMockUseCase(tenantMocks.NewMockGetTenantUseCase, tc.SetupSteps, t)
			controller := tenant.NewTenantController(zap.NewNop(), nil, nil, mockUseCase, nil, nil)
			helper.ExecuteControllerTest(t, tc, http.MethodPost, "/tenant/get", controller.GetTenant)
		})
	}
}

func TestController_GetTenantList(t *testing.T) {
	requester := identity.Requester{
		RequesterUserId: 1,
		RequesterRole:   identity.ROLE_SUPER_ADMIN,
	}
	tenantList := []tenant.Tenant{
		{Id: uuid.New(), Name: "Tenant A"},
		{Id: uuid.New(), Name: "Tenant B"},
	}
	total := uint(17)

	expectedCmd := tenant.GetTenantListCommand{Requester: requester, Page: 2, Limit: 10}

	useCaseOK := func(mockUC *tenantMocks.MockGetTenantListUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetTenantList(gomock.Eq(expectedCmd)).
			Return(tenantList, total, nil).
			Times(1)
	}
	useCaseNeverCalled := func(mockUC *tenantMocks.MockGetTenantListUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetTenantList(gomock.Any()).
			Times(0)
	}
	useCaseUnauthorized := func(mockUC *tenantMocks.MockGetTenantListUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetTenantList(gomock.Eq(expectedCmd)).
			Return(nil, uint(0), identity.ErrUnauthorizedAccess).
			Times(1)
	}
	errMock := errors.New("unexpected get list error")
	useCaseUnexpectedErr := func(mockUC *tenantMocks.MockGetTenantListUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetTenantList(gomock.Eq(expectedCmd)).
			Return(nil, uint(0), errMock).
			Times(1)
	}

	mountPath := "/tenants"
	baseURL := "/tenants?page=2&limit=10"
	cases := []helper.GenericControllerTestCase[any, tenantMocks.MockGetTenantListUseCase]{
		{
			Name:             "200 OK",
			Method:           http.MethodGet,
			Url:              baseURL,
			Requester:        requester,
			SetupSteps:       []helper.MockUseCaseSetupFunc[tenantMocks.MockGetTenantListUseCase]{useCaseOK},
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: tenant.NewTenantListResponseDTO(tenantList, total),
		},
		{
			Name:             "401 Unauthorized: missing identity",
			Method:           http.MethodGet,
			Url:              baseURL,
			OmitIdentity:     true,
			SetupSteps:       []helper.MockUseCaseSetupFunc[tenantMocks.MockGetTenantListUseCase]{useCaseNeverCalled},
			ExpectedStatus:   http.StatusUnauthorized,
			ExpectedResponse: gin.H{"error": transportHttp.ErrMissingIdentity.Error()},
		},
		{
			Name:             "400 Bad Request: invalid pagination",
			Method:           http.MethodGet,
			Url:              "/tenants?page=0&limit=10",
			Requester:        requester,
			SetupSteps:       []helper.MockUseCaseSetupFunc[tenantMocks.MockGetTenantListUseCase]{useCaseNeverCalled},
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedResponse: helper.HasError{},
		},
		{
			Name:             "401 Unauthorized: use case unauthorized",
			Method:           http.MethodGet,
			Url:              baseURL,
			Requester:        requester,
			SetupSteps:       []helper.MockUseCaseSetupFunc[tenantMocks.MockGetTenantListUseCase]{useCaseUnauthorized},
			ExpectedStatus:   http.StatusUnauthorized,
			ExpectedResponse: gin.H{"error": identity.ErrUnauthorizedAccess.Error()},
		},
		{
			Name:             "500 Internal Server Error",
			Method:           http.MethodGet,
			Url:              baseURL,
			Requester:        requester,
			SetupSteps:       []helper.MockUseCaseSetupFunc[tenantMocks.MockGetTenantListUseCase]{useCaseUnexpectedErr},
			ExpectedStatus:   http.StatusInternalServerError,
			ExpectedResponse: gin.H{"error": errMock.Error()},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			mockUseCase := helper.SetupMockUseCase(tenantMocks.NewMockGetTenantListUseCase, tc.SetupSteps, t)
			controller := tenant.NewTenantController(zap.NewNop(), nil, nil, nil, mockUseCase, nil)
			helper.ExecuteControllerTest(t, tc, http.MethodGet, mountPath, controller.GetTenantList)
		})
	}
}

func TestController_GetAllTenants(t *testing.T) {
	tenantList := []tenant.Tenant{{Id: uuid.New(), Name: "Tenant A"}, {Id: uuid.New(), Name: "Tenant B"}}

	useCaseOK := func(mockUC *tenantMocks.MockGetAllTenantsUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetAllTenants().
			Return(tenantList, nil).
			Times(1)
	}
	errMock := errors.New("unexpected get all error")
	useCaseUnexpectedErr := func(mockUC *tenantMocks.MockGetAllTenantsUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetAllTenants().
			Return(nil, errMock).
			Times(1)
	}

	baseURL := "/tenant/all"
	cases := []helper.GenericControllerTestCase[any, tenantMocks.MockGetAllTenantsUseCase]{
		{
			Name:             "200 OK",
			Method:           http.MethodGet,
			Url:              baseURL,
			SetupSteps:       []helper.MockUseCaseSetupFunc[tenantMocks.MockGetAllTenantsUseCase]{useCaseOK},
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: tenant.NewAllTenantsResponseDTO(tenantList),
		},
		{
			Name:             "500 Internal Server Error",
			Method:           http.MethodGet,
			Url:              baseURL,
			SetupSteps:       []helper.MockUseCaseSetupFunc[tenantMocks.MockGetAllTenantsUseCase]{useCaseUnexpectedErr},
			ExpectedStatus:   http.StatusInternalServerError,
			ExpectedResponse: gin.H{"error": errMock.Error()},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			mockUseCase := helper.SetupMockUseCase(tenantMocks.NewMockGetAllTenantsUseCase, tc.SetupSteps, t)
			controller := tenant.NewTenantController(zap.NewNop(), nil, nil, nil, nil, mockUseCase)
			helper.ExecuteControllerTest(t, tc, http.MethodGet, "/tenant/all", controller.GetAllTenants)
		})
	}
}
