package user_test

import (
	"testing"

	"backend/internal/shared/identity"
	"backend/internal/tenant"
	"backend/internal/user"
	tenantmocks "backend/tests/tenant/mocks"
	"backend/tests/user/mocks"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

type mockSetupFunc_DeleteUserService func(
	deleteUserPort *mocks.MockDeleteUserPort,
	getUserPort *mocks.MockGetUserPort,
	getTenantPort *tenantmocks.MockGetTenantPort,
) *gomock.Call

func newStepTenantOk_DeleteUserService(targetTenantId uuid.UUID, canImpersonate bool) mockSetupFunc_DeleteUserService {
	return func(
		deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantmocks.MockGetTenantPort,
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

func newStepTenantNotFound_DeleteUserService(targetTenantId uuid.UUID) mockSetupFunc_DeleteUserService {
	return func(
		deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantmocks.MockGetTenantPort,
	) *gomock.Call {
		return getTenantPort.EXPECT().
			GetTenant(targetTenantId).
			Return(tenant.Tenant{}, nil).
			Times(1)
	}
}

func newStepTenantError_DeleteUserService(targetTenantId uuid.UUID, err error) mockSetupFunc_DeleteUserService {
	return func(
		deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantmocks.MockGetTenantPort,
	) *gomock.Call {
		return getTenantPort.EXPECT().
			GetTenant(targetTenantId).
			Return(tenant.Tenant{}, err).
			Times(1)
	}
}

func TestService_DeleteTenantUser(t *testing.T) {
	// Dati test
	targetTenantId := uuid.New()
	otherTenantId := uuid.New()
	targetUserId := uint(1)
	expectedUser := user.User{
		Id:       targetUserId,
		Name:     "Test",
		TenantId: &targetTenantId,
	}

	type testCase struct {
		name          string
		input         user.DeleteTenantUserCommand
		setupSteps    []mockSetupFunc_DeleteUserService
		expectedError error
		expectedUser  user.User
	}

	// Steps
	// Step 1: get tenant
	step1TenantOk_CanImpersonate := newStepTenantOk_DeleteUserService(targetTenantId, true)

	step1TenantOk_CannotImpersonate := newStepTenantOk_DeleteUserService(targetTenantId, false)

	step1TenantFail := newStepTenantNotFound_DeleteUserService(targetTenantId)

	errMockStep1 := newMockError(1)
	step1TenantError := newStepTenantError_DeleteUserService(targetTenantId, errMockStep1)

	// Step 2: get user
	step2GetUserOk := func(
		deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantmocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantUser(targetTenantId, targetUserId).
			Return(expectedUser, nil).
			Times(1)
	}

	step2GetUserFail := func(
		deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantmocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantUser(targetTenantId, targetUserId).
			Return(user.User{}, nil).
			Times(1)
	}

	errMockStep2 := newMockError(2)
	step2GetUserError := func(
		deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantmocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantUser(targetTenantId, targetUserId).
			Return(user.User{}, errMockStep2).
			Times(1)
	}

	step2NeverCalled := func(
		deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantmocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantUser(gomock.Any(), gomock.Any()).
			Times(0)
	}

	// Step 3: delete user
	step3DeleteUserOk := func(
		deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantmocks.MockGetTenantPort,
	) *gomock.Call {
		return deleteUserPort.EXPECT().
			DeleteTenantUser(targetTenantId, targetUserId).
			Return(expectedUser, nil).
			Times(1)
	}

	errMockStep3 := newMockError(3)
	step3DeleteUserError := func(
		deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantmocks.MockGetTenantPort,
	) *gomock.Call {
		return deleteUserPort.EXPECT().
			DeleteTenantUser(targetTenantId, targetUserId).
			Return(user.User{}, errMockStep3).
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

	baseInput := user.DeleteTenantUserCommand{
		TenantId: targetTenantId,
		UserId:   targetUserId,
	}

	inputWith := func(requester identity.Requester) user.DeleteTenantUserCommand {
		return user.DeleteTenantUserCommand{
			Requester: requester,
			TenantId:  baseInput.TenantId,
			UserId:    baseInput.UserId,
		}
	}

	cases := []testCase{
		// Successo
		{
			name:  "(Super Admin) Success: delete tenant user, impersonation ok",
			input: inputWith(superAdminRequester),
			setupSteps: []mockSetupFunc_DeleteUserService{
				step1TenantOk_CanImpersonate,
				step2GetUserOk,
				step3DeleteUserOk,
			},
			expectedError: nil,
			expectedUser:  expectedUser,
		},
		{
			name:  "(Tenant Admin) Success: delete tenant user, authorization ok",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_DeleteUserService{
				step1TenantOk_CanImpersonate, // NOTA: impersonazione irrilevante
				step2GetUserOk,
				step3DeleteUserOk,
			},
			expectedError: nil,
			expectedUser:  expectedUser,
		},

		// Step 1: get tenant
		// NOTA: qua requester non conta
		{
			name:  "Fail (step 1): tenant not found",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_DeleteUserService{
				step1TenantFail,
			},
			expectedError: tenant.ErrTenantNotFound,
			expectedUser:  user.User{},
		},
		{
			name:  "Fail (step 1): unexpected error",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_DeleteUserService{
				step1TenantError,
			},
			expectedError: errMockStep1,
			expectedUser:  user.User{},
		},

		// Step 1: autorizzazione
		{
			name:  "(Super Admin) Fail (step 1 auth): impersonation fail",
			input: inputWith(superAdminRequester),
			setupSteps: []mockSetupFunc_DeleteUserService{
				step1TenantOk_CannotImpersonate,
				step2NeverCalled,
			},
			expectedError: identity.ErrUnauthorizedAccess,
			expectedUser:  user.User{},
		},
		{
			name:  "(Tenant Admin) Fail (step 1 auth): unauthorized access",
			input: inputWith(unauthorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_DeleteUserService{
				step1TenantOk_CanImpersonate, // NOTA: impersonazione irrilevante
				step2NeverCalled,
			},
			expectedError: identity.ErrUnauthorizedAccess,
			expectedUser:  user.User{},
		},
		{
			name:  "(Tenant User) Fail (step 1 auth): unauthorized access",
			input: inputWith(tenantUserRequester),
			setupSteps: []mockSetupFunc_DeleteUserService{
				step1TenantOk_CanImpersonate, // NOTA: impersonazione irrilevante
				step2NeverCalled,
			},
			expectedError: identity.ErrUnauthorizedAccess,
			expectedUser:  user.User{},
		},

		// Step 2: get user
		// NOTA: da qui il requester non importa
		{
			name:  "Fail (step 2): user not found",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_DeleteUserService{
				step1TenantOk_CanImpersonate, // NOTA: impersonazione irrilevante
				step2GetUserFail,
			},
			expectedError: user.ErrUserNotFound,
			expectedUser:  user.User{},
		},
		{
			name:  "Fail (step 2): unexpected error",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_DeleteUserService{
				step1TenantOk_CanImpersonate, // NOTA: impersonazione irrilevante
				step2GetUserError,
			},
			expectedError: errMockStep2,
			expectedUser:  user.User{},
		},

		// Step 3: delete user
		{
			name:  "Fail (step 3): unexpected error",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_DeleteUserService{
				step1TenantOk_CanImpersonate, // NOTA: impersonazione irrilevante
				step2GetUserOk,
				step3DeleteUserError,
			},
			expectedError: errMockStep3,
			expectedUser:  user.User{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// NOTA: il controller di gomock va inizializzato qua dentro!
			mockController := gomock.NewController(t)

			mockDeletePort := mocks.NewMockDeleteUserPort(mockController)
			mockGetUserPort := mocks.NewMockGetUserPort(mockController)
			mockGetTenantPort := tenantmocks.NewMockGetTenantPort(mockController)

			// Slice con chiamate da eseguire
			var expectedCalls []any // NOTA: Dovrebbe essere []*gomock.Call, però il compilatore non accetta

			// Collezione le chiamate per questo test case
			for _, step := range tc.setupSteps {
				call := step(mockDeletePort, mockGetUserPort, mockGetTenantPort)
				if call != nil {
					expectedCalls = append(expectedCalls, call)
				}
			}

			// Richiedi ordine nelle chiamate
			if len(expectedCalls) > 0 {
				gomock.InOrder(expectedCalls...)
			}

			// Crea servizio con porte mock
			deleteTenantUserUseCase, _, _ := user.NewDeleteUserService(
				mockDeletePort, mockGetUserPort, mockGetTenantPort,
			)

			// Esegui funzione in oggetto
			createdUser, err := deleteTenantUserUseCase.DeleteTenantUser(tc.input)

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

func TestService_DeleteTenantAdmin(t *testing.T) {
	// Dati test
	targetTenantId := uuid.New()
	otherTenantId := uuid.New()
	targetUserId := uint(1)
	expectedUser := user.User{
		Id:       targetUserId,
		Name:     "Test",
		TenantId: &targetTenantId,
	}

	type testCase struct {
		name          string
		input         user.DeleteTenantAdminCommand
		setupSteps    []mockSetupFunc_DeleteUserService
		expectedError error
		expectedUser  user.User
	}

	// Steps
	// Step 1: get tenant
	step1TenantOk_CanImpersonate := newStepTenantOk_DeleteUserService(targetTenantId, true)

	step1TenantOk_CannotImpersonate := newStepTenantOk_DeleteUserService(targetTenantId, false)

	step1TenantFail := newStepTenantNotFound_DeleteUserService(targetTenantId)

	errMockStep1 := newMockError(1)
	step1TenantError := newStepTenantError_DeleteUserService(targetTenantId, errMockStep1)

	// Step 2: get user 
	step2GetUserOk := func(
		deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantmocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantAdmin(targetTenantId, targetUserId).
			Return(expectedUser, nil).
			Times(1)
	}

	step2GetUserFail := func(
		deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantmocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantAdmin(targetTenantId, targetUserId).
			Return(user.User{}, nil).
			Times(1)
	}

	errMockStep2 := newMockError(2)
	step2GetUserError := func(
		deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantmocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantAdmin(targetTenantId, targetUserId).
			Return(user.User{}, errMockStep2).
			Times(1)
	}

	step2NeverCalled := func(
		deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantmocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantAdmin(gomock.Any(), gomock.Any()).
			Times(0)
	}

	// Step 3: controllo ultimo tenant admin
	step3CountOk := func(
		deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantmocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			CountTenantAdminsByTenant(targetTenantId).
			Return(uint(2), nil).
			Times(1)
	}

	// Tentativo di eliminare ultimo tenant
	step3Count_1Admin := func(
		deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantmocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			CountTenantAdminsByTenant(targetTenantId).
			Return(uint(1), nil).
			Times(1)
	}

	errMockStep3 := newMockError(3)
	step3CountError := func(
		deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantmocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			CountTenantAdminsByTenant(targetTenantId).
			Return(uint(0), errMockStep3).
			Times(1)
	}

	// Step 4: delete user
	step4DeleteUserOk := func(
		deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantmocks.MockGetTenantPort,
	) *gomock.Call {
		return deleteUserPort.EXPECT().
			DeleteTenantAdmin(targetTenantId, targetUserId).
			Return(expectedUser, nil).
			Times(1)
	}

	errMockStep4 := newMockError(4)

	step4DeleteUserError := func(
		deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantmocks.MockGetTenantPort,
	) *gomock.Call {
		return deleteUserPort.EXPECT().
			DeleteTenantAdmin(targetTenantId, targetUserId).
			Return(user.User{}, errMockStep4).
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

	baseInput := user.DeleteTenantUserCommand{
		TenantId: targetTenantId,
		UserId:   targetUserId,
	}

	inputWith := func(requester identity.Requester) user.DeleteTenantAdminCommand {
		return user.DeleteTenantAdminCommand{
			Requester: requester,
			TenantId:  baseInput.TenantId,
			UserId:    baseInput.UserId,
		}
	}

	cases := []testCase{
		// Successo
		{
			name:  "(Super Admin) Success: delete tenant admin, impersonation ok",
			input: inputWith(superAdminRequester),
			setupSteps: []mockSetupFunc_DeleteUserService{
				step1TenantOk_CanImpersonate,
				step2GetUserOk,
				step3CountOk,
				step4DeleteUserOk,
			},
			expectedError: nil,
			expectedUser:  expectedUser,
		},
		{
			name:  "(Tenant Admin) Success: delete tenant admin, authorization ok",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_DeleteUserService{
				step1TenantOk_CanImpersonate, // NOTA: impersonazione irrilevante
				step2GetUserOk,
				step3CountOk,
				step4DeleteUserOk,
			},
			expectedError: nil,
			expectedUser:  expectedUser,
		},

		// Step 1: get tenant
		// NOTA: qua requester non conta
		{
			name:  "Fail (step 1): tenant not found",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_DeleteUserService{
				step1TenantFail,
			},
			expectedError: tenant.ErrTenantNotFound,
			expectedUser:  user.User{},
		},
		{
			name:  "Fail (step 1): unexpected error",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_DeleteUserService{
				step1TenantError,
			},
			expectedError: errMockStep1,
			expectedUser:  user.User{},
		},

		// Step 1: autorizzazione
		{
			name:  "(Super Admin) Fail (step 1 auth): impersonation fail",
			input: inputWith(superAdminRequester),
			setupSteps: []mockSetupFunc_DeleteUserService{
				step1TenantOk_CannotImpersonate,
				step2NeverCalled,
			},
			expectedError: identity.ErrUnauthorizedAccess,
			expectedUser:  user.User{},
		},
		{
			name:  "(Tenant Admin) Fail (step 1 auth): unauthorized access",
			input: inputWith(unauthorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_DeleteUserService{
				step1TenantOk_CanImpersonate, // NOTA: impersonazione irrilevante
				step2NeverCalled,
			},
			expectedError: identity.ErrUnauthorizedAccess,
			expectedUser:  user.User{},
		},
		{
			name:  "(Tenant User) Fail (step 1 auth): unauthorized access",
			input: inputWith(tenantUserRequester),
			setupSteps: []mockSetupFunc_DeleteUserService{
				step1TenantOk_CanImpersonate, // NOTA: impersonazione irrilevante
				step2NeverCalled,
			},
			expectedError: identity.ErrUnauthorizedAccess,
			expectedUser:  user.User{},
		},

		// Step 2: get user
		// NOTA: da qui il requester non importa
		{
			name:  "Fail (step 2): user not found",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_DeleteUserService{
				step1TenantOk_CanImpersonate, // NOTA: impersonazione irrilevante
				step2GetUserFail,
			},
			expectedError: user.ErrUserNotFound,
			expectedUser:  user.User{},
		},
		{
			name:  "Fail (step 2): unexpected error",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_DeleteUserService{
				step1TenantOk_CanImpersonate, // NOTA: impersonazione irrilevante
				step2GetUserError,
			},
			expectedError: errMockStep2,
			expectedUser:  user.User{},
		},

		{
			name:  "Fail (step 3): cannot delete last admin",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_DeleteUserService{
				step1TenantOk_CanImpersonate, // NOTA: impersonazione irrilevante
				step2GetUserOk,
				step3Count_1Admin,
			},
			expectedError: user.ErrCannotDeleteLastAdmin,
			expectedUser:  user.User{},
		},
		{
			name:  "Fail (step 3): unexpected error",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_DeleteUserService{
				step1TenantOk_CanImpersonate, // NOTA: impersonazione irrilevante
				step2GetUserOk,
				step3CountError,
			},
			expectedError: errMockStep3,
			expectedUser:  user.User{},
		},

		// Step 4: delete user
		{
			name:  "Fail (step 4): unexpected error",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_DeleteUserService{
				step1TenantOk_CanImpersonate, // NOTA: impersonazione irrilevante
				step2GetUserOk,
				step3CountOk,
				step4DeleteUserError,
			},
			expectedError: errMockStep4,
			expectedUser:  user.User{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// NOTA: il controller di gomock va inizializzato qua dentro!
			mockController := gomock.NewController(t)

			mockDeletePort := mocks.NewMockDeleteUserPort(mockController)
			mockGetUserPort := mocks.NewMockGetUserPort(mockController)
			mockGetTenantPort := tenantmocks.NewMockGetTenantPort(mockController)

			// Slice con chiamate da eseguire
			var expectedCalls []any // NOTA: Dovrebbe essere []*gomock.Call, però il compilatore non accetta

			// Collezione le chiamate per questo test case
			for _, step := range tc.setupSteps {
				call := step(mockDeletePort, mockGetUserPort, mockGetTenantPort)
				if call != nil {
					expectedCalls = append(expectedCalls, call)
				}
			}

			// Richiedi ordine nelle chiamate
			if len(expectedCalls) > 0 {
				gomock.InOrder(expectedCalls...)
			}

			// Crea servizio con porte mock
			_, deleteTenantAdminUseCase, _ := user.NewDeleteUserService(
				mockDeletePort, mockGetUserPort, mockGetTenantPort,
			)

			// Esegui funzione in oggetto
			createdUser, err := deleteTenantAdminUseCase.DeleteTenantAdmin(tc.input)

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

func TestService_DeleteSuperAdmin(t *testing.T) {
	// Dati test
	targetTenantId := (*uuid.UUID)(nil)
	targetUserId := uint(1)
	targetUserEmail := "test@example.com"
	targetUserName := "Test"
	targetConfirmed := false
	targetRole := identity.ROLE_SUPER_ADMIN

	expectedUser := user.User{
		Id:        targetUserId,
		Name:      targetUserName,
		Email:     targetUserEmail,
		TenantId:  targetTenantId,
		Confirmed: targetConfirmed,
		Role:      targetRole,
	}

	type testCase struct {
		name          string
		input         user.DeleteSuperAdminCommand
		setupSteps    []mockSetupFunc_DeleteUserService
		expectedUser  user.User
		expectedError error
	}

	// Step 1: get user
	step1GetUserOk := func(
		deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantmocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetSuperAdmin(targetUserId).
			Return(expectedUser, nil).
			Times(1)
	}

	step1GetExistingUserFail := func(
		deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantmocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetSuperAdmin(targetUserId).
			Return(user.User{}, nil).
			Times(1)
	}

	errMockStep1 := newMockError(1)
	step1GetUserError := func(
		deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantmocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetSuperAdmin(targetUserId).
			Return(user.User{}, errMockStep1).
			Times(1)
	}

	step1NeverCalled := func(
		deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantmocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetSuperAdmin(gomock.Any()).
			Times(0)
	}

	// Step 2: controllo ultimo tenant admin
	step2CountOk := func(
		deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantmocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			CountSuperAdmins().
			Return(uint(2), nil).
			Times(1)
	}

	// Tentativo di eliminare ultimo tenant
	step2Count_1Admin := func(
		deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantmocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			CountSuperAdmins().
			Return(uint(1), nil).
			Times(1)
	}

	errMockStep2 := newMockError(3)
	step2CountError := func(
		deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantmocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			CountSuperAdmins().
			Return(uint(0), errMockStep2).
			Times(1)
	}


	// step 4: delete user
	step3DeleteUserOk := func(
		deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantmocks.MockGetTenantPort,
	) *gomock.Call {
		return deleteUserPort.EXPECT().
			DeleteSuperAdmin(targetUserId).
			Return(expectedUser, nil).
			Times(1)
	}

	errMockStep3 := newMockError(2)
	step3DeleteUserError := func(
		deleteUserPort *mocks.MockDeleteUserPort, getUserPort *mocks.MockGetUserPort, getTenantPort *tenantmocks.MockGetTenantPort,
	) *gomock.Call {
		return deleteUserPort.EXPECT().
			DeleteSuperAdmin(targetUserId).
			Return(user.User{}, errMockStep3).
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

	baseInput := user.DeleteSuperAdminCommand{
		UserId: targetUserId,
	}

	inputWith := func(requester identity.Requester) user.DeleteSuperAdminCommand {
		return user.DeleteSuperAdminCommand{
			Requester: requester,
			UserId:    baseInput.UserId,
		}
	}

	cases := []testCase{
		{
			name:  "Success: delete super admin",
			input: inputWith(superAdminRequester),
			setupSteps: []mockSetupFunc_DeleteUserService{
				step1GetUserOk,
				step2CountOk,
				step3DeleteUserOk,
			},
			expectedError: nil,
			expectedUser:  expectedUser,
		},

		// Test autorizzazione
		{
			name:  "(Tenant Admin) Fail: unauthorized access",
			input: inputWith(tenantAdminRequester),
			setupSteps: []mockSetupFunc_DeleteUserService{
				step1NeverCalled,
			},
			expectedError: identity.ErrUnauthorizedAccess,
			expectedUser:  user.User{},
		},
		{
			name:  "(Tenant User) Fail: unauthorized access",
			input: inputWith(tenantUserRequester),
			setupSteps: []mockSetupFunc_DeleteUserService{
				step1NeverCalled,
			},
			expectedError: identity.ErrUnauthorizedAccess,
			expectedUser:  user.User{},
		},

		// Step 1
		{
			name:  "Fail (step 1): unexpected error",
			input: inputWith(superAdminRequester),
			setupSteps: []mockSetupFunc_DeleteUserService{
				step1GetUserError,
			},
			expectedError: errMockStep1,
			expectedUser:  user.User{},
		},
		{
			name:  "Fail (step 1): user not found",
			input: inputWith(superAdminRequester),
			setupSteps: []mockSetupFunc_DeleteUserService{
				step1GetExistingUserFail,
			},
			expectedError: user.ErrUserNotFound,
			expectedUser:  user.User{},
		},

		// Step 2
		{
			name:  "Fail (step 2): cannot delete last admin",
			input: inputWith(superAdminRequester),
			setupSteps: []mockSetupFunc_DeleteUserService{
				step1GetUserOk,
				step2Count_1Admin,
			},
			expectedError: user.ErrCannotDeleteLastAdmin,
			expectedUser:  user.User{},
		},
		{
			name:  "Fail (step 2): unexpected error",
			input: inputWith(superAdminRequester),
			setupSteps: []mockSetupFunc_DeleteUserService{
				step1GetUserOk,
				step2CountError,
			},
			expectedError: errMockStep2,
			expectedUser:  user.User{},
		},


		// Step 3
		{
			name:  "Fail (step 3): unexpected error",
			input: inputWith(superAdminRequester),
			setupSteps: []mockSetupFunc_DeleteUserService{
				step1GetUserOk,
				step2CountOk,
				step3DeleteUserError,
			},
			expectedError: errMockStep3,
			expectedUser:  user.User{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// NOTA: il controller di gomock va inizializzato qua dentro!
			mockController := gomock.NewController(t)

			mockDeletePort := mocks.NewMockDeleteUserPort(mockController)
			mockGetUserPort := mocks.NewMockGetUserPort(mockController)
			mockGetTenantPort := tenantmocks.NewMockGetTenantPort(mockController)

			// Slice con chiamate da eseguire
			var expectedCalls []any // NOTA: Dovrebbe essere []*gomock.Call, però il compilatore non accetta

			// Collezione le chiamate per questo test case
			for _, step := range tc.setupSteps {
				call := step(mockDeletePort, mockGetUserPort, mockGetTenantPort)
				if call != nil {
					expectedCalls = append(expectedCalls, call)
				}
			}

			// Richiedi ordine nelle chiamate
			if len(expectedCalls) > 0 {
				gomock.InOrder(expectedCalls...)
			}

			// Crea servizio con porte mock
			_, _, deleteSuperAdminUseCase := user.NewDeleteUserService(
				mockDeletePort, mockGetUserPort, mockGetTenantPort,
			)

			// Esegui funzione in oggetto
			createdUser, err := deleteSuperAdminUseCase.DeleteSuperAdmin(tc.input)

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
