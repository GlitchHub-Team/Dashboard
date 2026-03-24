package user_test

import (
	"errors"
	"slices"
	"testing"

	"backend/internal/identity"
	"backend/internal/tenant"
	"backend/internal/user"
	tenantMocks "backend/tests/tenant/mocks"
	"backend/tests/user/mocks"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

type mockSetupFunc_GetUserService func(
	getUserPort *mocks.MockGetUserPort,
	getTenantPort *tenantMocks.MockGetTenantPort,
) *gomock.Call

func newStepTenantOk_GetUserService(
	targetTenantId uuid.UUID, canImpersonate bool,
) mockSetupFunc_GetUserService {
	return func(
		getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort,
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

func newStepTenantNotFound_GetUserService(targetTenantId uuid.UUID) mockSetupFunc_GetUserService {
	return func(
		getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort,
	) *gomock.Call {
		return getTenantPort.EXPECT().
			GetTenant(targetTenantId).
			Return(tenant.Tenant{}, nil).
			Times(1)
	}
}

func newStepTenantError_GetUserService(targetTenantId uuid.UUID, mockError error) mockSetupFunc_GetUserService {
	return func(
		getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort,
	) *gomock.Call {
		return getTenantPort.EXPECT().
			GetTenant(targetTenantId).
			Return(tenant.Tenant{}, mockError).
			Times(1)
	}
}

// Get singolo =============================================================================================================================
func TestService_GetTenantUser(t *testing.T) {
	// Dati test
	targetTenantId := uuid.New()
	otherTenantId := uuid.New()

	targetUserId := uint(1)
	otherUserId := uint(2)

	// expectedTenant := tenant.Tenant{Id: targetTenantId}

	expectedUser := user.User{
		Id:       targetUserId,
		Role:     identity.ROLE_TENANT_USER,
		TenantId: &targetTenantId,
	}

	type testCase struct {
		name          string
		input         user.GetTenantUserCommand
		setupSteps    []mockSetupFunc_GetUserService
		expectedUser  user.User
		expectedError error
	}

	// Test comportamentali
	// Step 1: get tenant
	step1TenantOk_CanImpersonate := newStepTenantOk_GetUserService(targetTenantId, true)
	step1TenantOk_CannotImpersonate := newStepTenantOk_GetUserService(targetTenantId, false)

	step1TenantFail := newStepTenantNotFound_GetUserService(targetTenantId)

	errMockStep1 := newMockError(1)
	step1TenantError := newStepTenantError_GetUserService(targetTenantId, errMockStep1)

	// Step 2: get user
	step2GetUserOk := func(
		getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantUser(targetTenantId, targetUserId).
			Return(expectedUser, nil).
			Times(1)
	}

	step2UserNotFoundFail := func(
		getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantUser(targetTenantId, targetUserId).
			Return(user.User{}, nil).
			Times(1)
	}

	errMockStep2 := newMockError(2)
	step2GetUserError := func(
		getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantUser(targetTenantId, targetUserId).
			Return(user.User{}, errMockStep2).
			Times(1)
	}

	step2NeverCalled := func(
		getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantUser(gomock.Any(), gomock.Any()).
			Times(0)
	}

	// Requesters:

	// Super admin ------------------------------------------------
	superAdminRequester := identity.Requester{
		RequesterUserId: uint(99), // NOTA: id non importa
		RequesterRole:   identity.ROLE_SUPER_ADMIN,
	}

	// Tenant admin ------------------------------------------------
	authorizedTenantAdminRequester := identity.Requester{
		RequesterUserId:   uint(99),
		RequesterTenantId: &targetTenantId, // NOTA: importa tenant id, non user id
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	unauthorizedTenantAdminRequester := identity.Requester{
		RequesterUserId:   uint(99),
		RequesterTenantId: &otherTenantId, // NOTA: importa tenant id, non user id
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	// Tenant User ------------------------------------------------
	authorizedTenantUserRequester := identity.Requester{
		RequesterUserId:   targetUserId, // NOTA: importano tenant id e user id
		RequesterTenantId: &targetTenantId,
		RequesterRole:     identity.ROLE_TENANT_USER,
	}

	unauthorizedTenantUserRequester_differentTenantId := identity.Requester{
		RequesterUserId:   otherUserId, // NOTA: importano tenant id e user id
		RequesterTenantId: &targetTenantId,
		RequesterRole:     identity.ROLE_TENANT_USER,
	}

	unauthorizedTenantUserRequester_differentUserId := identity.Requester{
		RequesterUserId:   otherUserId, // NOTA: importano tenant id e user id
		RequesterTenantId: &otherTenantId,
		RequesterRole:     identity.ROLE_TENANT_USER,
	}

	// Input
	baseInput := user.GetTenantUserCommand{
		TenantId: targetTenantId,
		UserId:   targetUserId,
	}

	inputWith := func(requester identity.Requester) user.GetTenantUserCommand {
		return user.GetTenantUserCommand{
			Requester: requester,
			UserId:    baseInput.UserId,
			TenantId:  baseInput.TenantId,
		}
	}

	cases := []testCase{
		// Successo
		{
			name:  "(Super Admin) Success: get tenant user OK",
			input: inputWith(superAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk_CanImpersonate,
				step2GetUserOk,
			},
			expectedUser:  expectedUser,
			expectedError: nil,
		},
		{
			name:  "(Tenant Admin) Success: get tenant user OK",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk_CanImpersonate, // NOTA: impersonazione non importa
				step2GetUserOk,
			},
			expectedUser:  expectedUser,
			expectedError: nil,
		},
		{
			name:  "(Tenant User) Success: get self OK",
			input: inputWith(authorizedTenantUserRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk_CanImpersonate, // NOTA: impersonazione non importa
				step2GetUserOk,
			},
			expectedUser:  expectedUser,
			expectedError: nil,
		},

		// Step 1: get tenant (NOTA: non importa ancora requester)
		{
			name:  "Fail (step 1): tenant not found",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantFail,
			},
			expectedUser:  user.User{},
			expectedError: tenant.ErrTenantNotFound,
		},
		{
			name:  "Fail (step 1): unexpected error",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantError,
			},
			expectedUser:  user.User{},
			expectedError: errMockStep1,
		},

		// Step 1: autorizzazione
		{
			name:  "(Super Admin) Fail (step 1): impersonation fail",
			input: inputWith(superAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk_CannotImpersonate,
				step2NeverCalled,
			},
			expectedUser:  user.User{},
			expectedError: identity.ErrUnauthorizedAccess,
		},
		{
			name:  "(Tenant Admin) Fail (step 1): unauthorized access",
			input: inputWith(unauthorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk_CanImpersonate, // NOTA: impersonazione non importa
				step2NeverCalled,
			},
			expectedUser:  user.User{},
			expectedError: identity.ErrUnauthorizedAccess,
		},
		{
			name:  "(Tenant User) Fail (step 1): unauthorized access, different user id",
			input: inputWith(unauthorizedTenantUserRequester_differentUserId),
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk_CanImpersonate, // NOTA: impersonazione non importa
				step2NeverCalled,
			},
			expectedUser:  user.User{},
			expectedError: identity.ErrUnauthorizedAccess,
		},
		{
			name:  "(Tenant User) Fail (step 1): unauthorized access, different tenant id",
			input: inputWith(unauthorizedTenantUserRequester_differentTenantId),
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk_CanImpersonate, // NOTA: impersonazione non importa
				step2NeverCalled,
			},
			expectedUser:  user.User{},
			expectedError: identity.ErrUnauthorizedAccess,
		},

		// Step 2: get user
		{
			name:  "Fail (step 2): user not found",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk_CanImpersonate, // NOTA: impersonazione non importa
				step2UserNotFoundFail,
			},
			expectedUser:  user.User{},
			expectedError: user.ErrUserNotFound,
		},
		{
			name:  "Fail (step 2): unexpected error",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk_CanImpersonate, // NOTA: impersonazione non importa
				step2GetUserError,
			},
			expectedUser:  user.User{},
			expectedError: errMockStep2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockController := gomock.NewController(t)

			mockGetUserPort := mocks.NewMockGetUserPort(mockController)
			mockGetTenantPort := tenantMocks.NewMockGetTenantPort(mockController)

			// Slice con chiamate da eseguire
			var expectedCalls []any // NOTA: Dovrebbe essere []*gomock.Call, però il compilatore non accetta

			// Collezione le chiamate per questo test case
			for _, step := range tc.setupSteps {
				call := step(mockGetUserPort, mockGetTenantPort)
				if call != nil {
					expectedCalls = append(expectedCalls, call)
				}
			}

			// Richiedi ordine nelle chiamate
			if len(expectedCalls) > 0 {
				gomock.InOrder(expectedCalls...)
			}

			// Crea servizio con porte mock
			getTenantUserUseCase, _, _, _, _, _ := user.NewGetUserService(
				mockGetUserPort, mockGetTenantPort,
			)

			// Esegui funzione in oggetto
			createdUser, err := getTenantUserUseCase.GetTenantUser(tc.input)

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

func TestService_GetTenantAdmin(t *testing.T) {
	// Dati test
	targetTenantId := uuid.New()
	otherTenantId := uuid.New()

	targetUserId := uint(1)

	// expectedTenant := tenant.Tenant{Id: targetTenantId}

	expectedUser := user.User{
		Id:       targetUserId,
		Role:     identity.ROLE_TENANT_USER,
		TenantId: &targetTenantId,
	}

	type testCase struct {
		name          string
		input         user.GetTenantAdminCommand
		setupSteps    []mockSetupFunc_GetUserService
		expectedUser  user.User
		expectedError error
	}

	// Test comportamentali
	// Step 1: get tenant
	step1TenantOk_CanImpersonate := newStepTenantOk_GetUserService(targetTenantId, true)
	step1TenantOk_CannotImpersonate := newStepTenantOk_GetUserService(targetTenantId, false)

	step1TenantFail := newStepTenantNotFound_GetUserService(targetTenantId)

	errMockStep1 := newMockError(1)
	step1TenantError := newStepTenantError_GetUserService(targetTenantId, errMockStep1)

	// Step 2: get user
	step2GetUserOk := func(
		getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantAdmin(targetTenantId, targetUserId).
			Return(expectedUser, nil).
			Times(1)
	}

	step2UserNotFoundFail := func(
		getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantAdmin(targetTenantId, targetUserId).
			Return(user.User{}, nil).
			Times(1)
	}

	errMockStep2 := newMockError(2)
	step2GetUserError := func(
		getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantAdmin(targetTenantId, targetUserId).
			Return(user.User{}, errMockStep2).
			Times(1)
	}

	step2NeverCalled := func(
		getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantAdmin(gomock.Any(), gomock.Any()).
			Times(0)
	}

	// Requesters
	superAdminRequester := identity.Requester{
		RequesterUserId: uint(99), // NOTA: id non importa
		RequesterRole:   identity.ROLE_SUPER_ADMIN,
	}

	authorizedTenantAdminRequester := identity.Requester{
		RequesterUserId:   uint(99),
		RequesterTenantId: &targetTenantId, // NOTA: importa tenant id, non user id
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	unauthorizedTenantAdminRequester := identity.Requester{
		RequesterUserId:   uint(99),
		RequesterTenantId: &otherTenantId, // NOTA: importa tenant id, non user id
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	tenantUserRequester := identity.Requester{
		RequesterUserId:   targetUserId,
		RequesterTenantId: &targetTenantId,
		RequesterRole:     identity.ROLE_TENANT_USER,
	}

	baseInput := user.GetTenantUserCommand{
		TenantId: targetTenantId,
		UserId:   targetUserId,
	}

	inputWith := func(requester identity.Requester) user.GetTenantAdminCommand {
		return user.GetTenantAdminCommand{
			Requester: requester,
			UserId:    baseInput.UserId,
			TenantId:  baseInput.TenantId,
		}
	}

	cases := []testCase{
		// Successo
		{
			name:  "(Super Admin) Success: get tenant user OK",
			input: inputWith(superAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk_CanImpersonate,
				step2GetUserOk,
			},
			expectedUser:  expectedUser,
			expectedError: nil,
		},
		{
			name:  "(Tenant Admin) Success: get tenant user OK",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk_CanImpersonate, // NOTA: impersonazione non importa
				step2GetUserOk,
			},
			expectedUser:  expectedUser,
			expectedError: nil,
		},

		// Step 1: get tenant (NOTA: non importa ancora requester)
		{
			name:  "Fail (step 1): tenant not found",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantFail,
			},
			expectedUser:  user.User{},
			expectedError: tenant.ErrTenantNotFound,
		},
		{
			name:  "Fail (step 1): unexpected error",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantError,
			},
			expectedUser:  user.User{},
			expectedError: errMockStep1,
		},

		// Step 1: autorizzazione
		{
			name:  "(Super Admin) Fail (step 1): impersonation fail",
			input: inputWith(superAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk_CannotImpersonate,
				step2NeverCalled,
			},
			expectedUser:  user.User{},
			expectedError: identity.ErrUnauthorizedAccess,
		},
		{
			name:  "(Tenant Admin) Fail (step 1): unauthorized access",
			input: inputWith(unauthorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk_CanImpersonate, // NOTA: impersonazione non importa
				step2NeverCalled,
			},
			expectedUser:  user.User{},
			expectedError: identity.ErrUnauthorizedAccess,
		},
		{
			name:  "(Tenant User) Fail (step 1): unauthorized access",
			input: inputWith(tenantUserRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk_CanImpersonate, // NOTA: impersonazione non importa
				step2NeverCalled,
			},
			expectedUser:  user.User{},
			expectedError: identity.ErrUnauthorizedAccess,
		},

		// Step 2: get user
		{
			name:  "Fail (step 2): user not found",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk_CanImpersonate, // NOTA: impersonazione non importa
				step2UserNotFoundFail,
			},
			expectedUser:  user.User{},
			expectedError: user.ErrUserNotFound,
		},
		{
			name:  "Fail (step 2): unexpected error",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk_CanImpersonate, // NOTA: impersonazione non importa
				step2GetUserError,
			},
			expectedUser:  user.User{},
			expectedError: errMockStep2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockController := gomock.NewController(t)

			mockGetUserPort := mocks.NewMockGetUserPort(mockController)
			mockGetTenantPort := tenantMocks.NewMockGetTenantPort(mockController)

			// Slice con chiamate da eseguire
			var expectedCalls []any // NOTA: Dovrebbe essere []*gomock.Call, però il compilatore non accetta

			// Collezione le chiamate per questo test case
			for _, step := range tc.setupSteps {
				call := step(mockGetUserPort, mockGetTenantPort)
				if call != nil {
					expectedCalls = append(expectedCalls, call)
				}
			}

			// Richiedi ordine nelle chiamate
			if len(expectedCalls) > 0 {
				gomock.InOrder(expectedCalls...)
			}

			// Crea servizio con porte mock
			_, getTenantAdminUseCase, _, _, _, _ := user.NewGetUserService(
				mockGetUserPort, mockGetTenantPort,
			)

			// Esegui funzione in oggetto
			createdUser, err := getTenantAdminUseCase.GetTenantAdmin(tc.input)

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

func TestService_GetSuperAdmin(t *testing.T) {
	// Dati test
	targetTenantId := uuid.New()

	targetUserId := uint(1)

	// expectedTenant := tenant.Tenant{Id: targetTenantId}

	expectedUser := user.User{
		Id:       targetUserId,
		Role:     identity.ROLE_TENANT_USER,
		TenantId: &targetTenantId,
	}

	type testCase struct {
		name          string
		input         user.GetSuperAdminCommand
		setupSteps    []mockSetupFunc_GetUserService
		expectedUser  user.User
		expectedError error
	}

	// Test comportamentali

	// Step 2: get user
	step1GetUserOk := func(
		getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetSuperAdmin(targetUserId).
			Return(expectedUser, nil).
			Times(1)
	}

	step1UserNotFoundFail := func(
		getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetSuperAdmin(targetUserId).
			Return(user.User{}, nil).
			Times(1)
	}

	errMockStep1 := newMockError(2)
	step1GetUserError := func(
		getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetSuperAdmin(targetUserId).
			Return(user.User{}, errMockStep1).
			Times(1)
	}

	step1NeverCalled := func(
		getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetSuperAdmin(gomock.Any()).
			Times(0)
	}

	// Requesters
	superAdminRequester := identity.Requester{
		RequesterUserId: uint(99), // NOTA: id non importa
		RequesterRole:   identity.ROLE_SUPER_ADMIN,
	}

	tenantAdminRequester := identity.Requester{
		RequesterUserId:   uint(99),
		RequesterTenantId: &targetTenantId, // NOTA: importa tenant id, non user id
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	tenantUserRequester := identity.Requester{
		RequesterUserId:   targetUserId, // NOTA: importano tenant id e user id
		RequesterTenantId: &targetTenantId,
		RequesterRole:     identity.ROLE_TENANT_USER,
	}

	baseInput := user.GetTenantUserCommand{
		TenantId: targetTenantId,
		UserId:   targetUserId,
	}

	inputWith := func(requester identity.Requester) user.GetSuperAdminCommand {
		return user.GetSuperAdminCommand{
			Requester: requester,
			UserId:    baseInput.UserId,
		}
	}

	cases := []testCase{
		// Successo
		{
			name:  "(Super Admin) Success: get super admin OK",
			input: inputWith(superAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1GetUserOk,
			},
			expectedUser:  expectedUser,
			expectedError: nil,
		},

		// Step 1: test autorizzazione
		{
			name:  "(Tenant Admin) Fail (auth): unauthorized access",
			input: inputWith(tenantAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1NeverCalled,
			},
			expectedUser:  user.User{},
			expectedError: identity.ErrUnauthorizedAccess,
		},
		{
			name:  "(Tenant User) Fail (auth): unauthorized access",
			input: inputWith(tenantUserRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1NeverCalled,
			},
			expectedUser:  user.User{},
			expectedError: identity.ErrUnauthorizedAccess,
		},

		// Step 1: get user
		{
			name:  "Fail (step 1): user not found",
			input: inputWith(superAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1UserNotFoundFail,
			},
			expectedUser:  user.User{},
			expectedError: user.ErrUserNotFound,
		},
		{
			name:  "Fail (step 1): unexpected error",
			input: inputWith(superAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1GetUserError,
			},
			expectedUser:  user.User{},
			expectedError: errMockStep1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockController := gomock.NewController(t)

			mockGetUserPort := mocks.NewMockGetUserPort(mockController)
			mockGetTenantPort := tenantMocks.NewMockGetTenantPort(mockController)

			// Slice con chiamate da eseguire
			var expectedCalls []any // NOTA: Dovrebbe essere []*gomock.Call, però il compilatore non accetta

			// Collezione le chiamate per questo test case
			for _, step := range tc.setupSteps {
				call := step(mockGetUserPort, mockGetTenantPort)
				if call != nil {
					expectedCalls = append(expectedCalls, call)
				}
			}

			// Richiedi ordine nelle chiamate
			if len(expectedCalls) > 0 {
				gomock.InOrder(expectedCalls...)
			}

			// Crea servizio con porte mock
			_, _, getSuperAdminUseCase_, _, _, _ := user.NewGetUserService(
				mockGetUserPort, mockGetTenantPort,
			)

			// Esegui funzione in oggetto
			createdUser, err := getSuperAdminUseCase_.GetSuperAdmin(tc.input)

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

// Get multiplo ===================================================================================================================
func TestService_GetTenantUsersByTenant(t *testing.T) {
	// Dati test
	targetTenantId := uuid.New()
	otherTenantId := uuid.New()

	targetUserId := uint(1)

	// - Caso 1: page=1, limit=1
	targetPageCase1 := 1
	targetLimitCase1 := 1
	expectedUsersCase1 := []user.User{
		{
			Id:       targetUserId,
			Role:     identity.ROLE_TENANT_USER,
			TenantId: &targetTenantId,
		},
	}
	expectedTotalCase1 := uint(1)

	// - Caso 2: page=2, limit=1 (pagina vuota)
	targetPageCase2 := 2
	targetLimitCase2 := 2
	expectedUsersCase2 := ([]user.User)(nil) // NOTA: slice vuoto
	expectedTotalCase2 := uint(1)

	type testCase struct {
		name          string
		input         user.GetTenantUsersByTenantCommand
		setupSteps    []mockSetupFunc_GetUserService
		expectedUsers []user.User
		expectedTotal uint
		expectedError error
	}

	// Test comportamentali
	// 1) Get tenant
	//   - OK -> 2
	//	 - Fail
	//   - Error
	// 2) Get users
	//	   - get page 1, limit 1: OK -> success
	//     - get page 1, limit 1: Error
	//     - get page 2, limit 1: OK (empty page) -> success
	//

	step1TenantOk_CanImpersonate := newStepTenantOk_GetUserService(targetTenantId, true)
	step1TenantOk_CannotImpersonate := newStepTenantOk_GetUserService(targetTenantId, false)

	step1TenantFail := newStepTenantNotFound_GetUserService(targetTenantId)

	errMockStep1 := errors.New("unexpected error in step 1")
	step1TenantError := newStepTenantError_GetUserService(targetTenantId, errMockStep1)

	step2Case1Ok := func(
		getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantUsersByTenant(targetTenantId, targetPageCase1, targetLimitCase1).
			Return(expectedUsersCase1, expectedTotalCase1, nil).
			Times(1)
	}

	step2Case2Ok := func(
		getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantUsersByTenant(targetTenantId, targetPageCase2, targetLimitCase2).
			Return(expectedUsersCase2, expectedTotalCase2, nil).
			Times(1)
	}

	errMockStep2 := errors.New("unexpected error in step 2")
	step2GetUsersError := func(
		getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantUsersByTenant(targetTenantId, targetPageCase1, targetLimitCase1).
			Return(nil, uint(0), errMockStep2).
			Times(1)
	}

	step2NeverCalled := func(
		getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantUsersByTenant(gomock.Any(), gomock.Any(), gomock.Any()).
			Times(0)
	}

	// Requesters
	superAdminRequester := identity.Requester{
		RequesterUserId: uint(99), // NOTA: id non importa
		RequesterRole:   identity.ROLE_SUPER_ADMIN,
	}

	// Tenant admin ------------------------------------------------
	authorizedTenantAdminRequester := identity.Requester{
		RequesterUserId:   uint(99),
		RequesterTenantId: &targetTenantId, // NOTA: importa tenant id, non user id
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	unauthorizedTenantAdminRequester := identity.Requester{
		RequesterUserId:   uint(99),
		RequesterTenantId: &otherTenantId, // NOTA: importa tenant id, non user id
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	// Tenant User ------------------------------------------------
	tenantUserRequester := identity.Requester{
		RequesterUserId:   targetUserId, // NOTA: importano tenant id e user id
		RequesterTenantId: &targetTenantId,
		RequesterRole:     identity.ROLE_TENANT_USER,
	}

	// Input
	baseInputCase1 := user.GetTenantUsersByTenantCommand{
		Page:     targetPageCase1,
		Limit:    targetLimitCase1,
		TenantId: targetTenantId,
	}

	baseInputCase2 := user.GetTenantUsersByTenantCommand{
		Page:     targetPageCase2,
		Limit:    targetLimitCase2,
		TenantId: targetTenantId,
	}

	inputWith := func(input user.GetTenantUsersByTenantCommand, requester identity.Requester) user.GetTenantUsersByTenantCommand {
		return user.GetTenantUsersByTenantCommand{
			Requester: requester,
			Page:      input.Page,
			Limit:     input.Limit,
			TenantId:  input.TenantId,
		}
	}

	testCases := []testCase{
		// Successo
		{
			name:  "(Super Admin) Success: case 1: page=1, limit=1",
			input: inputWith(baseInputCase1, superAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk_CanImpersonate,
				step2Case1Ok,
			},
			expectedUsers: expectedUsersCase1,
			expectedTotal: expectedTotalCase1,
			expectedError: nil,
		},
		{
			name:  "(Tenant Admin) Success: case 1: page=1, limit=1",
			input: inputWith(baseInputCase1, authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk_CanImpersonate,
				step2Case1Ok,
			},
			expectedUsers: expectedUsersCase1,
			expectedTotal: expectedTotalCase1,
			expectedError: nil,
		},
		{
			name:  "(Super Admin) Success: case 2: page=2, limit=1 (empty page)",
			input: inputWith(baseInputCase2, superAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk_CanImpersonate,
				step2Case2Ok,
			},
			expectedUsers: expectedUsersCase2,
			expectedTotal: expectedTotalCase2,
			expectedError: nil,
		},
		{
			name:  "(Tenant Admin) Success: case 2: page=2, limit=1 (empty page)",
			input: inputWith(baseInputCase2, authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk_CanImpersonate,
				step2Case2Ok,
			},
			expectedUsers: expectedUsersCase2,
			expectedTotal: expectedTotalCase2,
			expectedError: nil,
		},

		// Step 1: get tenant
		{
			name:  "Fail (step 1): tenant not found",
			input: inputWith(baseInputCase1, authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantFail,
			},
			expectedUsers: nil,
			expectedTotal: 0,
			expectedError: tenant.ErrTenantNotFound,
		},
		{
			name:  "Fail (step 1): unexpected error",
			input: inputWith(baseInputCase1, authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantError,
			},
			expectedUsers: nil,
			expectedTotal: 0,
			expectedError: errMockStep1,
		},

		// Step 1: autorizzazione
		{
			name:  "(Super Admin) Fail (auth): impersonation fail",
			input: inputWith(baseInputCase1, superAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk_CannotImpersonate,
				step2NeverCalled,
			},
			expectedUsers: nil,
			expectedTotal: 0,
			expectedError: identity.ErrUnauthorizedAccess,
		},
		{
			name:  "(Tenant Admin) Fail (auth): impersonation fail",
			input: inputWith(baseInputCase1, unauthorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk_CanImpersonate,
				step2NeverCalled,
			},
			expectedUsers: nil,
			expectedTotal: 0,
			expectedError: identity.ErrUnauthorizedAccess,
		},
		{
			name:  "(Tenant User) Fail (auth): impersonation fail",
			input: inputWith(baseInputCase1, tenantUserRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk_CanImpersonate,
				step2NeverCalled,
			},
			expectedUsers: nil,
			expectedTotal: 0,
			expectedError: identity.ErrUnauthorizedAccess,
		},

		// Step 2: get users
		{
			name:  "Fail (step 2): unexpected error",
			input: inputWith(baseInputCase1, authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk_CanImpersonate,
				step2GetUsersError,
			},
			expectedUsers: nil,
			expectedTotal: 0,
			expectedError: errMockStep2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockController := gomock.NewController(t)

			mockGetUserPort := mocks.NewMockGetUserPort(mockController)
			mockGetTenantPort := tenantMocks.NewMockGetTenantPort(mockController)

			// Slice con chiamate da eseguire
			var expectedCalls []any // NOTA: Dovrebbe essere []*gomock.Call, però il compilatore non accetta

			// Collezione le chiamate per questo test case
			for _, step := range tc.setupSteps {
				call := step(mockGetUserPort, mockGetTenantPort)
				if call != nil {
					expectedCalls = append(expectedCalls, call)
				}
			}

			// Richiedi ordine nelle chiamate
			if len(expectedCalls) > 0 {
				gomock.InOrder(expectedCalls...)
			}

			// Crea servizio con porte mock
			_, _, _, getTenantUsersByTenantUseCase, _, _ := user.NewGetUserService(
				mockGetUserPort, mockGetTenantPort,
			)

			// Esegui funzione in oggetto
			obtainedUsers, total, err := getTenantUsersByTenantUseCase.GetTenantUsersByTenant(tc.input)

			// Assertions
			if err != tc.expectedError {
				t.Errorf("expected error %v, got %v", tc.expectedError, err)
			}
			if !slices.Equal(obtainedUsers, tc.expectedUsers) {
				t.Errorf("expected users %v, got %v", tc.expectedUsers, obtainedUsers)
			}
			if total != tc.expectedTotal {
				t.Errorf("expected total %v, got %v", tc.expectedTotal, total)
			}
		})
	}
}

func TestService_GetTenantAdminsByTenant(t *testing.T) {
	// Dati test
	targetTenantId := uuid.New()
	otherTenantId := uuid.New()

	targetUserId := uint(1)

	// - Caso 1: page=1, limit=1
	targetPageCase1 := 1
	targetLimitCase1 := 1
	expectedUsersCase1 := []user.User{
		{
			Id:       targetUserId,
			Role:     identity.ROLE_TENANT_ADMIN,
			TenantId: &targetTenantId,
		},
	}
	expectedTotalCase1 := uint(1)

	// - Caso 2: page=2, limit=1 (pagina vuota)
	targetPageCase2 := 2
	targetLimitCase2 := 2
	expectedUsersCase2 := ([]user.User)(nil) // NOTA: slice vuoto
	expectedTotalCase2 := uint(1)

	type testCase struct {
		name          string
		input         user.GetTenantAdminsByTenantCommand
		setupSteps    []mockSetupFunc_GetUserService
		expectedUsers []user.User
		expectedTotal uint
		expectedError error
	}

	// Test comportamentali
	// 1) Get tenant
	//   - OK -> 2
	//	 - Fail
	//   - Error
	// 2) Get users
	//	   - get page 1, limit 1: OK -> success
	//     - get page 1, limit 1: Error
	//     - get page 2, limit 1: OK (empty page) -> success
	//

	step1TenantOk_CanImpersonate := newStepTenantOk_GetUserService(targetTenantId, true)
	step1TenantOk_CannotImpersonate := newStepTenantOk_GetUserService(targetTenantId, false)

	step1TenantFail := newStepTenantNotFound_GetUserService(targetTenantId)

	errMockStep1 := errors.New("unexpected error in step 1")
	step1TenantError := newStepTenantError_GetUserService(targetTenantId, errMockStep1)

	step2Case1Ok := func(
		getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantAdminsByTenant(targetTenantId, targetPageCase1, targetLimitCase1).
			Return(expectedUsersCase1, expectedTotalCase1, nil).
			Times(1)
	}

	step2Case2Ok := func(
		getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantAdminsByTenant(targetTenantId, targetPageCase2, targetLimitCase2).
			Return(expectedUsersCase2, expectedTotalCase2, nil).
			Times(1)
	}

	errMockStep2 := errors.New("unexpected error in step 2")
	step2GetUsersError := func(
		getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantAdminsByTenant(targetTenantId, targetPageCase1, targetLimitCase1).
			Return(nil, uint(0), errMockStep2).
			Times(1)
	}

	step2NeverCalled := func(
		getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantAdminsByTenant(gomock.Any(), gomock.Any(), gomock.Any()).
			Times(0)
	}

	// Requesters
	superAdminRequester := identity.Requester{
		RequesterUserId: uint(99), // NOTA: id non importa
		RequesterRole:   identity.ROLE_SUPER_ADMIN,
	}

	// Tenant admin ------------------------------------------------
	authorizedTenantAdminRequester := identity.Requester{
		RequesterUserId:   uint(99),
		RequesterTenantId: &targetTenantId, // NOTA: importa tenant id, non user id
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	unauthorizedTenantAdminRequester := identity.Requester{
		RequesterUserId:   uint(99),
		RequesterTenantId: &otherTenantId, // NOTA: importa tenant id, non user id
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	// Tenant User ------------------------------------------------
	tenantUserRequester := identity.Requester{
		RequesterUserId:   targetUserId, // NOTA: importano tenant id e user id
		RequesterTenantId: &targetTenantId,
		RequesterRole:     identity.ROLE_TENANT_USER,
	}

	// Input
	baseInputCase1 := user.GetTenantAdminsByTenantCommand{
		Page:     targetPageCase1,
		Limit:    targetLimitCase1,
		TenantId: targetTenantId,
	}

	baseInputCase2 := user.GetTenantAdminsByTenantCommand{
		Page:     targetPageCase2,
		Limit:    targetLimitCase2,
		TenantId: targetTenantId,
	}

	inputWith := func(input user.GetTenantAdminsByTenantCommand, requester identity.Requester) user.GetTenantAdminsByTenantCommand {
		return user.GetTenantAdminsByTenantCommand{
			Requester: requester,
			Page:      input.Page,
			Limit:     input.Limit,
			TenantId:  input.TenantId,
		}
	}

	testCases := []testCase{
		// Successo
		{
			name:  "(Super Admin) Success: case 1: page=1, limit=1",
			input: inputWith(baseInputCase1, superAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk_CanImpersonate,
				step2Case1Ok,
			},
			expectedUsers: expectedUsersCase1,
			expectedTotal: expectedTotalCase1,
			expectedError: nil,
		},
		{
			name:  "(Tenant Admin) Success: case 2: page=2, limit=1 (empty page)",
			input: inputWith(baseInputCase2, authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk_CanImpersonate,
				step2Case2Ok,
			},
			expectedUsers: expectedUsersCase2,
			expectedTotal: expectedTotalCase2,
			expectedError: nil,
		},
		{
			name:  "(Super Admin) Success: case 1: page=1, limit=1",
			input: inputWith(baseInputCase1, superAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk_CanImpersonate,
				step2Case1Ok,
			},
			expectedUsers: expectedUsersCase1,
			expectedTotal: expectedTotalCase1,
			expectedError: nil,
		},
		{
			name:  "(Tenant Admin) Success: case 2: page=2, limit=1 (empty page)",
			input: inputWith(baseInputCase2, authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk_CanImpersonate,
				step2Case2Ok,
			},
			expectedUsers: expectedUsersCase2,
			expectedTotal: expectedTotalCase2,
			expectedError: nil,
		},

		// Step 1: get tenant
		{
			name:  "Fail (step 1): tenant not found",
			input: inputWith(baseInputCase1, authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantFail,
			},
			expectedUsers: nil,
			expectedTotal: 0,
			expectedError: tenant.ErrTenantNotFound,
		},
		{
			name:  "Fail (step 1): unexpected error",
			input: inputWith(baseInputCase1, authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantError,
			},
			expectedUsers: nil,
			expectedTotal: 0,
			expectedError: errMockStep1,
		},

		// Step 1: autorizzazione
		{
			name:  "(Super Admin) Fail (auth): impersonation failed",
			input: inputWith(baseInputCase1, superAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk_CannotImpersonate,
				step2NeverCalled,
			},
			expectedUsers: nil,
			expectedTotal: 0,
			expectedError: identity.ErrUnauthorizedAccess,
		},
		{
			name:  "(Tenant Admin) Fail (auth): unauthorized access",
			input: inputWith(baseInputCase1, unauthorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk_CanImpersonate,
				step2NeverCalled,
			},
			expectedUsers: nil,
			expectedTotal: 0,
			expectedError: identity.ErrUnauthorizedAccess,
		},
		{
			name:  "(Tenant User) Fail (auth): unauthorized access",
			input: inputWith(baseInputCase1, tenantUserRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk_CanImpersonate,
				step2NeverCalled,
			},
			expectedUsers: nil,
			expectedTotal: 0,
			expectedError: identity.ErrUnauthorizedAccess,
		},

		// step 2: get users
		{
			name:  "Fail (step 2): unexpected error",
			input: inputWith(baseInputCase1, authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk_CanImpersonate,
				step2GetUsersError,
			},
			expectedUsers: nil,
			expectedTotal: 0,
			expectedError: errMockStep2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockController := gomock.NewController(t)

			mockGetUserPort := mocks.NewMockGetUserPort(mockController)
			mockGetTenantPort := tenantMocks.NewMockGetTenantPort(mockController)

			// Slice con chiamate da eseguire
			var expectedCalls []any // NOTA: Dovrebbe essere []*gomock.Call, però il compilatore non accetta

			// Collezione le chiamate per questo test case
			for _, step := range tc.setupSteps {
				call := step(mockGetUserPort, mockGetTenantPort)
				if call != nil {
					expectedCalls = append(expectedCalls, call)
				}
			}

			// Richiedi ordine nelle chiamate
			if len(expectedCalls) > 0 {
				gomock.InOrder(expectedCalls...)
			}

			// Crea servizio con porte mock
			_, _, _, _, getTenantAdminsByTenantUseCase, _ := user.NewGetUserService(
				mockGetUserPort, mockGetTenantPort,
			)

			// Esegui funzione in oggetto
			obtainedUsers, total, err := getTenantAdminsByTenantUseCase.GetTenantAdminsByTenant(tc.input)

			// Assertions
			if err != tc.expectedError {
				t.Errorf("expected error %v, got %v", tc.expectedError, err)
			}
			if !slices.Equal(obtainedUsers, tc.expectedUsers) {
				t.Errorf("expected users %v, got %v", tc.expectedUsers, obtainedUsers)
			}
			if total != tc.expectedTotal {
				t.Errorf("expected total %v, got %v", tc.expectedTotal, total)
			}
		})
	}
}

func TestService_GetSuperAdminList(t *testing.T) {
	// Dati test
	targetUserId := uint(1)

	targetTenantId := uuid.New()

	// - Caso 1: page=1, limit=1
	targetPageCase1 := 1
	targetLimitCase1 := 1
	expectedUsersCase1 := []user.User{
		{
			Id:   targetUserId,
			Role: identity.ROLE_SUPER_ADMIN,
		},
	}
	expectedTotalCase1 := uint(1)

	// - Caso 2: page=2, limit=1 (pagina vuota)
	targetPageCase2 := 2
	targetLimitCase2 := 2
	expectedUsersCase2 := ([]user.User)(nil) // NOTA: slice vuoto
	expectedTotalCase2 := uint(1)

	type testCase struct {
		name          string
		input         user.GetSuperAdminListCommand
		setupSteps    []mockSetupFunc_GetUserService
		expectedUsers []user.User
		expectedTotal uint
		expectedError error
	}

	// Test comportamentali
	// 1) Get tenant
	//   - OK -> 2
	//	 - Fail
	//   - Error
	// 2) Get users
	//	   - get page 1, limit 1: OK -> success
	//     - get page 1, limit 1: Error
	//     - get page 2, limit 1: OK (empty page) -> success
	//

	step1Case1Ok := func(
		getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetSuperAdminList(targetPageCase1, targetLimitCase1).
			Return(expectedUsersCase1, expectedTotalCase1, nil).
			Times(1)
	}

	step1Case2Ok := func(
		getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetSuperAdminList(targetPageCase2, targetLimitCase2).
			Return(expectedUsersCase2, expectedTotalCase2, nil).
			Times(1)
	}

	errMockStep1 := errors.New("unexpected error in step 1")
	step1GetUsersError := func(
		getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetSuperAdminList(targetPageCase1, targetLimitCase1).
			Return(nil, uint(0), errMockStep1).
			Times(1)
	}

	step1NeverCalled := func(
		getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetSuperAdminList(gomock.Any(), gomock.Any()).
			Times(0)
	}

	// Requesters
	superAdminRequester := identity.Requester{
		RequesterUserId: uint(99), // NOTA: id non importa
		RequesterRole:   identity.ROLE_SUPER_ADMIN,
	}

	// Tenant admin ------------------------------------------------
	tenantAdminRequester := identity.Requester{
		RequesterUserId:   uint(99),
		RequesterTenantId: &targetTenantId, // NOTA: importa tenant id, non user id
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	// Tenant User ------------------------------------------------
	tenantUserRequester := identity.Requester{
		RequesterUserId:   targetUserId, // NOTA: importano tenant id e user id
		RequesterTenantId: &targetTenantId,
		RequesterRole:     identity.ROLE_TENANT_USER,
	}

	// Input
	baseInputCase1 := user.GetSuperAdminListCommand{
		Page:  targetPageCase1,
		Limit: targetLimitCase1,
	}

	baseInputCase2 := user.GetSuperAdminListCommand{
		Page:  targetPageCase2,
		Limit: targetLimitCase2,
	}

	inputWith := func(input user.GetSuperAdminListCommand, requester identity.Requester) user.GetSuperAdminListCommand {
		return user.GetSuperAdminListCommand{
			Requester: requester,
			Page:      input.Page,
			Limit:     input.Limit,
		}
	}

	testCases := []testCase{
		// Successo
		{
			name:  "(Super Admin) Success: case 1: page=1, limit=1",
			input: inputWith(baseInputCase1, superAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1Case1Ok,
			},
			expectedUsers: expectedUsersCase1,
			expectedTotal: expectedTotalCase1,
			expectedError: nil,
		},
		{
			name:  "(Super Admin) Success: case 1: page=2, limit=1 (empty page)",
			input: inputWith(baseInputCase2, superAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1Case2Ok,
			},
			expectedUsers: expectedUsersCase2,
			expectedTotal: expectedTotalCase2,
			expectedError: nil,
		},

		// Autorizzazione
		{
			name:  "(Tenant Admin) Fail (auth): unauthorized access",
			input: inputWith(baseInputCase1, tenantAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1NeverCalled,
			},
			expectedUsers: nil,
			expectedTotal: 0,
			expectedError: identity.ErrUnauthorizedAccess,
		},
		{
			name:  "(Tenant User) Fail (auth): unauthorized access",
			input: inputWith(baseInputCase1, tenantUserRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1NeverCalled,
			},
			expectedUsers: nil,
			expectedTotal: 0,
			expectedError: identity.ErrUnauthorizedAccess,
		},

		// step 2: get users
		{
			name:  "Fail (step 2): unexpected error",
			input: inputWith(baseInputCase1, superAdminRequester),
			setupSteps: []mockSetupFunc_GetUserService{
				step1GetUsersError,
			},
			expectedUsers: nil,
			expectedTotal: 0,
			expectedError: errMockStep1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockController := gomock.NewController(t)

			mockGetUserPort := mocks.NewMockGetUserPort(mockController)
			mockGetTenantPort := tenantMocks.NewMockGetTenantPort(mockController)

			// Slice con chiamate da eseguire
			var expectedCalls []any // NOTA: Dovrebbe essere []*gomock.Call, però il compilatore non accetta

			// Collezione le chiamate per questo test case
			for _, step := range tc.setupSteps {
				call := step(mockGetUserPort, mockGetTenantPort)
				if call != nil {
					expectedCalls = append(expectedCalls, call)
				}
			}

			// Richiedi ordine nelle chiamate
			if len(expectedCalls) > 0 {
				gomock.InOrder(expectedCalls...)
			}

			// Crea servizio con porte mock
			_, _, _, _, _, getSuperAdminListUseCase := user.NewGetUserService(
				mockGetUserPort, mockGetTenantPort,
			)

			// Esegui funzione in oggetto
			obtainedUsers, total, err := getSuperAdminListUseCase.GetSuperAdminList(tc.input)

			// Assertions
			if err != tc.expectedError {
				t.Errorf("expected error %v, got %v", tc.expectedError, err)
			}
			if !slices.Equal(obtainedUsers, tc.expectedUsers) {
				t.Errorf("expected users %v, got %v", tc.expectedUsers, obtainedUsers)
			}
			if total != tc.expectedTotal {
				t.Errorf("expected total %v, got %v", tc.expectedTotal, total)
			}
		})
	}
}
