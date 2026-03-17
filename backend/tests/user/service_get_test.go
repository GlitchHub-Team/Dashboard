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

func newStepTenantOk(targetTenantId uuid.UUID, expectedTenant tenant.Tenant,) mockSetupFunc_GetUserService {
	return func(
		getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort,
	) *gomock.Call {
		return getTenantPort.EXPECT().
			GetTenant(targetTenantId).
			Return(expectedTenant, nil).
			Times(1)
	}
}

func newStepTenantNotFound(targetTenantId uuid.UUID, ) mockSetupFunc_GetUserService {
	return func(
		getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort,
	) *gomock.Call {
		return getTenantPort.EXPECT().
			GetTenant(targetTenantId).
			Return(tenant.Tenant{}, nil).
			Times(1)
	}
}

func newStepTenantError(targetTenantId uuid.UUID, mockError error) mockSetupFunc_GetUserService {
	return func(
		getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort,
	) *gomock.Call {
		return getTenantPort.EXPECT().
			GetTenant(targetTenantId).
			Return(tenant.Tenant{}, mockError).
			Times(1)
	}
}

func TestGetTenantUser(t *testing.T) {
	// Dati test
	targetTenantId := uuid.New()
	targetUserId := uint(1)

	expectedTenant := tenant.Tenant{Id: targetTenantId}

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

	baseInput := user.GetTenantUserCommand{
		TenantId: targetTenantId,
		UserId:   targetUserId,
	}

	// Test comportamentali
	step1TenantOk := newStepTenantOk(targetTenantId, expectedTenant)
	
	step1TenantFail := newStepTenantNotFound(targetTenantId)

	errMockStep1 := errors.New("unexpected error in step 1")
	step1TenantError := newStepTenantError(targetTenantId, errMockStep1)

	step2GetUserOk := func(
		getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantUser(targetTenantId, targetUserId).
			Return(expectedUser, nil).
			Times(1)
	}

	step2GetInexistentUserFail := func(
		getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantUser(targetTenantId, targetUserId).
			Return(user.User{}, nil).
			Times(1)
	}

	errMockStep2 := errors.New("unexpected error in step 2")
	step2GetUserError := func(
		getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantUser(targetTenantId, targetUserId).
			Return(user.User{}, errMockStep2).
			Times(1)
	}

	cases := []testCase{
		{
			name:  "Success: user created successfully",
			input: baseInput,
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk,
				step2GetUserOk,
			},
			expectedUser:  expectedUser,
			expectedError: nil,
		},

		{
			name:  "Fail (step 1): tenant not found",
			input: baseInput,
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantFail,
			},
			expectedUser:  user.User{},
			expectedError: tenant.ErrTenantNotFound,
		},
		{
			name:  "Fail (step 1): unexpected error",
			input: baseInput,
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantError,
			},
			expectedUser:  user.User{},
			expectedError: errMockStep1,
		},

		{
			name:  "Fail (step 2): user not found",
			input: baseInput,
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk,
				step2GetInexistentUserFail,
			},
			expectedUser:  user.User{},
			expectedError: user.ErrUserNotFound,
		},
		{
			name:  "Fail (step 2): unexpected error",
			input: baseInput,
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk,
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


func TestGetTenantAdmin(t *testing.T) {
	// Dati test
	targetTenantId := uuid.New()
	targetUserId := uint(1)

	expectedTenant := tenant.Tenant{Id: targetTenantId}

	expectedUser := user.User{
		Id:       targetUserId,
		Role:     identity.ROLE_TENANT_ADMIN,
		TenantId: &targetTenantId,
	}

	baseInput := user.GetTenantAdminCommand{
		TenantId: targetTenantId,
		UserId:   targetUserId,
	}

	type testCase struct {
		name          string
		input         user.GetTenantAdminCommand
		setupSteps    []mockSetupFunc_GetUserService
		expectedUser  user.User
		expectedError error
	}

	// Test comportamentali
	step1TenantOk := newStepTenantOk(targetTenantId, expectedTenant)
	
	step1TenantFail := newStepTenantNotFound(targetTenantId)

	errMockStep1 := errors.New("unexpected error in step 1")
	step1TenantError := newStepTenantError(targetTenantId, errMockStep1)

	step2GetUserOk := func(
		getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantAdmin(targetTenantId, targetUserId).
			Return(expectedUser, nil).
			Times(1)
	}

	step2GetInexistentUserFail := func(
		getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantAdmin(targetTenantId, targetUserId).
			Return(user.User{}, nil).
			Times(1)
	}

	errMockStep2 := errors.New("unexpected error in step 2")
	step2GetUserError := func(
		getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetTenantAdmin(targetTenantId, targetUserId).
			Return(user.User{}, errMockStep2).
			Times(1)
	}

	cases := []testCase{
		{
			name:  "Success: user created successfully",
			input: baseInput,
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk,
				step2GetUserOk,
			},
			expectedUser:  expectedUser,
			expectedError: nil,
		},

		{
			name:  "Fail (step 1): tenant not found",
			input: baseInput,
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantFail,
			},
			expectedUser:  user.User{},
			expectedError: tenant.ErrTenantNotFound,
		},
		{
			name:  "Fail (step 1): unexpected error",
			input: baseInput,
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantError,
			},
			expectedUser:  user.User{},
			expectedError: errMockStep1,
		},

		{
			name:  "Fail (step 2): user not found",
			input: baseInput,
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk,
				step2GetInexistentUserFail,
			},
			expectedUser:  user.User{},
			expectedError: user.ErrUserNotFound,
		},
		{
			name:  "Fail (step 2): unexpected error",
			input: baseInput,
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk,
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

func TestGetSuperAdmin(t *testing.T) {
	// Dati test
	targetUserId := uint(1)

	expectedUser := user.User{
		Id:       targetUserId,
		Role:     identity.ROLE_SUPER_ADMIN,
	}

	baseInput := user.GetSuperAdminCommand{
		UserId:   targetUserId,
	}

	type testCase struct {
		name          string
		input         user.GetSuperAdminCommand
		setupSteps    []mockSetupFunc_GetUserService
		expectedUser  user.User
		expectedError error
	}

	step1GetUserOk := func(
		getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetSuperAdmin(targetUserId).
			Return(expectedUser, nil).
			Times(1)
	}

	step1GetInexistentUserFail := func(
		getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetSuperAdmin(targetUserId).
			Return(user.User{}, nil).
			Times(1)
	}

	errMockStep1 := errors.New("unexpected error in step 2")
	step1GetUserError := func(
		getUserPort *mocks.MockGetUserPort, getTenantPort *tenantMocks.MockGetTenantPort,
	) *gomock.Call {
		return getUserPort.EXPECT().
			GetSuperAdmin(targetUserId).
			Return(user.User{}, errMockStep1).
			Times(1)
	}

	cases := []testCase{
		{
			name:  "Success: user created successfully",
			input: baseInput,
			setupSteps: []mockSetupFunc_GetUserService{
				step1GetUserOk,
			},
			expectedUser:  expectedUser,
			expectedError: nil,
		},
		{
			name:  "Fail: user not found",
			input: baseInput,
			setupSteps: []mockSetupFunc_GetUserService{
				step1GetInexistentUserFail,
			},
			expectedUser:  user.User{},
			expectedError: user.ErrUserNotFound,
		},
		{
			name:  "Fail: unexpected error",
			input: baseInput,
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
			_, _, getSuperAdminUseCase, _, _, _ := user.NewGetUserService(
				mockGetUserPort, mockGetTenantPort,
			)

			// Esegui funzione in oggetto
			createdUser, err := getSuperAdminUseCase.GetSuperAdmin(tc.input)

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


func TestGetTenantUsersByTenant(t *testing.T) {
	// Dati test
	targetTenantId := uuid.New()
	targetUserId := uint(1)

	// - Caso 1: page=1, limit=1
	targetPageCase1 := 1
	targetLimitCase1 := 1
	expectedUsersCase1 := []user.User{
		{
			Id: targetUserId,
			Role: identity.ROLE_TENANT_USER,
			TenantId: &targetTenantId,
		},
	}
	expectedTotalCase1 := uint(1)

	// - Caso 2: page=2, limit=1 (pagina vuota)
	targetPageCase2 := 2
	targetLimitCase2 := 2
	expectedUsersCase2 := ([]user.User)(nil) // NOTA: slice vuoto
	expectedTotalCase2 := uint(1)

	// - Valore atteso tenant
	expectedTenant := tenant.Tenant{Id: targetTenantId}
	

	type testCase struct {
		name string
		input user.GetTenantUsersByTenantCommand
		setupSteps []mockSetupFunc_GetUserService
		expectedUsers []user.User
		expectedTotal uint
		expectedError error
	}

	inputCase1 := user.GetTenantUsersByTenantCommand{
		Page: targetPageCase1,
		Limit: targetLimitCase1,
		TenantId: targetTenantId,
	}

	inputCase2 := user.GetTenantUsersByTenantCommand{
		Page: targetPageCase2,
		Limit: targetLimitCase2,
		TenantId: targetTenantId,
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

	step1TenantOk := newStepTenantOk(targetTenantId, expectedTenant)
	
	step1TenantFail := newStepTenantNotFound(targetTenantId)

	errMockStep1 := errors.New("unexpected error in step 1")
	step1TenantError := newStepTenantError(targetTenantId, errMockStep1)
 

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

	testCases := []testCase{
		{
			name: "Success: case 1: page=1, limit=1",
			input: inputCase1,
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk,
				step2Case1Ok,
			},
			expectedUsers: expectedUsersCase1,
			expectedTotal: expectedTotalCase1,
			expectedError: nil,
		},
		{
			name: "Success: case 2: page=2, limit=1 (empty page)",
			input: inputCase2,
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk,
				step2Case2Ok,
			},
			expectedUsers: expectedUsersCase2,
			expectedTotal: expectedTotalCase2,
			expectedError: nil,
		},
		{
			name: "Fail (step 1): tenant not found",
			input: inputCase1,
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantFail,
			},
			expectedUsers: nil,
			expectedTotal: 0,
			expectedError: tenant.ErrTenantNotFound,
		},
		{
			name: "Fail (step 1): unexpected error",
			input: inputCase1,
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantError,
			},
			expectedUsers: nil,
			expectedTotal: 0,
			expectedError: errMockStep1,
		},
		{
			name: "Fail (step 2): unexpected error",
			input: inputCase1,
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk,
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


func TestGetTenantAdminsByTenant(t *testing.T) {
	// Dati test
	targetTenantId := uuid.New()
	targetUserId := uint(1)

	// - Caso 1: page=1, limit=1
	targetPageCase1 := 1
	targetLimitCase1 := 1
	expectedUsersCase1 := []user.User{
		{
			Id: targetUserId,
			Role: identity.ROLE_TENANT_ADMIN,
			TenantId: &targetTenantId,
		},
	}
	expectedTotalCase1 := uint(1)

	// - Caso 2: page=2, limit=1 (pagina vuota)
	targetPageCase2 := 2
	targetLimitCase2 := 2
	expectedUsersCase2 := ([]user.User)(nil) // NOTA: slice vuoto
	expectedTotalCase2 := uint(1)

	// - Valore atteso tenant
	expectedTenant := tenant.Tenant{Id: targetTenantId}
	

	type testCase struct {
		name string
		input user.GetTenantAdminsByTenantCommand
		setupSteps []mockSetupFunc_GetUserService
		expectedUsers []user.User
		expectedTotal uint
		expectedError error
	}

	inputCase1 := user.GetTenantAdminsByTenantCommand{
		Page: targetPageCase1,
		Limit: targetLimitCase1,
		TenantId: targetTenantId,
	}

	inputCase2 := user.GetTenantAdminsByTenantCommand{
		Page: targetPageCase2,
		Limit: targetLimitCase2,
		TenantId: targetTenantId,
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

	step1TenantOk := newStepTenantOk(targetTenantId, expectedTenant)
	
	step1TenantFail := newStepTenantNotFound(targetTenantId)

	errMockStep1 := errors.New("unexpected error in step 1")
	step1TenantError := newStepTenantError(targetTenantId, errMockStep1)
 

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

	testCases := []testCase{
		{
			name: "Success: case 1: page=1, limit=1",
			input: inputCase1,
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk,
				step2Case1Ok,
			},
			expectedUsers: expectedUsersCase1,
			expectedTotal: expectedTotalCase1,
			expectedError: nil,
		},
		{
			name: "Success: case 2: page=2, limit=1 (empty page)",
			input: inputCase2,
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk,
				step2Case2Ok,
			},
			expectedUsers: expectedUsersCase2,
			expectedTotal: expectedTotalCase2,
			expectedError: nil,
		},
		{
			name: "Fail (step 1): tenant not found",
			input: inputCase1,
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantFail,
			},
			expectedUsers: nil,
			expectedTotal: 0,
			expectedError: tenant.ErrTenantNotFound,
		},
		{
			name: "Fail (step 1): unexpected error",
			input: inputCase1,
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantError,
			},
			expectedUsers: nil,
			expectedTotal: 0,
			expectedError: errMockStep1,
		},
		{
			name: "Fail (step 2): unexpected error",
			input: inputCase1,
			setupSteps: []mockSetupFunc_GetUserService{
				step1TenantOk,
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

func TestGetSuperAdminList(t *testing.T) {
	// Dati test
	targetUserId := uint(1)

	// - Caso 1: page=1, limit=1
	targetPageCase1 := 1
	targetLimitCase1 := 1
	expectedUsersCase1 := []user.User{
		{
			Id: targetUserId,
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
		name string
		input user.GetSuperAdminListCommand
		setupSteps []mockSetupFunc_GetUserService
		expectedUsers []user.User
		expectedTotal uint
		expectedError error
	}

	inputCase1 := user.GetSuperAdminListCommand{
		Page: targetPageCase1,
		Limit: targetLimitCase1,
	}

	inputCase2 := user.GetSuperAdminListCommand{
		Page: targetPageCase2,
		Limit: targetLimitCase2,
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

	testCases := []testCase{
		{
			name: "Success: case 1: page=1, limit=1",
			input: inputCase1,
			setupSteps: []mockSetupFunc_GetUserService{
				step1Case1Ok,
			},
			expectedUsers: expectedUsersCase1,
			expectedTotal: expectedTotalCase1,
			expectedError: nil,
		},
		{
			name: "Success: case 2: page=2, limit=1 (empty page)",
			input: inputCase2,
			setupSteps: []mockSetupFunc_GetUserService{
				step1Case2Ok,
			},
			expectedUsers: expectedUsersCase2,
			expectedTotal: expectedTotalCase2,
			expectedError: nil,
		},
		{
			name: "Fail (step 2): unexpected error",
			input: inputCase1,
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
