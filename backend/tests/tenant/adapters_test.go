package tenant_test

import (
	"errors"
	"reflect"
	"testing"

	"backend/internal/infra/database/pagination"
	"backend/internal/tenant"
	helper "backend/tests/helper"
	mocks "backend/tests/tenant/mocks"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

type tenantAdapterMocks struct {
	repo *mocks.MockTenantRepository
}

type mockSetupFunc_TenantAdapter = helper.AdapterMockSetupFunc[tenantAdapterMocks]

func setupTenantAdapter(
	t *testing.T,
	setupSteps []mockSetupFunc_TenantAdapter,
) *tenant.TenantPostgreAdapter {
	t.Helper()

	return helper.SetupAdapterWithOrderedSteps(
		t,
		func(ctrl *gomock.Controller) tenantAdapterMocks {
			return tenantAdapterMocks{
				repo: mocks.NewMockTenantRepository(ctrl),
			}
		},
		setupSteps,
		func(mockBundle tenantAdapterMocks) *tenant.TenantPostgreAdapter {
			return tenant.NewTenantPostgreAdapter(mockBundle.repo)
		},
	)
}

func TestTenantPostgreAdapter_CreateTenant(t *testing.T) {
	targetId := uuid.New()
	targetTenant := tenant.Tenant{
		Id:             targetId,
		Name:           "Tenant A",
		CanImpersonate: true,
	}

	stepSaveTenantOk := func(mockBundle tenantAdapterMocks) *gomock.Call {
		return mockBundle.repo.EXPECT().
			SaveTenant(gomock.AssignableToTypeOf(&tenant.TenantEntity{})).
			DoAndReturn(func(entity *tenant.TenantEntity) error {
				if entity.ID != targetId.String() {
					t.Fatalf("expected entity ID %s, got %s", targetId.String(), entity.ID)
				}
				if entity.Name != targetTenant.Name {
					t.Fatalf("expected entity Name %s, got %s", targetTenant.Name, entity.Name)
				}
				if entity.CanImpersonate != targetTenant.CanImpersonate {
					t.Fatalf("expected entity CanImpersonate %v, got %v", targetTenant.CanImpersonate, entity.CanImpersonate)
				}
				return nil
			}).
			Times(1)
	}

	stepSaveTenantDuplicatedKey := func(mockBundle tenantAdapterMocks) *gomock.Call {
		return mockBundle.repo.EXPECT().
			SaveTenant(gomock.AssignableToTypeOf(&tenant.TenantEntity{})).
			Return(gorm.ErrDuplicatedKey).
			Times(1)
	}

	mockSaveError := errors.New("unexpected save tenant error")
	stepSaveTenantErr := func(mockBundle tenantAdapterMocks) *gomock.Call {
		return mockBundle.repo.EXPECT().
			SaveTenant(gomock.AssignableToTypeOf(&tenant.TenantEntity{})).
			Return(mockSaveError).
			Times(1)
	}

	stepSaveTenantOk_ButEntityMutatedWithInvalidId := func(mockBundle tenantAdapterMocks) *gomock.Call {
		return mockBundle.repo.EXPECT().
			SaveTenant(gomock.AssignableToTypeOf(&tenant.TenantEntity{})).
			DoAndReturn(func(entity *tenant.TenantEntity) error {
				entity.ID = "invalid-uuid"
				return nil
			}).
			Times(1)
	}

	type testCase struct {
		name           string
		setupSteps     []mockSetupFunc_TenantAdapter
		expectedTenant tenant.Tenant
		expectedError  error
		expectAnyError bool
	}

	cases := []testCase{
		{
			name: "Success",
			setupSteps: []mockSetupFunc_TenantAdapter{
				stepSaveTenantOk,
			},
			expectedTenant: targetTenant,
		},
		{
			name: "Fail: duplicated key maps to ErrTenantAlreadyExists",
			setupSteps: []mockSetupFunc_TenantAdapter{
				stepSaveTenantDuplicatedKey,
			},
			expectedError: tenant.ErrTenantAlreadyExists,
		},
		{
			name: "Fail: save tenant returns unexpected error",
			setupSteps: []mockSetupFunc_TenantAdapter{
				stepSaveTenantErr,
			},
			expectedError: mockSaveError,
		},
		{
			name: "Fail: mapping error after save",
			setupSteps: []mockSetupFunc_TenantAdapter{
				stepSaveTenantOk_ButEntityMutatedWithInvalidId,
			},
			expectAnyError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			adapter := setupTenantAdapter(t, tc.setupSteps)

			createdTenant, err := adapter.CreateTenant(targetTenant)

			if tc.expectAnyError {
				if err == nil {
					t.Fatalf("expected an error, got nil")
				}
				if !createdTenant.IsZero() {
					t.Fatalf("expected zero tenant on error, got %+v", createdTenant)
				}
				return
			}

			if tc.expectedError != nil {
				if !errors.Is(err, tc.expectedError) {
					t.Fatalf("expected error %v, got %v", tc.expectedError, err)
				}
				if !createdTenant.IsZero() {
					t.Fatalf("expected zero tenant on error, got %+v", createdTenant)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if !reflect.DeepEqual(tc.expectedTenant, createdTenant) {
				t.Fatalf("unexpected tenant. expected %+v, got %+v", tc.expectedTenant, createdTenant)
			}
		})
	}
}

func TestTenantPostgreAdapter_DeleteTenant(t *testing.T) {
	targetTenantId := uuid.New()
	targetEntity := tenant.TenantEntity{
		ID:             targetTenantId.String(),
		Name:           "Tenant A",
		CanImpersonate: true,
	}
	targetDeleteEntity := &tenant.TenantEntity{
		ID: targetTenantId.String(),
	}
	expectedTenant := tenant.Tenant{
		Id:             targetTenantId,
		Name:           targetEntity.Name,
		CanImpersonate: targetEntity.CanImpersonate,
	}

	stepDeleteFail_NotFound := func(mockBundle tenantAdapterMocks) *gomock.Call {
		return mockBundle.repo.EXPECT().
			DeleteTenant(targetDeleteEntity).
			Return(tenant.ErrTenantNotFound).
			Times(1)
	}

	mockDeleteErr := errors.New("unexpected delete tenant error")
	stepDeleteErr := func(mockBundle tenantAdapterMocks) *gomock.Call {
		return mockBundle.repo.EXPECT().
			DeleteTenant(targetDeleteEntity).
			Return(mockDeleteErr).
			Times(1)
	}

	stepDeleteOk := func(mockBundle tenantAdapterMocks) *gomock.Call {
		return mockBundle.repo.EXPECT().
			DeleteTenant(targetDeleteEntity).
			Do(func(entity *tenant.TenantEntity) {
				*entity = targetEntity
			}).
			Return(nil).
			Times(1)
	}

	type testCase struct {
		name           string
		setupSteps     []mockSetupFunc_TenantAdapter
		expectedTenant tenant.Tenant
		expectedError  error
		expectAnyError bool
	}

	cases := []testCase{
		{
			name: "Success",
			setupSteps: []mockSetupFunc_TenantAdapter{
				stepDeleteOk,
			},
			expectedTenant: expectedTenant,
		},

		{
			name: "Fail: tenant not found",
			setupSteps: []mockSetupFunc_TenantAdapter{
				stepDeleteFail_NotFound,
			},
			expectedTenant: tenant.Tenant{},
			expectedError:  tenant.ErrTenantNotFound,
		},
		{
			name: "Fail: delete tenant returns unexpected error",
			setupSteps: []mockSetupFunc_TenantAdapter{
				stepDeleteErr,
			},
			expectedError: mockDeleteErr,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			adapter := setupTenantAdapter(t, tc.setupSteps)

			deletedTenant, err := adapter.DeleteTenant(targetTenantId)

			if tc.expectAnyError {
				if err == nil {
					t.Fatalf("expected an error, got nil")
				}
				if !deletedTenant.IsZero() {
					t.Fatalf("expected zero tenant on error, got %+v", deletedTenant)
				}
				return
			}

			if tc.expectedError != nil {
				if !errors.Is(err, tc.expectedError) {
					t.Fatalf("expected error %v, got %v", tc.expectedError, err)
				}
				if !deletedTenant.IsZero() {
					t.Fatalf("expected zero tenant on error, got %+v", deletedTenant)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if !reflect.DeepEqual(tc.expectedTenant, deletedTenant) {
				t.Fatalf("unexpected tenant. expected %+v, got %+v", tc.expectedTenant, deletedTenant)
			}
		})
	}
}

func TestTenantPostgreAdapter_GetTenant(t *testing.T) {
	targetTenantId := uuid.New()
	targetEntity := &tenant.TenantEntity{
		ID:             targetTenantId.String(),
		Name:           "Tenant A",
		CanImpersonate: true,
	}
	expectedTenant := tenant.Tenant{
		Id:             targetTenantId,
		Name:           targetEntity.Name,
		CanImpersonate: targetEntity.CanImpersonate,
	}

	stepGetTenantNotFound := func(mockBundle tenantAdapterMocks) *gomock.Call {
		return mockBundle.repo.EXPECT().
			GetTenant(targetTenantId.String()).
			Return(nil, gorm.ErrRecordNotFound).
			Times(1)
	}

	mockGetTenantErr := errors.New("unexpected get tenant error")
	stepGetTenantErr := func(mockBundle tenantAdapterMocks) *gomock.Call {
		return mockBundle.repo.EXPECT().
			GetTenant(targetTenantId.String()).
			Return(nil, mockGetTenantErr).
			Times(1)
	}

	stepGetTenantOk := func(mockBundle tenantAdapterMocks) *gomock.Call {
		return mockBundle.repo.EXPECT().
			GetTenant(targetTenantId.String()).
			Return(targetEntity, nil).
			Times(1)
	}

	stepGetTenantOk_ButInvalidId := func(mockBundle tenantAdapterMocks) *gomock.Call {
		return mockBundle.repo.EXPECT().
			GetTenant(targetTenantId.String()).
			Return(&tenant.TenantEntity{ID: "invalid-uuid"}, nil).
			Times(1)
	}

	type testCase struct {
		name           string
		setupSteps     []mockSetupFunc_TenantAdapter
		expectedTenant tenant.Tenant
		expectedError  error
		expectAnyError bool
	}

	cases := []testCase{
		{
			name: "Fail: tenant not found",
			setupSteps: []mockSetupFunc_TenantAdapter{
				stepGetTenantNotFound,
			},
			expectedError: tenant.ErrTenantNotFound,
		},
		{
			name: "Fail: get tenant returns unexpected error",
			setupSteps: []mockSetupFunc_TenantAdapter{
				stepGetTenantErr,
			},
			expectedError: mockGetTenantErr,
		},
		{
			name: "Success",
			setupSteps: []mockSetupFunc_TenantAdapter{
				stepGetTenantOk,
			},
			expectedTenant: expectedTenant,
		},
		{
			name: "Fail: mapping error",
			setupSteps: []mockSetupFunc_TenantAdapter{
				stepGetTenantOk_ButInvalidId,
			},
			expectAnyError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			adapter := setupTenantAdapter(t, tc.setupSteps)

			foundTenant, err := adapter.GetTenant(targetTenantId)

			if tc.expectAnyError {
				if err == nil {
					t.Fatalf("expected an error, got nil")
				}
				if !foundTenant.IsZero() {
					t.Fatalf("expected zero tenant on error, got %+v", foundTenant)
				}
				return
			}

			if tc.expectedError != nil {
				if !errors.Is(err, tc.expectedError) {
					t.Fatalf("expected error %v, got %v", tc.expectedError, err)
				}
				if !foundTenant.IsZero() {
					t.Fatalf("expected zero tenant on error, got %+v", foundTenant)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if !reflect.DeepEqual(tc.expectedTenant, foundTenant) {
				t.Fatalf("unexpected tenant. expected %+v, got %+v", tc.expectedTenant, foundTenant)
			}
		})
	}
}

func TestTenantPostgreAdapter_GetTenants(t *testing.T) {
	targetTenantIdA := uuid.New()
	targetTenantIdB := uuid.New()
	entityList := []tenant.TenantEntity{
		{
			ID:             targetTenantIdA.String(),
			Name:           "Tenant A",
			CanImpersonate: true,
		},
		{
			ID:             targetTenantIdB.String(),
			Name:           "Tenant B",
			CanImpersonate: false,
		},
	}
	expectedList := []tenant.Tenant{
		{
			Id:             targetTenantIdA,
			Name:           "Tenant A",
			CanImpersonate: true,
		},
		{
			Id:             targetTenantIdB,
			Name:           "Tenant B",
			CanImpersonate: false,
		},
	}

	stepGetTenantsNeverCalled := func(mockBundle tenantAdapterMocks) *gomock.Call {
		return mockBundle.repo.EXPECT().
			GetTenants(gomock.Any(), gomock.Any()).
			Times(0)
	}

	mockGetTenantsErr := errors.New("unexpected get tenants error")
	stepGetTenantsErr := func(mockBundle tenantAdapterMocks) *gomock.Call {
		return mockBundle.repo.EXPECT().
			GetTenants(10, 5).
			Return(nil, int64(0), mockGetTenantsErr).
			Times(1)
	}

	stepGetTenantsOk := func(mockBundle tenantAdapterMocks) *gomock.Call {
		return mockBundle.repo.EXPECT().
			GetTenants(5, 5).
			Return(entityList, int64(2), nil).
			Times(1)
	}

	stepGetTenantsOk_ButInvalidEntityId := func(mockBundle tenantAdapterMocks) *gomock.Call {
		return mockBundle.repo.EXPECT().
			GetTenants(5, 5).
			Return([]tenant.TenantEntity{{ID: "invalid-uuid"}}, int64(1), nil).
			Times(1)
	}

	type testCase struct {
		name           string
		page           int
		limit          int
		setupSteps     []mockSetupFunc_TenantAdapter
		expectedList   []tenant.Tenant
		expectedTotal  uint
		expectedError  error
		expectAnyError bool
	}

	cases := []testCase{
		{
			name:  "Fail: invalid page",
			page:  0,
			limit: 5,
			setupSteps: []mockSetupFunc_TenantAdapter{
				stepGetTenantsNeverCalled,
			},
			expectedError: pagination.ErrInvalidPage,
		},
		{
			name:  "Fail: invalid limit",
			page:  1,
			limit: 0,
			setupSteps: []mockSetupFunc_TenantAdapter{
				stepGetTenantsNeverCalled,
			},
			expectedError: pagination.ErrInvalidLimit,
		},
		{
			name:  "Fail: get tenants returns error",
			page:  3,
			limit: 5,
			setupSteps: []mockSetupFunc_TenantAdapter{
				stepGetTenantsErr,
			},
			expectedError: mockGetTenantsErr,
		},
		{
			name:  "Success",
			page:  2,
			limit: 5,
			setupSteps: []mockSetupFunc_TenantAdapter{
				stepGetTenantsOk,
			},
			expectedList:  expectedList,
			expectedTotal: uint(2),
		},
		{
			name:  "Fail: mapping error",
			page:  2,
			limit: 5,
			setupSteps: []mockSetupFunc_TenantAdapter{
				stepGetTenantsOk_ButInvalidEntityId,
			},
			expectAnyError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			adapter := setupTenantAdapter(t, tc.setupSteps)

			tenantList, total, err := adapter.GetTenants(tc.page, tc.limit)

			if tc.expectAnyError {
				if err == nil {
					t.Fatalf("expected an error, got nil")
				}
				if len(tenantList) != 0 {
					t.Fatalf("expected empty tenant list on error, got %+v", tenantList)
				}
				return
			}

			if tc.expectedError != nil {
				if !errors.Is(err, tc.expectedError) {
					t.Fatalf("expected error %v, got %v", tc.expectedError, err)
				}
				if tenantList != nil {
					t.Fatalf("expected nil tenant list on error, got %+v", tenantList)
				}
				if total != 0 {
					t.Fatalf("expected zero total on error, got %d", total)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if !reflect.DeepEqual(tc.expectedList, tenantList) {
				t.Fatalf("unexpected tenant list. expected %+v, got %+v", tc.expectedList, tenantList)
			}
			if total != tc.expectedTotal {
				t.Fatalf("unexpected total. expected %d, got %d", tc.expectedTotal, total)
			}
		})
	}
}

func TestTenantPostgreAdapter_GetAllTenants(t *testing.T) {
	targetTenantIdA := uuid.New()
	targetTenantIdB := uuid.New()
	entityList := []tenant.TenantEntity{
		{
			ID:             targetTenantIdA.String(),
			Name:           "Tenant A",
			CanImpersonate: true,
		},
		{
			ID:             targetTenantIdB.String(),
			Name:           "Tenant B",
			CanImpersonate: false,
		},
	}
	expectedList := []tenant.Tenant{
		{
			Id:             targetTenantIdA,
			Name:           "Tenant A",
			CanImpersonate: true,
		},
		{
			Id:             targetTenantIdB,
			Name:           "Tenant B",
			CanImpersonate: false,
		},
	}

	mockGetAllErr := errors.New("unexpected get all tenants error")
	stepGetAllTenantsErr := func(mockBundle tenantAdapterMocks) *gomock.Call {
		return mockBundle.repo.EXPECT().
			GetAllTenants().
			Return(nil, mockGetAllErr).
			Times(1)
	}

	stepGetAllTenantsOk := func(mockBundle tenantAdapterMocks) *gomock.Call {
		return mockBundle.repo.EXPECT().
			GetAllTenants().
			Return(entityList, nil).
			Times(1)
	}

	stepGetAllTenantsOk_ButInvalidEntityId := func(mockBundle tenantAdapterMocks) *gomock.Call {
		return mockBundle.repo.EXPECT().
			GetAllTenants().
			Return([]tenant.TenantEntity{{ID: "invalid-uuid"}}, nil).
			Times(1)
	}

	type testCase struct {
		name           string
		setupSteps     []mockSetupFunc_TenantAdapter
		expectedList   []tenant.Tenant
		expectedError  error
		expectAnyError bool
	}

	cases := []testCase{
		{
			name: "Fail: get all tenants returns error",
			setupSteps: []mockSetupFunc_TenantAdapter{
				stepGetAllTenantsErr,
			},
			expectedError: mockGetAllErr,
		},
		{
			name: "Success",
			setupSteps: []mockSetupFunc_TenantAdapter{
				stepGetAllTenantsOk,
			},
			expectedList: expectedList,
		},
		{
			name: "Fail: mapping error",
			setupSteps: []mockSetupFunc_TenantAdapter{
				stepGetAllTenantsOk_ButInvalidEntityId,
			},
			expectAnyError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			adapter := setupTenantAdapter(t, tc.setupSteps)

			tenantList, err := adapter.GetAllTenants()

			if tc.expectAnyError {
				if err == nil {
					t.Fatalf("expected an error, got nil")
				}
				if len(tenantList) != 0 {
					t.Fatalf("expected empty tenant list on error, got %+v", tenantList)
				}
				return
			}

			if tc.expectedError != nil {
				if !errors.Is(err, tc.expectedError) {
					t.Fatalf("expected error %v, got %v", tc.expectedError, err)
				}
				if tenantList != nil {
					t.Fatalf("expected nil tenant list on error, got %+v", tenantList)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if !reflect.DeepEqual(tc.expectedList, tenantList) {
				t.Fatalf("unexpected tenant list. expected %+v, got %+v", tc.expectedList, tenantList)
			}
		})
	}
}
