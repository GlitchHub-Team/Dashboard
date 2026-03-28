package tenant_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"backend/internal/infra/transport/http/dto"
	"backend/internal/shared/identity"
	"backend/internal/tenant"
	"backend/tests/tenant/mocks"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	//"go.uber.org/zap"

	"go.uber.org/mock/gomock"
)

type hasError struct{}

func setupMockUseCase[MockT any](
	constructor func(*gomock.Controller) *MockT,
	setupSteps []mockUseCaseSetupFunc[MockT],
	t *testing.T,
) *MockT {
	// Isola mock controller
	ctrl := gomock.NewController(t)
	mockUseCase := constructor(ctrl)

	// Applica step in ordine
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

func executeControllerTest[InputT any, MockT any](
	t *testing.T,
	tc genericControllerTestCase[InputT, MockT],
	mountMethod string,
	mountUrl string,
	controllerFunc func(*gin.Context),
) {
	// 1. Imposta router
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

	// 2. Inject identity con middleware inline
	router.Use(func(ctx *gin.Context) {
		if !tc.omitIdentity {
			ctx.Set("requester_user_id", tc.requester.RequesterUserId)

			if tc.requester.RequesterRole != identity.ROLE_SUPER_ADMIN {
				if tc.requester.RequesterTenantId == nil {
					panic("Requester passed to test router is tenant member with no tenant id")
				}
				ctx.Set("requester_tenant_id", tc.requester.RequesterTenantId.String())
			}

			ctx.Set("requester_role", string(tc.requester.RequesterRole))
		}
		ctx.Next()
	})

	// 3. Imposta la route
	router.Handle(mountMethod, mountUrl, controllerFunc)

	// 4. Ottieni il corpo della richiesta se presente
	// TODO: test con Body DTO e Query DTO
	var reqBody []byte
	reqBody, err := json.Marshal(tc.inputDto)
	if err != nil {
		t.Fatalf("error marshaling: %v", err)
	}

	// 5. Crea ed esegui richiesta
	req, err := http.NewRequest(tc.method, tc.url, bytes.NewBuffer(reqBody)) //nolint:noctx
	if err != nil {
		t.Fatalf("Error when creating request: %v", err)
	}
	if reflect.ValueOf(tc.inputDto) != reflect.Zero(reflect.TypeFor[InputT]()) {
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 6. Verifica status code
	if w.Code != tc.expectedStatus {
		t.Fatalf("expected status %d, got %d. Response: %s", tc.expectedStatus, w.Code, w.Body.String())
	}

	// 7. Controlla risposta
	checkHttpResponse(w, tc.expectedResponse, t)
}

func checkHttpResponse[T any](w *httptest.ResponseRecorder, expectedResponse T, t *testing.T) {
	if reflect.ValueOf(expectedResponse) != reflect.Zero(reflect.TypeFor[T]()) {
		// Byte della risposta vera
		actualBytes := w.Body.Bytes()

		// NOTA: si usa Marshal per garantire allineamento di tipi
		expectedBytes, err := json.Marshal(expectedResponse)
		if err != nil {
			t.Fatalf("failed to marshal expected response: %v", err)
		}

		// NOTA: Si rifa unmarshaling in strutture, poi verificate con reflection
		var actualObj, expectedObj any
		if err := json.Unmarshal(actualBytes, &actualObj); err != nil {
			t.Fatalf("failed to unmarshal actual response: %v.Body was: %s", err, string(actualBytes))
		}
		if err := json.Unmarshal(expectedBytes, &expectedObj); err != nil {
			t.Fatalf("failed to unmarshal expected response: %v", err)
		}

		// Se risposta attesa è di tipo hasError, allora controlla solamente che actualObj contenga la chiave "error"
		_, onlyCheckError := any(expectedResponse).(hasError)
		if onlyCheckError {
			actualMap, ok := actualObj.(map[string]any)
			if !ok {
				t.Fatalf("Expected JSON object, got %v", actualObj)
			}
			_, ok = actualMap["error"]
			if !ok {
				t.Fatalf("Expected object with 'error' key, got %#v", actualObj)
			}
			return
		}

		// Check uguaglianza strutturato
		if !reflect.DeepEqual(expectedObj, actualObj) {
			t.Errorf("Response body mismatch.\nExpected: %#v\nGot:      %#v", expectedObj, actualObj)
		}
	}
}

// Tipi comuni ----------------------------------------------------------------------------------------
type mockUseCaseSetupFunc[T any] func(*T) *gomock.Call

type genericControllerTestCase[InputT any, MockT any] struct {
	name             string
	method           string
	url              string
	inputDto         InputT
	requester        identity.Requester
	omitIdentity     bool
	setupSteps       []mockUseCaseSetupFunc[MockT]
	expectedStatus   int
	expectedResponse any
}

// Create =============================================================================================
func TestController_TestTenantCreate(t *testing.T) {
	targetTenantId := uuid.New()
	targetTenantName := "Stefano"
	targetCanImpersonate := false

	expectedResponse := tenant.TenantResponseDTO{}
	expectedResponse.TenantId = targetTenantId.String()
	expectedResponse.TenantName = targetTenantName
	expectedResponse.CanImpersonate = targetCanImpersonate

	expectedTenant := tenant.Tenant{
		Id:             targetTenantId,
		Name:           targetTenantName,
		CanImpersonate: targetCanImpersonate,
	}

	requestBody := tenant.CreateTenantDTO{}
	requestBody.TenantName = targetTenantName
	requestBody.CanImpersonate = targetCanImpersonate

	// Requester
	/*
		superAdminRequester := identity.Requester{
			RequesterUserId: uint(1),
			RequesterRole:   identity.ROLE_SUPER_ADMIN,
		}
	*/

	authTenantAdminRequester := identity.Requester{
		RequesterUserId:   uint(1),
		RequesterTenantId: &targetTenantId,
		RequesterRole:     identity.ROLE_SUPER_ADMIN,
	}

	validPayload := tenant.CreateTenantCommand{
		Name:           "Stefano",
		CanImpersonate: false,
		Requester: identity.Requester{
			RequesterUserId:   uint(1),
			RequesterTenantId: nil,
			RequesterRole:     identity.ROLE_SUPER_ADMIN,
		},
	}

	useCaseok := func(mockUC *mocks.MockCreateTenantUseCase) *gomock.Call {
		return mockUC.EXPECT().
			CreateTenant(validPayload).
			Return(expectedTenant, nil).
			Times(1)
	}
	//	invalidPayload := tenant.CreateTenantCommand{}

	validUrl := fmt.Sprintf("/tenant/%v", targetTenantId.String())
	//	invalidUrl := "/tenant/123"

	cases := []genericControllerTestCase[tenant.CreateTenantDTO, mocks.MockCreateTenantUseCase]{
		{
			name:      "200 OK",
			method:    "POST",
			url:       validUrl,
			inputDto:  requestBody,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockCreateTenantUseCase]{
				useCaseok,
			},
			expectedStatus:   http.StatusOK,
			expectedResponse: expectedResponse,
		},
	}

	mountMethod := "POST"
	mountUrl := "/tenant/:tenant_id"

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockUseCase := setupMockUseCase(
				mocks.NewMockCreateTenantUseCase,
				tc.setupSteps, t,
			)
			tenantController := tenant.NewTenantController(
				nil, mockUseCase, nil, nil, nil, nil,
			)

			executeControllerTest[tenant.CreateTenantDTO, mocks.MockCreateTenantUseCase](
				t, tc,
				mountMethod, mountUrl,
				tenantController.CreateTenant,
			)
		})
	}
}

// Delete =============================================================================================
func TestController_DeleteTenant(t *testing.T) {
	targetTenantId := uuid.New()
	superAdminUserId := uint(1)

	expectedTenant := tenant.Tenant{
		Id:             targetTenantId,
		Name:           "Stefano",
		CanImpersonate: false,
	}

	expectedResponse := tenant.TenantResponseDTO{}
	expectedResponse.TenantId = targetTenantId.String()
	expectedResponse.TenantName = expectedTenant.Name
	expectedResponse.CanImpersonate = expectedTenant.CanImpersonate

	superAdminRequester := identity.Requester{
		RequesterUserId: superAdminUserId,
		RequesterRole:   identity.ROLE_SUPER_ADMIN,
	}

	validBody := tenant.DeleteTenantDTO{
		TenantIdField: dto.TenantIdField{TenantId: targetTenantId.String()},
	}

	validCmd := tenant.DeleteTenantCommand{
		Requester: identity.Requester{
			RequesterUserId: superAdminUserId,
			RequesterRole:   identity.ROLE_SUPER_ADMIN,
		},
		TenantId: targetTenantId,
	}

	validUrl := fmt.Sprintf("/tenant/%v", targetTenantId.String())
	mountMethod := "DELETE"
	mountUrl := "/tenant/:tenant_id"

	useCaseOK := func(m *mocks.MockDeleteTenantUseCase) *gomock.Call {
		return m.EXPECT().DeleteTenant(validCmd).Return(expectedTenant, nil).Times(1)
	}
	useCaseUnauthorized := func(m *mocks.MockDeleteTenantUseCase) *gomock.Call {
		return m.EXPECT().DeleteTenant(validCmd).Return(tenant.Tenant{}, identity.ErrUnauthorizedAccess).Times(1)
	}
	useCaseNotFound := func(m *mocks.MockDeleteTenantUseCase) *gomock.Call {
		return m.EXPECT().DeleteTenant(validCmd).Return(tenant.Tenant{}, tenant.ErrTenantNotFound).Times(1)
	}
	useCaseServerError := func(m *mocks.MockDeleteTenantUseCase) *gomock.Call {
		return m.EXPECT().DeleteTenant(validCmd).Return(tenant.Tenant{}, newMockError(1)).Times(1)
	}

	cases := []genericControllerTestCase[tenant.DeleteTenantDTO, mocks.MockDeleteTenantUseCase]{
		{
			name:             "200 OK",
			method:           mountMethod,
			url:              validUrl,
			inputDto:         validBody,
			requester:        superAdminRequester,
			setupSteps:       []mockUseCaseSetupFunc[mocks.MockDeleteTenantUseCase]{useCaseOK},
			expectedStatus:   http.StatusOK,
			expectedResponse: expectedResponse,
		},
		{
			name:             "401 - missing identity",
			method:           mountMethod,
			url:              validUrl,
			inputDto:         validBody,
			omitIdentity:     true,
			expectedStatus:   http.StatusUnauthorized,
			expectedResponse: hasError{},
		},
		{
			name:             "400 - invalid body",
			method:           mountMethod,
			url:              validUrl,
			inputDto:         tenant.DeleteTenantDTO{},
			requester:        superAdminRequester,
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: hasError{},
		},
		{
			name:             "401 - use case unauthorized",
			method:           mountMethod,
			url:              validUrl,
			inputDto:         validBody,
			requester:        superAdminRequester,
			setupSteps:       []mockUseCaseSetupFunc[mocks.MockDeleteTenantUseCase]{useCaseUnauthorized},
			expectedStatus:   http.StatusUnauthorized,
			expectedResponse: hasError{},
		},
		{
			name:             "400 - tenant not found",
			method:           mountMethod,
			url:              validUrl,
			inputDto:         validBody,
			requester:        superAdminRequester,
			setupSteps:       []mockUseCaseSetupFunc[mocks.MockDeleteTenantUseCase]{useCaseNotFound},
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: hasError{},
		},
		{
			name:             "500 - use case server error",
			method:           mountMethod,
			url:              validUrl,
			inputDto:         validBody,
			requester:        superAdminRequester,
			setupSteps:       []mockUseCaseSetupFunc[mocks.MockDeleteTenantUseCase]{useCaseServerError},
			expectedStatus:   http.StatusInternalServerError,
			expectedResponse: hasError{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockUseCase := setupMockUseCase(mocks.NewMockDeleteTenantUseCase, tc.setupSteps, t)
			controller := tenant.NewTenantController(nil, nil, mockUseCase, nil, nil, nil)
			executeControllerTest[tenant.DeleteTenantDTO, mocks.MockDeleteTenantUseCase](
				t, tc,
				mountMethod, mountUrl,
				controller.DeleteTenant,
			)
		})
	}
}

// Get Tenant =========================================================================================
func TestController_GetTenant(t *testing.T) {
	targetTenantId := uuid.New()
	superAdminUserId := uint(1)

	expectedTenant := tenant.Tenant{
		Id:             targetTenantId,
		Name:           "Stefano",
		CanImpersonate: true,
	}

	expectedResponse := tenant.TenantResponseDTO{}
	expectedResponse.TenantId = targetTenantId.String()
	expectedResponse.TenantName = expectedTenant.Name
	expectedResponse.CanImpersonate = expectedTenant.CanImpersonate

	superAdminRequester := identity.Requester{
		RequesterUserId: superAdminUserId,
		RequesterRole:   identity.ROLE_SUPER_ADMIN,
	}

	validBody := tenant.GetTenantDTO{
		TenantIdField: dto.TenantIdField{TenantId: targetTenantId.String()},
	}

	validCmd := tenant.GetTenantCommand{
		Requester: identity.Requester{
			RequesterUserId: superAdminUserId,
			RequesterRole:   identity.ROLE_SUPER_ADMIN,
		},
		TenantId: targetTenantId,
	}

	validUrl := fmt.Sprintf("/tenant/%v", targetTenantId.String())
	mountMethod := "GET"
	mountUrl := "/tenant/:tenant_id"

	useCaseOK := func(m *mocks.MockGetTenantUseCase) *gomock.Call {
		return m.EXPECT().GetTenant(validCmd).Return(expectedTenant, nil).Times(1)
	}
	useCaseUnauthorized := func(m *mocks.MockGetTenantUseCase) *gomock.Call {
		return m.EXPECT().GetTenant(validCmd).Return(tenant.Tenant{}, identity.ErrUnauthorizedAccess).Times(1)
	}
	useCaseNotFound := func(m *mocks.MockGetTenantUseCase) *gomock.Call {
		return m.EXPECT().GetTenant(validCmd).Return(tenant.Tenant{}, tenant.ErrTenantNotFound).Times(1)
	}
	useCaseServerError := func(m *mocks.MockGetTenantUseCase) *gomock.Call {
		return m.EXPECT().GetTenant(validCmd).Return(tenant.Tenant{}, newMockError(1)).Times(1)
	}

	cases := []genericControllerTestCase[tenant.GetTenantDTO, mocks.MockGetTenantUseCase]{
		{
			name:             "200 OK",
			method:           mountMethod,
			url:              validUrl,
			inputDto:         validBody,
			requester:        superAdminRequester,
			setupSteps:       []mockUseCaseSetupFunc[mocks.MockGetTenantUseCase]{useCaseOK},
			expectedStatus:   http.StatusOK,
			expectedResponse: expectedResponse,
		},
		{
			name:             "401 - missing identity",
			method:           mountMethod,
			url:              validUrl,
			inputDto:         validBody,
			omitIdentity:     true,
			expectedStatus:   http.StatusUnauthorized,
			expectedResponse: hasError{},
		},
		{
			name:             "400 - invalid body",
			method:           mountMethod,
			url:              validUrl,
			inputDto:         tenant.GetTenantDTO{},
			requester:        superAdminRequester,
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: hasError{},
		},
		{
			name:             "401 - use case unauthorized",
			method:           mountMethod,
			url:              validUrl,
			inputDto:         validBody,
			requester:        superAdminRequester,
			setupSteps:       []mockUseCaseSetupFunc[mocks.MockGetTenantUseCase]{useCaseUnauthorized},
			expectedStatus:   http.StatusUnauthorized,
			expectedResponse: hasError{},
		},
		{
			name:             "400 - tenant not found",
			method:           mountMethod,
			url:              validUrl,
			inputDto:         validBody,
			requester:        superAdminRequester,
			setupSteps:       []mockUseCaseSetupFunc[mocks.MockGetTenantUseCase]{useCaseNotFound},
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: hasError{},
		},
		{
			name:             "500 - use case server error",
			method:           mountMethod,
			url:              validUrl,
			inputDto:         validBody,
			requester:        superAdminRequester,
			setupSteps:       []mockUseCaseSetupFunc[mocks.MockGetTenantUseCase]{useCaseServerError},
			expectedStatus:   http.StatusInternalServerError,
			expectedResponse: hasError{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockUseCase := setupMockUseCase(mocks.NewMockGetTenantUseCase, tc.setupSteps, t)
			controller := tenant.NewTenantController(nil, nil, nil, mockUseCase, nil, nil)
			executeControllerTest[tenant.GetTenantDTO, mocks.MockGetTenantUseCase](
				t, tc,
				mountMethod, mountUrl,
				controller.GetTenant,
			)
		})
	}
}

// Get Tenant List ====================================================================================

func TestController_GetTenants(t *testing.T) {
	superAdminUserId := uint(1)

	tenantA := tenant.Tenant{Id: uuid.New(), Name: "Stefano", CanImpersonate: false}
	tenantB := tenant.Tenant{Id: uuid.New(), Name: "Tullio", CanImpersonate: true}
	tenantList := []tenant.Tenant{tenantA, tenantB}

	superAdminRequester := identity.Requester{
		RequesterUserId: superAdminUserId,
		RequesterRole:   identity.ROLE_SUPER_ADMIN,
	}

	validCmd := tenant.GetTenantListCommand{
		Requester: identity.Requester{
			RequesterUserId: superAdminUserId,
			RequesterRole:   identity.ROLE_SUPER_ADMIN,
		},
		Page:  1,
		Limit: 10,
	}

	validUrl := "/tenants?page=1&limit=10"
	mountMethod := "GET"
	mountUrl := "/tenants"

	useCaseOK := func(m *mocks.MockGetTenantListUseCase) *gomock.Call {
		return m.EXPECT().GetTenantList(validCmd).Return(tenantList, nil).Times(1)
	}
	useCaseUnauthorized := func(m *mocks.MockGetTenantListUseCase) *gomock.Call {
		return m.EXPECT().GetTenantList(validCmd).Return(nil, identity.ErrUnauthorizedAccess).Times(1)
	}
	useCaseServerError := func(m *mocks.MockGetTenantListUseCase) *gomock.Call {
		return m.EXPECT().GetTenantList(validCmd).Return(nil, newMockError(1)).Times(1)
	}

	cases := []genericControllerTestCase[struct{}, mocks.MockGetTenantListUseCase]{
		{
			name:           "200 OK",
			method:         mountMethod,
			url:            validUrl,
			requester:      superAdminRequester,
			setupSteps:     []mockUseCaseSetupFunc[mocks.MockGetTenantListUseCase]{useCaseOK},
			expectedStatus: http.StatusOK,
			expectedResponse: gin.H{
				"list_info": gin.H{
					"page":  1,
					"limit": 10,
				},
				"tenants": tenant.NewTenantListResponseDTO(tenantList, 10),
			},
		},
		{
			name:             "401 - missing identity",
			method:           mountMethod,
			url:              validUrl,
			omitIdentity:     true,
			expectedStatus:   http.StatusUnauthorized,
			expectedResponse: hasError{},
		},
		{
			name:             "401 - use case unauthorized",
			method:           mountMethod,
			url:              validUrl,
			requester:        superAdminRequester,
			setupSteps:       []mockUseCaseSetupFunc[mocks.MockGetTenantListUseCase]{useCaseUnauthorized},
			expectedStatus:   http.StatusUnauthorized,
			expectedResponse: hasError{},
		},
		{
			name:             "500 - use case server error",
			method:           mountMethod,
			url:              validUrl,
			requester:        superAdminRequester,
			setupSteps:       []mockUseCaseSetupFunc[mocks.MockGetTenantListUseCase]{useCaseServerError},
			expectedStatus:   http.StatusInternalServerError,
			expectedResponse: hasError{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockUseCase := setupMockUseCase(mocks.NewMockGetTenantListUseCase, tc.setupSteps, t)
			controller := tenant.NewTenantController(nil, nil, nil, nil, mockUseCase, nil)
			executeControllerTest[struct{}, mocks.MockGetTenantListUseCase](
				t, tc,
				mountMethod, mountUrl,
				controller.GetTenants,
			)
		})
	}
}

// Get Tenant By User =================================================================================
//
// /*
/*
func TestController_GetTenantByUser(t *testing.T) {
	targetTenantId := uuid.New()
	superAdminUserId := uint(1)

	expectedTenant := tenant.Tenant{
		Id:             targetTenantId,
		Name:           "Stefano",
		CanImpersonate: false,
	}

	superAdminRequester := identity.Requester{
		RequesterUserId: superAdminUserId,
		RequesterRole:   identity.ROLE_SUPER_ADMIN,
	}

	targetUserId := uint(2)

	validBody := tenant.GetTenantByUserDTO{
		UserIdField: dto.UserIdField{UserId: targetUserId},
	}

	validCmd := tenant.GetTenantByUserCommand{
		Requester: identity.Requester{
			RequesterUserId: superAdminUserId,
			RequesterRole:   identity.ROLE_SUPER_ADMIN,
		},
		UserId: targetUserId,
	}
	validUrl := "/tenant/by-user"
	mountMethod := "GET"
	mountUrl := "/tenant/by-user"

	useCaseOK := func(m *mocks.MockGetTenantByUserUseCase) *gomock.Call {
		return m.EXPECT().GetTenantByUser(validCmd).Return(expectedTenant, nil).Times(1)
	}
	useCaseNotFound := func(m *mocks.MockGetTenantByUserUseCase) *gomock.Call {
		return m.EXPECT().GetTenantByUser(validCmd).Return(tenant.Tenant{}, tenant.ErrTenantNotFound).Times(1)
	}
	useCaseServerError := func(m *mocks.MockGetTenantByUserUseCase) *gomock.Call {
		return m.EXPECT().GetTenantByUser(validCmd).Return(tenant.Tenant{}, newMockError(1)).Times(1)
	}

	cases := []genericControllerTestCase[tenant.GetTenantByUserDTO, mocks.MockGetTenantByUserUseCase]{
		{
			name:             "200 OK",
			method:           mountMethod,
			url:              validUrl,
			inputDto:         validBody,
			requester:        superAdminRequester,
			setupSteps:       []mockUseCaseSetupFunc[mocks.MockGetTenantByUserUseCase]{useCaseOK},
			expectedStatus:   http.StatusOK,
			expectedResponse: tenant.NewTenantResponseDTO(expectedTenant),
		},
		{
			name:             "401 - missing identity",
			method:           mountMethod,
			url:              validUrl,
			inputDto:         validBody,
			omitIdentity:     true,
			expectedStatus:   http.StatusUnauthorized,
			expectedResponse: hasError{},
		},
		{
			name:             "400 - invalid body (missing user_id)",
			method:           mountMethod,
			url:              validUrl,
			inputDto:         tenant.GetTenantByUserDTO{},
			requester:        superAdminRequester,
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: hasError{},
		},
		{
			name:             "400 - tenant not found",
			method:           mountMethod,
			url:              validUrl,
			inputDto:         validBody,
			requester:        superAdminRequester,
			setupSteps:       []mockUseCaseSetupFunc[mocks.MockGetTenantByUserUseCase]{useCaseNotFound},
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: hasError{},
		},
		{
			name:             "500 - server error",
			method:           mountMethod,
			url:              validUrl,
			inputDto:         validBody,
			requester:        superAdminRequester,
			setupSteps:       []mockUseCaseSetupFunc[mocks.MockGetTenantByUserUseCase]{useCaseServerError},
			expectedStatus:   http.StatusInternalServerError,
			expectedResponse: hasError{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockUseCase := setupMockUseCase(mocks.NewMockGetTenantByUserUseCase, tc.setupSteps, t)
			controller := tenant.NewTenantController(zap.NewNop(), nil, nil, nil, nil, mockUseCase)
			executeControllerTest[tenant.GetTenantByUserDTO, mocks.MockGetTenantByUserUseCase](
				t, tc,
				mountMethod, mountUrl,
				controller.GetTenantByUser,
			)
		})
	}
}
*/
