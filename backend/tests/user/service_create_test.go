package user_test

import (
	"errors"
	"testing"

	"backend/internal/tenant"
	"backend/internal/user"

	authMocks "backend/tests/auth/mocks"
	emailMocks "backend/tests/email/mocks"
	tenantMocks "backend/tests/tenant/mocks"
	"backend/tests/user/mocks"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestCreateTenantUser(t *testing.T) {
	// Dati test
	targetTenantId := uuid.New()
	targetUserId := uint(1)
	targetUserEmail := "test@example.com"
	targetUserName := "Test"
	targetConfirmed := false
	targetRole := user.ROLE_TENANT_USER

	targetCreatedUser := user.User{
		Name:      targetUserName,
		Email:     targetUserEmail,
		Role:      targetRole,
		TenantId:  &targetTenantId,
		Confirmed: targetConfirmed,
	}

	expectedUser := user.User{
		Id:        targetUserId,
		Name:      targetUserName,
		Email:     targetUserEmail,
		TenantId:  &targetTenantId,
		Confirmed: targetConfirmed,
		Role:      targetRole,
	}
	expectedToken := "token"

	type mockSetupFunc func(
		createUserPort *mocks.MockCreateUserPort,
		deleteUserPort *mocks.MockDeleteUserPort,
		getUserPort *mocks.MockGetUserPort,
		getTenantPort *tenantMocks.MockGetTenantPort,
		confirmAccountTokenPort *authMocks.MockConfirmTokenPort,
		sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call
	type testCase struct {
		name          string
		input         user.CreateTenantUserCommand
		setupSteps    []mockSetupFunc
		expectedUser  user.User
		expectedError error
	}

	// Step 1: Cercare tenant
	step1TenantOk := func(
		createUserPort *mocks.MockCreateUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return getTenantPort.EXPECT().
			GetTenant(targetTenantId).
			Return(tenant.Tenant{Id: targetTenantId}, nil).
			Times(1)
	}

	step1TenantFail := func(
		createUserPort *mocks.MockCreateUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return getTenantPort.EXPECT().
			GetTenant(targetTenantId).
			Return(tenant.Tenant{}, tenant.ErrTenantNotFound).
			Times(1)
	}

	errMockStep1 := errors.New("unexpected error in step 1")
	step1TenantError := func(
		createUserPort *mocks.MockCreateUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return getTenantPort.EXPECT().
			GetTenant(targetTenantId).
			Return(tenant.Tenant{}, errMockStep1).
			Times(1)
	}

	step2GetUserOk := func(
		createUserPort *mocks.MockCreateUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantUserByEmail(targetTenantId, targetUserEmail).
			Return(user.User{}, nil).
			Times(1)
	}

	step2GetExistingUserFail := func(
		createUserPort *mocks.MockCreateUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantUserByEmail(targetTenantId, targetUserEmail).
			Return(expectedUser, nil).
			Times(1)
	}

	errMockStep2 := errors.New("unexpected error in step 2")
	step2GetUserError := func(
		createUserPort *mocks.MockCreateUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantUserByEmail(targetTenantId, targetUserEmail).
			Return(user.User{}, errMockStep2).
			Times(1)
	}

	step3CreateUserOk := func(
		createUserPort *mocks.MockCreateUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return createUserPort.EXPECT().
			CreateUser(targetCreatedUser).
			Return(expectedUser, nil).
			Times(1)
	}

	errMockStep3 := errors.New("unexpected error in step 3")
	step3CreateUserError := func(
		createUserPort *mocks.MockCreateUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return createUserPort.EXPECT().
			CreateUser(targetCreatedUser).
			Return(user.User{}, errMockStep3).
			Times(1)
	}

	step4CreateTokenOk := func(
		createUserPort *mocks.MockCreateUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return confirmAccountTokenPort.EXPECT().
			NewConfirmAccountToken(targetUserId).
			Return(expectedToken, nil).
			Times(1)
	}

	errMockStep4 := errors.New("unexpected error in step 4")
	step4CreateTokenError := func(
		createUserPort *mocks.MockCreateUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return confirmAccountTokenPort.EXPECT().
			NewConfirmAccountToken(targetUserId).
			Return("", errMockStep4).
			Times(1)
	}

	step5SendEmailOk := func(
		createUserPort *mocks.MockCreateUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return sendEmailPort.EXPECT().
			SendConfirmAccountEmail(targetUserEmail, expectedToken).
			Return(nil).
			Times(1)
	}

	errMockStep5 := errors.New("unexpected error in step 5")
	step5SendEmailError := func(
		createUserPort *mocks.MockCreateUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return sendEmailPort.EXPECT().
			SendConfirmAccountEmail(targetUserEmail, expectedToken).
			Return(errMockStep5).
			Times(1)
	}

	step6DeleteUserOk := func(
		createUserPort *mocks.MockCreateUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return deleteUserPort.EXPECT().
			DeleteTenantUser(targetTenantId, targetUserId).
			Return(expectedUser, nil).
			Times(1)
	}

	errMockStep6 := errors.New("unexpected error in step 6")
	step6DeleteUserError := func(
		createUserPort *mocks.MockCreateUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return deleteUserPort.EXPECT().
			DeleteTenantUser(targetTenantId, targetUserId).
			Return(user.User{}, errMockStep6).
			Times(1)
	}

	baseInput := user.CreateTenantUserCommand{
		Email:    targetUserEmail,
		Username: targetUserName,
		TenantId: targetTenantId,
	}
	cases := []testCase{
		{
			name: "Success: tenant user created successfully",
			input: baseInput,
			setupSteps: []mockSetupFunc{
				step1TenantOk,
				step2GetUserOk,
				step3CreateUserOk,
				step4CreateTokenOk,
				step5SendEmailOk,
			},
			expectedError: nil,
			expectedUser:  expectedUser,
		},
		{
			name: "Fail (step 1): tenant not found",
			input: baseInput,
			setupSteps: []mockSetupFunc{
				step1TenantFail,
			},
			expectedError: tenant.ErrTenantNotFound,
			expectedUser:  user.User{},
		},
		{
			name: "Fail (step 1): unexpected error",
			input: baseInput,
			setupSteps: []mockSetupFunc{
				step1TenantError,
			},
			expectedError: errMockStep1,
			expectedUser:  user.User{},
		},
		{
			name: "Fail (step 2): user already exists",
			input: baseInput,
			setupSteps: []mockSetupFunc{
				step1TenantOk,
				step2GetUserError,
			},
			expectedError: errMockStep2,
			expectedUser:  user.User{},
		},
		{
			name: "Fail (step 2): unexpected error",
			input: baseInput,
			setupSteps: []mockSetupFunc{
				step1TenantOk,
				step2GetExistingUserFail,
			},
			expectedError: user.ErrUserAlreadyExists,
			expectedUser:  user.User{},
		},
		{
			name: "Fail (step 3): unexpected error",
			input: baseInput,
			setupSteps: []mockSetupFunc{
				step1TenantOk,
				step2GetUserOk,
				step3CreateUserError,
			},
			expectedError: errMockStep3,
			expectedUser:  user.User{},
		},
		{
			name: "Fail (step 4): unexpected error",
			input: baseInput,
			setupSteps: []mockSetupFunc{
				step1TenantOk,
				step2GetUserOk,
				step3CreateUserOk,
				step4CreateTokenError,
			},
			expectedError: errMockStep4,
			expectedUser:  user.User{},
		},

		{
			name: "Fail (step 5): unexpected error -> Success (step 6): rolled back user",
			input: baseInput,
			setupSteps: []mockSetupFunc{
				step1TenantOk,
				step2GetUserOk,
				step3CreateUserOk,
				step4CreateTokenOk,
				step5SendEmailError,
				step6DeleteUserOk,
			},
			expectedError: user.ErrCannotSendEmail,
			expectedUser:  user.User{},
		},
		{
			name: "Fail (step 5): unexpected error -> Fail (step 6): cannot roll back user",
			input: baseInput,
			setupSteps: []mockSetupFunc{
				step1TenantOk,
				step2GetUserOk,
				step3CreateUserOk,
				step4CreateTokenOk,
				step5SendEmailError,
				step6DeleteUserError,
			},
			expectedError: errMockStep6,
			expectedUser:  user.User{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// NOTA: il controller di gomock va inizializzato qua dentro!
			mockController := gomock.NewController(t)

			mockCreatePort := mocks.NewMockCreateUserPort(mockController)
			mockDeletePort := mocks.NewMockDeleteUserPort(mockController)
			mockGetPort := mocks.NewMockGetUserPort(mockController)
			mockTenantPort := tenantMocks.NewMockGetTenantPort(mockController)
			mockConfirmTokenPort := authMocks.NewMockConfirmTokenPort(mockController)
			mockSendEmailPort := emailMocks.NewMockSendEmailPort(mockController)

			// Slice con chiamate da eseguire
			var expectedCalls []any // NOTA: Dovrebbe essere []*gomock.Call, però il compilatore non accetta

			// Collezione le chiamate per questo test case
			for _, step := range tc.setupSteps {
				call := step(mockCreatePort, mockDeletePort, mockGetPort, mockTenantPort, mockConfirmTokenPort, mockSendEmailPort)
				if call != nil {
					expectedCalls = append(expectedCalls, call)
				}
			}

			// Richiedi ordine nelle chiamate
			if len(expectedCalls) > 0 {
				gomock.InOrder(expectedCalls...)
			}
			// Initialize service
			createTenantUserUseCase, _, _ := user.NewCreateUserService(
				mockCreatePort, mockDeletePort, mockGetPort, mockTenantPort, mockConfirmTokenPort, mockSendEmailPort,
			)

			// Execute function
			createdUser, err := createTenantUserUseCase.CreateTenantUser(tc.input)

			// Assertions
			if err != tc.expectedError {
				t.Errorf("expected error %v, got %v", tc.expectedError, err)
			}
			if createdUser != tc.expectedUser {
				t.Errorf("expected user %v, got %v", tc.expectedUser, createdUser)
			}
		})
	}
}
