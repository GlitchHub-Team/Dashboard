package tenant_test

import (
	"errors"
	"testing"

	"backend/internal/identity"
	"backend/internal/tenant"
	"backend/tests/tenant/mocks"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestCreateTenant(t *testing.T) {
	targetTenantID := uuid.New()
	targetTenantName := "Stefano"
	targetCanImpersonate := true

	superAdminRequester := identity.Requester{
		RequesterRole: identity.ROLE_SUPER_ADMIN,
	}

	targetCreateTenant := tenant.Tenant{
		Name:           targetTenantName,
		CanImpersonate: targetCanImpersonate,
	}

	expectedTenant := tenant.Tenant{
		Id:             targetTenantID,
		Name:           targetTenantName,
		CanImpersonate: targetCanImpersonate,
	}

	type mockSetupFunc func(
		createTenantPort *mocks.MockCreateTenantPort,
	) *gomock.Call

	type testCase struct {
		name           string
		input          tenant.CreateTenantCommand
		setupSteps     []mockSetupFunc
		expectedTenant tenant.Tenant
		expectedError  error
	}

	step1CreateTenantOK := func(
		createTenantPort *mocks.MockCreateTenantPort,
	) *gomock.Call {
		return createTenantPort.EXPECT().
			CreateTenant(targetCreateTenant).
			Return(expectedTenant, nil).
			Times(1)
	}

	errMockStep1 := errors.New("unexpected error in step 1")
	step1CreateTenantError := func(
		createTenantPort *mocks.MockCreateTenantPort,
	) *gomock.Call {
		return createTenantPort.EXPECT().
			CreateTenant(targetCreateTenant).
			Return(tenant.Tenant{}, errMockStep1).
			Times(1)
	}

	baseInput := tenant.CreateTenantCommand{
		Name:           targetTenantName,
		CanImpersonate: targetCanImpersonate,
		Requester:      superAdminRequester,
	}

	cases := []testCase{
		{
			name:  "Success: tenant created successfully",
			input: baseInput,
			setupSteps: []mockSetupFunc{
				step1CreateTenantOK,
			},
			expectedTenant: expectedTenant,
			expectedError:  nil,
		},
		{
			name:  "Fail (step 1): unexpected error from port",
			input: baseInput,
			setupSteps: []mockSetupFunc{
				step1CreateTenantError,
			},
			expectedTenant: tenant.Tenant{},
			expectedError:  errMockStep1,
		},
		{
			name: "Fail: requester is not superadmin",
			input: tenant.CreateTenantCommand{
				Name:           targetTenantName,
				CanImpersonate: targetCanImpersonate,
				Requester:      identity.Requester{},
			},
			setupSteps:     []mockSetupFunc{},
			expectedTenant: tenant.Tenant{},
			expectedError:  tenant.ErrUnauthorized,
		},
		{
			name: "Fail: canImpersonate is false",
			input: tenant.CreateTenantCommand{
				Name:           targetTenantName,
				CanImpersonate: false,
				Requester:      superAdminRequester,
			},
			setupSteps:     []mockSetupFunc{},
			expectedTenant: tenant.Tenant{},
			expectedError:  tenant.ErrImpersonationFailded,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockController := gomock.NewController(t)

			mockCreateTenantPort := mocks.NewMockCreateTenantPort(mockController)

			var expectedCalls []any
			for _, step := range tc.setupSteps {
				call := step(mockCreateTenantPort)
				if call != nil {
					expectedCalls = append(expectedCalls, call)
				}
			}
			if len(expectedCalls) > 0 {
				gomock.InOrder(expectedCalls...)
			}

			createUseCase, _, _, _, _ := tenant.NewCreateTenantService(
				mockCreateTenantPort,
				nil,
				nil,
				nil,
				nil,
			)

			createdTenant, err := createUseCase.CreateTenant(tc.input)

			if err != tc.expectedError {
				t.Errorf("expected error %v, got %v", tc.expectedError, err)
			}
			if createdTenant != tc.expectedTenant {
				t.Errorf("expected tenant %v, got %v", tc.expectedTenant, createdTenant)
			}
		})
	}
}

// 2) delete tenant ===================================================================================
func TestDeleteTenant(t *testing.T) {
	targetTenantID := uuid.New()
	targetTenantName := "Stefano"
	superAdminRequester := identity.Requester{RequesterRole: identity.ROLE_SUPER_ADMIN}

	expectedTenant := tenant.Tenant{
		Id:             targetTenantID,
		Name:           targetTenantName,
		CanImpersonate: true,
	}

	type mockSetupFunc func(
		deletePort *mocks.MockDeleteTenantPort,
		getPort *mocks.MockGetTenantPort,
	) *gomock.Call

	type testCase struct {
		name           string
		input          tenant.DeleteTenantCommand
		setupSteps     []mockSetupFunc
		expectedTenant tenant.Tenant
		expectedError  error
	}

	step1GetTenantNotFound := func(deletePort *mocks.MockDeleteTenantPort, getPort *mocks.MockGetTenantPort) *gomock.Call {
		return getPort.EXPECT().GetTenant(targetTenantID).Return(tenant.Tenant{}, nil).Times(1)
	}

	stepGetTenantNonImpersonable := func(deletePort *mocks.MockDeleteTenantPort, getPort *mocks.MockGetTenantPort) *gomock.Call {
		return getPort.EXPECT().GetTenant(targetTenantID).Return(tenant.Tenant{
			Id:             targetTenantID,
			Name:           targetTenantName,
			CanImpersonate: false,
		}, nil).Times(1)
	}

	//  1) SUCCESS ====================================================================================
	stepGetTenantOK := func(deletePort *mocks.MockDeleteTenantPort, getPort *mocks.MockGetTenantPort) *gomock.Call {
		return getPort.EXPECT().GetTenant(targetTenantID).Return(expectedTenant, nil).Times(1)
	}

	stepDeleteTenantOK := func(deletePort *mocks.MockDeleteTenantPort, getPort *mocks.MockGetTenantPort) *gomock.Call {
		return deletePort.EXPECT().DeleteTenant(targetTenantID).Return(expectedTenant, nil).Times(1)
	}

	// 3) GET ERROR ===================================================================================
	errGetPort := errors.New("unexpected db error on get")
	stepGetTenantPortError := func(deletePort *mocks.MockDeleteTenantPort, getPort *mocks.MockGetTenantPort) *gomock.Call {
		return getPort.EXPECT().GetTenant(targetTenantID).Return(tenant.Tenant{}, errGetPort).Times(1)
	}

	// 4) DELETE ERROR ================================================================================
	errDeletePort := errors.New("unexpected db error on delete")
	stepDeleteTenantPortError := func(deletePort *mocks.MockDeleteTenantPort, getPort *mocks.MockGetTenantPort) *gomock.Call {
		return deletePort.EXPECT().DeleteTenant(targetTenantID).Return(tenant.Tenant{}, errDeletePort).Times(1)
	}

	// 5) NOT SUPER ADMIN =============================================================================
	stepNotSuperAdmin := func(deletePort *mocks.MockDeleteTenantPort, getPort *mocks.MockGetTenantPort) *gomock.Call {
		return getPort.EXPECT().GetTenant(targetTenantID).Return(expectedTenant, nil).Times(1)
	}

	cases := []testCase{
		{
			name:           "Success: tenant deleted successfully",
			input:          tenant.DeleteTenantCommand{TenantId: targetTenantID, Requester: superAdminRequester},
			setupSteps:     []mockSetupFunc{stepGetTenantOK, stepDeleteTenantOK},
			expectedTenant: expectedTenant,
			expectedError:  nil,
		},
		{
			name:           "Fail: tenant not found",
			input:          tenant.DeleteTenantCommand{TenantId: targetTenantID, Requester: superAdminRequester},
			setupSteps:     []mockSetupFunc{step1GetTenantNotFound},
			expectedTenant: tenant.Tenant{},
			expectedError:  tenant.ErrTenantNotFound,
		},
		{
			name:           "Fail: unexpected error from GetTenant port",
			input:          tenant.DeleteTenantCommand{TenantId: targetTenantID, Requester: superAdminRequester},
			setupSteps:     []mockSetupFunc{stepGetTenantPortError},
			expectedTenant: tenant.Tenant{},
			expectedError:  errGetPort,
		},
		{
			name:           "Fail: unexpected error from DeleteTenant port",
			input:          tenant.DeleteTenantCommand{TenantId: targetTenantID, Requester: superAdminRequester},
			setupSteps:     []mockSetupFunc{stepGetTenantOK, stepDeleteTenantPortError},
			expectedTenant: tenant.Tenant{},
			expectedError:  errDeletePort,
		},
		{
			name: "Fail: requester is not superadmin",
			input: tenant.DeleteTenantCommand{
				TenantId:  targetTenantID,
				Requester: identity.Requester{},
			},
			setupSteps:     []mockSetupFunc{stepNotSuperAdmin},
			expectedTenant: tenant.Tenant{},
			expectedError:  tenant.ErrUnauthorized,
		},

		{
			name:           "Fail: tenant not found (IsZero)",
			input:          tenant.DeleteTenantCommand{TenantId: targetTenantID, Requester: superAdminRequester},
			setupSteps:     []mockSetupFunc{step1GetTenantNotFound},
			expectedTenant: tenant.Tenant{},
			expectedError:  tenant.ErrTenantNotFound,
		},
		{
			name:           "Fail: tenant cannot be impersonated",
			input:          tenant.DeleteTenantCommand{TenantId: targetTenantID, Requester: superAdminRequester},
			setupSteps:     []mockSetupFunc{stepGetTenantNonImpersonable},
			expectedTenant: tenant.Tenant{},
			expectedError:  tenant.ErrImpersonationFailded,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockController := gomock.NewController(t)
			defer mockController.Finish()

			mockDeletePort := mocks.NewMockDeleteTenantPort(mockController)
			mockGetPort := mocks.NewMockGetTenantPort(mockController)

			for _, step := range tc.setupSteps {
				step(mockDeletePort, mockGetPort)
			}

			_, deleteUseCase, _, _, _ := tenant.NewCreateTenantService(
				nil,
				mockDeletePort,
				mockGetPort,
				nil,
				nil,
			)

			deletedTenant, err := deleteUseCase.DeleteTenant(tc.input)

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("expected error %v, got %v", tc.expectedError, err)
			}
			if deletedTenant != tc.expectedTenant {
				t.Errorf("expected tenant %v, got %v", tc.expectedTenant, deletedTenant)
			}
		})
	}
}

func TestGetTenant(t *testing.T) {
	targetTenantID := uuid.New()
	targetTenantName := "Stefano"
	superAdminRequester := identity.Requester{RequesterRole: identity.ROLE_SUPER_ADMIN}
	authorizedTenantAdminRequester := identity.Requester{
    RequesterRole:     identity.ROLE_TENANT_ADMIN,
    RequesterTenantId: &targetTenantID,
}

	expectedTenant := tenant.Tenant{
		Id:             targetTenantID,
		Name:           targetTenantName,
		CanImpersonate: true,
	}

	type mockSetupFunc func(
		getPort *mocks.MockGetTenantPort,
	) *gomock.Call

	type testCase struct {
		name           string
		input          tenant.GetTenantCommand
		setupSteps     []mockSetupFunc
		expectedTenant tenant.Tenant
		expectedError  error
	}

	// 1) ==============================================================================================
	stepGetTenantOK := func(getPort *mocks.MockGetTenantPort) *gomock.Call {
		return getPort.EXPECT().GetTenant(targetTenantID).Return(expectedTenant, nil).Times(1)
	}

	// 2) ==============================================================================================
	stepGetTenantNotFound := func(getPort *mocks.MockGetTenantPort) *gomock.Call {
		return getPort.EXPECT().GetTenant(targetTenantID).Return(tenant.Tenant{}, nil).Times(1)
	}

	// 3) ==============================================================================================
	errGetPort := errors.New("unexpected db error on get")
	stepGetTenantPortError := func(getPort *mocks.MockGetTenantPort) *gomock.Call {
		return getPort.EXPECT().GetTenant(targetTenantID).Return(tenant.Tenant{}, errGetPort).Times(1)
	}

	cases := []testCase{
		{
			name:           "Success: tenant retrieved successfully",
			input:          tenant.GetTenantCommand{TenantId: targetTenantID, Requester: superAdminRequester},
			setupSteps:     []mockSetupFunc{stepGetTenantOK},
			expectedTenant: expectedTenant,
			expectedError:  nil,
		},
		{
			name:           "Fail: tenant not found",
			input:          tenant.GetTenantCommand{TenantId: targetTenantID, Requester: superAdminRequester},
			setupSteps:     []mockSetupFunc{stepGetTenantNotFound},
			expectedTenant: tenant.Tenant{},
			expectedError:  tenant.ErrTenantNotFound,
		},
		{
			name:           "Fail: unexpected error from GetTenant port",
			input:          tenant.GetTenantCommand{TenantId: targetTenantID, Requester: superAdminRequester},
			setupSteps:     []mockSetupFunc{stepGetTenantPortError},
			expectedTenant: tenant.Tenant{},
			expectedError:  errGetPort,
		},
		{
			name: "Fail: requester is not superadmin",
			input: tenant.GetTenantCommand{
				TenantId:  targetTenantID,
				Requester: identity.Requester{},
			},
			setupSteps:     []mockSetupFunc{stepGetTenantOK},
			expectedTenant: tenant.Tenant{},
			expectedError:  tenant.ErrUnauthorized,
		},
		{
			name: "Fail: requester cannot access tenant",
			input: tenant.GetTenantCommand{
				TenantId: targetTenantID,
				Requester: identity.Requester{
					RequesterRole: identity.ROLE_TENANT_ADMIN,
				},
			},
			setupSteps:     []mockSetupFunc{stepGetTenantOK},
			expectedTenant: tenant.Tenant{},
			expectedError:  tenant.ErrUnauthorized,
		},
		{
			name:           "Success: tenant retrieved by authorized Tenant Admin",
			input:          tenant.GetTenantCommand{TenantId: targetTenantID, Requester: authorizedTenantAdminRequester},
			setupSteps:     []mockSetupFunc{stepGetTenantOK},
			expectedTenant: expectedTenant,
			expectedError:  nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockController := gomock.NewController(t)
			defer mockController.Finish()

			mockGetPort := mocks.NewMockGetTenantPort(mockController)

			for _, step := range tc.setupSteps {
				step(mockGetPort)
			}

			_, _, getUseCase, _, _ := tenant.NewCreateTenantService(
				nil,
				nil,
				mockGetPort,
				nil,
				nil,
			)

			retrievedTenant, err := getUseCase.GetTenant(tc.input)

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("expected error %v, got %v", tc.expectedError, err)
			}
			if retrievedTenant != tc.expectedTenant {
				t.Errorf("expected tenant %v, got %v", tc.expectedTenant, retrievedTenant)
			}
		})
	}
}

func TestGetTenantLisit(t *testing.T) {
	targetTenantID := uuid.New()
	targetTenantName := "Stefano"
	superAdminRequester := identity.Requester{RequesterRole: identity.ROLE_SUPER_ADMIN}

	expectedTenant := tenant.Tenant{
		Id:             targetTenantID,
		Name:           targetTenantName,
		CanImpersonate: true,
	}

	type mockSetupFunc func(
		getListPort *mocks.MockGetTenantsPort,
	) *gomock.Call

	type testCase struct {
		name               string
		input              tenant.GetTenantListCommand
		setupSteps         []mockSetupFunc
		expectedTenantList []tenant.Tenant
		expectedError      error
	}

	step1GetTenantListOK := func(getListPort *mocks.MockGetTenantsPort) *gomock.Call {
		return getListPort.EXPECT().GetTenants().Return([]tenant.Tenant{expectedTenant}, nil).Times(1)
	}

	errGetListPort := errors.New("unexpected db error on get list")
	stepGetTenantListPortError := func(getListPort *mocks.MockGetTenantsPort) *gomock.Call {
		return getListPort.EXPECT().GetTenants().Return(nil, errGetListPort).Times(1)
	}

	stepGetTenantListNotCalled := func(getListPort *mocks.MockGetTenantsPort) *gomock.Call {
		return nil
	}

	stepGetTenantListWithNonImpersonable := func(getListPort *mocks.MockGetTenantsPort) *gomock.Call {
		return getListPort.EXPECT().GetTenants().Return([]tenant.Tenant{
			{
				Id:             targetTenantID,
				Name:           targetTenantName,
				CanImpersonate: false,
			},
		}, nil).Times(1)
	}

	caseTest := []testCase{
		{
			name:               "Success: tenant list retrieved successfully",
			input:              tenant.GetTenantListCommand{Requester: superAdminRequester},
			setupSteps:         []mockSetupFunc{step1GetTenantListOK},
			expectedTenantList: []tenant.Tenant{expectedTenant},
			expectedError:      nil,
		},
		{
			name:               "Fail: unexpected error from GetTenants port",
			input:              tenant.GetTenantListCommand{Requester: superAdminRequester},
			setupSteps:         []mockSetupFunc{stepGetTenantListPortError},
			expectedTenantList: nil,
			expectedError:      errGetListPort,
		},
		{
			name:               "Fail: requester is not superadmin",
			input:              tenant.GetTenantListCommand{Requester: identity.Requester{}},
			setupSteps:         []mockSetupFunc{stepGetTenantListNotCalled},
			expectedTenantList: nil,
			expectedError:      tenant.ErrUnauthorized,
		},
		{
			name:               "Fail: tenant list contains non-impersonable tenant",
			input:              tenant.GetTenantListCommand{Requester: superAdminRequester},
			setupSteps:         []mockSetupFunc{stepGetTenantListWithNonImpersonable},
			expectedTenantList: nil,
			expectedError:      tenant.ErrUnauthorized,
		},
	}

	for _, tc := range caseTest {
		t.Run(tc.name, func(t *testing.T) {
			mockController := gomock.NewController(t)

			mockGetListPort := mocks.NewMockGetTenantsPort(mockController)

			var expectedCalls []any
			for _, step := range tc.setupSteps {
				call := step(mockGetListPort)
				if call != nil {
					expectedCalls = append(expectedCalls, call)
				}
			}
			if len(expectedCalls) > 0 {
				gomock.InOrder(expectedCalls...)
			}

			_, _, _, getListUseCase, _ := tenant.NewCreateTenantService(
				nil,
				nil,
				nil,
				mockGetListPort,
				nil,
			)

			retrievedTenantList, err := getListUseCase.GetTenantList(tc.input)

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("expected error %v, got %v", tc.expectedError, err)
			}

			if len(retrievedTenantList) != len(tc.expectedTenantList) {
				t.Errorf("expected tenant list length %v, got %v", len(tc.expectedTenantList), len(retrievedTenantList))
			} else {
				for i, ten := range retrievedTenantList {
					if ten != tc.expectedTenantList[i] {
						t.Errorf("expected tenant %v at index %d, got %v", tc.expectedTenantList[i], i, ten)
					}
				}
			}
		})
	}
}

func TestGetTenantByUser(t *testing.T) {
	targetTenantID := uuid.New()
	userId := uuid.New()

	superAdminRequester := identity.Requester{RequesterRole: identity.ROLE_SUPER_ADMIN}

	unauthorizedRequester := identity.Requester{}

	authorizedTenantAdminRequester := identity.Requester{
    RequesterRole:     identity.ROLE_TENANT_ADMIN,
    RequesterTenantId: &targetTenantID, 
}

	expectedTenant := tenant.Tenant{
		Id:             targetTenantID,
		Name:           "Stefano",
		CanImpersonate: true,
	}

	type mockSetupFunc func(
		getByUserPort *mocks.MockGetTenantByUserPort,
	) *gomock.Call

	type testCase struct {
		name           string
		input          tenant.GetTenantByUserCommand
		setupSteps     []mockSetupFunc
		expectedTenant tenant.Tenant
		expectedError  error
	}

	stepGetTenantByUserOK := func(getByUserPort *mocks.MockGetTenantByUserPort) *gomock.Call {
		return getByUserPort.EXPECT().GetTenantByUser(userId).Return(expectedTenant, nil).Times(1)
	}

	errGetByUserPort := errors.New("unexpected db error on get by user")
	stepGetTenantByUserPortError := func(getByUserPort *mocks.MockGetTenantByUserPort) *gomock.Call {
		return getByUserPort.EXPECT().GetTenantByUser(userId).Return(tenant.Tenant{}, errGetByUserPort).Times(1)
	}

	stepGetTenantByUserNotFound := func(getByUserPort *mocks.MockGetTenantByUserPort) *gomock.Call {
		return getByUserPort.EXPECT().GetTenantByUser(userId).Return(tenant.Tenant{}, nil).Times(1)
	}

	cases := []testCase{
		{
			name:           "Success: tenant retrieved successfully by user (SuperAdmin)",
			input:          tenant.GetTenantByUserCommand{UserId: userId, Requester: superAdminRequester},
			setupSteps:     []mockSetupFunc{stepGetTenantByUserOK},
			expectedTenant: expectedTenant,
			expectedError:  nil,
		},
		{
			name:           "Fail: unexpected error from GetTenantByUser port",
			input:          tenant.GetTenantByUserCommand{UserId: userId, Requester: superAdminRequester},
			setupSteps:     []mockSetupFunc{stepGetTenantByUserPortError},
			expectedTenant: tenant.Tenant{},
			expectedError:  errGetByUserPort,
		},
		{
			name:           "Fail: tenant is zero (not found)",
			input:          tenant.GetTenantByUserCommand{UserId: userId, Requester: superAdminRequester},
			setupSteps:     []mockSetupFunc{stepGetTenantByUserNotFound},
			expectedTenant: tenant.Tenant{},
			expectedError:  tenant.ErrTenantNotFound,
		},
		{
			name:           "Fail: requester is not superadmin and cannot access tenant admin",
			input:          tenant.GetTenantByUserCommand{UserId: userId, Requester: unauthorizedRequester},
			setupSteps:     []mockSetupFunc{stepGetTenantByUserOK}, // Il DB ritorna il tenant, ma il service blocca l'utente
			expectedTenant: tenant.Tenant{},
			expectedError:  tenant.ErrUnauthorized,
		},
		{
			name:           "Success: tenant retrieved by authorized Tenant Admin",
			input:          tenant.GetTenantByUserCommand{UserId: userId, Requester: authorizedTenantAdminRequester},
			setupSteps:     []mockSetupFunc{stepGetTenantByUserOK},
			expectedTenant: expectedTenant,
			expectedError:  nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockController := gomock.NewController(t)
			defer mockController.Finish()

			mockGetByUserPort := mocks.NewMockGetTenantByUserPort(mockController)

			for _, step := range tc.setupSteps {
				step(mockGetByUserPort)
			}

			_, _, _, _, getByUserUseCase := tenant.NewCreateTenantService(
				nil,
				nil,
				nil,
				nil,
				mockGetByUserPort,
			)

			retrievedTenant, err := getByUserUseCase.GetTenantByUser(tc.input)

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("expected error %v, got %v", tc.expectedError, err)
			}
			if retrievedTenant != tc.expectedTenant {
				t.Errorf("expected tenant %v, got %v", tc.expectedTenant, retrievedTenant)
			}
		})
	}
}
