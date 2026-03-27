package user_test

import (
	"errors"
	"reflect"
	"testing"

	"backend/internal/infra/database/pagination"
	"backend/internal/shared/identity"
	"backend/internal/user"
	"backend/tests/user/mocks"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	// "backend/internal/user"
	// "go.uber.org/mock/gomock"
)

type mockSetupFunc_userPostgreAdapter func(
	tenantMemberRepo *mocks.MockTenantMemberRepository,
	superAdminRepo *mocks.MockSuperAdminRepository,
) *gomock.Call

func setupMockSteps_UserPostgreAdapter(
	t *testing.T,
	setupSteps []mockSetupFunc_userPostgreAdapter,
) (*mocks.MockTenantMemberRepository, *mocks.MockSuperAdminRepository) {
	ctrl := gomock.NewController(t)

	mockTenantMemberRepo := mocks.NewMockTenantMemberRepository(ctrl)
	mockSuperAdminRepo := mocks.NewMockSuperAdminRepository(ctrl)

	var expectedCalls []any
	for _, step := range setupSteps {
		if call := step(mockTenantMemberRepo, mockSuperAdminRepo); call != nil {
			expectedCalls = append(expectedCalls, call)
		}
	}

	if len(expectedCalls) > 0 {
		gomock.InOrder(expectedCalls...)
	}

	return mockTenantMemberRepo, mockSuperAdminRepo
}

func TestUserPostgreAdapter_SaveUser(t *testing.T) {
	type testCase struct {
		name          string
		inputUser     user.User
		setupSteps    []mockSetupFunc_userPostgreAdapter
		expectedUser  user.User
		expectedError error
	}

	// Input ----------------------------------------------------------------------------------
	tenantMemberUuid := uuid.New()
	tenantMemberUser := user.User{
		Id:           uint(1),
		Name:         "username",
		Email:        "info@example.com",
		PasswordHash: new(string),
		Role:         identity.ROLE_TENANT_ADMIN,
		TenantId:     &tenantMemberUuid,
		Confirmed:    true,
	}

	superAdminUser := user.User{
		Id:           uint(1),
		Name:         "username",
		Email:        "info@example.com",
		PasswordHash: new(string),
		Role:         identity.ROLE_SUPER_ADMIN,
		TenantId:     nil,
		Confirmed:    true,
	}

	invalidUser := user.User{
		Role: "",
	}

	mockTenantMemberEntity := gomock.AssignableToTypeOf(&user.TenantMemberEntity{})
	mockSuperAdminEntity := gomock.AssignableToTypeOf(&user.SuperAdminEntity{})

	// SaveTenantMember() ---------------------------------------------------------------------

	saveTenantMemberOk := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			SaveTenantMember(mockTenantMemberEntity).
			Return(nil).
			Times(1)
	}

	mockSaveTenantMemberError := errors.New("unexpected error (tenant member)")
	saveTenantMemberError := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			SaveTenantMember(mockTenantMemberEntity).
			Return(mockSaveTenantMemberError).
			Times(1)
	}

	saveTenantMemberNeverCalled := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			SaveTenantMember(gomock.Any()).
			Times(0)
	}

	// SaveSuperAdmin() ---------------------------------------------------------------------

	saveSuperAdminOk := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return superAdminRepo.EXPECT().
			SaveSuperAdmin(mockSuperAdminEntity).
			Return(nil).
			Times(1)
	}

	mockSaveSuperAdminError := errors.New("unexpected error (super admin)")
	saveSuperAdminError := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return superAdminRepo.EXPECT().
			SaveSuperAdmin(mockSuperAdminEntity).
			Return(mockSaveSuperAdminError).
			Times(1)
	}

	saveSuperAdminNeverCalled := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return superAdminRepo.EXPECT().
			SaveSuperAdmin(gomock.Any()).
			Times(0)
	}

	cases := []testCase{
		// Tenant Member
		{
			name:      "(Tenant Member) Success",
			inputUser: tenantMemberUser,
			setupSteps: []mockSetupFunc_userPostgreAdapter{
				saveTenantMemberOk,
				saveSuperAdminNeverCalled,
			},
			expectedUser:  tenantMemberUser,
			expectedError: nil,
		},
		{
			name:      "(Tenant Member) Fail: Unexpected error",
			inputUser: tenantMemberUser,
			setupSteps: []mockSetupFunc_userPostgreAdapter{
				saveTenantMemberError,
				saveSuperAdminNeverCalled,
			},
			expectedUser:  user.User{},
			expectedError: mockSaveTenantMemberError,
		},

		// Super Admin
		{
			name:      "(Super Admin) Success",
			inputUser: superAdminUser,
			setupSteps: []mockSetupFunc_userPostgreAdapter{
				saveSuperAdminOk,
				saveTenantMemberNeverCalled,
			},
			expectedUser:  superAdminUser,
			expectedError: nil,
		},
		{
			name:      "(Super Admin) Fail: Unexpected error",
			inputUser: superAdminUser,
			setupSteps: []mockSetupFunc_userPostgreAdapter{
				saveSuperAdminError,
				saveTenantMemberNeverCalled,
			},
			expectedUser:  user.User{},
			expectedError: mockSaveSuperAdminError,
		},

		// Invalid role
		{
			name:      "Fail: Invalid role",
			inputUser: invalidUser,
			setupSteps: []mockSetupFunc_userPostgreAdapter{
				saveSuperAdminNeverCalled,
				saveTenantMemberNeverCalled,
			},
			expectedUser:  user.User{},
			expectedError: identity.ErrUnknownRole,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockTenantMemberRepo, mockSuperAdminRepo := setupMockSteps_UserPostgreAdapter(t, tc.setupSteps)

			adapter := user.NewUserPostgreAdapter(nil, mockTenantMemberRepo, mockSuperAdminRepo)

			savedUser, err := adapter.SaveUser(tc.inputUser)

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("want error %v, got error %v", tc.expectedError, err)
			}

			if !reflect.DeepEqual(tc.expectedUser, savedUser) {
				t.Errorf("want %+#v, got %#+v", tc.expectedUser, savedUser)
			}
		})
	}
}

func TestUserPostgreAdapter_DeleteTenantUser(t *testing.T) {
	type testCase struct {
		name          string
		inputTenantId uuid.UUID
		inputUserId   uint
		setupSteps    []mockSetupFunc_userPostgreAdapter
		expectedUser  user.User
		expectedError error
	}

	// Input ----------------------------------------------------------------------------------
	targetUserId := uint(1)
	targetName := "username"
	targetEmail := "info@example.com"
	targetPassword := "123"
	targetRole := identity.ROLE_TENANT_USER
	targetTenantId := uuid.New()
	targetConfirmed := true

	expectedTenantUser := user.User{
		Id:           targetUserId,
		Name:         targetName,
		Email:        targetEmail,
		PasswordHash: &targetPassword,
		Role:         targetRole,
		TenantId:     &targetTenantId,
		Confirmed:    targetConfirmed,
	}

	expectedEntity := user.TenantMemberEntity{
		ID:        targetUserId,
		Name:      targetName,
		Email:     targetEmail,
		Password:  &targetPassword,
		Role:      string(targetRole),
		TenantId:  targetTenantId.String(),
		Confirmed: targetConfirmed,
	}

	mockTenantMemberEntity := gomock.AssignableToTypeOf(&user.TenantMemberEntity{})

	stepDeleteTenantMemberOk := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			DeleteTenantMember(mockTenantMemberEntity).
			Do(func(entity *user.TenantMemberEntity) {
				*entity = expectedEntity
			}).
			Return(nil).
			Times(1)
	}

	mockError := errors.New("unexpected error (tenant member)")
	stepDeleteTenantMemberError := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			DeleteTenantMember(mockTenantMemberEntity).
			Return(mockError).
			Times(1)
	}

	cases := []testCase{
		{
			name:          "Success",
			inputTenantId: targetTenantId,
			inputUserId:   targetUserId,
			setupSteps: []mockSetupFunc_userPostgreAdapter{
				stepDeleteTenantMemberOk,
			},
			expectedUser:  expectedTenantUser,
			expectedError: nil,
		},
		{
			name:          "Fail: Unexpected error",
			inputTenantId: targetTenantId,
			inputUserId:   targetUserId,
			setupSteps: []mockSetupFunc_userPostgreAdapter{
				stepDeleteTenantMemberError,
			},
			expectedUser:  user.User{},
			expectedError: mockError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockTenantMemberRepo, mockSuperAdminRepo := setupMockSteps_UserPostgreAdapter(t, tc.setupSteps)

			adapter := user.NewUserPostgreAdapter(nil, mockTenantMemberRepo, mockSuperAdminRepo)

			savedUser, err := adapter.DeleteTenantUser(tc.inputTenantId, tc.inputUserId)

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("want error %v, got error %v", tc.expectedError, err)
			}

			if !reflect.DeepEqual(tc.expectedUser, savedUser) {
				t.Errorf("want %+#v, got %#+v", tc.expectedUser, savedUser)
			}
		})
	}
}

func TestUserPostgreAdapter_DeleteTenantAdmin(t *testing.T) {
	type testCase struct {
		name          string
		inputTenantId uuid.UUID
		inputUserId   uint
		setupSteps    []mockSetupFunc_userPostgreAdapter
		expectedUser  user.User
		expectedError error
	}

	// Input ----------------------------------------------------------------------------------
	targetUserId := uint(1)
	targetName := "username"
	targetEmail := "info@example.com"
	targetPassword := "123"
	targetRole := identity.ROLE_TENANT_ADMIN
	targetTenantId := uuid.New()
	targetConfirmed := true

	expectedTenantAdmin := user.User{
		Id:           targetUserId,
		Name:         targetName,
		Email:        targetEmail,
		PasswordHash: &targetPassword,
		Role:         targetRole,
		TenantId:     &targetTenantId,
		Confirmed:    targetConfirmed,
	}

	expectedEntity := user.TenantMemberEntity{
		ID:        targetUserId,
		Name:      targetName,
		Email:     targetEmail,
		Password:  &targetPassword,
		Role:      string(targetRole),
		TenantId:  targetTenantId.String(),
		Confirmed: targetConfirmed,
	}

	mockTenantMemberEntity := gomock.AssignableToTypeOf(&user.TenantMemberEntity{})

	stepDeleteTenantMemberOk := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			DeleteTenantMember(mockTenantMemberEntity).
			Do(func(entity *user.TenantMemberEntity) {
				*entity = expectedEntity
			}).
			Return(nil).
			Times(1)
	}

	mockError := errors.New("unexpected error (tenant member)")
	stepDeleteTenantMemberError := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			DeleteTenantMember(mockTenantMemberEntity).
			Return(mockError).
			Times(1)
	}

	cases := []testCase{
		{
			name:          "Success",
			inputTenantId: targetTenantId,
			inputUserId:   targetUserId,
			setupSteps: []mockSetupFunc_userPostgreAdapter{
				stepDeleteTenantMemberOk,
			},
			expectedUser:  expectedTenantAdmin,
			expectedError: nil,
		},
		{
			name:          "Fail: Unexpected error",
			inputTenantId: targetTenantId,
			inputUserId:   targetUserId,
			setupSteps: []mockSetupFunc_userPostgreAdapter{
				stepDeleteTenantMemberError,
			},
			expectedUser:  user.User{},
			expectedError: mockError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockTenantMemberRepo, mockSuperAdminRepo := setupMockSteps_UserPostgreAdapter(t, tc.setupSteps)

			adapter := user.NewUserPostgreAdapter(nil, mockTenantMemberRepo, mockSuperAdminRepo)

			savedUser, err := adapter.DeleteTenantAdmin(tc.inputTenantId, tc.inputUserId)

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("want error %v, got error %v", tc.expectedError, err)
			}

			if !reflect.DeepEqual(tc.expectedUser, savedUser) {
				t.Errorf("want %+#v, got %#+v", tc.expectedUser, savedUser)
			}
		})
	}
}

func TestUserPostgreAdapter_DeleteSuperAdmin(t *testing.T) {
	type testCase struct {
		name          string
		inputUserId   uint
		setupSteps    []mockSetupFunc_userPostgreAdapter
		expectedUser  user.User
		expectedError error
	}

	// Input ----------------------------------------------------------------------------------
	targetUserId := uint(1)
	targetName := "username"
	targetEmail := "info@example.com"
	targetPassword := "123"
	targetRole := identity.ROLE_SUPER_ADMIN
	targetConfirmed := true

	expectedSuperAdmin := user.User{
		Id:           targetUserId,
		Name:         targetName,
		Email:        targetEmail,
		PasswordHash: &targetPassword,
		Role:         targetRole,
		TenantId:     nil,
		Confirmed:    targetConfirmed,
	}

	expectedEntity := user.SuperAdminEntity{
		ID:        targetUserId,
		Name:      targetName,
		Email:     targetEmail,
		Password:  &targetPassword,
		Confirmed: targetConfirmed,
	}

	mockSuperAdminEntity := gomock.AssignableToTypeOf(&user.SuperAdminEntity{})

	stepDeleteTenantMemberOk := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return superAdminRepo.EXPECT().
			DeleteSuperAdmin(mockSuperAdminEntity).
			Do(func(entity *user.SuperAdminEntity) {
				*entity = expectedEntity
			}).
			Return(nil).
			Times(1)
	}

	mockError := errors.New("unexpected error (tenant member)")
	stepDeleteTenantMemberError := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return superAdminRepo.EXPECT().
			DeleteSuperAdmin(mockSuperAdminEntity).
			Return(mockError).
			Times(1)
	}

	cases := []testCase{
		{
			name:        "Success",
			inputUserId: targetUserId,
			setupSteps: []mockSetupFunc_userPostgreAdapter{
				stepDeleteTenantMemberOk,
			},
			expectedUser:  expectedSuperAdmin,
			expectedError: nil,
		},
		{
			name:        "Fail: Unexpected error",
			inputUserId: targetUserId,
			setupSteps: []mockSetupFunc_userPostgreAdapter{
				stepDeleteTenantMemberError,
			},
			expectedUser:  user.User{},
			expectedError: mockError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockTenantMemberRepo, mockSuperAdminRepo := setupMockSteps_UserPostgreAdapter(t, tc.setupSteps)

			adapter := user.NewUserPostgreAdapter(nil, mockTenantMemberRepo, mockSuperAdminRepo)

			savedUser, err := adapter.DeleteSuperAdmin(tc.inputUserId)

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("want error %v, got error %v", tc.expectedError, err)
			}

			if !reflect.DeepEqual(tc.expectedUser, savedUser) {
				t.Errorf("want %+#v, got %#+v", tc.expectedUser, savedUser)
			}
		})
	}
}

func TestUserPostgreAdapter_GetUser(t *testing.T) {
	type testCase struct {
		name          string
		inputTenantId *uuid.UUID
		inputUserId   uint
		setupSteps    []mockSetupFunc_userPostgreAdapter
		expectedUser  user.User
		expectedError error
	}

	// Input ----------------------------------------------------------------------------------
	targetUserId := uint(1)
	targetName := "username"
	targetEmail := "info@example.com"
	targetPassword := "123"
	targetConfirmed := true
	targetTenantId := uuid.New()

	targetTenantMemberRole := identity.ROLE_TENANT_ADMIN
	tenantMemberUser := user.User{
		Id:           targetUserId,
		Name:         targetName,
		Email:        targetEmail,
		PasswordHash: &targetPassword,
		Role:         targetTenantMemberRole,
		TenantId:     &targetTenantId,
		Confirmed:    true,
	}

	tenantMemberEntity := &user.TenantMemberEntity{
		ID:        targetUserId,
		Email:     targetEmail,
		Name:      targetName,
		Password:  &targetPassword,
		Confirmed: targetConfirmed,
		Role:      string(targetTenantMemberRole),
		TenantId:  targetTenantId.String(),
	}

	targetSuperAdminRole := identity.ROLE_SUPER_ADMIN
	superAdminUser := user.User{
		Id:           targetUserId,
		Name:         targetName,
		Email:        targetEmail,
		PasswordHash: &targetPassword,
		Role:         targetSuperAdminRole,
		TenantId:     nil,
		Confirmed:    true,
	}

	superAdminEntity := &user.SuperAdminEntity{
		ID:        targetUserId,
		Email:     targetEmail,
		Name:      targetName,
		Password:  &targetPassword,
		Confirmed: targetConfirmed,
	}

	// invalidUser := user.User{
	// 	Role: "",
	// }

	// mockTenantMemberEntity := gomock.AssignableToTypeOf(&user.TenantMemberEntity{})
	// mockSuperAdminEntity := gomock.AssignableToTypeOf(&user.SuperAdminEntity{})
	mockBy := gomock.AssignableToTypeOf(user.UserRepositoryGetUserBy{})

	// GetTenantMember() ---------------------------------------------------------------------

	getTenantMemberOk := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			GetTenantMember(targetTenantId.String(), mockBy).
			Return(tenantMemberEntity, nil).
			Times(1)
	}

	getTenantMemberOk_ReturnNil := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			GetTenantMember(targetTenantId.String(), mockBy).
			Return(nil, nil).
			Times(1)
	}

	mockGetTenantMemberError := errors.New("unexpected error (tenant member)")
	getTenantMemberError := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			GetTenantMember(targetTenantId.String(), mockBy).
			Return(&user.TenantMemberEntity{}, mockGetTenantMemberError).
			Times(1)
	}

	getTenantMemberNeverCalled := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			GetTenantMember(gomock.Any(), gomock.Any()).
			Times(0)
	}

	// GetSuperAdmin() ---------------------------------------------------------------------

	getSuperAdminOk := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return superAdminRepo.EXPECT().
			GetSuperAdmin(mockBy).
			Return(superAdminEntity, nil).
			Times(1)
	}

	getSuperAdminOk_NilReturned := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return superAdminRepo.EXPECT().
			GetSuperAdmin(mockBy).
			Return(nil, nil).
			Times(1)
	}


	mockGetSuperAdminError := errors.New("unexpected error (super admin)")
	getSuperAdminError := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return superAdminRepo.EXPECT().
			GetSuperAdmin(mockBy).
			Return(&user.SuperAdminEntity{}, mockGetSuperAdminError).
			Times(1)
	}

	getSuperAdminNeverCalled := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return superAdminRepo.EXPECT().
			GetSuperAdmin(gomock.Any()).
			Times(0)
	}

	cases := []testCase{
		// Tenant Member
		{
			name:          "(Tenant Member) Success",
			inputTenantId: &targetTenantId,
			inputUserId:   targetUserId,
			setupSteps: []mockSetupFunc_userPostgreAdapter{
				getTenantMemberOk,
				getSuperAdminNeverCalled,
			},
			expectedUser:  tenantMemberUser,
			expectedError: nil,
		},
		{
			name:          "(Tenant Member) Success (nil returned)",
			inputTenantId: &targetTenantId,
			inputUserId:   targetUserId,
			setupSteps: []mockSetupFunc_userPostgreAdapter{
				getTenantMemberOk_ReturnNil,
				getSuperAdminNeverCalled,
			},
			expectedUser:  user.User{},
			expectedError: nil,
		},
		{
			name:          "(Tenant Member) Fail: Unexpected error",
			inputTenantId: &targetTenantId,
			inputUserId:   targetUserId,
			setupSteps: []mockSetupFunc_userPostgreAdapter{
				getTenantMemberError,
				getSuperAdminNeverCalled,
			},
			expectedUser:  user.User{},
			expectedError: mockGetTenantMemberError,
		},

		// Super Admin
		{
			name:          "(Super Admin) Success",
			inputTenantId: nil,
			inputUserId:   targetUserId,
			setupSteps: []mockSetupFunc_userPostgreAdapter{
				getTenantMemberNeverCalled,
				getSuperAdminOk,
			},
			expectedUser:  superAdminUser,
			expectedError: nil,
		},
		{
			name:          "(Super Admin) Success (nil returned)",
			inputTenantId: nil,
			inputUserId:   targetUserId,
			setupSteps: []mockSetupFunc_userPostgreAdapter{
				getTenantMemberNeverCalled,
				getSuperAdminOk_NilReturned,
			},
			expectedUser:  user.User{},
			expectedError: nil,
		},
		{
			name:          "(Super Admin) Fail: Unexpected error",
			inputTenantId: nil,
			inputUserId:   targetUserId,
			setupSteps: []mockSetupFunc_userPostgreAdapter{
				getTenantMemberNeverCalled,
				getSuperAdminError,
			},
			expectedUser:  user.User{},
			expectedError: mockGetSuperAdminError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockTenantMemberRepo, mockSuperAdminRepo := setupMockSteps_UserPostgreAdapter(t, tc.setupSteps)

			adapter := user.NewUserPostgreAdapter(nil, mockTenantMemberRepo, mockSuperAdminRepo)

			savedUser, err := adapter.GetUser(tc.inputTenantId, tc.inputUserId)

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("want error %v, got error %v", tc.expectedError, err)
			}

			if !reflect.DeepEqual(tc.expectedUser, savedUser) {
				t.Errorf("want %+#v, got %#+v", tc.expectedUser, savedUser)
			}
		})
	}
}

func TestUserPostgreAdapter_GetUserByEmail(t *testing.T) {
	type testCase struct {
		name          string
		inputTenantId *uuid.UUID
		inputEmail    string
		setupSteps    []mockSetupFunc_userPostgreAdapter
		expectedUser  user.User
		expectedError error
	}

	// Input ----------------------------------------------------------------------------------
	targetUserId := uint(1)
	targetName := "username"
	targetEmail := "info@example.com"
	targetPassword := "123"
	targetConfirmed := true
	targetTenantId := uuid.New()

	targetTenantMemberRole := identity.ROLE_TENANT_ADMIN
	tenantMemberUser := user.User{
		Id:           targetUserId,
		Name:         targetName,
		Email:        targetEmail,
		PasswordHash: &targetPassword,
		Role:         targetTenantMemberRole,
		TenantId:     &targetTenantId,
		Confirmed:    true,
	}

	tenantMemberEntity := &user.TenantMemberEntity{
		ID:        targetUserId,
		Email:     targetEmail,
		Name:      targetName,
		Password:  &targetPassword,
		Confirmed: targetConfirmed,
		Role:      string(targetTenantMemberRole),
		TenantId:  targetTenantId.String(),
	}

	targetSuperAdminRole := identity.ROLE_SUPER_ADMIN
	superAdminUser := user.User{
		Id:           targetUserId,
		Name:         targetName,
		Email:        targetEmail,
		PasswordHash: &targetPassword,
		Role:         targetSuperAdminRole,
		TenantId:     nil,
		Confirmed:    true,
	}

	superAdminEntity := &user.SuperAdminEntity{
		ID:        targetUserId,
		Email:     targetEmail,
		Name:      targetName,
		Password:  &targetPassword,
		Confirmed: targetConfirmed,
	}

	mockBy := gomock.AssignableToTypeOf(user.UserRepositoryGetUserBy{})

	// GetTenantMember() ---------------------------------------------------------------------

	getTenantMemberOk := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			GetTenantMember(targetTenantId.String(), mockBy).
			Return(tenantMemberEntity, nil).
			Times(1)
	}

	mockGetTenantMemberError := errors.New("unexpected error (tenant member)")
	getTenantMemberError := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			GetTenantMember(targetTenantId.String(), mockBy).
			Return(&user.TenantMemberEntity{}, mockGetTenantMemberError).
			Times(1)
	}

	getTenantMemberNeverCalled := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			GetTenantMember(gomock.Any(), gomock.Any()).
			Times(0)
	}

	// GetSuperAdmin() ---------------------------------------------------------------------

	getSuperAdminOk := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return superAdminRepo.EXPECT().
			GetSuperAdmin(mockBy).
			Return(superAdminEntity, nil).
			Times(1)
	}

	mockGetSuperAdminError := errors.New("unexpected error (super admin)")
	getSuperAdminError := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return superAdminRepo.EXPECT().
			GetSuperAdmin(mockBy).
			Return(&user.SuperAdminEntity{}, mockGetSuperAdminError).
			Times(1)
	}

	getSuperAdminNeverCalled := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return superAdminRepo.EXPECT().
			GetSuperAdmin(gomock.Any()).
			Times(0)
	}

	cases := []testCase{
		// Tenant Member
		{
			name:          "(Tenant Member) Success",
			inputTenantId: &targetTenantId,
			inputEmail:    targetEmail,
			setupSteps: []mockSetupFunc_userPostgreAdapter{
				getTenantMemberOk,
				getSuperAdminNeverCalled,
			},
			expectedUser:  tenantMemberUser,
			expectedError: nil,
		},
		{
			name:          "(Tenant Member) Fail: Unexpected error",
			inputTenantId: &targetTenantId,
			inputEmail:    targetEmail,
			setupSteps: []mockSetupFunc_userPostgreAdapter{
				getTenantMemberError,
				getSuperAdminNeverCalled,
			},
			expectedUser:  user.User{},
			expectedError: mockGetTenantMemberError,
		},

		// Super Admin
		{
			name:          "(Super Admin) Success",
			inputTenantId: nil,
			inputEmail:    targetEmail,
			setupSteps: []mockSetupFunc_userPostgreAdapter{
				getTenantMemberNeverCalled,
				getSuperAdminOk,
			},
			expectedUser:  superAdminUser,
			expectedError: nil,
		},
		{
			name:          "(Super Admin) Fail: Unexpected error",
			inputTenantId: nil,
			inputEmail:    targetEmail,
			setupSteps: []mockSetupFunc_userPostgreAdapter{
				getTenantMemberNeverCalled,
				getSuperAdminError,
			},
			expectedUser:  user.User{},
			expectedError: mockGetSuperAdminError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockTenantMemberRepo, mockSuperAdminRepo := setupMockSteps_UserPostgreAdapter(t, tc.setupSteps)

			adapter := user.NewUserPostgreAdapter(nil, mockTenantMemberRepo, mockSuperAdminRepo)

			savedUser, err := adapter.GetUserByEmail(tc.inputTenantId, tc.inputEmail)

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("want error %v, got error %v", tc.expectedError, err)
			}

			if !reflect.DeepEqual(tc.expectedUser, savedUser) {
				t.Errorf("want %+#v, got %#+v", tc.expectedUser, savedUser)
			}
		})
	}
}

func TestUserPostgreAdapter_GetTenantUsersByTenant(t *testing.T) {
	type testCase struct {
		name          string
		inputTenantId *uuid.UUID
		inputPage     int
		inputLimit    int
		setupSteps    []mockSetupFunc_userPostgreAdapter
		expectedUsers []user.User
		expectedTotal uint
		expectedError error
	}

	// Input ----------------------------------------------------------------------------------
	targetPassword1 := "123"
	targetPassword2 := "456"
	targetTenantId := uuid.New()
	targetEntityList := []user.TenantMemberEntity{
		{
			ID: uint(1),
			Email: "email@email.com",
			Name: "Username",
			Password: &targetPassword1,
			Confirmed: true,
			Role: string(identity.ROLE_TENANT_USER),
			TenantId: targetTenantId.String(),
		},
		{
			ID: uint(2),
			Email: "email2@email.com",
			Name: "Username2",
			Password: &targetPassword2,
			Confirmed: true,
			Role: string(identity.ROLE_TENANT_USER),
			TenantId: targetTenantId.String(),
		},
	}

	expectedUserList := []user.User{
		{
			Id: uint(1),
			Email: "email@email.com",
			Name: "Username",
			PasswordHash: &targetPassword1,
			Confirmed: true,
			Role: identity.ROLE_TENANT_USER,
			TenantId: &targetTenantId,
		},
		{
			Id: uint(2),
			Email: "email2@email.com",
			Name: "Username2",
			PasswordHash: &targetPassword2,
			Confirmed: true,
			Role: identity.ROLE_TENANT_USER,
			TenantId: &targetTenantId,
		},
	}

	targetEmptyEntityList := make([]user.TenantMemberEntity, 0)
	expectedEmptyUserList := make([]user.User, 0)

	anyInt := gomock.AssignableToTypeOf(0)

	getTenantUsersOk_PopulatedList := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			GetTenantUsers(targetTenantId.String(), anyInt, anyInt).
			Return(targetEntityList, int64(len(targetEntityList)), nil).
			Times(1)
	}

	getTenantUsersOk_EmptyList := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			GetTenantUsers(targetTenantId.String(), anyInt, anyInt).
			Return(targetEmptyEntityList, int64(len(targetEmptyEntityList)),  nil).
			Times(1)
	}

	mockError := errors.New("unexpected error")
	getTenantUsersError := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			GetTenantUsers(targetTenantId.String(), anyInt, anyInt).
			Return(nil, int64(0), mockError).
			Times(1)
	}

	getTenantUsersNeverCalled := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			GetTenantUsers(targetTenantId.String(), anyInt, anyInt).
			Times(0)
	}

	cases := []testCase{
		{
			name: "Success: populated list",
			inputTenantId: &targetTenantId,
			inputPage: 1,
			inputLimit: 10,
			setupSteps: []mockSetupFunc_userPostgreAdapter{
				getTenantUsersOk_PopulatedList,
			},
			expectedUsers: expectedUserList,
			expectedTotal: uint(2),
			expectedError: nil,
		},
		{
			name: "Success: empty list",
			inputTenantId: &targetTenantId,
			inputPage: 1,
			inputLimit: 10,
			setupSteps: []mockSetupFunc_userPostgreAdapter{
				getTenantUsersOk_EmptyList,
			},
			expectedUsers: expectedEmptyUserList,
			expectedTotal: uint(0),
			expectedError: nil,
		},
		{
			name: "Fail: pagination error",
			inputTenantId: &targetTenantId,
			inputPage: 0,
			inputLimit: 10,
			setupSteps: []mockSetupFunc_userPostgreAdapter{
				getTenantUsersNeverCalled,
			},
			expectedUsers: expectedEmptyUserList,
			expectedTotal: uint(0),
			expectedError: pagination.ErrInvalidPage,
		},
		{
			name: "Fail: unexpected error",
			inputTenantId: &targetTenantId,
			inputPage: 1,
			inputLimit: 10,
			setupSteps: []mockSetupFunc_userPostgreAdapter{
				getTenantUsersError,
			},
			expectedUsers: expectedEmptyUserList,
			expectedTotal: uint(0),
			expectedError: mockError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockTenantMemberRepo, mockSuperAdminRepo := setupMockSteps_UserPostgreAdapter(t, tc.setupSteps)

			adapter := user.NewUserPostgreAdapter(nil, mockTenantMemberRepo, mockSuperAdminRepo)

			users, total, err := adapter.GetTenantUsersByTenant(*tc.inputTenantId, tc.inputPage, tc.inputLimit)

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("want error %v, got error %v", tc.expectedError, err)
			}

			if !reflect.DeepEqual(tc.expectedUsers, users) {
				t.Errorf("want %+#v, got %#+v", tc.expectedUsers, users)
			}

			if total != tc.expectedTotal {
				t.Errorf("want total %v, got %v", tc.expectedTotal, total)
			}
		})
	}

}

func TestUserPostgreAdapter_GetTenantAdminsByTenant(t *testing.T) {
	type testCase struct {
		name          string
		inputTenantId *uuid.UUID
		inputPage     int
		inputLimit    int
		setupSteps    []mockSetupFunc_userPostgreAdapter
		expectedUsers []user.User
		expectedTotal uint
		expectedError error
	}

	// Input ----------------------------------------------------------------------------------
	targetPassword1 := "123"
	targetPassword2 := "456"
	targetTenantId := uuid.New()
	targetRole := identity.ROLE_TENANT_ADMIN
	targetEntityList := []user.TenantMemberEntity{
		{
			ID: uint(1),
			Email: "email@email.com",
			Name: "Username",
			Password: &targetPassword1,
			Confirmed: true,
			Role: string(targetRole),
			TenantId: targetTenantId.String(),
		},
		{
			ID: uint(2),
			Email: "email2@email.com",
			Name: "Username2",
			Password: &targetPassword2,
			Confirmed: true,
			Role: string(targetRole),
			TenantId: targetTenantId.String(),
		},
	}

	expectedUserList := []user.User{
		{
			Id: uint(1),
			Email: "email@email.com",
			Name: "Username",
			PasswordHash: &targetPassword1,
			Confirmed: true,
			Role: targetRole,
			TenantId: &targetTenantId,
		},
		{
			Id: uint(2),
			Email: "email2@email.com",
			Name: "Username2",
			PasswordHash: &targetPassword2,
			Confirmed: true,
			Role: targetRole,
			TenantId: &targetTenantId,
		},
	}

	targetEmptyEntityList := make([]user.TenantMemberEntity, 0)
	expectedEmptyUserList := make([]user.User, 0)

	anyInt := gomock.AssignableToTypeOf(0)

	getTenantAdminsOk_PopulatedList := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			GetTenantAdmins(targetTenantId.String(), anyInt, anyInt).
			Return(targetEntityList, int64(len(targetEntityList)), nil).
			Times(1)
	}

	getTenantAdminsOk_EmptyList := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			GetTenantAdmins(targetTenantId.String(), anyInt, anyInt).
			Return(targetEmptyEntityList, int64(len(targetEmptyEntityList)),  nil).
			Times(1)
	}

	mockError := errors.New("unexpected error")
	getTenantAdminsError := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			GetTenantAdmins(targetTenantId.String(), anyInt, anyInt).
			Return(nil, int64(0), mockError).
			Times(1)
	}

	getTenantAdminsNeverCalled := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			GetTenantAdmins(targetTenantId.String(), anyInt, anyInt).
			Times(0)
	}

	cases := []testCase{
		{
			name: "Success: populated list",
			inputTenantId: &targetTenantId,
			inputPage: 1,
			inputLimit: 10,
			setupSteps: []mockSetupFunc_userPostgreAdapter{
				getTenantAdminsOk_PopulatedList,
			},
			expectedUsers: expectedUserList,
			expectedTotal: uint(2),
			expectedError: nil,
		},
		{
			name: "Success: empty list",
			inputTenantId: &targetTenantId,
			inputPage: 1,
			inputLimit: 10,
			setupSteps: []mockSetupFunc_userPostgreAdapter{
				getTenantAdminsOk_EmptyList,
			},
			expectedUsers: expectedEmptyUserList,
			expectedTotal: uint(0),
			expectedError: nil,
		},
		{
			name: "Fail: pagination error",
			inputTenantId: &targetTenantId,
			inputPage: 0,
			inputLimit: 10,
			setupSteps: []mockSetupFunc_userPostgreAdapter{
				getTenantAdminsNeverCalled,
			},
			expectedUsers: expectedEmptyUserList,
			expectedTotal: uint(0),
			expectedError: pagination.ErrInvalidPage,
		},
		{
			name: "Fail: unexpected error",
			inputTenantId: &targetTenantId,
			inputPage: 1,
			inputLimit: 10,
			setupSteps: []mockSetupFunc_userPostgreAdapter{
				getTenantAdminsError,
			},
			expectedUsers: expectedEmptyUserList,
			expectedTotal: uint(0),
			expectedError: mockError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockTenantMemberRepo, mockSuperAdminRepo := setupMockSteps_UserPostgreAdapter(t, tc.setupSteps)

			adapter := user.NewUserPostgreAdapter(nil, mockTenantMemberRepo, mockSuperAdminRepo)

			users, total, err := adapter.GetTenantAdminsByTenant(*tc.inputTenantId, tc.inputPage, tc.inputLimit)

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("want error %v, got error %v", tc.expectedError, err)
			}

			if !reflect.DeepEqual(tc.expectedUsers, users) {
				t.Errorf("want %+#v, got %#+v", tc.expectedUsers, users)
			}

			if total != tc.expectedTotal {
				t.Errorf("want total %v, got %v", tc.expectedTotal, total)
			}
		})
	}

}

func TestUserPostgreAdapter_GetSuperAdmins(t *testing.T) {
	type testCase struct {
		name          string
		inputPage     int
		inputLimit    int
		setupSteps    []mockSetupFunc_userPostgreAdapter
		expectedUsers []user.User
		expectedTotal uint
		expectedError error
	}

	// Input ----------------------------------------------------------------------------------
	targetPassword1 := "123"
	targetPassword2 := "456"
	targetRole := identity.ROLE_SUPER_ADMIN
	targetEntityList := []user.SuperAdminEntity{
		{
			ID: uint(1),
			Email: "email@email.com",
			Name: "Username",
			Password: &targetPassword1,
			Confirmed: true,
		},
		{
			ID: uint(2),
			Email: "email2@email.com",
			Name: "Username2",
			Password: &targetPassword2,
			Confirmed: true,
		},
	}

	expectedUserList := []user.User{
		{
			Id: uint(1),
			Email: "email@email.com",
			Name: "Username",
			PasswordHash: &targetPassword1,
			Confirmed: true,
			TenantId: nil,
			Role: targetRole,
		},
		{
			Id: uint(2),
			Email: "email2@email.com",
			Name: "Username2",
			PasswordHash: &targetPassword2,
			Confirmed: true,
			TenantId: nil,
			Role: targetRole,
		},
	}

	targetEmptyEntityList := make([]user.SuperAdminEntity, 0)
	expectedEmptyUserList := make([]user.User, 0)

	anyInt := gomock.AssignableToTypeOf(0)

	getSuperAdminsOk_PopulatedList := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return superAdminRepo.EXPECT().
			GetSuperAdmins(anyInt, anyInt).
			Return(targetEntityList, int64(len(targetEntityList)), nil).
			Times(1)
	}

	getSuperAdminsOk_EmptyList := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return superAdminRepo.EXPECT().
			GetSuperAdmins(anyInt, anyInt).
			Return(targetEmptyEntityList, int64(len(targetEmptyEntityList)),  nil).
			Times(1)
	}

	mockError := errors.New("unexpected error")
	getSuperAdminsError := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return superAdminRepo.EXPECT().
			GetSuperAdmins(anyInt, anyInt).
			Return(nil, int64(0), mockError).
			Times(1)
	}

	getSuperAdminsNeverCalled := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return superAdminRepo.EXPECT().
			GetSuperAdmins(anyInt, anyInt).
			Times(0)
	}

	cases := []testCase{
		{
			name: "Success: populated list",
			inputPage: 1,
			inputLimit: 10,
			setupSteps: []mockSetupFunc_userPostgreAdapter{
				getSuperAdminsOk_PopulatedList,
			},
			expectedUsers: expectedUserList,
			expectedTotal: uint(2),
			expectedError: nil,
		},
		{
			name: "Success: empty list",
			inputPage: 1,
			inputLimit: 10,
			setupSteps: []mockSetupFunc_userPostgreAdapter{
				getSuperAdminsOk_EmptyList,
			},
			expectedUsers: expectedEmptyUserList,
			expectedTotal: uint(0),
			expectedError: nil,
		},
		{
			name: "Fail: pagination error",
			inputPage: 0,
			inputLimit: 10,
			setupSteps: []mockSetupFunc_userPostgreAdapter{
				getSuperAdminsNeverCalled,
			},
			expectedUsers: expectedEmptyUserList,
			expectedTotal: uint(0),
			expectedError: pagination.ErrInvalidPage,
		},
		{
			name: "Fail: unexpected error",
			inputPage: 1,
			inputLimit: 10,
			setupSteps: []mockSetupFunc_userPostgreAdapter{
				getSuperAdminsError,
			},
			expectedUsers: expectedEmptyUserList,
			expectedTotal: uint(0),
			expectedError: mockError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockTenantMemberRepo, mockSuperAdminRepo := setupMockSteps_UserPostgreAdapter(t, tc.setupSteps)

			adapter := user.NewUserPostgreAdapter(nil, mockTenantMemberRepo, mockSuperAdminRepo)

			users, total, err := adapter.GetSuperAdminList(tc.inputPage, tc.inputLimit)

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("want error %v, got error %v", tc.expectedError, err)
			}

			if !reflect.DeepEqual(tc.expectedUsers, users) {
				t.Errorf("want %+#v, got %#+v", tc.expectedUsers, users)
			}

			if total != tc.expectedTotal {
				t.Errorf("want total %v, got %v", tc.expectedTotal, total)
			}
		})
	}

}

func TestUserPostgreAdapter_CountTenantAdminsByTenant(t *testing.T) {
	type testCase struct {
		name          string
		inputTenantId uuid.UUID
		setupSteps    []mockSetupFunc_userPostgreAdapter
		expectedTotal uint
		expectedError error
	}

	// Input ----------------------------------------------------------------------------------
	targetTenantId := uuid.New()
	targetTotal := int64(10)

	countOk := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			CountTenantAdminsByTenant(targetTenantId.String()).
			Return(targetTotal, nil).
			Times(1)
	}

	mockError := errors.New("unexpected error")
	countError := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			CountTenantAdminsByTenant(targetTenantId.String()).
			Return(int64(0), mockError).
			Times(1)
	}

	cases := []testCase{
		{
			name: "Success",
			inputTenantId: targetTenantId,
			setupSteps: []mockSetupFunc_userPostgreAdapter{
				countOk,
			},
			expectedTotal: uint(targetTotal),
			expectedError: nil,
		},
		{
			name: "Fail: unexpected error",
			inputTenantId: targetTenantId,
			setupSteps: []mockSetupFunc_userPostgreAdapter{
				countError,
			},
			expectedTotal: uint(0),
			expectedError: mockError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockTenantMemberRepo, mockSuperAdminRepo := setupMockSteps_UserPostgreAdapter(t, tc.setupSteps)

			adapter := user.NewUserPostgreAdapter(nil, mockTenantMemberRepo, mockSuperAdminRepo)

			total, err := adapter.CountTenantAdminsByTenant(tc.inputTenantId)

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("want error %v, got error %v", tc.expectedError, err)
			}

			if total != tc.expectedTotal {
				t.Errorf("want total %v, got %v", tc.expectedTotal, total)
			}
		})
	}

}

func TestUserPostgreAdapter_CountSuperAdmins(t *testing.T) {
	type testCase struct {
		name          string
		setupSteps    []mockSetupFunc_userPostgreAdapter
		expectedTotal uint
		expectedError error
	}

	// Input ----------------------------------------------------------------------------------
	targetTotal := int64(10)

	countOk := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return superAdminRepo.EXPECT().
			CountSuperAdmins().
			Return(targetTotal, nil).
			Times(1)
	}

	mockError := errors.New("unexpected error")
	countError := func(
		tenantMemberRepo *mocks.MockTenantMemberRepository, superAdminRepo *mocks.MockSuperAdminRepository,
	) *gomock.Call {
		return superAdminRepo.EXPECT().
			CountSuperAdmins().
			Return(int64(0), mockError).
			Times(1)
	}

	cases := []testCase{
		{
			name: "Success",
			setupSteps: []mockSetupFunc_userPostgreAdapter{
				countOk,
			},
			expectedTotal: uint(targetTotal),
			expectedError: nil,
		},
		{
			name: "Fail: unexpected error",
			setupSteps: []mockSetupFunc_userPostgreAdapter{
				countError,
			},
			expectedTotal: uint(0),
			expectedError: mockError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockTenantMemberRepo, mockSuperAdminRepo := setupMockSteps_UserPostgreAdapter(t, tc.setupSteps)

			adapter := user.NewUserPostgreAdapter(nil, mockTenantMemberRepo, mockSuperAdminRepo)

			total, err := adapter.CountSuperAdmins()

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("want error %v, got error %v", tc.expectedError, err)
			}

			if total != tc.expectedTotal {
				t.Errorf("want total %v, got %v", tc.expectedTotal, total)
			}
		})
	}

}
