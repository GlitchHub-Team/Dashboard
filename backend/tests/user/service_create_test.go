package user_test

import (
	"errors"
	"testing"

	"backend/internal/shared/identity"
	"backend/internal/tenant"
	"backend/internal/user"
	authMocks "backend/tests/auth/mocks"
	emailMocks "backend/tests/email/mocks"
	tenantMocks "backend/tests/tenant/mocks"
	"backend/tests/user/mocks"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

type mockSetupFunc_CreateUserService func(
	createUserPort *mocks.MockSaveUserPort,
	deleteUserPort *mocks.MockDeleteUserPort,
	getUserPort *mocks.MockGetUserPort,
	getTenantPort *tenantMocks.MockGetTenantPort,
	confirmAccountTokenPort *authMocks.MockConfirmAccountTokenPort,
	sendEmailPort *emailMocks.MockSendEmailPort,
) *gomock.Call

func newStepTenantOk_CreateUserService(targetTenantId uuid.UUID, canImpersonate bool) mockSetupFunc_CreateUserService {
	return func(
		createUserPort *mocks.MockSaveUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmAccountTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return getTenantPort.EXPECT().
			GetTenant(targetTenantId).
			Return(tenant.Tenant{
				Id:             targetTenantId,
				CanImpersonate: canImpersonate,
			}, nil).
			Times(1)
	}
}

func newStepTenantNotFound_CreateUserService(targetTenantId uuid.UUID) mockSetupFunc_CreateUserService {
	return func(
		createUserPort *mocks.MockSaveUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmAccountTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return getTenantPort.EXPECT().
			GetTenant(targetTenantId).
			Return(tenant.Tenant{}, tenant.ErrTenantNotFound).
			Times(1)
	}
}

func newStepTenantError_CreateUserService(targetTenantId uuid.UUID, err error) mockSetupFunc_CreateUserService {
	return func(
		createUserPort *mocks.MockSaveUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmAccountTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return getTenantPort.EXPECT().
			GetTenant(targetTenantId).
			Return(tenant.Tenant{}, err).
			Times(1)
	}
}

func TestService_CreateTenantUser(t *testing.T) {
	// Dati test
	targetTenantId := uuid.New()
	otherTenantId := uuid.New()
	targetUserId := uint(100)
	targetUserEmail := "test@example.com"
	targetUserName := "Test"
	targetConfirmed := false
	targetRole := identity.ROLE_TENANT_USER

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

	type testCase struct {
		name          string
		input         user.CreateTenantUserCommand
		setupSteps    []mockSetupFunc_CreateUserService
		expectedUser  user.User
		expectedError error
	}

	// Step 1: Cercare tenant
	step1TenantOk_CanImpersonate := newStepTenantOk_CreateUserService(targetTenantId, true)
	step1TenantOk_CannotImpersonate := newStepTenantOk_CreateUserService(targetTenantId, false)

	step1TenantNotFound := newStepTenantNotFound_CreateUserService(targetTenantId)

	errMockStep1 := errors.New("unexpected error in step 1")
	step1TenantError := newStepTenantError_CreateUserService(targetTenantId, errMockStep1)

	// Step 2: Get user
	step2GetUserOk := func(
		createUserPort *mocks.MockSaveUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmAccountTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantUserByEmail(targetTenantId, targetUserEmail).
			Return(user.User{}, nil).
			Times(1)
	}

	step2UserExistsFail := func(
		createUserPort *mocks.MockSaveUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmAccountTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantUserByEmail(targetTenantId, targetUserEmail).
			Return(expectedUser, nil).
			Times(1)
	}

	step2NeverCalled := func(
		createUserPort *mocks.MockSaveUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmAccountTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantUserByEmail(gomock.Any(), gomock.Any()).
			Times(0)
	}

	errMockStep2 := errors.New("unexpected error in step 2")
	step2GetUserError := func(
		createUserPort *mocks.MockSaveUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmAccountTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantUserByEmail(targetTenantId, targetUserEmail).
			Return(user.User{}, errMockStep2).
			Times(1)
	}

	// Step 3: create user
	step3CreateUserOk := func(
		createUserPort *mocks.MockSaveUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmAccountTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return createUserPort.EXPECT().
			SaveUser(targetCreatedUser).
			Return(expectedUser, nil).
			Times(1)
	}

	errMockStep3 := errors.New("unexpected error in step 3")
	step3CreateUserError := func(
		createUserPort *mocks.MockSaveUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmAccountTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return createUserPort.EXPECT().
			SaveUser(targetCreatedUser).
			Return(user.User{}, errMockStep3).
			Times(1)
	}

	// Step 4: create token
	step4CreateTokenOk := func(
		createUserPort *mocks.MockSaveUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmAccountTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return confirmAccountTokenPort.EXPECT().
			NewConfirmAccountToken(&targetTenantId, targetUserId).
			Return(expectedToken, nil).
			Times(1)
	}

	errMockStep4 := errors.New("unexpected error in step 4")
	step4CreateTokenError := func(
		createUserPort *mocks.MockSaveUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmAccountTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return confirmAccountTokenPort.EXPECT().
			NewConfirmAccountToken(&targetTenantId, targetUserId).
			Return("", errMockStep4).
			Times(1)
	}

	// Step 5: Send email
	step5SendEmailOk := func(
		createUserPort *mocks.MockSaveUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmAccountTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return sendEmailPort.EXPECT().
			SendConfirmAccountEmail(targetUserEmail, expectedToken).
			Return(nil).
			Times(1)
	}

	errMockStep5 := errors.New("unexpected error in step 5")
	step5SendEmailError := func(
		createUserPort *mocks.MockSaveUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmAccountTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return sendEmailPort.EXPECT().
			SendConfirmAccountEmail(targetUserEmail, expectedToken).
			Return(errMockStep5).
			Times(1)
	}

	// Step 6: rollback user
	step6DeleteUserOk := func(
		createUserPort *mocks.MockSaveUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmAccountTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return deleteUserPort.EXPECT().
			DeleteTenantUser(targetTenantId, targetUserId).
			Return(expectedUser, nil).
			Times(1)
	}

	errMockStep6 := errors.New("unexpected error in step 6")
	step6DeleteUserError := func(
		createUserPort *mocks.MockSaveUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmAccountTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return deleteUserPort.EXPECT().
			DeleteTenantUser(targetTenantId, targetUserId).
			Return(user.User{}, errMockStep6).
			Times(1)
	}

	// Requesters
	superAdminRequester := identity.Requester{
		RequesterUserId: uint(1),
		RequesterRole:   identity.ROLE_SUPER_ADMIN,
	}

	authorizedTenantAdminRequester := identity.Requester{
		RequesterUserId:   uint(1),
		RequesterTenantId: &targetTenantId,
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	unauthorizedTenantAdminRequester := identity.Requester{
		RequesterUserId:   uint(1),
		RequesterTenantId: &otherTenantId,
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	tenantUserRequester := identity.Requester{
		RequesterUserId:   uint(1),
		RequesterTenantId: &targetTenantId,
		RequesterRole:     identity.ROLE_TENANT_USER,
	}

	baseInput := user.CreateTenantUserCommand{
		Email:    targetUserEmail,
		Username: targetUserName,
		TenantId: targetTenantId,
	}

	inputWith := func(requester identity.Requester) user.CreateTenantUserCommand {
		return user.CreateTenantUserCommand{
			Requester: requester,
			Email:     baseInput.Email,
			Username:  baseInput.Username,
			TenantId:  baseInput.TenantId,
		}
	}

	cases := []testCase{
		// Successo
		{
			name:  "(Super Admin) Success: impersonation OK",
			input: inputWith(superAdminRequester),
			setupSteps: []mockSetupFunc_CreateUserService{
				step1TenantOk_CanImpersonate,
				step2GetUserOk,
				step3CreateUserOk,
				step4CreateTokenOk,
				step5SendEmailOk,
			},
			expectedError: nil,
			expectedUser:  expectedUser,
		},
		{
			name:  "(Tenant Admin) Success: authorized, OK",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_CreateUserService{
				step1TenantOk_CanImpersonate, // NOTA: impersonazione qui irrilevante
				step2GetUserOk,
				step3CreateUserOk,
				step4CreateTokenOk,
				step5SendEmailOk,
			},
			expectedError: nil,
			expectedUser:  expectedUser,
		},
		// Step 1: test get tenant (non importa il requester)
		{
			name:  "Fail (step 1): tenant not found",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_CreateUserService{
				step1TenantNotFound,
			},
			expectedError: tenant.ErrTenantNotFound,
			expectedUser:  user.User{},
		},
		{
			name:  "Fail (step 1): unexpected error",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_CreateUserService{
				step1TenantError,
			},
			expectedError: errMockStep1,
			expectedUser:  user.User{},
		},

		// Step 1: test autorizzazione
		{
			name:  "(Super Admin) Fail (step 1): impersonation fail",
			input: inputWith(superAdminRequester),
			setupSteps: []mockSetupFunc_CreateUserService{
				step1TenantOk_CannotImpersonate,
				step2NeverCalled,
			},
			expectedError: identity.ErrUnauthorizedAccess,
			expectedUser:  user.User{},
		},
		{
			name:  "(Tenant Admin) Fail (step 1): unauthorized access",
			input: inputWith(unauthorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_CreateUserService{
				step1TenantOk_CanImpersonate, // NOTA: impersonazione qui irrilevante
				step2NeverCalled,
			},
			expectedError: identity.ErrUnauthorizedAccess,
			expectedUser:  user.User{},
		},
		{
			name:  "(Tenant User) Fail (step 1): unauthorized access",
			input: inputWith(tenantUserRequester),
			setupSteps: []mockSetupFunc_CreateUserService{
				step1TenantOk_CanImpersonate, // NOTA: impersonazione qui irrilevante
				step2NeverCalled,
			},
			expectedError: identity.ErrUnauthorizedAccess,
			expectedUser:  user.User{},
		},

		// Step 2: test get user
		// NOTA: Da qui in poi, non importa se il requester è tenant admin o super admin impersonante,
		// usiamo tenant admin
		{
			name:  "Fail (step 2): unexpected error",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_CreateUserService{
				step1TenantOk_CanImpersonate,
				step2GetUserError,
			},
			expectedError: errMockStep2,
			expectedUser:  user.User{},
		},
		{
			name:  "Fail (step 2): user already exists",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_CreateUserService{
				step1TenantOk_CanImpersonate,
				step2UserExistsFail,
			},
			expectedError: user.ErrUserAlreadyExists,
			expectedUser:  user.User{},
		},

		// Step 3: test create user
		{
			name:  "Fail (step 3): unexpected error",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_CreateUserService{
				step1TenantOk_CanImpersonate,
				step2GetUserOk,
				step3CreateUserError,
			},
			expectedError: errMockStep3,
			expectedUser:  user.User{},
		},

		// Step 4: test create token
		{
			name:  "Fail (step 4): unexpected error",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_CreateUserService{
				step1TenantOk_CanImpersonate,
				step2GetUserOk,
				step3CreateUserOk,
				step4CreateTokenError,
			},
			expectedError: errMockStep4,
			expectedUser:  user.User{},
		},

		{
			name:  "Fail (step 5): unexpected error -> Success (step 6): rolled back user",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_CreateUserService{
				step1TenantOk_CanImpersonate,
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
			name:  "Fail (step 5): unexpected error -> Fail (step 6): cannot roll back user",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_CreateUserService{
				step1TenantOk_CanImpersonate,
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

			mockCreatePort := mocks.NewMockSaveUserPort(mockController)
			mockDeletePort := mocks.NewMockDeleteUserPort(mockController)
			mockGetPort := mocks.NewMockGetUserPort(mockController)
			mockTenantPort := tenantMocks.NewMockGetTenantPort(mockController)
			mockConfirmTokenPort := authMocks.NewMockConfirmAccountTokenPort(mockController)
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

			// Crea servizio con porte mock
			createTenantUserUseCase, _, _ := user.NewCreateUserService(
				mockCreatePort, mockDeletePort, mockGetPort, mockTenantPort, mockConfirmTokenPort, mockSendEmailPort,
			)

			// Esegui funzione in oggetto
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

func TestService_CreateTenantAdmin(t *testing.T) {
	// Dati test
	targetTenantId := uuid.New()
	otherTenantId := uuid.New()
	targetUserId := uint(100)
	targetUserEmail := "test@example.com"
	targetUserName := "Test"
	targetConfirmed := false
	targetRole := identity.ROLE_TENANT_ADMIN

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

	type testCase struct {
		name          string
		input         user.CreateTenantAdminCommand
		setupSteps    []mockSetupFunc_CreateUserService
		expectedUser  user.User
		expectedError error
	}

	// Step 1: get tenant
	step1TenantOk_CanImpersonate := newStepTenantOk_CreateUserService(targetTenantId, true)
	step1TenantOk_CannotImpersonate := newStepTenantOk_CreateUserService(targetTenantId, false)

	step1TenantNotFound := newStepTenantNotFound_CreateUserService(targetTenantId)

	errMockStep1 := errors.New("unexpected error in step 1")
	step1TenantError := newStepTenantError_CreateUserService(targetTenantId, errMockStep1)

	// Step 2: get user
	step2GetUserOk := func(
		createUserPort *mocks.MockSaveUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmAccountTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantAdminByEmail(targetTenantId, targetUserEmail).
			Return(user.User{}, nil).
			Times(1)
	}

	step2UserExistsFail := func(
		createUserPort *mocks.MockSaveUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmAccountTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantAdminByEmail(targetTenantId, targetUserEmail).
			Return(expectedUser, nil).
			Times(1)
	}

	step2NeverCalled := func(
		createUserPort *mocks.MockSaveUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmAccountTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantAdminByEmail(gomock.Any(), gomock.Any()).
			Times(0)
	}

	errMockStep2 := errors.New("unexpected error in step 2")
	step2GetUserError := func(
		createUserPort *mocks.MockSaveUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmAccountTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantAdminByEmail(targetTenantId, targetUserEmail).
			Return(user.User{}, errMockStep2).
			Times(1)
	}

	// Step 3: create user
	step3CreateUserOk := func(
		createUserPort *mocks.MockSaveUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmAccountTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return createUserPort.EXPECT().
			SaveUser(targetCreatedUser).
			Return(expectedUser, nil).
			Times(1)
	}

	errMockStep3 := errors.New("unexpected error in step 3")
	step3CreateUserError := func(
		createUserPort *mocks.MockSaveUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmAccountTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return createUserPort.EXPECT().
			SaveUser(targetCreatedUser).
			Return(user.User{}, errMockStep3).
			Times(1)
	}

	// Step 4: create token
	step4CreateTokenOk := func(
		createUserPort *mocks.MockSaveUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmAccountTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return confirmAccountTokenPort.EXPECT().
			NewConfirmAccountToken(&targetTenantId, targetUserId).
			Return(expectedToken, nil).
			Times(1)
	}

	errMockStep4 := errors.New("unexpected error in step 4")
	step4CreateTokenError := func(
		createUserPort *mocks.MockSaveUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmAccountTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return confirmAccountTokenPort.EXPECT().
			NewConfirmAccountToken(&targetTenantId, targetUserId).
			Return("", errMockStep4).
			Times(1)
	}

	// Step 5: send email
	step5SendEmailOk := func(
		createUserPort *mocks.MockSaveUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmAccountTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return sendEmailPort.EXPECT().
			SendConfirmAccountEmail(targetUserEmail, expectedToken).
			Return(nil).
			Times(1)
	}

	errMockStep5 := errors.New("unexpected error in step 5")
	step5SendEmailError := func(
		createUserPort *mocks.MockSaveUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmAccountTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return sendEmailPort.EXPECT().
			SendConfirmAccountEmail(targetUserEmail, expectedToken).
			Return(errMockStep5).
			Times(1)
	}

	// Step 6: rollback user
	step6DeleteUserOk := func(
		createUserPort *mocks.MockSaveUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmAccountTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return deleteUserPort.EXPECT().
			DeleteTenantAdmin(targetTenantId, targetUserId).
			Return(expectedUser, nil).
			Times(1)
	}

	errMockStep6 := errors.New("unexpected error in step 6")
	step6DeleteUserError := func(
		createUserPort *mocks.MockSaveUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmAccountTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return deleteUserPort.EXPECT().
			DeleteTenantAdmin(targetTenantId, targetUserId).
			Return(user.User{}, errMockStep6).
			Times(1)
	}

	// Requesters
	superAdminRequester := identity.Requester{
		RequesterUserId: uint(1),
		RequesterRole:   identity.ROLE_SUPER_ADMIN,
	}

	authorizedTenantAdminRequester := identity.Requester{
		RequesterUserId:   uint(1),
		RequesterTenantId: &targetTenantId,
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	unauthorizedTenantAdminRequester := identity.Requester{
		RequesterUserId:   uint(1),
		RequesterTenantId: &otherTenantId,
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	tenantUserRequester := identity.Requester{
		RequesterUserId:   uint(1),
		RequesterTenantId: &targetTenantId,
		RequesterRole:     identity.ROLE_TENANT_USER,
	}

	baseInput := user.CreateTenantAdminCommand{
		Email:    targetUserEmail,
		Username: targetUserName,
		TenantId: targetTenantId,
	}

	inputWith := func(requester identity.Requester) user.CreateTenantAdminCommand {
		return user.CreateTenantAdminCommand{
			Requester: requester,
			Email:     baseInput.Email,
			Username:  baseInput.Username,
			TenantId:  baseInput.TenantId,
		}
	}

	cases := []testCase{
		// Successo
		{
			name:  "(Super Admin) Success: impersonation OK",
			input: inputWith(superAdminRequester),
			setupSteps: []mockSetupFunc_CreateUserService{
				step1TenantOk_CanImpersonate,
				step2GetUserOk,
				step3CreateUserOk,
				step4CreateTokenOk,
				step5SendEmailOk,
			},
			expectedError: nil,
			expectedUser:  expectedUser,
		},
		{
			name:  "(Tenant Admin) Success: authorized, OK",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_CreateUserService{
				step1TenantOk_CanImpersonate, // NOTA: impersonazione qui irrilevante
				step2GetUserOk,
				step3CreateUserOk,
				step4CreateTokenOk,
				step5SendEmailOk,
			},
			expectedError: nil,
			expectedUser:  expectedUser,
		},
		// Step 1: test get tenant (non importa il requester)
		{
			name:  "Fail (step 1): tenant not found",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_CreateUserService{
				step1TenantNotFound,
			},
			expectedError: tenant.ErrTenantNotFound,
			expectedUser:  user.User{},
		},
		{
			name:  "Fail (step 1): unexpected error",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_CreateUserService{
				step1TenantError,
			},
			expectedError: errMockStep1,
			expectedUser:  user.User{},
		},

		// Step 1: test autorizzazione
		{
			name:  "(Super Admin) Fail (step 1): impersonation fail",
			input: inputWith(superAdminRequester),
			setupSteps: []mockSetupFunc_CreateUserService{
				step1TenantOk_CannotImpersonate,
				step2NeverCalled,
			},
			expectedError: identity.ErrUnauthorizedAccess,
			expectedUser:  user.User{},
		},
		{
			name:  "(Tenant Admin) Fail (step 1): unauthorized access",
			input: inputWith(unauthorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_CreateUserService{
				step1TenantOk_CanImpersonate, // NOTA: impersonazione qui irrilevante
				step2NeverCalled,
			},
			expectedError: identity.ErrUnauthorizedAccess,
			expectedUser:  user.User{},
		},
		{
			name:  "(Tenant User) Fail (step 1): unauthorized access",
			input: inputWith(tenantUserRequester),
			setupSteps: []mockSetupFunc_CreateUserService{
				step1TenantOk_CanImpersonate, // NOTA: impersonazione qui irrilevante
				step2NeverCalled,
			},
			expectedError: identity.ErrUnauthorizedAccess,
			expectedUser:  user.User{},
		},

		// Step 2: test get user
		// NOTA: Da qui in poi, non importa se il requester è tenant admin o super admin impersonante,
		// usiamo tenant admin
		{
			name:  "Fail (step 2): unexpected error",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_CreateUserService{
				step1TenantOk_CanImpersonate,
				step2GetUserError,
			},
			expectedError: errMockStep2,
			expectedUser:  user.User{},
		},
		{
			name:  "Fail (step 2): user already exists",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_CreateUserService{
				step1TenantOk_CanImpersonate,
				step2UserExistsFail,
			},
			expectedError: user.ErrUserAlreadyExists,
			expectedUser:  user.User{},
		},

		// Step 3: test create user
		{
			name:  "Fail (step 3): unexpected error",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_CreateUserService{
				step1TenantOk_CanImpersonate,
				step2GetUserOk,
				step3CreateUserError,
			},
			expectedError: errMockStep3,
			expectedUser:  user.User{},
		},

		// Step 4: test create token
		{
			name:  "Fail (step 4): unexpected error",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_CreateUserService{
				step1TenantOk_CanImpersonate,
				step2GetUserOk,
				step3CreateUserOk,
				step4CreateTokenError,
			},
			expectedError: errMockStep4,
			expectedUser:  user.User{},
		},

		{
			name:  "Fail (step 5): unexpected error -> Success (step 6): rolled back user",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_CreateUserService{
				step1TenantOk_CanImpersonate,
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
			name:  "Fail (step 5): unexpected error -> Fail (step 6): cannot roll back user",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_CreateUserService{
				step1TenantOk_CanImpersonate,
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

			mockCreatePort := mocks.NewMockSaveUserPort(mockController)
			mockDeletePort := mocks.NewMockDeleteUserPort(mockController)
			mockGetPort := mocks.NewMockGetUserPort(mockController)
			mockTenantPort := tenantMocks.NewMockGetTenantPort(mockController)
			mockConfirmTokenPort := authMocks.NewMockConfirmAccountTokenPort(mockController)
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

			// Crea servizio con porte mock
			_, createTenantAdminUseCase, _ := user.NewCreateUserService(
				mockCreatePort, mockDeletePort, mockGetPort, mockTenantPort, mockConfirmTokenPort, mockSendEmailPort,
			)

			// Esegui funzione in oggetto
			createdUser, err := createTenantAdminUseCase.CreateTenantAdmin(tc.input)

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

func TestService_CreateSuperAdmin(t *testing.T) {
	// Dati test
	targetTenantId := (*uuid.UUID)(nil)
	targetUserId := uint(1)
	targetUserEmail := "test@example.com"
	targetUserName := "Test"
	targetConfirmed := false
	targetRole := identity.ROLE_SUPER_ADMIN

	targetCreatedUser := user.User{
		Name:      targetUserName,
		Email:     targetUserEmail,
		Role:      targetRole,
		TenantId:  targetTenantId,
		Confirmed: targetConfirmed,
	}

	expectedUser := user.User{
		Id:        targetUserId,
		Name:      targetUserName,
		Email:     targetUserEmail,
		TenantId:  targetTenantId,
		Confirmed: targetConfirmed,
		Role:      targetRole,
	}
	expectedToken := "token"

	type testCase struct {
		name          string
		input         user.CreateSuperAdminCommand
		setupSteps    []mockSetupFunc_CreateUserService
		expectedUser  user.User
		expectedError error
	}

	// Step 1: get user
	step1GetUserOk := func(
		createUserPort *mocks.MockSaveUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmAccountTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetSuperAdminByEmail(targetUserEmail).
			Return(user.User{}, nil).
			Times(1)
	}

	step1GetExistingUserFail := func(
		createUserPort *mocks.MockSaveUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmAccountTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetSuperAdminByEmail(targetUserEmail).
			Return(expectedUser, nil).
			Times(1)
	}

	errMockStep1 := errors.New("unexpected error in step 1")
	step1GetUserError := func(
		createUserPort *mocks.MockSaveUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmAccountTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetSuperAdminByEmail(targetUserEmail).
			Return(user.User{}, errMockStep1).
			Times(1)
	}

	step1NeverCalled := func(
		createUserPort *mocks.MockSaveUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmAccountTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetSuperAdminByEmail(gomock.Any()).
			Times(0)
	}

	// step 2: create user
	step2CreateUserOk := func(
		createUserPort *mocks.MockSaveUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmAccountTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return createUserPort.EXPECT().
			SaveUser(targetCreatedUser).
			Return(expectedUser, nil).
			Times(1)
	}

	errMockStep2 := errors.New("unexpected error in step 3")
	step2CreateUserError := func(
		createUserPort *mocks.MockSaveUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmAccountTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return createUserPort.EXPECT().
			SaveUser(targetCreatedUser).
			Return(user.User{}, errMockStep2).
			Times(1)
	}

	// Step 3: create token
	step3CreateTokenOk := func(
		createUserPort *mocks.MockSaveUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmAccountTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return confirmAccountTokenPort.EXPECT().
			NewConfirmAccountToken(targetTenantId, targetUserId).
			Return(expectedToken, nil).
			Times(1)
	}

	errMockStep3 := errors.New("unexpected error in step 4")
	step3CreateTokenError := func(
		createUserPort *mocks.MockSaveUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmAccountTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return confirmAccountTokenPort.EXPECT().
			NewConfirmAccountToken(targetTenantId, targetUserId).
			Return("", errMockStep3).
			Times(1)
	}

	// Step 4: send email
	step4SendEmailOk := func(
		createUserPort *mocks.MockSaveUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmAccountTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return sendEmailPort.EXPECT().
			SendConfirmAccountEmail(targetUserEmail, expectedToken).
			Return(nil).
			Times(1)
	}

	errMockStep4 := errors.New("unexpected error in step 5")
	step4SendEmailError := func(
		createUserPort *mocks.MockSaveUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmAccountTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return sendEmailPort.EXPECT().
			SendConfirmAccountEmail(targetUserEmail, expectedToken).
			Return(errMockStep4).
			Times(1)
	}

	// Step 5: rollback user
	step5DeleteUserOk := func(
		createUserPort *mocks.MockSaveUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmAccountTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return deleteUserPort.EXPECT().
			DeleteSuperAdmin(targetUserId).
			Return(expectedUser, nil).
			Times(1)
	}

	errMockStep5 := errors.New("unexpected error in step 6")
	step5DeleteUserError := func(
		createUserPort *mocks.MockSaveUserPort, deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort, confirmAccountTokenPort *authMocks.MockConfirmAccountTokenPort, sendEmailPort *emailMocks.MockSendEmailPort,
	) *gomock.Call {
		return deleteUserPort.EXPECT().
			DeleteSuperAdmin(targetUserId).
			Return(user.User{}, errMockStep5).
			Times(1)
	}

	// Requesters

	superAdminRequester := identity.Requester{
		RequesterUserId: uint(1),
		RequesterRole:   identity.ROLE_SUPER_ADMIN,
	}

	exampleTenantId := uuid.New()
	tenantAdminRequester := identity.Requester{
		RequesterUserId:   uint(1),
		RequesterTenantId: &exampleTenantId,
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	tenantUserRequester := identity.Requester{
		RequesterUserId:   uint(1),
		RequesterTenantId: &exampleTenantId,
		RequesterRole:     identity.ROLE_TENANT_USER,
	}

	baseInput := user.CreateTenantAdminCommand{
		Email:    targetUserEmail,
		Username: targetUserName,
	}

	inputWith := func(requester identity.Requester) user.CreateSuperAdminCommand {
		return user.CreateSuperAdminCommand{
			Requester: requester,
			Email:     baseInput.Email,
			Username:  baseInput.Username,
		}
	}

	cases := []testCase{
		{
			name:  "Success",
			input: inputWith(superAdminRequester),
			setupSteps: []mockSetupFunc_CreateUserService{
				step1GetUserOk,
				step2CreateUserOk,
				step3CreateTokenOk,
				step4SendEmailOk,
			},
			expectedError: nil,
			expectedUser:  expectedUser,
		},
		// Test autorizzazione
		{
			name:  "(Tenant Admin) Fail: unauthorized access",
			input: inputWith(tenantAdminRequester),
			setupSteps: []mockSetupFunc_CreateUserService{
				step1NeverCalled,
			},
			expectedError: identity.ErrUnauthorizedAccess,
			expectedUser:  user.User{},
		},
		{
			name:  "(Tenant User) Fail: unauthorized access",
			input: inputWith(tenantUserRequester),
			setupSteps: []mockSetupFunc_CreateUserService{
				step1NeverCalled,
			},
			expectedError: identity.ErrUnauthorizedAccess,
			expectedUser:  user.User{},
		},

		// Test 1
		{
			name:  "Fail (step 1): unexpected error",
			input: inputWith(superAdminRequester),
			setupSteps: []mockSetupFunc_CreateUserService{
				step1GetUserError,
			},
			expectedError: errMockStep1,
			expectedUser:  user.User{},
		},
		{
			name:  "Fail (step 1): user already exists",
			input: inputWith(superAdminRequester),
			setupSteps: []mockSetupFunc_CreateUserService{
				step1GetExistingUserFail,
			},
			expectedError: user.ErrUserAlreadyExists,
			expectedUser:  user.User{},
		},
		{
			name:  "Fail (step 2): unexpected error",
			input: inputWith(superAdminRequester),
			setupSteps: []mockSetupFunc_CreateUserService{
				step1GetUserOk,
				step2CreateUserError,
			},
			expectedError: errMockStep2,
			expectedUser:  user.User{},
		},
		{
			name:  "Fail (step 3): unexpected error",
			input: inputWith(superAdminRequester),
			setupSteps: []mockSetupFunc_CreateUserService{
				step1GetUserOk,
				step2CreateUserOk,
				step3CreateTokenError,
			},
			expectedError: errMockStep3,
			expectedUser:  user.User{},
		},

		{
			name:  "Fail (step 4): unexpected error -> Success (step 5): rolled back user",
			input: inputWith(superAdminRequester),
			setupSteps: []mockSetupFunc_CreateUserService{
				step1GetUserOk,
				step2CreateUserOk,
				step3CreateTokenOk,
				step4SendEmailError,
				step5DeleteUserOk,
			},
			expectedError: user.ErrCannotSendEmail,
			expectedUser:  user.User{},
		},
		{
			name:  "Fail (step 4): unexpected error -> Fail (step 5): cannot roll back user",
			input: inputWith(superAdminRequester),
			setupSteps: []mockSetupFunc_CreateUserService{
				step1GetUserOk,
				step2CreateUserOk,
				step3CreateTokenOk,
				step4SendEmailError,
				step5DeleteUserError,
			},
			expectedError: errMockStep5,
			expectedUser:  user.User{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// NOTA: il controller di gomock va inizializzato qua dentro!
			mockController := gomock.NewController(t)

			mockCreatePort := mocks.NewMockSaveUserPort(mockController)
			mockDeletePort := mocks.NewMockDeleteUserPort(mockController)
			mockGetPort := mocks.NewMockGetUserPort(mockController)
			mockTenantPort := tenantMocks.NewMockGetTenantPort(mockController)
			mockConfirmTokenPort := authMocks.NewMockConfirmAccountTokenPort(mockController)
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
			// Crea servizio con porte mock
			_, _, createSuperAdminUseCase := user.NewCreateUserService(
				mockCreatePort, mockDeletePort, mockGetPort, mockTenantPort, mockConfirmTokenPort, mockSendEmailPort,
			)

			// Esegui funzione in oggetto
			createdUser, err := createSuperAdminUseCase.CreateSuperAdmin(tc.input)

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
