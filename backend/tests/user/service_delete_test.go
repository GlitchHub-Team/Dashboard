package user_test

import (
	"testing"

	"backend/internal/tenant"
	"backend/internal/user"
	tenantMocks "backend/tests/tenant/mocks"
	"backend/tests/user/mocks"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestDeleteTenantUser(t *testing.T) {
	// Dati test
	targetTenantId := uuid.New()
	targetUserId := uint(1)
	expectedUser := user.User{
		Id:       targetUserId,
		Name:     "Test",
		TenantId: &targetTenantId,
	}

	type mockSetupFunc func(
		deletePort *mocks.MockDeleteUserPort,
		getUserPort *mocks.MockGetUserPort,
		getTenantPort *tenantMocks.MockGetTenantPort,
	)
	type testCase struct {
		name          string
		tenantId      uuid.UUID
		userId        uint
		setupMocks    mockSetupFunc
		expectedError error
		expectedUser  user.User
	}

	cases := []testCase{
		{
			name:     "Success: user deleted successfully",
			tenantId: targetTenantId,
			userId:   targetUserId,
			setupMocks: func(
				deletePort *mocks.MockDeleteUserPort,
				getUserPort *mocks.MockGetUserPort,
				getTenantPort *tenantMocks.MockGetTenantPort,
			) {
				gomock.InOrder(
					getTenantPort.EXPECT().
						GetTenant(targetTenantId).
						Return(tenant.Tenant{Id: targetTenantId}, nil).
						Times(1),

					getUserPort.EXPECT().
						GetTenantUser(targetTenantId, targetUserId).
						Return(expectedUser, nil).
						Times(1),

					deletePort.EXPECT().
						DeleteTenantUser(targetTenantId, targetUserId).
						Return(expectedUser, nil).
						Times(1),
				)
			},
			expectedError: nil,
			expectedUser:  expectedUser,
		},
		{
			name:     "Fail: tenant not found",
			tenantId: targetTenantId,
			userId:   targetUserId,
			setupMocks: func(
				deletePort *mocks.MockDeleteUserPort,
				getUserPort *mocks.MockGetUserPort,
				getTenantPort *tenantMocks.MockGetTenantPort,
			) {
				getTenantPort.EXPECT().
					GetTenant(targetTenantId).
					Return(tenant.Tenant{}, tenant.ErrTenantNotFound).
					Times(1)
			},
			expectedError: tenant.ErrTenantNotFound,
			expectedUser:  user.User{},
		},
		{
			name:     "Fail: tenant user not found",
			tenantId: targetTenantId,
			userId:   targetUserId,
			setupMocks: func(
				deletePort *mocks.MockDeleteUserPort,
				getUserPort *mocks.MockGetUserPort,
				getTenantPort *tenantMocks.MockGetTenantPort,
			) {
				gomock.InOrder(
					getTenantPort.EXPECT().
						GetTenant(targetTenantId).
						Return(tenant.Tenant{Id: targetTenantId}, nil).
						Times(1),

					getUserPort.EXPECT().
						GetTenantUser(targetTenantId, targetUserId).
						Return(user.User{}, user.ErrUserNotFound).
						Times(1),
				)
			},
			expectedError: user.ErrUserNotFound,
			expectedUser:  user.User{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// NOTA: il controller di gomock va inizializzato qua dentro!
			mockController := gomock.NewController(t)

			mockDeletePort := mocks.NewMockDeleteUserPort(mockController)
			mockGetPort := mocks.NewMockGetUserPort(mockController)
			mockTenantPort := tenantMocks.NewMockGetTenantPort(mockController)

			// Execute the mock setup specific to this test case
			tc.setupMocks(mockDeletePort, mockGetPort, mockTenantPort)

			// Crea servizio con porte mock
			deleteTenantUserUseCase, _, _ := user.NewDeleteUserService(mockDeletePort, mockGetPort, mockTenantPort)

			// Esegui funzione in oggetto
			deletedUser, err := deleteTenantUserUseCase.DeleteTenantUser(user.DeleteTenantUserCommand{
				TenantId: tc.tenantId,
				UserId:   tc.userId,
			})

			// Assertions
			if err != tc.expectedError {
				t.Errorf("expected error %v, got %v", tc.expectedError, err)
			}
			if deletedUser != tc.expectedUser {
				t.Errorf("expected user %v, got %v", tc.expectedUser, deletedUser)
			}
		})
	}
}

func TestDeleteTenantAdmin(t *testing.T) {
	// Dati test
	targetTenantId := uuid.New()
	targetUserId := uint(1)
	expectedUser := user.User{
		Id:       targetUserId,
		Name:     "Test",
		TenantId: &targetTenantId,
	}

	type testCaseMockSetup func(
		deletePort *mocks.MockDeleteUserPort,
		getUserPort *mocks.MockGetUserPort,
		getTenantPort *tenantMocks.MockGetTenantPort,
	)
	type testCase struct {
		name          string
		tenantId      uuid.UUID
		userId        uint
		setupMocks    testCaseMockSetup
		expectedError error
		expectedUser  user.User
	}

	cases := []testCase{
		{
			name:     "Success: tenant admin deleted successfully",
			tenantId: targetTenantId,
			userId:   targetUserId,
			setupMocks: func(
				deletePort *mocks.MockDeleteUserPort,
				getUserPort *mocks.MockGetUserPort,
				getTenantPort *tenantMocks.MockGetTenantPort,
			) {
				gomock.InOrder(
					getTenantPort.EXPECT().
						GetTenant(targetTenantId).
						Return(tenant.Tenant{Id: targetTenantId}, nil).
						Times(1),

					getUserPort.EXPECT().
						GetTenantAdmin(targetTenantId, targetUserId).
						Return(expectedUser, nil).
						Times(1),

					deletePort.EXPECT().
						DeleteTenantAdmin(targetTenantId, targetUserId).
						Return(expectedUser, nil).
						Times(1),
				)
			},
			expectedError: nil,
			expectedUser:  expectedUser,
		},
		{
			name:     "Fail: Tenant not found",
			tenantId: targetTenantId,
			userId:   targetUserId,
			setupMocks: func(
				deletePort *mocks.MockDeleteUserPort,
				getUserPort *mocks.MockGetUserPort,
				getTenantPort *tenantMocks.MockGetTenantPort,
			) {
				getTenantPort.EXPECT().
					GetTenant(targetTenantId).
					Return(tenant.Tenant{}, tenant.ErrTenantNotFound).
					Times(1)
			},
			expectedError: tenant.ErrTenantNotFound,
			expectedUser:  user.User{},
		},
		{
			name:     "Fail: tenant admin not found",
			tenantId: targetTenantId,
			userId:   targetUserId,
			setupMocks: func(
				deletePort *mocks.MockDeleteUserPort,
				getUserPort *mocks.MockGetUserPort,
				getTenantPort *tenantMocks.MockGetTenantPort,
			) {
				gomock.InOrder(
					getTenantPort.EXPECT().
						GetTenant(targetTenantId).
						Return(tenant.Tenant{Id: targetTenantId}, nil).
						Times(1),

					getUserPort.EXPECT().
						GetTenantAdmin(targetTenantId, targetUserId).
						Return(user.User{}, user.ErrUserNotFound).
						Times(1),
				)
			},
			expectedError: user.ErrUserNotFound,
			expectedUser:  user.User{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// NOTA: il controller di gomock va inizializzato qua dentro!
			mockController := gomock.NewController(t)

			mockDeletePort := mocks.NewMockDeleteUserPort(mockController)
			mockGetPort := mocks.NewMockGetUserPort(mockController)
			mockTenantPort := tenantMocks.NewMockGetTenantPort(mockController)

			// Execute the mock setup specific to this test case
			tc.setupMocks(mockDeletePort, mockGetPort, mockTenantPort)

			// Crea servizio con porte mock
			_, deleteTenantAdminUseCase, _ := user.NewDeleteUserService(mockDeletePort, mockGetPort, mockTenantPort)

			// Esegui funzione in oggetto
			deletedUser, err := deleteTenantAdminUseCase.DeleteTenantAdmin(user.DeleteTenantAdminCommand{
				TenantId: tc.tenantId,
				UserId:   tc.userId,
			})

			// Assertions
			if err != tc.expectedError {
				t.Errorf("expected error %v, got %v", tc.expectedError, err)
			}
			if deletedUser != tc.expectedUser {
				t.Errorf("expected user %v, got %v", tc.expectedUser, deletedUser)
			}
		})
	}
}

func TestDeleteSuperAdmin(t *testing.T) {
	// Dati test
	targetUserId := uint(1)
	expectedUser := user.User{
		Id:       targetUserId,
		Name:     "Test",
		TenantId: nil,
	}

	type testCaseMockSetup func(
		deletePort *mocks.MockDeleteUserPort,
		getUserPort *mocks.MockGetUserPort,
	)
	type testCase struct {
		name          string
		userId        uint
		setupMocks    testCaseMockSetup
		expectedError error
		expectedUser  user.User
	}

	cases := []testCase{
		{
			name:   "Success: super admin deleted successfully",
			userId: targetUserId,
			setupMocks: func(
				deletePort *mocks.MockDeleteUserPort,
				getUserPort *mocks.MockGetUserPort,
			) {
				gomock.InOrder(
					getUserPort.EXPECT().
						GetSuperAdmin(targetUserId).
						Return(expectedUser, nil).
						Times(1),

					deletePort.EXPECT().
						DeleteSuperAdmin(targetUserId).
						Return(expectedUser, nil).
						Times(1),
				)
			},
			expectedError: nil,
			expectedUser:  expectedUser,
		},
		{
			name:   "Fail: super admin not found",
			userId: targetUserId,
			setupMocks: func(
				deletePort *mocks.MockDeleteUserPort,
				getUserPort *mocks.MockGetUserPort,
			) {
				gomock.InOrder(
					getUserPort.EXPECT().
						GetSuperAdmin(targetUserId).
						Return(user.User{}, user.ErrUserNotFound).
						Times(1),
				)
			},
			expectedError: user.ErrUserNotFound,
			expectedUser:  user.User{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// NOTA: il controller di gomock va inizializzato qua dentro!
			mockController := gomock.NewController(t)

			mockDeletePort := mocks.NewMockDeleteUserPort(mockController)
			mockGetPort := mocks.NewMockGetUserPort(mockController)

			// Execute the mock setup specific to this test case
			tc.setupMocks(mockDeletePort, mockGetPort)

			// Crea servizio con porte mock
			_, _, deleteSuperAdminUseCase := user.NewDeleteUserService(mockDeletePort, mockGetPort, nil)

			// Esegui funzione in oggetto
			deletedUser, err := deleteSuperAdminUseCase.DeleteSuperAdmin(user.DeleteSuperAdminCommand{
				UserId: tc.userId,
			})

			// Assertions
			if err != tc.expectedError {
				t.Errorf("expected error %v, got %v", tc.expectedError, err)
			}
			if deletedUser != tc.expectedUser {
				t.Errorf("expected user %v, got %v", tc.expectedUser, deletedUser)
			}
		})
	}
}
