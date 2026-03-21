package user_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"backend/internal/common/dto"
	"backend/internal/identity"
	"backend/internal/tenant"
	transportHttp "backend/internal/transport/http"
	"backend/internal/user"
	"backend/tests/user/mocks"

	// "fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

// Funzioni comuni ------------------------------------------------------------------------------------

/*
	Passa questo struct nell'expectedResponse per verificare che ci sia un errore generico
*/
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
	inputDto         InputT             // Will be serialized to JSON
	requester        identity.Requester // To simulate the extracted identity
	omitIdentity     bool
	setupSteps       []mockUseCaseSetupFunc[MockT]
	expectedStatus   int
	expectedResponse any
}

// CREATE ==============================================================================================================
func TestController_CreateTenantUser(t *testing.T) {
	// type testCase genericControllerTestCase[user.CreateUserBodyDTO, mocks.MockCreateTenantUserUseCase, user.UserResponseDTO]

	// Data
	targetTenantId := uuid.New()

	targetUserEmail := "test@example.com"
	targetUserName := "Test"
	targetUserId := uint(1)
	targetUserRole := identity.ROLE_TENANT_USER
	targetConfirmed := false

	expectedResponse := user.UserResponseDTO{}
	expectedResponse.UserId = targetUserId
	expectedResponse.Username = targetUserName
	expectedResponse.Email = targetUserEmail
	expectedResponse.UserRole = string(targetUserRole)
	expectedResponse.TenantId = targetTenantId.String()

	expectedUser := user.User{
		Id:        targetUserId,
		Name:      targetUserName,
		Email:     targetUserEmail,
		TenantId:  &targetTenantId,
		Confirmed: targetConfirmed,
		Role:      targetUserRole,
	}

	// Setup -----
	useCaseOk := func(mockUC *mocks.MockCreateTenantUserUseCase) *gomock.Call {
		return mockUC.EXPECT().
			CreateTenantUser(gomock.Any()).
			Return(expectedUser, nil).
			Times(1)
	}

	useCaseNeverCalled := func(mockUC *mocks.MockCreateTenantUserUseCase) *gomock.Call {
		return mockUC.EXPECT().
			CreateTenantUser(gomock.Any()).
			Times(0)
	}

	useCaseTenantNotFound := func(mockUC *mocks.MockCreateTenantUserUseCase) *gomock.Call {
		return mockUC.EXPECT().
			CreateTenantUser(gomock.Any()).
			Return(user.User{}, tenant.ErrTenantNotFound).
			Times(1)
	}

	useCaseAlreadyExists := func(mockUC *mocks.MockCreateTenantUserUseCase) *gomock.Call {
		return mockUC.EXPECT().
			CreateTenantUser(gomock.Any()).
			Return(user.User{}, user.ErrUserAlreadyExists).
			Times(1)
	}

	useCaseUnauthorizedAccess := func(mockUC *mocks.MockCreateTenantUserUseCase) *gomock.Call {
		return mockUC.EXPECT().
			CreateTenantUser(gomock.Any()).
			Return(user.User{}, identity.ErrUnauthorizedAccess).
			Times(1)
	}

	useCaseCannotSendEmail := func(mockUC *mocks.MockCreateTenantUserUseCase) *gomock.Call {
		return mockUC.EXPECT().
			CreateTenantUser(gomock.Any()).
			Return(user.User{}, user.ErrCannotSendEmail).
			Times(1)
	}

	errMock := errors.New("unexpected error")
	useCaseUnexpectedErr := func(mockUC *mocks.MockCreateTenantUserUseCase) *gomock.Call {
		return mockUC.EXPECT().
			CreateTenantUser(gomock.Any()).
			Return(user.User{}, errMock).
			Times(1)
	}

	// Requester
	superAdminRequester := identity.Requester{
		RequesterUserId: uint(1),
		RequesterRole:   identity.ROLE_SUPER_ADMIN,
	}

	authTenantAdminRequester := identity.Requester{
		RequesterUserId:   uint(1),
		RequesterTenantId: &targetTenantId,
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	// Input
	validPayload := user.CreateUserBodyDTO{
		EmailField:    dto.EmailField{Email: targetUserEmail},
		UsernameField: dto.UsernameField{Username: targetUserName},
	}
	invalidPayload := user.CreateUserBodyDTO{}

	validUrl := fmt.Sprintf("/tenant/%v/tenant_user", targetTenantId.String())
	invalidUrl := "/tenant/123/tenant_user"

	cases := []genericControllerTestCase[user.CreateUserBodyDTO, mocks.MockCreateTenantUserUseCase]{
		{
			name:      "200 OK",
			method:    "POST",
			url:       validUrl,
			inputDto:  validPayload,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockCreateTenantUserUseCase]{
				useCaseOk,
			},
			expectedStatus:   http.StatusOK,
			expectedResponse: expectedResponse,
		},
		{
			name:      "400 Bad Request: Bad URI",
			method:    "POST",
			url:       invalidUrl,
			inputDto:  validPayload,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockCreateTenantUserUseCase]{
				useCaseNeverCalled,
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: gin.H{
				"error": "invalid format",
				"fields": gin.H{
					"tenant_id": "uuid4",
				},
			},
		},
		{
			name:      "400 Bad Request: Bad body",
			method:    "POST",
			url:       validUrl,
			inputDto:  invalidPayload,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockCreateTenantUserUseCase]{
				useCaseNeverCalled,
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: gin.H{
				"error": "invalid format",
				"fields": gin.H{
					"email":    "required",
					"username": "required",
				},
			},
		},
		{
			name:      "400 Bad request: User already exists",
			method:    "POST",
			url:       validUrl,
			inputDto:  validPayload,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockCreateTenantUserUseCase]{
				useCaseAlreadyExists,
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: gin.H{
				"error": user.ErrUserAlreadyExists.Error(),
			},
		},
		{
			name:         "401 Unauthorized: No identity",
			method:       "POST",
			url:          validUrl,
			inputDto:     validPayload,
			requester:    authTenantAdminRequester,
			omitIdentity: true,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockCreateTenantUserUseCase]{
				useCaseNeverCalled,
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: gin.H{
				"error": transportHttp.ErrMissingIdentity.Error(),
			},
		},
		{
			name:      "404 Not found: Tenant not found",
			method:    "POST",
			url:       validUrl,
			inputDto:  validPayload,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockCreateTenantUserUseCase]{
				useCaseTenantNotFound,
			},
			expectedStatus: http.StatusNotFound,
			expectedResponse: gin.H{
				"error": tenant.ErrTenantNotFound.Error(),
			},
		},
		{
			name:      "404 Not found: Unauthorized access (obfuscated)",
			method:    "POST",
			url:       validUrl,
			inputDto:  validPayload,
			requester: superAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockCreateTenantUserUseCase]{
				useCaseUnauthorizedAccess,
			},
			expectedStatus: http.StatusNotFound,
			expectedResponse: gin.H{
				"error": tenant.ErrTenantNotFound.Error(),
			},
		},
		{
			name:      "500 Server Error: Cannot create user (cannot send email)",
			method:    "POST",
			url:       validUrl,
			inputDto:  validPayload,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockCreateTenantUserUseCase]{
				useCaseCannotSendEmail,
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResponse: gin.H{
				"error": user.ErrCannotSendEmail.Error(),
			},
		},
		{
			name:      "500 Server Error: Unexpected error",
			method:    "POST",
			url:       validUrl,
			inputDto:  validPayload,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockCreateTenantUserUseCase]{
				useCaseUnexpectedErr,
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResponse: gin.H{
				"error": errMock.Error(),
			},
		},
	}

	// Parametri di test
	mountMethod := "POST"
	mountUrl := "/tenant/:tenant_id/tenant_user"

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockUseCase := setupMockUseCase(
				mocks.NewMockCreateTenantUserUseCase,
				tc.setupSteps, t,
			)
			userController := user.NewUserController(
				nil, mockUseCase, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
			)

			executeControllerTest(
				t, tc,
				mountMethod, mountUrl,
				userController.CreateTenantUser,
			)
		})
	}
}

func TestController_CreateTenantAdmin(t *testing.T) {
	// type testCase genericControllerTestCase[user.CreateUserBodyDTO, mocks.MockCreateTenantUserUseCase, user.UserResponseDTO]

	// Data
	targetTenantId := uuid.New()

	targetUserEmail := "test@example.com"
	targetUserName := "Test"
	targetUserId := uint(1)
	targetUserRole := identity.ROLE_TENANT_ADMIN
	targetConfirmed := false

	expectedResponse := user.UserResponseDTO{}
	expectedResponse.UserId = targetUserId
	expectedResponse.Username = targetUserName
	expectedResponse.Email = targetUserEmail
	expectedResponse.UserRole = string(targetUserRole)
	expectedResponse.TenantId = targetTenantId.String()

	expectedUser := user.User{
		Id:        targetUserId,
		Name:      targetUserName,
		Email:     targetUserEmail,
		TenantId:  &targetTenantId,
		Confirmed: targetConfirmed,
		Role:      targetUserRole,
	}

	// Setup -----
	useCaseOk := func(mockUC *mocks.MockCreateTenantAdminUseCase) *gomock.Call {
		return mockUC.EXPECT().
			CreateTenantAdmin(gomock.Any()).
			Return(expectedUser, nil).
			Times(1)
	}

	useCaseNeverCalled := func(mockUC *mocks.MockCreateTenantAdminUseCase) *gomock.Call {
		return mockUC.EXPECT().
			CreateTenantAdmin(gomock.Any()).
			Times(0)
	}

	useCaseTenantNotFound := func(mockUC *mocks.MockCreateTenantAdminUseCase) *gomock.Call {
		return mockUC.EXPECT().
			CreateTenantAdmin(gomock.Any()).
			Return(user.User{}, tenant.ErrTenantNotFound).
			Times(1)
	}

	useCaseAlreadyExists := func(mockUC *mocks.MockCreateTenantAdminUseCase) *gomock.Call {
		return mockUC.EXPECT().
			CreateTenantAdmin(gomock.Any()).
			Return(user.User{}, user.ErrUserAlreadyExists).
			Times(1)
	}

	useCaseUnauthorizedAccess := func(mockUC *mocks.MockCreateTenantAdminUseCase) *gomock.Call {
		return mockUC.EXPECT().
			CreateTenantAdmin(gomock.Any()).
			Return(user.User{}, identity.ErrUnauthorizedAccess).
			Times(1)
	}

	useCaseCannotSendEmail := func(mockUC *mocks.MockCreateTenantAdminUseCase) *gomock.Call {
		return mockUC.EXPECT().
			CreateTenantAdmin(gomock.Any()).
			Return(user.User{}, user.ErrCannotSendEmail).
			Times(1)
	}

	errMock := errors.New("unexpected error")
	useCaseUnexpectedErr := func(mockUC *mocks.MockCreateTenantAdminUseCase) *gomock.Call {
		return mockUC.EXPECT().
			CreateTenantAdmin(gomock.Any()).
			Return(user.User{}, errMock).
			Times(1)
	}

	// Requester
	superAdminRequester := identity.Requester{
		RequesterUserId: uint(1),
		RequesterRole:   identity.ROLE_SUPER_ADMIN,
	}

	authTenantAdminRequester := identity.Requester{
		RequesterUserId:   uint(1),
		RequesterTenantId: &targetTenantId,
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	// Input
	validPayload := user.CreateUserBodyDTO{
		EmailField:    dto.EmailField{Email: targetUserEmail},
		UsernameField: dto.UsernameField{Username: targetUserName},
	}
	invalidPayload := user.CreateUserBodyDTO{}

	validUrl := fmt.Sprintf("/tenant/%v/tenant_admin", targetTenantId.String())
	invalidUrl := "/tenant/123/tenant_admin"

	cases := []genericControllerTestCase[user.CreateUserBodyDTO, mocks.MockCreateTenantAdminUseCase]{
		{
			name:      "200 OK",
			method:    "POST",
			url:       validUrl,
			inputDto:  validPayload,
			requester: superAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockCreateTenantAdminUseCase]{
				useCaseOk,
			},
			expectedStatus:   http.StatusOK,
			expectedResponse: expectedResponse,
		},
		{
			name:      "400 Bad Request: Bad URI",
			method:    "POST",
			url:       invalidUrl,
			inputDto:  validPayload,
			requester: superAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockCreateTenantAdminUseCase]{
				useCaseNeverCalled,
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: gin.H{
				"error": "invalid format",
				"fields": gin.H{
					"tenant_id": "uuid4",
				},
			},
		},
		{
			name:      "400 Bad Request: Bad body",
			method:    "POST",
			url:       validUrl,
			inputDto:  invalidPayload,
			requester: superAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockCreateTenantAdminUseCase]{
				useCaseNeverCalled,
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: gin.H{
				"error": "invalid format",
				"fields": gin.H{
					"email":    "required",
					"username": "required",
				},
			},
		},
		{
			name:      "400 Bad request: User already exists",
			method:    "POST",
			url:       validUrl,
			inputDto:  validPayload,
			requester: superAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockCreateTenantAdminUseCase]{
				useCaseAlreadyExists,
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: gin.H{
				"error": user.ErrUserAlreadyExists.Error(),
			},
		},
		{
			name:         "401 Unauthorized: No identity",
			method:       "POST",
			url:          validUrl,
			inputDto:     validPayload,
			requester:    superAdminRequester,
			omitIdentity: true,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockCreateTenantAdminUseCase]{
				useCaseNeverCalled,
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: gin.H{
				"error": transportHttp.ErrMissingIdentity.Error(),
			},
		},
		{
			name:      "404 Not found: Tenant not found",
			method:    "POST",
			url:       validUrl,
			inputDto:  validPayload,
			requester: superAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockCreateTenantAdminUseCase]{
				useCaseTenantNotFound,
			},
			expectedStatus: http.StatusNotFound,
			expectedResponse: gin.H{
				"error": tenant.ErrTenantNotFound.Error(),
			},
		},
		{
			name:      "404 Not found: Unauthorized access (obfuscated)",
			method:    "POST",
			url:       validUrl,
			inputDto:  validPayload,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockCreateTenantAdminUseCase]{
				useCaseUnauthorizedAccess,
			},
			expectedStatus: http.StatusNotFound,
			expectedResponse: gin.H{
				"error": tenant.ErrTenantNotFound.Error(),
			},
		},
		{
			name:      "500 Server Error: Cannot create user (cannot send email)",
			method:    "POST",
			url:       validUrl,
			inputDto:  validPayload,
			requester: superAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockCreateTenantAdminUseCase]{
				useCaseCannotSendEmail,
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResponse: gin.H{
				"error": user.ErrCannotSendEmail.Error(),
			},
		},
		{
			name:      "500 Server Error: Unexpected error",
			method:    "POST",
			url:       validUrl,
			inputDto:  validPayload,
			requester: superAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockCreateTenantAdminUseCase]{
				useCaseUnexpectedErr,
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResponse: gin.H{
				"error": errMock.Error(),
			},
		},
	}

	// Parametri di test
	mountMethod := "POST"
	mountUrl := "/tenant/:tenant_id/tenant_admin"

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockUseCase := setupMockUseCase(
				mocks.NewMockCreateTenantAdminUseCase,
				tc.setupSteps, t,
			)
			userController := user.NewUserController(
				nil, nil, mockUseCase, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
			)

			executeControllerTest(
				t, tc,
				mountMethod, mountUrl,
				userController.CreateTenantAdmin,
			)
		})
	}
}

func TestController_CreateSuperAdmin(t *testing.T) {
	// type testCase genericControllerTestCase[user.CreateUserBodyDTO, mocks.MockCreateTenantUserUseCase, user.UserResponseDTO]

	// Data
	targetUserEmail := "test@example.com"
	targetUserName := "Test"
	targetUserId := uint(1)
	targetUserRole := identity.ROLE_SUPER_ADMIN
	targetConfirmed := false

	expectedResponse := user.UserResponseDTO{}
	expectedResponse.UserId = targetUserId
	expectedResponse.Username = targetUserName
	expectedResponse.Email = targetUserEmail
	expectedResponse.UserRole = string(targetUserRole)
	// expectedResponse.TenantId = targetTenantId.String()

	expectedUser := user.User{
		Id:    targetUserId,
		Name:  targetUserName,
		Email: targetUserEmail,
		// TenantId:  &targetTenantId,
		Confirmed: targetConfirmed,
		Role:      targetUserRole,
	}

	// Setup -----
	useCaseOk := func(mockUC *mocks.MockCreateSuperAdminUseCase) *gomock.Call {
		return mockUC.EXPECT().
			CreateSuperAdmin(gomock.Any()).
			Return(expectedUser, nil).
			Times(1)
	}

	useCaseNeverCalled := func(mockUC *mocks.MockCreateSuperAdminUseCase) *gomock.Call {
		return mockUC.EXPECT().
			CreateSuperAdmin(gomock.Any()).
			Times(0)
	}

	useCaseTenantNotFound := func(mockUC *mocks.MockCreateSuperAdminUseCase) *gomock.Call {
		return mockUC.EXPECT().
			CreateSuperAdmin(gomock.Any()).
			Return(user.User{}, tenant.ErrTenantNotFound).
			Times(1)
	}

	useCaseAlreadyExists := func(mockUC *mocks.MockCreateSuperAdminUseCase) *gomock.Call {
		return mockUC.EXPECT().
			CreateSuperAdmin(gomock.Any()).
			Return(user.User{}, user.ErrUserAlreadyExists).
			Times(1)
	}

	useCaseUnauthorizedAccess := func(mockUC *mocks.MockCreateSuperAdminUseCase) *gomock.Call {
		return mockUC.EXPECT().
			CreateSuperAdmin(gomock.Any()).
			Return(user.User{}, identity.ErrUnauthorizedAccess).
			Times(1)
	}

	useCaseCannotSendEmail := func(mockUC *mocks.MockCreateSuperAdminUseCase) *gomock.Call {
		return mockUC.EXPECT().
			CreateSuperAdmin(gomock.Any()).
			Return(user.User{}, user.ErrCannotSendEmail).
			Times(1)
	}

	errMock := errors.New("unexpected error")
	useCaseUnexpectedErr := func(mockUC *mocks.MockCreateSuperAdminUseCase) *gomock.Call {
		return mockUC.EXPECT().
			CreateSuperAdmin(gomock.Any()).
			Return(user.User{}, errMock).
			Times(1)
	}

	// Requester
	superAdminRequester := identity.Requester{
		RequesterUserId: uint(1),
		RequesterRole:   identity.ROLE_SUPER_ADMIN,
	}

	tenantId := uuid.New()
	authTenantAdminRequester := identity.Requester{
		RequesterUserId:   uint(1),
		RequesterTenantId: &tenantId,
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	// Input
	validPayload := user.CreateUserBodyDTO{
		EmailField:    dto.EmailField{Email: targetUserEmail},
		UsernameField: dto.UsernameField{Username: targetUserName},
	}
	invalidPayload := user.CreateUserBodyDTO{}

	// validUrl := fmt.Sprintf("/super_admin", targetTenantId.String())
	// invalidUrl := "/tenant/123/super_admin"
	validUrl := "/super_admin"

	cases := []genericControllerTestCase[user.CreateUserBodyDTO, mocks.MockCreateSuperAdminUseCase]{
		{
			name:      "200 OK",
			method:    "POST",
			url:       validUrl,
			inputDto:  validPayload,
			requester: superAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockCreateSuperAdminUseCase]{
				useCaseOk,
			},
			expectedStatus:   http.StatusOK,
			expectedResponse: expectedResponse,
		},
		{
			name:      "400 Bad Request: Bad body",
			method:    "POST",
			url:       validUrl,
			inputDto:  invalidPayload,
			requester: superAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockCreateSuperAdminUseCase]{
				useCaseNeverCalled,
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: gin.H{
				"error": "invalid format",
				"fields": gin.H{
					"email":    "required",
					"username": "required",
				},
			},
		},
		{
			name:      "400 Bad request: User already exists",
			method:    "POST",
			url:       validUrl,
			inputDto:  validPayload,
			requester: superAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockCreateSuperAdminUseCase]{
				useCaseAlreadyExists,
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: gin.H{
				"error": user.ErrUserAlreadyExists.Error(),
			},
		},
		{
			name:         "401 Unauthorized: No identity",
			method:       "POST",
			url:          validUrl,
			inputDto:     validPayload,
			requester:    superAdminRequester,
			omitIdentity: true,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockCreateSuperAdminUseCase]{
				useCaseNeverCalled,
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: gin.H{
				"error": transportHttp.ErrMissingIdentity.Error(),
			},
		},
		{
			name:      "404 Not found: Tenant not found",
			method:    "POST",
			url:       validUrl,
			inputDto:  validPayload,
			requester: superAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockCreateSuperAdminUseCase]{
				useCaseTenantNotFound,
			},
			expectedStatus: http.StatusNotFound,
			expectedResponse: gin.H{
				"error": tenant.ErrTenantNotFound.Error(),
			},
		},
		{
			name:      "404 Not found: Unauthorized access (obfuscated)",
			method:    "POST",
			url:       validUrl,
			inputDto:  validPayload,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockCreateSuperAdminUseCase]{
				useCaseUnauthorizedAccess,
			},
			expectedStatus: http.StatusNotFound,
			expectedResponse: gin.H{
				"error": tenant.ErrTenantNotFound.Error(),
			},
		},
		{
			name:      "500 Server Error: Cannot create user (cannot send email)",
			method:    "POST",
			url:       validUrl,
			inputDto:  validPayload,
			requester: superAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockCreateSuperAdminUseCase]{
				useCaseCannotSendEmail,
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResponse: gin.H{
				"error": user.ErrCannotSendEmail.Error(),
			},
		},
		{
			name:      "500 Server Error: Unexpected error",
			method:    "POST",
			url:       validUrl,
			inputDto:  validPayload,
			requester: superAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockCreateSuperAdminUseCase]{
				useCaseUnexpectedErr,
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResponse: gin.H{
				"error": errMock.Error(),
			},
		},
	}

	// Parametri di test
	mountMethod := "POST"
	mountUrl := "/super_admin"

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockUseCase := setupMockUseCase(
				mocks.NewMockCreateSuperAdminUseCase,
				tc.setupSteps, t,
			)
			userController := user.NewUserController(
				nil, nil, nil, mockUseCase, nil, nil, nil, nil, nil, nil, nil, nil, nil,
			)

			executeControllerTest(
				t, tc,
				mountMethod, mountUrl,
				userController.CreateSuperAdmin,
			)
		})
	}
}

// DELETE ==============================================================================================================
func TestController_DeleteTenantUser(t *testing.T) {
	// Data
	targetTenantId := uuid.New()

	targetUserEmail := "test@example.com"
	targetUserName := "Test"
	targetUserId := uint(1)
	targetUserRole := identity.ROLE_TENANT_USER
	targetConfirmed := false

	expectedResponse := user.UserResponseDTO{}
	expectedResponse.UserId = targetUserId
	expectedResponse.Username = targetUserName
	expectedResponse.Email = targetUserEmail
	expectedResponse.UserRole = string(targetUserRole)
	expectedResponse.TenantId = targetTenantId.String()

	expectedUser := user.User{
		Id:        targetUserId,
		Name:      targetUserName,
		Email:     targetUserEmail,
		TenantId:  &targetTenantId,
		Confirmed: targetConfirmed,
		Role:      targetUserRole,
	}

	// Setup -----
	useCaseOk := func(mockUC *mocks.MockDeleteTenantUserUseCase) *gomock.Call {
		return mockUC.EXPECT().
			DeleteTenantUser(gomock.Any()).
			Return(expectedUser, nil).
			Times(1)
	}

	useCaseNeverCalled := func(mockUC *mocks.MockDeleteTenantUserUseCase) *gomock.Call {
		return mockUC.EXPECT().
			DeleteTenantUser(gomock.Any()).
			Times(0)
	}

	useCaseTenantNotFound := func(mockUC *mocks.MockDeleteTenantUserUseCase) *gomock.Call {
		return mockUC.EXPECT().
			DeleteTenantUser(gomock.Any()).
			Return(user.User{}, tenant.ErrTenantNotFound).
			Times(1)
	}

	useCaseNotFound := func(mockUC *mocks.MockDeleteTenantUserUseCase) *gomock.Call {
		return mockUC.EXPECT().
			DeleteTenantUser(gomock.Any()).
			Return(user.User{}, user.ErrUserNotFound).
			Times(1)
	}

	useCaseUnauthorizedAccess := func(mockUC *mocks.MockDeleteTenantUserUseCase) *gomock.Call {
		return mockUC.EXPECT().
			DeleteTenantUser(gomock.Any()).
			Return(user.User{}, identity.ErrUnauthorizedAccess).
			Times(1)
	}

	errMock := errors.New("unexpected error")
	useCaseUnexpectedErr := func(mockUC *mocks.MockDeleteTenantUserUseCase) *gomock.Call {
		return mockUC.EXPECT().
			DeleteTenantUser(gomock.Any()).
			Return(user.User{}, errMock).
			Times(1)
	}

	// Requester
	superAdminRequester := identity.Requester{
		RequesterUserId: uint(1),
		RequesterRole:   identity.ROLE_SUPER_ADMIN,
	}

	authTenantAdminRequester := identity.Requester{
		RequesterUserId:   uint(1),
		RequesterTenantId: &targetTenantId,
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	// Input
	type Empty struct{}
	validPayload := Empty{}

	validUrl := fmt.Sprintf("/tenant/%v/tenant_user/%v", targetTenantId.String(), targetUserId)
	invalidUrl_tenantId := "/tenant/123/tenant_user/1"
	invalidUrl_userId := fmt.Sprintf("/tenant/%v/tenant_user/0", targetTenantId.String())

	cases := []genericControllerTestCase[Empty, mocks.MockDeleteTenantUserUseCase]{
		{
			name:      "200 OK",
			method:    "DELETE",
			url:       validUrl,
			inputDto:  validPayload,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockDeleteTenantUserUseCase]{
				useCaseOk,
			},
			expectedStatus:   http.StatusOK,
			expectedResponse: expectedResponse,
		},
		{
			name:      "400 Bad Request: Bad URI (tenant Id)",
			method:    "DELETE",
			url:       invalidUrl_tenantId,
			inputDto:  validPayload,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockDeleteTenantUserUseCase]{
				useCaseNeverCalled,
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: gin.H{
				"error": "invalid format",
				"fields": gin.H{
					"tenant_id": "uuid4",
				},
			},
		},
		{
			name:      "400 Bad Request: Bad URI (user Id)",
			method:    "DELETE",
			url:       invalidUrl_userId,
			inputDto:  validPayload,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockDeleteTenantUserUseCase]{
				useCaseNeverCalled,
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: gin.H{
				"error": "invalid format",
				"fields": gin.H{
					"user_id": "required",
				},
			},
		},
		{
			name:         "401 Unauthorized: No identity",
			method:       "DELETE",
			url:          validUrl,
			inputDto:     validPayload,
			requester:    authTenantAdminRequester,
			omitIdentity: true,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockDeleteTenantUserUseCase]{
				useCaseNeverCalled,
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: gin.H{
				"error": transportHttp.ErrMissingIdentity.Error(),
			},
		},
		{
			name:      "404 Not found: Tenant not found",
			method:    "DELETE",
			url:       validUrl,
			inputDto:  validPayload,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockDeleteTenantUserUseCase]{
				useCaseTenantNotFound,
			},
			expectedStatus: http.StatusNotFound,
			expectedResponse: gin.H{
				"error": tenant.ErrTenantNotFound.Error(),
			},
		},
		{
			name:      "404 Not found: Unauthorized access (obfuscated)",
			method:    "DELETE",
			url:       validUrl,
			inputDto:  validPayload,
			requester: superAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockDeleteTenantUserUseCase]{
				useCaseUnauthorizedAccess,
			},
			expectedStatus: http.StatusNotFound,
			expectedResponse: gin.H{
				"error": tenant.ErrTenantNotFound.Error(),
			},
		},
		{
			name:      "404 Not found: User not found",
			method:    "DELETE",
			url:       validUrl,
			inputDto:  validPayload,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockDeleteTenantUserUseCase]{
				useCaseNotFound,
			},
			expectedStatus: http.StatusNotFound,
			expectedResponse: gin.H{
				"error": user.ErrUserNotFound.Error(),
			},
		},
		{
			name:      "500 Server Error: Unexpected error",
			method:    "DELETE",
			url:       validUrl,
			inputDto:  validPayload,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockDeleteTenantUserUseCase]{
				useCaseUnexpectedErr,
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResponse: gin.H{
				"error": errMock.Error(),
			},
		},
	}

	// Parametri di test
	mountMethod := "DELETE"
	mountUrl := "/tenant/:tenant_id/tenant_user/:user_id"

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockUseCase := setupMockUseCase(
				mocks.NewMockDeleteTenantUserUseCase,
				tc.setupSteps, t,
			)
			userController := user.NewUserController(
				nil, nil, nil, nil, mockUseCase, nil, nil, nil, nil, nil, nil, nil, nil,
			)

			executeControllerTest(
				t, tc,
				mountMethod, mountUrl,
				userController.DeleteTenantUser,
			)
		})
	}
}

func TestController_DeleteTenantAdmin(t *testing.T) {
	// Data
	targetTenantId := uuid.New()

	targetUserEmail := "test@example.com"
	targetUserName := "Test"
	targetUserId := uint(1)
	targetUserRole := identity.ROLE_TENANT_ADMIN
	targetConfirmed := false

	expectedResponse := user.UserResponseDTO{}
	expectedResponse.UserId = targetUserId
	expectedResponse.Username = targetUserName
	expectedResponse.Email = targetUserEmail
	expectedResponse.UserRole = string(targetUserRole)
	expectedResponse.TenantId = targetTenantId.String()

	expectedUser := user.User{
		Id:        targetUserId,
		Name:      targetUserName,
		Email:     targetUserEmail,
		TenantId:  &targetTenantId,
		Confirmed: targetConfirmed,
		Role:      targetUserRole,
	}

	// Setup -----
	useCaseOk := func(mockUC *mocks.MockDeleteTenantAdminUseCase) *gomock.Call {
		return mockUC.EXPECT().
			DeleteTenantAdmin(gomock.Any()).
			Return(expectedUser, nil).
			Times(1)
	}

	useCaseNeverCalled := func(mockUC *mocks.MockDeleteTenantAdminUseCase) *gomock.Call {
		return mockUC.EXPECT().
			DeleteTenantAdmin(gomock.Any()).
			Times(0)
	}

	useCaseTenantNotFound := func(mockUC *mocks.MockDeleteTenantAdminUseCase) *gomock.Call {
		return mockUC.EXPECT().
			DeleteTenantAdmin(gomock.Any()).
			Return(user.User{}, tenant.ErrTenantNotFound).
			Times(1)
	}

	useCaseNotFound := func(mockUC *mocks.MockDeleteTenantAdminUseCase) *gomock.Call {
		return mockUC.EXPECT().
			DeleteTenantAdmin(gomock.Any()).
			Return(user.User{}, user.ErrUserNotFound).
			Times(1)
	}

	useCaseUnauthorizedAccess := func(mockUC *mocks.MockDeleteTenantAdminUseCase) *gomock.Call {
		return mockUC.EXPECT().
			DeleteTenantAdmin(gomock.Any()).
			Return(user.User{}, identity.ErrUnauthorizedAccess).
			Times(1)
	}

	errMock := errors.New("unexpected error")
	useCaseUnexpectedErr := func(mockUC *mocks.MockDeleteTenantAdminUseCase) *gomock.Call {
		return mockUC.EXPECT().
			DeleteTenantAdmin(gomock.Any()).
			Return(user.User{}, errMock).
			Times(1)
	}

	// Requester
	superAdminRequester := identity.Requester{
		RequesterUserId: uint(1),
		RequesterRole:   identity.ROLE_SUPER_ADMIN,
	}

	authTenantAdminRequester := identity.Requester{
		RequesterUserId:   uint(1),
		RequesterTenantId: &targetTenantId,
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	// Input
	type Empty struct{}
	validPayload := Empty{}

	validUrl := fmt.Sprintf("/tenant/%v/tenant_admin/%v", targetTenantId.String(), targetUserId)
	invalidUrl_tenantId := "/tenant/123/tenant_admin/1"
	invalidUrl_userId := fmt.Sprintf("/tenant/%v/tenant_admin/0", targetTenantId.String())

	cases := []genericControllerTestCase[Empty, mocks.MockDeleteTenantAdminUseCase]{
		{
			name:      "200 OK",
			method:    "DELETE",
			url:       validUrl,
			inputDto:  validPayload,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockDeleteTenantAdminUseCase]{
				useCaseOk,
			},
			expectedStatus:   http.StatusOK,
			expectedResponse: expectedResponse,
		},
		{
			name:      "400 Bad Request: Bad URI (tenant Id)",
			method:    "DELETE",
			url:       invalidUrl_tenantId,
			inputDto:  validPayload,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockDeleteTenantAdminUseCase]{
				useCaseNeverCalled,
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: gin.H{
				"error": "invalid format",
				"fields": gin.H{
					"tenant_id": "uuid4",
				},
			},
		},
		{
			name:      "400 Bad Request: Bad URI (user Id)",
			method:    "DELETE",
			url:       invalidUrl_userId,
			inputDto:  validPayload,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockDeleteTenantAdminUseCase]{
				useCaseNeverCalled,
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: gin.H{
				"error": "invalid format",
				"fields": gin.H{
					"user_id": "required",
				},
			},
		},
		{
			name:         "401 Unauthorized: No identity",
			method:       "DELETE",
			url:          validUrl,
			inputDto:     validPayload,
			requester:    authTenantAdminRequester,
			omitIdentity: true,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockDeleteTenantAdminUseCase]{
				useCaseNeverCalled,
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: gin.H{
				"error": transportHttp.ErrMissingIdentity.Error(),
			},
		},
		{
			name:      "404 Not found: Tenant not found",
			method:    "DELETE",
			url:       validUrl,
			inputDto:  validPayload,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockDeleteTenantAdminUseCase]{
				useCaseTenantNotFound,
			},
			expectedStatus: http.StatusNotFound,
			expectedResponse: gin.H{
				"error": tenant.ErrTenantNotFound.Error(),
			},
		},
		{
			name:      "404 Not found: Unauthorized access (obfuscated)",
			method:    "DELETE",
			url:       validUrl,
			inputDto:  validPayload,
			requester: superAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockDeleteTenantAdminUseCase]{
				useCaseUnauthorizedAccess,
			},
			expectedStatus: http.StatusNotFound,
			expectedResponse: gin.H{
				"error": tenant.ErrTenantNotFound.Error(),
			},
		},
		{
			name:      "404 Not found: User not found",
			method:    "DELETE",
			url:       validUrl,
			inputDto:  validPayload,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockDeleteTenantAdminUseCase]{
				useCaseNotFound,
			},
			expectedStatus: http.StatusNotFound,
			expectedResponse: gin.H{
				"error": user.ErrUserNotFound.Error(),
			},
		},
		{
			name:      "500 Server Error: Unexpected error",
			method:    "DELETE",
			url:       validUrl,
			inputDto:  validPayload,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockDeleteTenantAdminUseCase]{
				useCaseUnexpectedErr,
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResponse: gin.H{
				"error": errMock.Error(),
			},
		},
	}

	// Parametri di test
	mountMethod := "DELETE"
	mountUrl := "/tenant/:tenant_id/tenant_admin/:user_id"

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockUseCase := setupMockUseCase(
				mocks.NewMockDeleteTenantAdminUseCase,
				tc.setupSteps, t,
			)
			userController := user.NewUserController(
				nil, nil, nil, nil, nil, mockUseCase, nil, nil, nil, nil, nil, nil, nil,
			)

			executeControllerTest(
				t, tc,
				mountMethod, mountUrl,
				userController.DeleteTenantAdmin,
			)
		})
	}
}

func TestController_DeleteSuperAdmin(t *testing.T) {
	// Data
	// targetTenantId := uuid.New()

	targetUserEmail := "test@example.com"
	targetUserName := "Test"
	targetUserId := uint(1)
	targetUserRole := identity.ROLE_SUPER_ADMIN
	targetConfirmed := false

	expectedResponse := user.UserResponseDTO{}
	expectedResponse.UserId = targetUserId
	expectedResponse.Username = targetUserName
	expectedResponse.Email = targetUserEmail
	expectedResponse.UserRole = string(targetUserRole)
	expectedResponse.TenantId = ""
	// expectedResponse.TenantId = targetTenantId.String()

	expectedUser := user.User{
		Id:       targetUserId,
		Name:     targetUserName,
		Email:    targetUserEmail,
		TenantId: nil,
		// TenantId:  &targetTenantId,
		Confirmed: targetConfirmed,
		Role:      targetUserRole,
	}

	// Setup -----
	useCaseOk := func(mockUC *mocks.MockDeleteSuperAdminUseCase) *gomock.Call {
		return mockUC.EXPECT().
			DeleteSuperAdmin(gomock.Any()).
			Return(expectedUser, nil).
			Times(1)
	}

	useCaseNeverCalled := func(mockUC *mocks.MockDeleteSuperAdminUseCase) *gomock.Call {
		return mockUC.EXPECT().
			DeleteSuperAdmin(gomock.Any()).
			Times(0)
	}

	useCaseNotFound := func(mockUC *mocks.MockDeleteSuperAdminUseCase) *gomock.Call {
		return mockUC.EXPECT().
			DeleteSuperAdmin(gomock.Any()).
			Return(user.User{}, user.ErrUserNotFound).
			Times(1)
	}

	useCaseUnauthorizedAccess := func(mockUC *mocks.MockDeleteSuperAdminUseCase) *gomock.Call {
		return mockUC.EXPECT().
			DeleteSuperAdmin(gomock.Any()).
			Return(user.User{}, identity.ErrUnauthorizedAccess).
			Times(1)
	}

	errMock := errors.New("unexpected error")
	useCaseUnexpectedErr := func(mockUC *mocks.MockDeleteSuperAdminUseCase) *gomock.Call {
		return mockUC.EXPECT().
			DeleteSuperAdmin(gomock.Any()).
			Return(user.User{}, errMock).
			Times(1)
	}

	// Requester
	superAdminRequester := identity.Requester{
		RequesterUserId: uint(1),
		RequesterRole:   identity.ROLE_SUPER_ADMIN,
	}

	tenantId := uuid.New()
	tenantAdminRequester := identity.Requester{
		RequesterUserId:   uint(1),
		RequesterTenantId: &tenantId,
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	// Input
	type Empty struct{}
	validPayload := Empty{}

	validUrl := fmt.Sprintf("/super_admin/%v", targetUserId)
	invalidUrl_userId := "/super_admin/0"

	cases := []genericControllerTestCase[Empty, mocks.MockDeleteSuperAdminUseCase]{
		{
			name:      "200 OK",
			method:    "DELETE",
			url:       validUrl,
			inputDto:  validPayload,
			requester: superAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockDeleteSuperAdminUseCase]{
				useCaseOk,
			},
			expectedStatus:   http.StatusOK,
			expectedResponse: expectedResponse,
		},
		{
			name:      "400 Bad Request: Bad URI (user Id)",
			method:    "DELETE",
			url:       invalidUrl_userId,
			inputDto:  validPayload,
			requester: superAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockDeleteSuperAdminUseCase]{
				useCaseNeverCalled,
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: gin.H{
				"error": "invalid format",
				"fields": gin.H{
					"user_id": "required",
				},
			},
		},
		{
			name:         "401 Unauthorized: No identity",
			method:       "DELETE",
			url:          validUrl,
			inputDto:     validPayload,
			requester:    superAdminRequester,
			omitIdentity: true,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockDeleteSuperAdminUseCase]{
				useCaseNeverCalled,
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: gin.H{
				"error": transportHttp.ErrMissingIdentity.Error(),
			},
		},
		{
			name:      "404 Not found: Unauthorized access (obfuscated)",
			method:    "DELETE",
			url:       validUrl,
			inputDto:  validPayload,
			requester: tenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockDeleteSuperAdminUseCase]{
				useCaseUnauthorizedAccess,
			},
			expectedStatus: http.StatusNotFound,
			expectedResponse: gin.H{
				"error": user.ErrUserNotFound.Error(),
			},
		},
		{
			name:      "404 Not found: User not found",
			method:    "DELETE",
			url:       validUrl,
			inputDto:  validPayload,
			requester: superAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockDeleteSuperAdminUseCase]{
				useCaseNotFound,
			},
			expectedStatus: http.StatusNotFound,
			expectedResponse: gin.H{
				"error": user.ErrUserNotFound.Error(),
			},
		},
		{
			name:      "500 Server Error: Unexpected error",
			method:    "DELETE",
			url:       validUrl,
			inputDto:  validPayload,
			requester: superAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockDeleteSuperAdminUseCase]{
				useCaseUnexpectedErr,
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResponse: gin.H{
				"error": errMock.Error(),
			},
		},
	}

	// Parametri di test
	mountMethod := "DELETE"
	mountUrl := "/super_admin/:user_id"

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockUseCase := setupMockUseCase(
				mocks.NewMockDeleteSuperAdminUseCase,
				tc.setupSteps, t,
			)
			userController := user.NewUserController(
				nil,
				nil, nil, nil,
				nil, nil, mockUseCase,
				nil, nil, nil,
				nil, nil, nil,
			)

			executeControllerTest(
				t, tc,
				mountMethod, mountUrl,
				userController.DeleteSuperAdmin,
			)
		})
	}
}

// GET SINGOLO ==============================================================================================================

func TestController_GetTenantUser(t *testing.T) {
	// Data
	targetTenantId := uuid.New()

	targetUserEmail := "test@example.com"
	targetUserName := "Test"
	targetUserId := uint(1)
	targetUserRole := identity.ROLE_TENANT_USER
	targetConfirmed := false

	expectedResponse := user.UserResponseDTO{}
	expectedResponse.UserId = targetUserId
	expectedResponse.Username = targetUserName
	expectedResponse.Email = targetUserEmail
	expectedResponse.UserRole = string(targetUserRole)
	expectedResponse.TenantId = targetTenantId.String()

	expectedUser := user.User{
		Id:        targetUserId,
		Name:      targetUserName,
		Email:     targetUserEmail,
		TenantId:  &targetTenantId,
		Confirmed: targetConfirmed,
		Role:      targetUserRole,
	}

	// Setup -----
	useCaseOk := func(mockUC *mocks.MockGetTenantUserUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetTenantUser(gomock.Any()).
			Return(expectedUser, nil).
			Times(1)
	}

	useCaseNeverCalled := func(mockUC *mocks.MockGetTenantUserUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetTenantUser(gomock.Any()).
			Times(0)
	}

	useCaseTenantNotFound := func(mockUC *mocks.MockGetTenantUserUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetTenantUser(gomock.Any()).
			Return(user.User{}, tenant.ErrTenantNotFound).
			Times(1)
	}

	useCaseNotFound := func(mockUC *mocks.MockGetTenantUserUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetTenantUser(gomock.Any()).
			Return(user.User{}, user.ErrUserNotFound).
			Times(1)
	}

	useCaseUnauthorizedAccess := func(mockUC *mocks.MockGetTenantUserUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetTenantUser(gomock.Any()).
			Return(user.User{}, identity.ErrUnauthorizedAccess).
			Times(1)
	}

	errMock := errors.New("unexpected error")
	useCaseUnexpectedErr := func(mockUC *mocks.MockGetTenantUserUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetTenantUser(gomock.Any()).
			Return(user.User{}, errMock).
			Times(1)
	}

	// Requester
	superAdminRequester := identity.Requester{
		RequesterUserId: uint(1),
		RequesterRole:   identity.ROLE_SUPER_ADMIN,
	}

	authTenantAdminRequester := identity.Requester{
		RequesterUserId:   uint(1),
		RequesterTenantId: &targetTenantId,
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	// Input
	validUrl := fmt.Sprintf("/tenant/%v/tenant_user/%v", targetTenantId.String(), targetUserId)
	invalidUrl_tenantId := "/tenant/123/tenant_user/1"
	invalidUrl_userId := fmt.Sprintf("/tenant/%v/tenant_user/0", targetTenantId.String())

	cases := []genericControllerTestCase[any, mocks.MockGetTenantUserUseCase]{
		{
			name:      "200 OK",
			method:    "GET",
			url:       validUrl,
			inputDto:  nil,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetTenantUserUseCase]{
				useCaseOk,
			},
			expectedStatus:   http.StatusOK,
			expectedResponse: expectedResponse,
		},
		{
			name:      "400 Bad Request: Bad URI (tenant Id)",
			method:    "GET",
			url:       invalidUrl_tenantId,
			inputDto:  nil,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetTenantUserUseCase]{
				useCaseNeverCalled,
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: gin.H{
				"error": "invalid format",
				"fields": gin.H{
					"tenant_id": "uuid4",
				},
			},
		},
		{
			name:      "400 Bad Request: Bad URI (user Id)",
			method:    "GET",
			url:       invalidUrl_userId,
			inputDto:  nil,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetTenantUserUseCase]{
				useCaseNeverCalled,
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: gin.H{
				"error": "invalid format",
				"fields": gin.H{
					"user_id": "required",
				},
			},
		},
		{
			name:         "401 Unauthorized: No identity",
			method:       "GET",
			url:          validUrl,
			inputDto:     nil,
			requester:    authTenantAdminRequester,
			omitIdentity: true,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetTenantUserUseCase]{
				useCaseNeverCalled,
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: gin.H{
				"error": transportHttp.ErrMissingIdentity.Error(),
			},
		},
		{
			name:      "404 Not found: Tenant not found",
			method:    "GET",
			url:       validUrl,
			inputDto:  nil,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetTenantUserUseCase]{
				useCaseTenantNotFound,
			},
			expectedStatus: http.StatusNotFound,
			expectedResponse: gin.H{
				"error": tenant.ErrTenantNotFound.Error(),
			},
		},
		{
			name:      "404 Not found: Unauthorized access (obfuscated)",
			method:    "GET",
			url:       validUrl,
			inputDto:  nil,
			requester: superAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetTenantUserUseCase]{
				useCaseUnauthorizedAccess,
			},
			expectedStatus: http.StatusNotFound,
			expectedResponse: gin.H{
				"error": tenant.ErrTenantNotFound.Error(),
			},
		},
		{
			name:      "404 Not found: User not found",
			method:    "GET",
			url:       validUrl,
			inputDto:  nil,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetTenantUserUseCase]{
				useCaseNotFound,
			},
			expectedStatus: http.StatusNotFound,
			expectedResponse: gin.H{
				"error": user.ErrUserNotFound.Error(),
			},
		},
		{
			name:      "500 Server Error: Unexpected error",
			method:    "GET",
			url:       validUrl,
			inputDto:  nil,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetTenantUserUseCase]{
				useCaseUnexpectedErr,
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResponse: gin.H{
				"error": errMock.Error(),
			},
		},
	}

	// Parametri di test
	mountMethod := "GET"
	mountUrl := "/tenant/:tenant_id/tenant_user/:user_id"

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockUseCase := setupMockUseCase(
				mocks.NewMockGetTenantUserUseCase,
				tc.setupSteps, t,
			)
			userController := user.NewUserController(
				nil,
				nil, nil, nil,
				nil, nil, nil,
				mockUseCase, nil, nil,
				nil, nil, nil,
			)

			executeControllerTest(
				t, tc,
				mountMethod, mountUrl,
				userController.GetTenantUser,
			)
		})
	}
}

func TestController_GetTenantAdmin(t *testing.T) {
	// Data
	targetTenantId := uuid.New()

	targetUserEmail := "test@example.com"
	targetUserName := "Test"
	targetUserId := uint(1)
	targetUserRole := identity.ROLE_TENANT_ADMIN
	targetConfirmed := false

	expectedResponse := user.UserResponseDTO{}
	expectedResponse.UserId = targetUserId
	expectedResponse.Username = targetUserName
	expectedResponse.Email = targetUserEmail
	expectedResponse.UserRole = string(targetUserRole)
	expectedResponse.TenantId = targetTenantId.String()

	expectedUser := user.User{
		Id:        targetUserId,
		Name:      targetUserName,
		Email:     targetUserEmail,
		TenantId:  &targetTenantId,
		Confirmed: targetConfirmed,
		Role:      targetUserRole,
	}

	// Setup -----
	useCaseOk := func(mockUC *mocks.MockGetTenantAdminUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetTenantAdmin(gomock.Any()).
			Return(expectedUser, nil).
			Times(1)
	}

	useCaseNeverCalled := func(mockUC *mocks.MockGetTenantAdminUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetTenantAdmin(gomock.Any()).
			Times(0)
	}

	useCaseTenantNotFound := func(mockUC *mocks.MockGetTenantAdminUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetTenantAdmin(gomock.Any()).
			Return(user.User{}, tenant.ErrTenantNotFound).
			Times(1)
	}

	useCaseNotFound := func(mockUC *mocks.MockGetTenantAdminUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetTenantAdmin(gomock.Any()).
			Return(user.User{}, user.ErrUserNotFound).
			Times(1)
	}

	useCaseUnauthorizedAccess := func(mockUC *mocks.MockGetTenantAdminUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetTenantAdmin(gomock.Any()).
			Return(user.User{}, identity.ErrUnauthorizedAccess).
			Times(1)
	}

	errMock := errors.New("unexpected error")
	useCaseUnexpectedErr := func(mockUC *mocks.MockGetTenantAdminUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetTenantAdmin(gomock.Any()).
			Return(user.User{}, errMock).
			Times(1)
	}

	// Requester
	superAdminRequester := identity.Requester{
		RequesterUserId: uint(1),
		RequesterRole:   identity.ROLE_SUPER_ADMIN,
	}

	authTenantAdminRequester := identity.Requester{
		RequesterUserId:   uint(1),
		RequesterTenantId: &targetTenantId,
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	// Input

	validUrl := fmt.Sprintf("/tenant/%v/tenant_admin/%v", targetTenantId.String(), targetUserId)
	invalidUrl_tenantId := "/tenant/123/tenant_admin/1"
	invalidUrl_userId := fmt.Sprintf("/tenant/%v/tenant_admin/0", targetTenantId.String())

	cases := []genericControllerTestCase[any, mocks.MockGetTenantAdminUseCase]{
		{
			name:      "200 OK",
			method:    "GET",
			url:       validUrl,
			inputDto:  nil,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetTenantAdminUseCase]{
				useCaseOk,
			},
			expectedStatus:   http.StatusOK,
			expectedResponse: expectedResponse,
		},
		{
			name:      "400 Bad Request: Bad URI (tenant Id)",
			method:    "GET",
			url:       invalidUrl_tenantId,
			inputDto:  nil,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetTenantAdminUseCase]{
				useCaseNeverCalled,
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: gin.H{
				"error": "invalid format",
				"fields": gin.H{
					"tenant_id": "uuid4",
				},
			},
		},
		{
			name:      "400 Bad Request: Bad URI (user Id)",
			method:    "GET",
			url:       invalidUrl_userId,
			inputDto:  nil,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetTenantAdminUseCase]{
				useCaseNeverCalled,
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: gin.H{
				"error": "invalid format",
				"fields": gin.H{
					"user_id": "required",
				},
			},
		},
		{
			name:         "401 Unauthorized: No identity",
			method:       "GET",
			url:          validUrl,
			inputDto:     nil,
			requester:    authTenantAdminRequester,
			omitIdentity: true,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetTenantAdminUseCase]{
				useCaseNeverCalled,
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: gin.H{
				"error": transportHttp.ErrMissingIdentity.Error(),
			},
		},
		{
			name:      "404 Not found: Tenant not found",
			method:    "GET",
			url:       validUrl,
			inputDto:  nil,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetTenantAdminUseCase]{
				useCaseTenantNotFound,
			},
			expectedStatus: http.StatusNotFound,
			expectedResponse: gin.H{
				"error": tenant.ErrTenantNotFound.Error(),
			},
		},
		{
			name:      "404 Not found: Unauthorized access (obfuscated)",
			method:    "GET",
			url:       validUrl,
			inputDto:  nil,
			requester: superAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetTenantAdminUseCase]{
				useCaseUnauthorizedAccess,
			},
			expectedStatus: http.StatusNotFound,
			expectedResponse: gin.H{
				"error": tenant.ErrTenantNotFound.Error(),
			},
		},
		{
			name:      "404 Not found: User not found",
			method:    "GET",
			url:       validUrl,
			inputDto:  nil,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetTenantAdminUseCase]{
				useCaseNotFound,
			},
			expectedStatus: http.StatusNotFound,
			expectedResponse: gin.H{
				"error": user.ErrUserNotFound.Error(),
			},
		},
		{
			name:      "500 Server Error: Unexpected error",
			method:    "GET",
			url:       validUrl,
			inputDto:  nil,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetTenantAdminUseCase]{
				useCaseUnexpectedErr,
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResponse: gin.H{
				"error": errMock.Error(),
			},
		},
	}

	// Parametri di test
	mountMethod := "GET"
	mountUrl := "/tenant/:tenant_id/tenant_admin/:user_id"

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockUseCase := setupMockUseCase(
				mocks.NewMockGetTenantAdminUseCase,
				tc.setupSteps, t,
			)
			userController := user.NewUserController(
				nil,
				nil, nil, nil,
				nil, nil, nil,
				nil, mockUseCase, nil,
				nil, nil, nil,
			)

			executeControllerTest(
				t, tc,
				mountMethod, mountUrl,
				userController.GetTenantAdmin,
			)
		})
	}
}

func TestController_GetSuperAdmin(t *testing.T) {
	// Data
	// targetTenantId := uuid.New()

	targetUserEmail := "test@example.com"
	targetUserName := "Test"
	targetUserId := uint(1)
	targetUserRole := identity.ROLE_SUPER_ADMIN
	targetConfirmed := false

	expectedResponse := user.UserResponseDTO{}
	expectedResponse.UserId = targetUserId
	expectedResponse.Username = targetUserName
	expectedResponse.Email = targetUserEmail
	expectedResponse.UserRole = string(targetUserRole)
	expectedResponse.TenantId = ""
	// expectedResponse.TenantId = targetTenantId.String()

	expectedUser := user.User{
		Id:       targetUserId,
		Name:     targetUserName,
		Email:    targetUserEmail,
		TenantId: nil,
		// TenantId:  &targetTenantId,
		Confirmed: targetConfirmed,
		Role:      targetUserRole,
	}

	// Setup -----
	useCaseOk := func(mockUC *mocks.MockGetSuperAdminUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetSuperAdmin(gomock.Any()).
			Return(expectedUser, nil).
			Times(1)
	}

	useCaseNeverCalled := func(mockUC *mocks.MockGetSuperAdminUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetSuperAdmin(gomock.Any()).
			Times(0)
	}

	useCaseNotFound := func(mockUC *mocks.MockGetSuperAdminUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetSuperAdmin(gomock.Any()).
			Return(user.User{}, user.ErrUserNotFound).
			Times(1)
	}

	useCaseUnauthorizedAccess := func(mockUC *mocks.MockGetSuperAdminUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetSuperAdmin(gomock.Any()).
			Return(user.User{}, identity.ErrUnauthorizedAccess).
			Times(1)
	}

	errMock := errors.New("unexpected error")
	useCaseUnexpectedErr := func(mockUC *mocks.MockGetSuperAdminUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetSuperAdmin(gomock.Any()).
			Return(user.User{}, errMock).
			Times(1)
	}

	// Requester
	superAdminRequester := identity.Requester{
		RequesterUserId: uint(1),
		RequesterRole:   identity.ROLE_SUPER_ADMIN,
	}

	tenantId := uuid.New()
	tenantAdminRequester := identity.Requester{
		RequesterUserId:   uint(1),
		RequesterTenantId: &tenantId,
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	// Input
	validUrl := fmt.Sprintf("/super_admin/%v", targetUserId)
	invalidUrl_userId := "/super_admin/0"

	cases := []genericControllerTestCase[any, mocks.MockGetSuperAdminUseCase]{
		{
			name:      "200 OK",
			method:    "GET",
			url:       validUrl,
			inputDto:  nil,
			requester: superAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetSuperAdminUseCase]{
				useCaseOk,
			},
			expectedStatus:   http.StatusOK,
			expectedResponse: expectedResponse,
		},
		{
			name:      "400 Bad Request: Bad URI (user Id)",
			method:    "GET",
			url:       invalidUrl_userId,
			inputDto:  nil,
			requester: superAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetSuperAdminUseCase]{
				useCaseNeverCalled,
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: gin.H{
				"error": "invalid format",
				"fields": gin.H{
					"user_id": "required",
				},
			},
		},
		{
			name:         "401 Unauthorized: No identity",
			method:       "GET",
			url:          validUrl,
			inputDto:     nil,
			requester:    superAdminRequester,
			omitIdentity: true,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetSuperAdminUseCase]{
				useCaseNeverCalled,
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: gin.H{
				"error": transportHttp.ErrMissingIdentity.Error(),
			},
		},
		{
			name:      "404 Not found: Unauthorized access (obfuscated)",
			method:    "GET",
			url:       validUrl,
			inputDto:  nil,
			requester: tenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetSuperAdminUseCase]{
				useCaseUnauthorizedAccess,
			},
			expectedStatus: http.StatusNotFound,
			expectedResponse: gin.H{
				"error": user.ErrUserNotFound.Error(),
			},
		},
		{
			name:      "404 Not found: User not found",
			method:    "GET",
			url:       validUrl,
			inputDto:  nil,
			requester: superAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetSuperAdminUseCase]{
				useCaseNotFound,
			},
			expectedStatus: http.StatusNotFound,
			expectedResponse: gin.H{
				"error": user.ErrUserNotFound.Error(),
			},
		},
		{
			name:      "500 Server Error: Unexpected error",
			method:    "GET",
			url:       validUrl,
			inputDto:  nil,
			requester: superAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetSuperAdminUseCase]{
				useCaseUnexpectedErr,
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResponse: gin.H{
				"error": errMock.Error(),
			},
		},
	}

	// Parametri di test
	mountMethod := "GET"
	mountUrl := "/super_admin/:user_id"

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockUseCase := setupMockUseCase(
				mocks.NewMockGetSuperAdminUseCase,
				tc.setupSteps, t,
			)
			userController := user.NewUserController(
				nil,
				nil, nil, nil,
				nil, nil, nil,
				nil, nil, mockUseCase,
				nil, nil, nil,
			)

			executeControllerTest(
				t, tc,
				mountMethod, mountUrl,
				userController.GetSuperAdmin,
			)
		})
	}
}

// GET MULTIPLO ==============================================================================================================

func TestController_GetTenantUsers(t *testing.T) {
	// Data
	targetTenantId := uuid.New()

	targetUserEmail := "test@example.com"
	targetUsername := "Test"
	targetUserId := uint(1)
	targetUserRole := identity.ROLE_TENANT_USER
	targetConfirmed := true

	expectedDto := user.UserResponseDTO{}
	expectedDto.UserId = targetUserId
	expectedDto.Username = targetUsername
	expectedDto.Email = targetUserEmail
	expectedDto.UserRole = string(targetUserRole)
	expectedDto.TenantId = targetTenantId.String()

	expectedResponse := user.UserListResponseDTO{
		ListInfo: dto.ListInfo{
			Count: uint(1),
			Total: uint(1),
		},
		Users: []user.UserResponseDTO{expectedDto},
		// expectedDto,
	}
	expectedResponseEmpty := user.UserListResponseDTO{
		ListInfo: dto.ListInfo{
			Count: uint(0),
			Total: uint(0),
		},
		Users: make([]user.UserResponseDTO, 0),
		// expectedDto,
	}

	expectedUser := user.User{
		Id:        targetUserId,
		Name:      targetUsername,
		Email:     targetUserEmail,
		TenantId:  &targetTenantId,
		Confirmed: targetConfirmed,
		Role:      targetUserRole,
	}

	emptySlice := make([]user.User, 0)
	// Setup -----
	useCaseOk := func(mockUC *mocks.MockGetTenantUsersByTenantUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetTenantUsersByTenant(gomock.Any()).
			Return([]user.User{expectedUser,}, uint(1), nil).
			Times(1)
	}

	useCaseOkEmpty := func(mockUC *mocks.MockGetTenantUsersByTenantUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetTenantUsersByTenant(gomock.Any()).
			Return(emptySlice, uint(0), nil).
			Times(1)
	}


	useCaseNeverCalled := func(mockUC *mocks.MockGetTenantUsersByTenantUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetTenantUsersByTenant(gomock.Any()).
			Times(0)
	}

	useCaseTenantNotFound := func(mockUC *mocks.MockGetTenantUsersByTenantUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetTenantUsersByTenant(gomock.Any()).
			Return(emptySlice, uint(0), tenant.ErrTenantNotFound).
			Times(1)
	}

	useCaseUnauthorizedAccess := func(mockUC *mocks.MockGetTenantUsersByTenantUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetTenantUsersByTenant(gomock.Any()).
			Return(emptySlice, uint(0), identity.ErrUnauthorizedAccess).
			Times(1)
	}

	errMock := errors.New("unexpected error")
	useCaseUnexpectedErr := func(mockUC *mocks.MockGetTenantUsersByTenantUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetTenantUsersByTenant(gomock.Any()).
			Return(emptySlice, uint(0), errMock).
			Times(1)
	}

	// Requester
	superAdminRequester := identity.Requester{
		RequesterUserId: uint(1),
		RequesterRole:   identity.ROLE_SUPER_ADMIN,
	}

	authTenantAdminRequester := identity.Requester{
		RequesterUserId:   uint(1),
		RequesterTenantId: &targetTenantId,
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	// Input

	validUrl := fmt.Sprintf("/tenant/%v/tenant_users", targetTenantId.String())
	invalidUrl_tenantId := "/tenant/123/tenant_users"
	invalidUrl_pagination := validUrl + "?page=invalid&limit=invalid"
	// invalidUrl_userId := fmt.Sprintf("/tenant/%v/tenant_users/", targetTenantId.String())

	cases := []genericControllerTestCase[any, mocks.MockGetTenantUsersByTenantUseCase]{
		{
			name:      "200 OK: Populated list",
			method:    "GET",
			url:       validUrl,
			inputDto:  nil,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetTenantUsersByTenantUseCase]{
				useCaseOk,
			},
			expectedStatus:   http.StatusOK,
			expectedResponse: expectedResponse,
		},
		{
			name:      "200 OK: Empty list",
			method:    "GET",
			url:       validUrl,
			inputDto:  nil,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetTenantUsersByTenantUseCase]{
				useCaseOkEmpty,
			},
			expectedStatus:   http.StatusOK,
			expectedResponse: expectedResponseEmpty,
		},
		{
			name:      "400 Bad Request: Bad URI (tenant Id)",
			method:    "GET",
			url:       invalidUrl_tenantId,
			inputDto:  nil,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetTenantUsersByTenantUseCase]{
				useCaseNeverCalled,
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: gin.H{
				"error": "invalid format",
				"fields": gin.H{
					"tenant_id": "uuid4",
				},
			},
		},
		{
			name:      "400 Bad Request: Invalid page parameters",
			method:    "GET",
			url:       invalidUrl_pagination,
			inputDto:  nil,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetTenantUsersByTenantUseCase]{
				useCaseNeverCalled,
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: hasError{},
		},
		{
			name:         "401 Unauthorized: No identity",
			method:       "GET",
			url:          validUrl,
			inputDto:     nil,
			requester:    authTenantAdminRequester,
			omitIdentity: true,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetTenantUsersByTenantUseCase]{
				useCaseNeverCalled,
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: gin.H{
				"error": transportHttp.ErrMissingIdentity.Error(),
			},
		},
		{
			name:      "404 Not found: Tenant not found",
			method:    "GET",
			url:       validUrl,
			inputDto:  nil,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetTenantUsersByTenantUseCase]{
				useCaseTenantNotFound,
			},
			expectedStatus: http.StatusNotFound,
			expectedResponse: gin.H{
				"error": tenant.ErrTenantNotFound.Error(),
			},
		},
		{
			name:      "404 Not found: Unauthorized access (obfuscated)",
			method:    "GET",
			url:       validUrl,
			inputDto:  nil,
			requester: superAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetTenantUsersByTenantUseCase]{
				useCaseUnauthorizedAccess,
			},
			expectedStatus: http.StatusNotFound,
			expectedResponse: gin.H{
				"error": tenant.ErrTenantNotFound.Error(),
			},
		},
		{
			name:      "500 Server Error: Unexpected error",
			method:    "GET",
			url:       validUrl,
			inputDto:  nil,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetTenantUsersByTenantUseCase]{
				useCaseUnexpectedErr,
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResponse: gin.H{
				"error": errMock.Error(),
			},
		},
	}

	// Parametri di test
	mountMethod := "GET"
	mountUrl := "/tenant/:tenant_id/tenant_users"

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockUseCase := setupMockUseCase(
				mocks.NewMockGetTenantUsersByTenantUseCase,
				tc.setupSteps, t,
			)
			userController := user.NewUserController(
				nil,
				nil, nil, nil,
				nil, nil, nil,
				nil, nil, nil,
				mockUseCase, nil, nil,
			)

			executeControllerTest(
				t, tc,
				mountMethod, mountUrl,
				userController.GetTenantUsers,
			)
		})
	}
}

func TestController_GetTenantAdmins(t *testing.T) {
	// Data
	targetTenantId := uuid.New()

	targetUserEmail := "test@example.com"
	targetUsername := "Test"
	targetUserId := uint(1)
	targetUserRole := identity.ROLE_TENANT_ADMIN
	targetConfirmed := true

	expectedDto := user.UserResponseDTO{}
	expectedDto.UserId = targetUserId
	expectedDto.Username = targetUsername
	expectedDto.Email = targetUserEmail
	expectedDto.UserRole = string(targetUserRole)
	expectedDto.TenantId = targetTenantId.String()

	expectedResponse := user.UserListResponseDTO{
		ListInfo: dto.ListInfo{
			Count: uint(1),
			Total: uint(1),
		},
		Users: []user.UserResponseDTO{expectedDto},
		// expectedDto,
	}
	expectedResponseEmpty := user.UserListResponseDTO{
		ListInfo: dto.ListInfo{
			Count: uint(0),
			Total: uint(0),
		},
		Users: make([]user.UserResponseDTO, 0),
		// expectedDto,
	}

	expectedUser := user.User{
		Id:        targetUserId,
		Name:      targetUsername,
		Email:     targetUserEmail,
		TenantId:  &targetTenantId,
		Confirmed: targetConfirmed,
		Role:      targetUserRole,
	}

	emptySlice := make([]user.User, 0)
	// Setup -----
	useCaseOk := func(mockUC *mocks.MockGetTenantAdminsByTenantUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetTenantAdminsByTenant(gomock.Any()).
			Return([]user.User{expectedUser,}, uint(1), nil).
			Times(1)
	}

	useCaseOkEmpty := func(mockUC *mocks.MockGetTenantAdminsByTenantUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetTenantAdminsByTenant(gomock.Any()).
			Return(emptySlice, uint(0), nil).
			Times(1)
	}


	useCaseNeverCalled := func(mockUC *mocks.MockGetTenantAdminsByTenantUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetTenantAdminsByTenant(gomock.Any()).
			Times(0)
	}

	useCaseTenantNotFound := func(mockUC *mocks.MockGetTenantAdminsByTenantUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetTenantAdminsByTenant(gomock.Any()).
			Return(emptySlice, uint(0), tenant.ErrTenantNotFound).
			Times(1)
	}

	useCaseUnauthorizedAccess := func(mockUC *mocks.MockGetTenantAdminsByTenantUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetTenantAdminsByTenant(gomock.Any()).
			Return(emptySlice, uint(0), identity.ErrUnauthorizedAccess).
			Times(1)
	}

	errMock := errors.New("unexpected error")
	useCaseUnexpectedErr := func(mockUC *mocks.MockGetTenantAdminsByTenantUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetTenantAdminsByTenant(gomock.Any()).
			Return(emptySlice, uint(0), errMock).
			Times(1)
	}

	// Requester
	superAdminRequester := identity.Requester{
		RequesterUserId: uint(1),
		RequesterRole:   identity.ROLE_SUPER_ADMIN,
	}

	authTenantAdminRequester := identity.Requester{
		RequesterUserId:   uint(1),
		RequesterTenantId: &targetTenantId,
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	// Input

	validUrl := fmt.Sprintf("/tenant/%v/tenant_users", targetTenantId.String())
	invalidUrl_tenantId := "/tenant/123/tenant_users"
	invalidUrl_pagination := validUrl + "?page=invalid&limit=invalid"
	// invalidUrl_userId := fmt.Sprintf("/tenant/%v/tenant_users/", targetTenantId.String())

	cases := []genericControllerTestCase[any, mocks.MockGetTenantAdminsByTenantUseCase]{
		{
			name:      "200 OK: Populated list",
			method:    "GET",
			url:       validUrl,
			inputDto:  nil,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetTenantAdminsByTenantUseCase]{
				useCaseOk,
			},
			expectedStatus:   http.StatusOK,
			expectedResponse: expectedResponse,
		},
		{
			name:      "200 OK: Empty list",
			method:    "GET",
			url:       validUrl,
			inputDto:  nil,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetTenantAdminsByTenantUseCase]{
				useCaseOkEmpty,
			},
			expectedStatus:   http.StatusOK,
			expectedResponse: expectedResponseEmpty,
		},
		{
			name:      "400 Bad Request: Bad URI (tenant Id)",
			method:    "GET",
			url:       invalidUrl_tenantId,
			inputDto:  nil,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetTenantAdminsByTenantUseCase]{
				useCaseNeverCalled,
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: gin.H{
				"error": "invalid format",
				"fields": gin.H{
					"tenant_id": "uuid4",
				},
			},
		},
		{
			name:      "400 Bad Request: Invalid page parameters",
			method:    "GET",
			url:       invalidUrl_pagination,
			inputDto:  nil,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetTenantAdminsByTenantUseCase]{
				useCaseNeverCalled,
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: hasError{},
		},
		{
			name:         "401 Unauthorized: No identity",
			method:       "GET",
			url:          validUrl,
			inputDto:     nil,
			requester:    authTenantAdminRequester,
			omitIdentity: true,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetTenantAdminsByTenantUseCase]{
				useCaseNeverCalled,
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: gin.H{
				"error": transportHttp.ErrMissingIdentity.Error(),
			},
		},
		{
			name:      "404 Not found: Tenant not found",
			method:    "GET",
			url:       validUrl,
			inputDto:  nil,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetTenantAdminsByTenantUseCase]{
				useCaseTenantNotFound,
			},
			expectedStatus: http.StatusNotFound,
			expectedResponse: gin.H{
				"error": tenant.ErrTenantNotFound.Error(),
			},
		},
		{
			name:      "404 Not found: Unauthorized access (obfuscated)",
			method:    "GET",
			url:       validUrl,
			inputDto:  nil,
			requester: superAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetTenantAdminsByTenantUseCase]{
				useCaseUnauthorizedAccess,
			},
			expectedStatus: http.StatusNotFound,
			expectedResponse: gin.H{
				"error": tenant.ErrTenantNotFound.Error(),
			},
		},
		{
			name:      "500 Server Error: Unexpected error",
			method:    "GET",
			url:       validUrl,
			inputDto:  nil,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetTenantAdminsByTenantUseCase]{
				useCaseUnexpectedErr,
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResponse: gin.H{
				"error": errMock.Error(),
			},
		},
	}

	// Parametri di test
	mountMethod := "GET"
	mountUrl := "/tenant/:tenant_id/tenant_users"

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockUseCase := setupMockUseCase(
				mocks.NewMockGetTenantAdminsByTenantUseCase,
				tc.setupSteps, t,
			)
			userController := user.NewUserController(
				nil,
				nil, nil, nil,
				nil, nil, nil,
				nil, nil, nil,
				nil, mockUseCase, nil, 
			)

			executeControllerTest(
				t, tc,
				mountMethod, mountUrl,
				userController.GetTenantAdmins,
			)
		})
	}
}

func TestController_GetSuperAdmins(t *testing.T) {
	// Data
	// targetTenantId := uuid.New()

	targetUserEmail := "test@example.com"
	targetUsername := "Test"
	targetUserId := uint(1)
	targetUserRole := identity.ROLE_SUPER_ADMIN
	targetConfirmed := true

	expectedDto := user.UserResponseDTO{}
	expectedDto.UserId = targetUserId
	expectedDto.Username = targetUsername
	expectedDto.Email = targetUserEmail
	expectedDto.UserRole = string(targetUserRole)
	// expectedDto.TenantId = targetTenantId.String()

	expectedResponse := user.UserListResponseDTO{
		ListInfo: dto.ListInfo{
			Count: uint(1),
			Total: uint(1),
		},
		Users: []user.UserResponseDTO{expectedDto},
		// expectedDto,
	}
	expectedResponseEmpty := user.UserListResponseDTO{
		ListInfo: dto.ListInfo{
			Count: uint(0),
			Total: uint(0),
		},
		Users: make([]user.UserResponseDTO, 0),
		// expectedDto,
	}

	expectedUser := user.User{
		Id:        targetUserId,
		Name:      targetUsername,
		Email:     targetUserEmail,
		// TenantId:  &targetTenantId,
		Confirmed: targetConfirmed,
		Role:      targetUserRole,
	}

	emptySlice := make([]user.User, 0)
	// Setup -----
	useCaseOk := func(mockUC *mocks.MockGetSuperAdminListUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetSuperAdminList(gomock.Any()).
			Return([]user.User{expectedUser,}, uint(1), nil).
			Times(1)
	}

	useCaseOkEmpty := func(mockUC *mocks.MockGetSuperAdminListUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetSuperAdminList(gomock.Any()).
			Return(emptySlice, uint(0), nil).
			Times(1)
	}


	useCaseNeverCalled := func(mockUC *mocks.MockGetSuperAdminListUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetSuperAdminList(gomock.Any()).
			Times(0)
	}

	useCaseUnauthorizedAccess := func(mockUC *mocks.MockGetSuperAdminListUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetSuperAdminList(gomock.Any()).
			Return(emptySlice, uint(0), identity.ErrUnauthorizedAccess).
			Times(1)
	}

	errMock := errors.New("unexpected error")
	useCaseUnexpectedErr := func(mockUC *mocks.MockGetSuperAdminListUseCase) *gomock.Call {
		return mockUC.EXPECT().
			GetSuperAdminList(gomock.Any()).
			Return(emptySlice, uint(0), errMock).
			Times(1)
	}

	// Requester
	superAdminRequester := identity.Requester{
		RequesterUserId: uint(1),
		RequesterRole:   identity.ROLE_SUPER_ADMIN,
	}

	tenantId := uuid.New()
	authTenantAdminRequester := identity.Requester{
		RequesterUserId:   uint(1),
		RequesterTenantId: &tenantId,
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	// Input

	validUrl := "/super_admins"
	invalidUrl_pagination := "/super_admins?page=invalid&limit=invalid"
	// invalidUrl_userId := fmt.Sprintf("/tenant/%v/tenant_users/", targetTenantId.String())

	cases := []genericControllerTestCase[any, mocks.MockGetSuperAdminListUseCase]{
		{
			name:      "200 OK: Populated list",
			method:    "GET",
			url:       validUrl,
			inputDto:  nil,
			requester: superAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetSuperAdminListUseCase]{
				useCaseOk,
			},
			expectedStatus:   http.StatusOK,
			expectedResponse: expectedResponse,
		},
		{
			name:      "200 OK: Empty list",
			method:    "GET",
			url:       validUrl,
			inputDto:  nil,
			requester: superAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetSuperAdminListUseCase]{
				useCaseOkEmpty,
			},
			expectedStatus:   http.StatusOK,
			expectedResponse: expectedResponseEmpty,
		},
		{
			name:      "400 Bad Request: Invalid page parameters",
			method:    "GET",
			url:       invalidUrl_pagination,
			inputDto:  nil,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetSuperAdminListUseCase]{
				useCaseNeverCalled,
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: hasError{},
		},
		{
			name:         "401 Unauthorized: No identity",
			method:       "GET",
			url:          validUrl,
			inputDto:     nil,
			requester:    authTenantAdminRequester,
			omitIdentity: true,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetSuperAdminListUseCase]{
				useCaseNeverCalled,
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: gin.H{
				"error": transportHttp.ErrMissingIdentity.Error(),
			},
		},
		{
			name:      "404 Not found: Unauthorized access (obfuscated)",
			method:    "GET",
			url:       validUrl,
			inputDto:  nil,
			requester: authTenantAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetSuperAdminListUseCase]{
				useCaseUnauthorizedAccess,
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: gin.H{
				"error": identity.ErrUnauthorizedAccess.Error(),
			},
		},
		{
			name:      "500 Server Error: Unexpected error",
			method:    "GET",
			url:       validUrl,
			inputDto:  nil,
			requester: superAdminRequester,
			setupSteps: []mockUseCaseSetupFunc[mocks.MockGetSuperAdminListUseCase]{
				useCaseUnexpectedErr,
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResponse: gin.H{
				"error": errMock.Error(),
			},
		},
	}

	// Parametri di test
	mountMethod := "GET"
	mountUrl := "/super_admins"

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockUseCase := setupMockUseCase(
				mocks.NewMockGetSuperAdminListUseCase,
				tc.setupSteps, t,
			)
			userController := user.NewUserController(
				nil,
				nil, nil, nil,
				nil, nil, nil,
				nil, nil, nil,
				nil, nil, mockUseCase, 
			)

			executeControllerTest(
				t, tc,
				mountMethod, mountUrl,
				userController.GetSuperAdmins,
			)
		})
	}
}