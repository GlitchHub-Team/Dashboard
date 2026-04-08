package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"backend/internal/auth"
	transportHttp "backend/internal/infra/transport/http"
	"backend/internal/infra/transport/http/dto"
	"backend/internal/shared/identity"
	"backend/internal/user"
	"backend/tests/auth/mocks"
	"backend/tests/helper"
	cryptoMocks "backend/tests/shared/crypto/mocks"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

func TestAuthController_LoginUser(t *testing.T) {
	type mockSetupFunc func(
		mockUC *mocks.MockLoginUserUseCase,
		authTokenManager *cryptoMocks.MockAuthTokenManager,
	) *gomock.Call

	// NOTA: privato e interno a funzione perché questo metodo del controller utilizza due interfacce
	type testCase struct {
		Name             string
		Method           string
		Url              string
		InputDto         auth.LoginUserDTO
		SetupSteps       []mockSetupFunc
		ExpectedStatus   int
		ExpectedResponse any
	}

	targetTenantId := uuid.New()
	targetTenantIdString := targetTenantId.String()
	targetUserId := uint(1)
	targetUserRole := identity.ROLE_TENANT_USER
	targetEmail := "a@example.com"
	targetPassword := "password-123"
	targetPasswordHash := "password-123-hash"

	expectedConfirmedUser := user.User{
		Id:           targetUserId,
		Name:         "a",
		Email:        targetEmail,
		PasswordHash: &targetPasswordHash,
		Role:         targetUserRole,
		TenantId:     &targetTenantId,
		Confirmed:    true,
	}

	expectedCommand := auth.LoginUserCommand{
		TenantId: &targetTenantId,
		Email:    targetEmail,
		Password: targetPassword,
	}

	targetRequester := identity.Requester{
		RequesterUserId:   targetUserId,
		RequesterTenantId: &targetTenantId,
		RequesterRole:     targetUserRole,
	}
	expectedJwt := "jwt-123123"

	expectedResponse := auth.LoginResponseDTO{
		JWT: expectedJwt,
	}

	// Step 2: esecuzione login
	step2UseCaseOk := func(
		mockUC *mocks.MockLoginUserUseCase, authTokenManager *cryptoMocks.MockAuthTokenManager,
	) *gomock.Call {
		return mockUC.EXPECT().
			LoginUser(gomock.Eq(expectedCommand)).
			Return(expectedConfirmedUser, nil).
			Times(1)
	}

	step2UseCaseNeverCalled := func(
		mockUC *mocks.MockLoginUserUseCase, authTokenManager *cryptoMocks.MockAuthTokenManager,
	) *gomock.Call {
		return mockUC.EXPECT().
			LoginUser(gomock.Any()).
			Times(0)
	}

	step2UseCaseAccountNotConfirmed := func(
		mockUC *mocks.MockLoginUserUseCase, authTokenManager *cryptoMocks.MockAuthTokenManager,
	) *gomock.Call {
		return mockUC.EXPECT().
			LoginUser(gomock.Eq(expectedCommand)).
			Return(user.User{}, auth.ErrAccountNotConfirmed).
			Times(1)
	}

	step2UseCaseWrongCredentials := func(
		mockUC *mocks.MockLoginUserUseCase, authTokenManager *cryptoMocks.MockAuthTokenManager,
	) *gomock.Call {
		return mockUC.EXPECT().
			LoginUser(gomock.Eq(expectedCommand)).
			Return(user.User{}, auth.ErrWrongCredentials).
			Times(1)
	}

	errMockStep2 := errors.New("unexpected error step 2")
	step2UseCaseUnexpectedErr := func(
		mockUC *mocks.MockLoginUserUseCase, authTokenManager *cryptoMocks.MockAuthTokenManager,
	) *gomock.Call {
		return mockUC.EXPECT().
			LoginUser(gomock.Eq(expectedCommand)).
			Return(user.User{}, errMockStep2).
			Times(1)
	}

	// Step 3: crea token
	step3GenerateTokenOk := func(
		mockUC *mocks.MockLoginUserUseCase, authTokenManager *cryptoMocks.MockAuthTokenManager,
	) *gomock.Call {
		return authTokenManager.EXPECT().
			GenerateForRequester(targetRequester).
			Return(expectedJwt, nil).
			Times(1)
	}

	errMockStep3 := errors.New("unexpected error step 3")
	step3GenerateTokenError := func(
		mockUC *mocks.MockLoginUserUseCase, authTokenManager *cryptoMocks.MockAuthTokenManager,
	) *gomock.Call {
		return authTokenManager.EXPECT().
			GenerateForRequester(targetRequester).
			Return("", errMockStep3).
			Times(1)
	}

	// Input
	validPayload := auth.LoginUserDTO{
		TenantIdField_NotRequired: dto.TenantIdField_NotRequired{
			TenantId: &targetTenantIdString,
		},
		EmailField: dto.EmailField{
			Email: targetEmail,
		},
		PasswordField: dto.PasswordField{
			Password: targetPassword,
		},
		// UserRoleField: dto.UserRoleField{
		// 	UserRole: string(targetUserRole),
		// },
	}

	invalidUuidString := "not-a-uuid"
	invalidPayload := auth.LoginUserDTO{
		TenantIdField_NotRequired: dto.TenantIdField_NotRequired{
			TenantId: &invalidUuidString,
		},
		EmailField: dto.EmailField{
			Email: "",
		},
		PasswordField: dto.PasswordField{
			Password: targetPassword,
		},
		// UserRoleField: dto.UserRoleField{
		// 	UserRole: "invalid_role",
		// },
	}

	baseUrl := "/auth/login"

	cases := []testCase{
		{
			Name:     "200 OK",
			Method:   "POST",
			Url:      baseUrl,
			InputDto: validPayload,
			SetupSteps: []mockSetupFunc{
				step2UseCaseOk,
				step3GenerateTokenOk,
			},
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: expectedResponse,
		},
		{
			Name:     "400 Bad Request (step 1): Invalid body",
			Method:   "POST",
			Url:      baseUrl,
			InputDto: invalidPayload,
			SetupSteps: []mockSetupFunc{
				step2UseCaseNeverCalled,
			},
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedResponse: helper.HasError{},
		},
		{
			Name:     "404 Not Found (step 2): Account not confirmed",
			Method:   "POST",
			Url:      baseUrl,
			InputDto: validPayload,
			SetupSteps: []mockSetupFunc{
				step2UseCaseAccountNotConfirmed,
			},
			ExpectedStatus: http.StatusNotFound,
			ExpectedResponse: gin.H{
				"error": auth.ErrAccountNotConfirmed.Error(),
			},
		},
		{
			Name:     "404 Not Found (step 2): Wrong credentials",
			Method:   "POST",
			Url:      baseUrl,
			InputDto: validPayload,
			SetupSteps: []mockSetupFunc{
				step2UseCaseWrongCredentials,
			},
			ExpectedStatus: http.StatusNotFound,
			ExpectedResponse: gin.H{
				"error": auth.ErrWrongCredentials.Error(),
			},
		},
		{
			Name:     "500 Server Error (step 2): unexpected error",
			Method:   "POST",
			Url:      baseUrl,
			InputDto: validPayload,
			SetupSteps: []mockSetupFunc{
				step2UseCaseUnexpectedErr,
			},
			ExpectedStatus: http.StatusInternalServerError,
			ExpectedResponse: gin.H{
				"error": errMockStep2.Error(),
			},
		},
		{
			Name:     "500 Server Error (step 3): unexpected error",
			Method:   "POST",
			Url:      baseUrl,
			InputDto: validPayload,
			SetupSteps: []mockSetupFunc{
				step2UseCaseOk,
				step3GenerateTokenError,
			},
			ExpectedStatus: http.StatusInternalServerError,
			ExpectedResponse: gin.H{
				"error": errMockStep3.Error(),
			},
		},
	}

	// NOTA: qua non uso helper perché LoginUser ha due dipendenze. Non dipende solo dal suo use case
	mountUrl := "/auth/login"
	mountMethod := "POST"
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			// Setup mocks
			ctrl := gomock.NewController(t)
			mockUseCase := mocks.NewMockLoginUserUseCase(ctrl)
			mockAuthTokenManager := cryptoMocks.NewMockAuthTokenManager(ctrl)

			var expectedCalls []any
			for _, step := range tc.SetupSteps {
				if call := step(mockUseCase, mockAuthTokenManager); call != nil {
					expectedCalls = append(expectedCalls, call)
				}
			}
			if len(expectedCalls) > 0 {
				gomock.InOrder(expectedCalls...)
			}

			// Init controller
			authController := auth.NewController(
				nil, mockAuthTokenManager,
				mockUseCase,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
			)

			// Esegui test su controller
			gin.SetMode(gin.TestMode)
			router := gin.New()

			if validatorEngine, ok := binding.Validator.Engine().(*validator.Validate); ok {
				validatorEngine.RegisterTagNameFunc(func(fld reflect.StructField) string {
					name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
					if name == "-" {
						return ""
					}
					return name
				})
			}

			// NOTA: Tolgo direttamente il middleware per autorizzazione

			router.Handle(mountMethod, mountUrl, authController.LoginUser)

			reqBody, err := json.Marshal(tc.InputDto)
			if err != nil {
				t.Fatalf("error marshaling request body: %v", err)
			}

			req, err := http.NewRequest(tc.Method, tc.Url, bytes.NewBuffer(reqBody)) //nolint:noctx
			if err != nil {
				t.Fatalf("error creating request: %v", err)
			}
			if reflect.ValueOf(tc.InputDto) != reflect.Zero(reflect.TypeFor[auth.LoginUserCommand]()) {
				req.Header.Set("Content-Type", "application/json")
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tc.ExpectedStatus {
				t.Fatalf("expected status %d, got %d. Response: %s", tc.ExpectedStatus, w.Code, w.Body.String())
			}

			helper.CheckHttpResponse(w, tc.ExpectedResponse, t)
		})
	}
}

func TestController_VerifyConfirmAccountToken(t *testing.T) {
	targetTenantId := uuid.New()
	targetTenantIdStr := targetTenantId.String()
	targetToken := "verify-token-123"

	inputDto := auth.VerifyConfirmAccountTokenBodyDTO{
		TokenFields: dto.TokenFields{
			TenantIdField_NotRequired: dto.TenantIdField_NotRequired{
				TenantId: &targetTenantIdStr,
			},
			Token: targetToken,
		},
	}

	expectedCommand := auth.VerifyConfirmAccountTokenCommand{
		TenantId: &targetTenantId,
		Token:    targetToken,
	}

	useCaseOk := func(mockUC *mocks.MockVerifyConfirmAccountTokenUseCase) *gomock.Call {
		return mockUC.EXPECT().
			VerifyConfirmAccountToken(gomock.Eq(expectedCommand)).
			Return(nil).
			Times(1)
	}

	useCaseTokenNotFound := func(mockUC *mocks.MockVerifyConfirmAccountTokenUseCase) *gomock.Call {
		return mockUC.EXPECT().
			VerifyConfirmAccountToken(gomock.Eq(expectedCommand)).
			Return(auth.ErrTokenNotFound).
			Times(1)
	}

	baseUrl := "/auth/confirm_account/verify_token"
	cases := []helper.GenericControllerTestCase[any, mocks.MockVerifyConfirmAccountTokenUseCase]{
		{
			Name:     "200 OK: Token verified successfully",
			Method:   "POST",
			Url:      baseUrl,
			InputDto: inputDto,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockVerifyConfirmAccountTokenUseCase]{
				useCaseOk,
			},
			ExpectedStatus: http.StatusOK,
			ExpectedResponse: gin.H{
				"result": true,
			},
		},
		{
			Name:     "404 Not Found: Invalid token",
			Method:   "POST",
			Url:      baseUrl,
			InputDto: inputDto,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockVerifyConfirmAccountTokenUseCase]{
				useCaseTokenNotFound,
			},
			ExpectedStatus: http.StatusNotFound,
			ExpectedResponse: gin.H{
				"error": auth.ErrTokenNotFound.Error(),
			},
		},
	}

	mountMethod := "POST"
	mountURL := "/auth/confirm_account/verify_token"

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			mockUseCase := helper.SetupMockUseCase(
				mocks.NewMockVerifyConfirmAccountTokenUseCase,
				tc.SetupSteps,
				t,
			)

			mockLogger := zaptest.NewLogger(t)

			controller := auth.NewController(
				mockLogger,
				nil, nil, nil, nil,
				mockUseCase,
				nil, nil, nil, nil,
			)

			helper.ExecuteControllerTest(
				t,
				tc,
				mountMethod,
				mountURL,
				controller.VerifyConfirmAccountToken,
			)
		})
	}
}

func TestAuthController_ConfirmAccount(t *testing.T) {
	type mockSetupFunc func(
		mockUC *mocks.MockConfirmAccountUseCase,
		authTokenManager *cryptoMocks.MockAuthTokenManager,
	) *gomock.Call

	// NOTA: privato e interno a funzione perché questo metodo del controller utilizza due interfacce
	type testCase struct {
		Name             string
		Method           string
		Url              string
		InputDto         auth.ConfirmUserAccountBodyDTO
		SetupSteps       []mockSetupFunc
		ExpectedStatus   int
		ExpectedResponse any
	}

	targetTenantId := uuid.New()
	targetTenantIdString := targetTenantId.String()
	targetUserId := uint(1)
	targetUserRole := identity.ROLE_TENANT_USER
	targetEmail := "a@example.com"
	targetPasswordHash := "password-123-hash"
	targetToken := "token"
	targetNewPassword := "new-password"

	expectedConfirmedUser := user.User{
		Id:           targetUserId,
		Name:         "a",
		Email:        targetEmail,
		PasswordHash: &targetPasswordHash,
		Role:         targetUserRole,
		TenantId:     &targetTenantId,
		Confirmed:    true,
	}

	expectedCommand := auth.ConfirmAccountCommand{
		TenantId:    &targetTenantId,
		Token:       targetToken,
		NewPassword: targetNewPassword,
	}

	targetRequester := identity.Requester{
		RequesterUserId:   targetUserId,
		RequesterTenantId: &targetTenantId,
		RequesterRole:     targetUserRole,
	}
	expectedJwt := "jwt-123123"

	expectedResponse := auth.LoginResponseDTO{
		JWT: expectedJwt,
	}

	// Step 2: esecuzione login
	step2UseCaseOk := func(
		mockUC *mocks.MockConfirmAccountUseCase, authTokenManager *cryptoMocks.MockAuthTokenManager,
	) *gomock.Call {
		return mockUC.EXPECT().
			ConfirmAccount(gomock.Eq(expectedCommand)).
			Return(expectedConfirmedUser, nil).
			Times(1)
	}

	step2UseCaseNeverCalled := func(
		mockUC *mocks.MockConfirmAccountUseCase, authTokenManager *cryptoMocks.MockAuthTokenManager,
	) *gomock.Call {
		return mockUC.EXPECT().
			ConfirmAccount(gomock.Any()).
			Times(0)
	}

	step2UseCaseAccountAlreadyConfirmed := func(
		mockUC *mocks.MockConfirmAccountUseCase, authTokenManager *cryptoMocks.MockAuthTokenManager,
	) *gomock.Call {
		return mockUC.EXPECT().
			ConfirmAccount(gomock.Eq(expectedCommand)).
			Return(user.User{}, auth.ErrAccountAlreadyConfirmed).
			Times(1)
	}

	step2UseCaseTokenNotFound := func(
		mockUC *mocks.MockConfirmAccountUseCase, authTokenManager *cryptoMocks.MockAuthTokenManager,
	) *gomock.Call {
		return mockUC.EXPECT().
			ConfirmAccount(gomock.Eq(expectedCommand)).
			Return(user.User{}, auth.ErrTokenNotFound).
			Times(1)
	}

	step2UseCaseTokenExpired := func(
		mockUC *mocks.MockConfirmAccountUseCase, authTokenManager *cryptoMocks.MockAuthTokenManager,
	) *gomock.Call {
		return mockUC.EXPECT().
			ConfirmAccount(gomock.Eq(expectedCommand)).
			Return(user.User{}, auth.ErrTokenExpired).
			Times(1)
	}

	errMockStep2 := errors.New("unexpected error step 2")
	step2UseCaseUnexpectedErr := func(
		mockUC *mocks.MockConfirmAccountUseCase, authTokenManager *cryptoMocks.MockAuthTokenManager,
	) *gomock.Call {
		return mockUC.EXPECT().
			ConfirmAccount(gomock.Eq(expectedCommand)).
			Return(user.User{}, errMockStep2).
			Times(1)
	}

	// Step 3: crea token
	step3GenerateTokenOk := func(
		mockUC *mocks.MockConfirmAccountUseCase, authTokenManager *cryptoMocks.MockAuthTokenManager,
	) *gomock.Call {
		return authTokenManager.EXPECT().
			GenerateForRequester(targetRequester).
			Return(expectedJwt, nil).
			Times(1)
	}

	errMockStep3 := errors.New("unexpected error step 3")
	step3GenerateTokenError := func(
		mockUC *mocks.MockConfirmAccountUseCase, authTokenManager *cryptoMocks.MockAuthTokenManager,
	) *gomock.Call {
		return authTokenManager.EXPECT().
			GenerateForRequester(targetRequester).
			Return("", errMockStep3).
			Times(1)
	}

	// Input
	validPayload := auth.ConfirmUserAccountBodyDTO{
		TokenFields: dto.TokenFields{
			Token: targetToken,
			TenantIdField_NotRequired: dto.TenantIdField_NotRequired{
				TenantId: &targetTenantIdString,
			},
		},
		NewPasswordField: dto.NewPasswordField{
			NewPassword: targetNewPassword,
		},
	}

	invalidUuidString := "not-a-uuid"
	invalidPayload := auth.ConfirmUserAccountBodyDTO{
		TokenFields: dto.TokenFields{
			Token: targetToken,
			TenantIdField_NotRequired: dto.TenantIdField_NotRequired{
				TenantId: &invalidUuidString,
			},
		},
		NewPasswordField: dto.NewPasswordField{
			NewPassword: targetNewPassword,
		},
	}

	baseUrl := "/auth/confirm_account"

	cases := []testCase{
		{
			Name:     "200 OK",
			Method:   "POST",
			Url:      baseUrl,
			InputDto: validPayload,
			SetupSteps: []mockSetupFunc{
				step2UseCaseOk,
				step3GenerateTokenOk,
			},
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: expectedResponse,
		},
		{
			Name:     "400 Bad Request (step 1): Invalid body",
			Method:   "POST",
			Url:      baseUrl,
			InputDto: invalidPayload,
			SetupSteps: []mockSetupFunc{
				step2UseCaseNeverCalled,
			},
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedResponse: helper.HasError{},
		},
		{
			Name:     "404 Not Found (step 2): Account already confirmed",
			Method:   "POST",
			Url:      baseUrl,
			InputDto: validPayload,
			SetupSteps: []mockSetupFunc{
				step2UseCaseAccountAlreadyConfirmed,
			},
			ExpectedStatus: http.StatusNotFound,
			ExpectedResponse: gin.H{
				"error": auth.ErrAccountAlreadyConfirmed.Error(),
			},
		},
		{
			Name:     "404 Not Found (step 2): Token not found (obfuscated)",
			Method:   "POST",
			Url:      baseUrl,
			InputDto: validPayload,
			SetupSteps: []mockSetupFunc{
				step2UseCaseTokenNotFound,
			},
			ExpectedStatus: http.StatusNotFound,
			ExpectedResponse: gin.H{
				"error": auth.ErrTokenNotFound.Error(),
			},
		},
		{
			Name:     "404 Not Found (step 2): Token expired (obfuscated)",
			Method:   "POST",
			Url:      baseUrl,
			InputDto: validPayload,
			SetupSteps: []mockSetupFunc{
				step2UseCaseTokenExpired,
			},
			ExpectedStatus: http.StatusNotFound,
			ExpectedResponse: gin.H{
				"error": auth.ErrTokenNotFound.Error(),
			},
		},
		{
			Name:     "500 Server Error (step 2): unexpected error",
			Method:   "POST",
			Url:      baseUrl,
			InputDto: validPayload,
			SetupSteps: []mockSetupFunc{
				step2UseCaseUnexpectedErr,
			},
			ExpectedStatus: http.StatusInternalServerError,
			ExpectedResponse: gin.H{
				"error": errMockStep2.Error(),
			},
		},
		{
			Name:     "500 Server Error (step 3): unexpected error",
			Method:   "POST",
			Url:      baseUrl,
			InputDto: validPayload,
			SetupSteps: []mockSetupFunc{
				step2UseCaseOk,
				step3GenerateTokenError,
			},
			ExpectedStatus: http.StatusInternalServerError,
			ExpectedResponse: gin.H{
				"error": errMockStep3.Error(),
			},
		},
	}

	// NOTA: qua non uso helper perché LoginUser ha due dipendenze. Non dipende solo dal suo use case
	mountUrl := "/auth/confirm_account"
	mountMethod := "POST"
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			// Setup mocks
			ctrl := gomock.NewController(t)
			mockUseCase := mocks.NewMockConfirmAccountUseCase(ctrl)
			mockAuthTokenManager := cryptoMocks.NewMockAuthTokenManager(ctrl)

			var expectedCalls []any
			for _, step := range tc.SetupSteps {
				if call := step(mockUseCase, mockAuthTokenManager); call != nil {
					expectedCalls = append(expectedCalls, call)
				}
			}
			if len(expectedCalls) > 0 {
				gomock.InOrder(expectedCalls...)
			}

			// Init controller
			authController := auth.NewController(
				nil, mockAuthTokenManager,
				nil,
				nil,
				mockUseCase,
				nil,
				nil,
				nil,
				nil,
				nil,
			)

			// Esegui test su controller
			gin.SetMode(gin.TestMode)
			router := gin.New()

			if validatorEngine, ok := binding.Validator.Engine().(*validator.Validate); ok {
				validatorEngine.RegisterTagNameFunc(func(fld reflect.StructField) string {
					name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
					if name == "-" {
						return ""
					}
					return name
				})
			}

			// NOTA: Tolgo direttamente il middleware per autorizzazione

			router.Handle(mountMethod, mountUrl, authController.ConfirmAccount)

			reqBody, err := json.Marshal(tc.InputDto)
			if err != nil {
				t.Fatalf("error marshaling request body: %v", err)
			}

			req, err := http.NewRequest(tc.Method, tc.Url, bytes.NewBuffer(reqBody)) //nolint:noctx
			if err != nil {
				t.Fatalf("error creating request: %v", err)
			}
			if reflect.ValueOf(tc.InputDto) != reflect.Zero(reflect.TypeFor[auth.LoginUserCommand]()) {
				req.Header.Set("Content-Type", "application/json")
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tc.ExpectedStatus {
				t.Fatalf("expected status %d, got %d. Response: %s", tc.ExpectedStatus, w.Code, w.Body.String())
			}

			helper.CheckHttpResponse(w, tc.ExpectedResponse, t)
		})
	}
}

func TestController_VerifyForgotPasswordToken(t *testing.T) {
	targetTenantId := uuid.New()
	targetTenantIdStr := targetTenantId.String()
	targetToken := "verify-token-123"

	inputDto := auth.VerifyForgotPasswordTokenBodyDTO{
		TokenFields: dto.TokenFields{
			TenantIdField_NotRequired: dto.TenantIdField_NotRequired{
				TenantId: &targetTenantIdStr,
			},
			Token: targetToken,
		},
	}

	incorrectUuidString := "not-a-uuid"
	invalidInputDto := auth.VerifyForgotPasswordTokenBodyDTO{
		TokenFields: dto.TokenFields{
			TenantIdField_NotRequired: dto.TenantIdField_NotRequired{
				TenantId: &incorrectUuidString,
			},
			Token: targetToken,
		},
	}

	expectedCommand := auth.VerifyForgotPasswordTokenCommand{
		TenantId: &targetTenantId,
		Token:    targetToken,
	}

	useCaseOk := func(mockUC *mocks.MockVerifyForgotPasswordTokenUseCase) *gomock.Call {
		return mockUC.EXPECT().
			VerifyForgotPasswordToken(gomock.Eq(expectedCommand)).
			Return(nil).
			Times(1)
	}

	useCaseTokenNotFound := func(mockUC *mocks.MockVerifyForgotPasswordTokenUseCase) *gomock.Call {
		return mockUC.EXPECT().
			VerifyForgotPasswordToken(gomock.Eq(expectedCommand)).
			Return(auth.ErrTokenNotFound).
			Times(1)
	}

	useCaseTokenNeverCalled := func(mockUC *mocks.MockVerifyForgotPasswordTokenUseCase) *gomock.Call {
		return mockUC.EXPECT().
			VerifyForgotPasswordToken(gomock.Any()).
			Times(0)
	}

	baseUrl := "/auth/forgot_password/verify_token"
	cases := []helper.GenericControllerTestCase[any, mocks.MockVerifyForgotPasswordTokenUseCase]{
		{
			Name:     "200 OK: Token verified successfully",
			Method:   "POST",
			Url:      baseUrl,
			InputDto: inputDto,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockVerifyForgotPasswordTokenUseCase]{
				useCaseOk,
			},
			ExpectedStatus: http.StatusOK,
			ExpectedResponse: gin.H{
				"result": "ok",
			},
		},
		{
			Name:     "404 Not Found: Invalid body (obfuscated)",
			Method:   "POST",
			Url:      baseUrl,
			InputDto: invalidInputDto,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockVerifyForgotPasswordTokenUseCase]{
				useCaseTokenNeverCalled,
			},
			ExpectedStatus: http.StatusNotFound,
			ExpectedResponse: gin.H{
				"error": auth.ErrTokenNotFound.Error(),
			},
		},

		{
			Name:     "404 Not Found: Invalid token",
			Method:   "POST",
			Url:      baseUrl,
			InputDto: inputDto,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockVerifyForgotPasswordTokenUseCase]{
				useCaseTokenNotFound,
			},
			ExpectedStatus: http.StatusNotFound,
			ExpectedResponse: gin.H{
				"error": auth.ErrTokenNotFound.Error(),
			},
		},
	}

	mountMethod := "POST"
	mountURL := "/auth/forgot_password/verify_token"

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			mockUseCase := helper.SetupMockUseCase(
				mocks.NewMockVerifyForgotPasswordTokenUseCase,
				tc.SetupSteps,
				t,
			)

			mockLogger := zaptest.NewLogger(t)

			controller := auth.NewController(
				mockLogger,
				nil,
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
				controller.VerifyForgotPasswordToken,
			)
		})
	}
}

func TestController_RequestForgotPasswordToken(t *testing.T) {
	targetTenantId := uuid.New()
	targetTenantIdStr := targetTenantId.String()
	targetEmail := "info@example.com"

	inputDto := auth.RequestForgotPasswordBodyDTO{
		TenantIdField_NotRequired: dto.TenantIdField_NotRequired{
			TenantId: &targetTenantIdStr,
		},
		EmailField: dto.EmailField{
			Email: targetEmail,
		},
	}

	incorrectUuidString := "not-a-uuid"
	invalidInputDto := auth.RequestForgotPasswordBodyDTO{
		TenantIdField_NotRequired: dto.TenantIdField_NotRequired{
			TenantId: &incorrectUuidString,
		},
		EmailField: dto.EmailField{
			Email: targetEmail,
		},
	}

	expectedCommand := auth.RequestForgotPasswordCommand{
		TenantId: &targetTenantId,
		Email:    targetEmail,
	}

	useCaseOk := func(mockUC *mocks.MockRequestForgotPasswordUseCase) *gomock.Call {
		return mockUC.EXPECT().
			RequestForgotPassword(gomock.Eq(expectedCommand)).
			Return(nil).
			Times(1)
	}

	useCaseUserNotFound := func(mockUC *mocks.MockRequestForgotPasswordUseCase) *gomock.Call {
		return mockUC.EXPECT().
			RequestForgotPassword(gomock.Eq(expectedCommand)).
			Return(user.ErrUserNotFound).
			Times(1)
	}

	useCaseUserAccountNotConfirmed := func(mockUC *mocks.MockRequestForgotPasswordUseCase) *gomock.Call {
		return mockUC.EXPECT().
			RequestForgotPassword(gomock.Eq(expectedCommand)).
			Return(auth.ErrAccountNotConfirmed).
			Times(1)
	}

	errMock := errors.New("unexpected error")
	useCaseUserError := func(mockUC *mocks.MockRequestForgotPasswordUseCase) *gomock.Call {
		return mockUC.EXPECT().
			RequestForgotPassword(gomock.Eq(expectedCommand)).
			Return(errMock).
			Times(1)
	}

	useCaseTokenNeverCalled := func(mockUC *mocks.MockRequestForgotPasswordUseCase) *gomock.Call {
		return mockUC.EXPECT().
			RequestForgotPassword(gomock.Any()).
			Times(0)
	}

	baseUrl := "/auth/forgot_password/request"
	cases := []helper.GenericControllerTestCase[any, mocks.MockRequestForgotPasswordUseCase]{
		{
			Name:     "200 OK",
			Method:   "POST",
			Url:      baseUrl,
			InputDto: inputDto,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockRequestForgotPasswordUseCase]{
				useCaseOk,
			},
			ExpectedStatus: http.StatusOK,
			ExpectedResponse: gin.H{
				"result": "ok",
			},
		},
		{
			Name:     "400 Bad request: Invalid body",
			Method:   "POST",
			Url:      baseUrl,
			InputDto: invalidInputDto,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockRequestForgotPasswordUseCase]{
				useCaseTokenNeverCalled,
			},
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedResponse: helper.HasError{},
		},

		{
			Name:     "404 Not Found: User not found",
			Method:   "POST",
			Url:      baseUrl,
			InputDto: inputDto,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockRequestForgotPasswordUseCase]{
				useCaseUserNotFound,
			},
			ExpectedStatus: http.StatusNotFound,
			ExpectedResponse: gin.H{
				"error": user.ErrUserNotFound.Error(),
			},
		},
		{
			Name:     "404 Not Found: Account not confirmed",
			Method:   "POST",
			Url:      baseUrl,
			InputDto: inputDto,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockRequestForgotPasswordUseCase]{
				useCaseUserAccountNotConfirmed,
			},
			ExpectedStatus: http.StatusNotFound,
			ExpectedResponse: gin.H{
				"error": auth.ErrAccountNotConfirmed.Error(),
			},
		},
		{
			Name:     "500 Server Error: unexpected error",
			Method:   "POST",
			Url:      baseUrl,
			InputDto: inputDto,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockRequestForgotPasswordUseCase]{
				useCaseUserError,
			},
			ExpectedStatus: http.StatusInternalServerError,
			ExpectedResponse: gin.H{
				"error": errMock.Error(),
			},
		},
	}

	mountMethod := "POST"
	mountURL := "/auth/forgot_password/request"

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			mockUseCase := helper.SetupMockUseCase(
				mocks.NewMockRequestForgotPasswordUseCase,
				tc.SetupSteps,
				t,
			)

			mockLogger := zaptest.NewLogger(t)

			controller := auth.NewController(
				mockLogger,
				nil,
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
				controller.RequestForgotPasswordToken,
			)
		})
	}
}

func TestController_ConfirmForgotPasswordToken(t *testing.T) {
	targetTenantId := uuid.New()
	targetTenantIdStr := targetTenantId.String()
	targetToken := "token-123"
	targetNewPassword := "new-password-123"

	inputDto := auth.ConfirmForgotPasswordBodyDTO{
		TokenFields: dto.TokenFields{
			Token: targetToken,
			TenantIdField_NotRequired: dto.TenantIdField_NotRequired{
				TenantId: &targetTenantIdStr,
			},
		},
		NewPasswordField: dto.NewPasswordField{
			NewPassword: targetNewPassword,
		},
	}

	incorrectUuidStr := "not-a-uuid"
	invalidInputDto := auth.ConfirmForgotPasswordBodyDTO{
		TokenFields: dto.TokenFields{
			Token: targetToken,
			TenantIdField_NotRequired: dto.TenantIdField_NotRequired{
				TenantId: &incorrectUuidStr,
			},
		},
		NewPasswordField: dto.NewPasswordField{
			NewPassword: targetNewPassword,
		},
	}

	expectedCommand := auth.ConfirmForgotPasswordCommand{
		TenantId:    &targetTenantId,
		Token:       targetToken,
		NewPassword: targetNewPassword,
	}

	useCaseOk := func(mockUC *mocks.MockConfirmForgotPasswordUseCase) *gomock.Call {
		return mockUC.EXPECT().
			ConfirmForgotPassword(gomock.Eq(expectedCommand)).
			Return(nil).
			Times(1)
	}

	useCaseUserAccountNotConfirmed := func(mockUC *mocks.MockConfirmForgotPasswordUseCase) *gomock.Call {
		return mockUC.EXPECT().
			ConfirmForgotPassword(gomock.Eq(expectedCommand)).
			Return(auth.ErrAccountNotConfirmed).
			Times(1)
	}

	errMock := errors.New("unexpected error")
	useCaseUserError := func(mockUC *mocks.MockConfirmForgotPasswordUseCase) *gomock.Call {
		return mockUC.EXPECT().
			ConfirmForgotPassword(gomock.Eq(expectedCommand)).
			Return(errMock).
			Times(1)
	}

	useCaseTokenNeverCalled := func(mockUC *mocks.MockConfirmForgotPasswordUseCase) *gomock.Call {
		return mockUC.EXPECT().
			ConfirmForgotPassword(gomock.Any()).
			Times(0)
	}

	baseUrl := "/auth/forgot_password/confirm"
	cases := []helper.GenericControllerTestCase[any, mocks.MockConfirmForgotPasswordUseCase]{
		{
			Name:     "200 OK",
			Method:   "POST",
			Url:      baseUrl,
			InputDto: inputDto,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockConfirmForgotPasswordUseCase]{
				useCaseOk,
			},
			ExpectedStatus: http.StatusOK,
			ExpectedResponse: gin.H{
				"result": "ok",
			},
		},
		{
			Name:     "400 Bad request: Invalid body",
			Method:   "POST",
			Url:      baseUrl,
			InputDto: invalidInputDto,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockConfirmForgotPasswordUseCase]{
				useCaseTokenNeverCalled,
			},
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedResponse: helper.HasError{},
		},
		{
			Name:     "404 Not Found: Account not confirmed",
			Method:   "POST",
			Url:      baseUrl,
			InputDto: inputDto,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockConfirmForgotPasswordUseCase]{
				useCaseUserAccountNotConfirmed,
			},
			ExpectedStatus: http.StatusNotFound,
			ExpectedResponse: gin.H{
				"error": auth.ErrAccountNotConfirmed.Error(),
			},
		},
		{
			Name:     "500 Server Error: unexpected error",
			Method:   "POST",
			Url:      baseUrl,
			InputDto: inputDto,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockConfirmForgotPasswordUseCase]{
				useCaseUserError,
			},
			ExpectedStatus: http.StatusInternalServerError,
			ExpectedResponse: gin.H{
				"error": errMock.Error(),
			},
		},
	}

	mountMethod := "POST"
	mountURL := "/auth/forgot_password/confirm"

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			mockUseCase := helper.SetupMockUseCase(
				mocks.NewMockConfirmForgotPasswordUseCase,
				tc.SetupSteps,
				t,
			)

			mockLogger := zaptest.NewLogger(t)

			controller := auth.NewController(
				mockLogger,
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

			helper.ExecuteControllerTest(
				t,
				tc,
				mountMethod,
				mountURL,
				controller.ConfirmForgotPasswordToken,
			)
		})
	}
}

func TestController_ChangePassword(t *testing.T) {
	targetTenantId := uuid.New()
	targetOldPassword := "old-password-123"
	targetNewPassword := "new-password-123"

	inputDto := auth.ChangePasswordBodyDTO{
		ChangePasswordFields: dto.ChangePasswordFields{
			OldPassword: targetOldPassword,
			NewPassword: targetNewPassword,
		},
	}

	wrongOldPassword := "wrong-old-pw-123"
	wrongInputDto := auth.ChangePasswordBodyDTO{
		ChangePasswordFields: dto.ChangePasswordFields{
			OldPassword: wrongOldPassword,
			NewPassword: targetNewPassword,
		},
	}

	requester := identity.Requester{
		RequesterUserId:   uint(1),
		RequesterTenantId: &targetTenantId,
		RequesterRole:     identity.ROLE_TENANT_USER, // NOTA: ruolo irrilevante
	}

	targetCommand := auth.ChangePasswordCommand{
		Requester:   requester,
		OldPassword: targetOldPassword,
		NewPassword: targetNewPassword,
	}

	targetCommand_WrongCredentials := auth.ChangePasswordCommand{
		Requester:   requester,
		OldPassword: wrongOldPassword,
		NewPassword: targetNewPassword,
	}

	useCaseOk := func(mockUC *mocks.MockChangePasswordUseCase) *gomock.Call {
		return mockUC.EXPECT().
			ChangePassword(gomock.Eq(targetCommand)).
			Return(nil).
			Times(1)
	}

	useCaseAccountNotConfirmed := func(mockUC *mocks.MockChangePasswordUseCase) *gomock.Call {
		return mockUC.EXPECT().
			ChangePassword(gomock.Eq(targetCommand)).
			Return(auth.ErrAccountNotConfirmed).
			Times(1)
	}

	useCaseWrongCredentials := func(mockUC *mocks.MockChangePasswordUseCase) *gomock.Call {
		return mockUC.EXPECT().
			ChangePassword(gomock.Eq(targetCommand_WrongCredentials)).
			Return(auth.ErrWrongCredentials).
			Times(1)
	}

	errMock := errors.New("unexpected error")
	useCaseUserError := func(mockUC *mocks.MockChangePasswordUseCase) *gomock.Call {
		return mockUC.EXPECT().
			ChangePassword(gomock.Eq(targetCommand)).
			Return(errMock).
			Times(1)
	}

	useCaseTokenNeverCalled := func(mockUC *mocks.MockChangePasswordUseCase) *gomock.Call {
		return mockUC.EXPECT().
			ChangePassword(gomock.Any()).
			Times(0)
	}

	baseUrl := "/auth/forgot_password/confirm"
	cases := []helper.GenericControllerTestCase[any, mocks.MockChangePasswordUseCase]{
		{
			Name:      "200 OK",
			Method:    "POST",
			Url:       baseUrl,
			InputDto:  inputDto,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockChangePasswordUseCase]{
				useCaseOk,
			},
			ExpectedStatus: http.StatusOK,
			ExpectedResponse: gin.H{
				"result": "ok",
			},
		},
		{
			Name:         "401 Unauthorized: no requester",
			Method:       "POST",
			Url:          baseUrl,
			InputDto:     wrongInputDto,
			OmitIdentity: true,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockChangePasswordUseCase]{
				useCaseTokenNeverCalled,
			},
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedResponse: gin.H{
				"error": transportHttp.ErrMissingIdentity.Error(),
			},
		},
		{
			Name:      "404 Not Found: Account not confirmed",
			Method:    "POST",
			Url:       baseUrl,
			InputDto:  inputDto,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockChangePasswordUseCase]{
				useCaseAccountNotConfirmed,
			},
			ExpectedStatus: http.StatusNotFound,
			ExpectedResponse: gin.H{
				"error": auth.ErrAccountNotConfirmed.Error(),
			},
		},
		{
			Name:      "404 Not Found: Wrong credentials",
			Method:    "POST",
			Url:       baseUrl,
			InputDto:  wrongInputDto,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockChangePasswordUseCase]{
				useCaseWrongCredentials,
			},
			ExpectedStatus: http.StatusNotFound,
			ExpectedResponse: gin.H{
				"error": auth.ErrWrongCredentials.Error(),
			},
		},
		{
			Name:      "500 Server Error: unexpected error",
			Method:    "POST",
			Url:       baseUrl,
			InputDto:  inputDto,
			Requester: requester,
			SetupSteps: []helper.MockUseCaseSetupFunc[mocks.MockChangePasswordUseCase]{
				useCaseUserError,
			},
			ExpectedStatus: http.StatusInternalServerError,
			ExpectedResponse: gin.H{
				"error": errMock.Error(),
			},
		},
	}

	mountMethod := "POST"
	mountURL := "/auth/forgot_password/confirm"

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			mockUseCase := helper.SetupMockUseCase(
				mocks.NewMockChangePasswordUseCase,
				tc.SetupSteps,
				t,
			)

			mockLogger := zaptest.NewLogger(t)

			controller := auth.NewController(
				mockLogger,
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

			helper.ExecuteControllerTest(
				t,
				tc,
				mountMethod,
				mountURL,
				controller.ChangePassword,
			)
		})
	}
}
