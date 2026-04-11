package tenant_test

import (
	"errors"
	"testing"

	"backend/internal/shared/identity"
	"backend/internal/tenant"
	"backend/tests/helper"
	tenantMocks "backend/tests/tenant/mocks"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

type tenantServiceMockBundle struct {
	createTenantPort *tenantMocks.MockCreateTenantPort
	deleteTenantPort *tenantMocks.MockDeleteTenantPort
	getTenantPort    *tenantMocks.MockGetTenantPort
	getTenantsPort   *tenantMocks.MockGetTenantsPort
}

func newTenantServiceMockBundle(ctrl *gomock.Controller) tenantServiceMockBundle {
	return tenantServiceMockBundle{
		createTenantPort: tenantMocks.NewMockCreateTenantPort(ctrl),
		deleteTenantPort: tenantMocks.NewMockDeleteTenantPort(ctrl),
		getTenantPort:    tenantMocks.NewMockGetTenantPort(ctrl),
		getTenantsPort:   tenantMocks.NewMockGetTenantsPort(ctrl),
	}
}

func newTenantService(bundle tenantServiceMockBundle) *tenant.TenantService {
	return tenant.NewCreateTenantService(
		bundle.createTenantPort,
		bundle.deleteTenantPort,
		bundle.getTenantPort,
		bundle.getTenantsPort,
	)
}

func assertErr(t *testing.T, got error, expected error) {
	t.Helper()
	if expected == nil {
		if got != nil {
			t.Fatalf("expected nil error, got %v", got)
		}
		return
	}

	if !errors.Is(got, expected) {
		t.Fatalf("expected error %v, got %v", expected, got)
	}
}

func TestService_CreateTenant(t *testing.T) {
	targetName := "Tenant A"
	superAdminRequester := identity.Requester{
		RequesterUserId: 1,
		RequesterRole:   identity.ROLE_SUPER_ADMIN,
	}
	tenantAdminRequester := identity.Requester{
		RequesterUserId:   2,
		RequesterTenantId: ptrUUID(uuid.New()),
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	type testCase struct {
		name        string
		input       tenant.CreateTenantCommand
		setupSteps  []helper.ServiceMockSetupFunc[tenantServiceMockBundle]
		expectedErr error
		check       func(t *testing.T, out tenant.Tenant)
	}

	errMock := newMockError(1)

	cases := []testCase{
		{
			name: "(Super Admin) Success",
			input: tenant.CreateTenantCommand{
				Requester:      superAdminRequester,
				Name:           targetName,
				CanImpersonate: true,
			},
			setupSteps: []helper.ServiceMockSetupFunc[tenantServiceMockBundle]{
				func(m tenantServiceMockBundle) *gomock.Call {
					return m.createTenantPort.EXPECT().
						CreateTenant(gomock.AssignableToTypeOf(tenant.Tenant{})).
						DoAndReturn(func(newTenant tenant.Tenant) (tenant.Tenant, error) {
							if newTenant.Id == uuid.Nil {
								t.Errorf("expected generated tenant ID")
							}
							if newTenant.Name != targetName {
								t.Errorf("expected name %q, got %q", targetName, newTenant.Name)
							}
							if !newTenant.CanImpersonate {
								t.Errorf("expected CanImpersonate=true")
							}
							return newTenant, nil
						}).
						Times(1)
				},
			},
			expectedErr: nil,
			check: func(t *testing.T, out tenant.Tenant) {
				t.Helper()
				if out.Id == uuid.Nil {
					t.Fatalf("expected non-nil tenant id")
				}
				if out.Name != targetName {
					t.Fatalf("expected tenant name %q, got %q", targetName, out.Name)
				}
				if !out.CanImpersonate {
					t.Fatalf("expected CanImpersonate=true")
				}
			},
		},
		{
			name: "(Tenant Admin) Fail: unauthorized",
			input: tenant.CreateTenantCommand{
				Requester:      tenantAdminRequester,
				Name:           targetName,
				CanImpersonate: true,
			},
			setupSteps: []helper.ServiceMockSetupFunc[tenantServiceMockBundle]{
				func(m tenantServiceMockBundle) *gomock.Call {
					return m.createTenantPort.EXPECT().
						CreateTenant(gomock.Any()).
						Times(0)
				},
			},
			expectedErr: identity.ErrUnauthorizedAccess,
		},
		{
			name: "Fail: create port returns error",
			input: tenant.CreateTenantCommand{
				Requester:      superAdminRequester,
				Name:           targetName,
				CanImpersonate: false,
			},
			setupSteps: []helper.ServiceMockSetupFunc[tenantServiceMockBundle]{
				func(m tenantServiceMockBundle) *gomock.Call {
					return m.createTenantPort.EXPECT().
						CreateTenant(gomock.AssignableToTypeOf(tenant.Tenant{})).
						Return(tenant.Tenant{}, errMock).
						Times(1)
				},
			},
			expectedErr: errMock,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := helper.SetupServiceWithOrderedSteps(
				t,
				newTenantServiceMockBundle,
				tc.setupSteps,
				newTenantService,
			)

			out, err := service.CreateTenant(tc.input)
			assertErr(t, err, tc.expectedErr)
			if tc.check != nil {
				tc.check(t, out)
			}
		})
	}
}

func TestService_DeleteTenant(t *testing.T) {
	tenantID := uuid.New()
	superAdminRequester := identity.Requester{
		RequesterUserId: 1,
		RequesterRole:   identity.ROLE_SUPER_ADMIN,
	}
	tenantUserRequester := identity.Requester{
		RequesterUserId:   2,
		RequesterTenantId: ptrUUID(tenantID),
		RequesterRole:     identity.ROLE_TENANT_USER,
	}

	expectedDeleted := tenant.Tenant{Id: tenantID, Name: "Tenant A", CanImpersonate: true}

	type testCase struct {
		name        string
		input       tenant.DeleteTenantCommand
		setupSteps  []helper.ServiceMockSetupFunc[tenantServiceMockBundle]
		expectedOut tenant.Tenant
		expectedErr error
	}

	// Step 1: get tenant
	getTenantOk := func(m tenantServiceMockBundle) *gomock.Call {
		return m.getTenantPort.EXPECT().
			GetTenant(tenantID).
			Return(expectedDeleted, nil).
			Times(1)
	}

	errMockGetTenant := newMockError(1)
	getTenantError := func(m tenantServiceMockBundle) *gomock.Call {
		return m.getTenantPort.EXPECT().
			GetTenant(tenantID).
			Return(tenant.Tenant{}, errMockGetTenant).
			Times(1)
	}

	getTenantNeverCalled := func(m tenantServiceMockBundle) *gomock.Call {
		return m.getTenantPort.EXPECT().
			GetTenant(gomock.Any()).
			Times(0)
	}

	// Step 2: delete tenant
	deleteTenantOk := func(m tenantServiceMockBundle) *gomock.Call {
		return m.deleteTenantPort.EXPECT().
			DeleteTenant(tenantID).
			Return(expectedDeleted, nil).
			Times(1)
	}

	errMock := newMockError(2)
	deleteTenantError := func(m tenantServiceMockBundle) *gomock.Call {
		return m.deleteTenantPort.EXPECT().
			DeleteTenant(tenantID).
			Return(tenant.Tenant{}, errMock).
			Times(1)
	}

	cases := []testCase{
		{
			name: "(Super Admin) Success",
			input: tenant.DeleteTenantCommand{
				Requester: superAdminRequester,
				TenantId:  tenantID,
			},
			setupSteps: []helper.ServiceMockSetupFunc[tenantServiceMockBundle]{
				getTenantOk,
				deleteTenantOk,
			},
			expectedOut: expectedDeleted,
		},

		// Autorizzazione
		{
			name: "(Tenant User) Fail: unauthorized",
			input: tenant.DeleteTenantCommand{
				Requester: tenantUserRequester,
				TenantId:  tenantID,
			},
			setupSteps: []helper.ServiceMockSetupFunc[tenantServiceMockBundle]{
				getTenantNeverCalled,
			},
			expectedErr: identity.ErrUnauthorizedAccess,
		},

		// Step 1: get tenant
		{
			name: "Fail: get tenant error",
			input: tenant.DeleteTenantCommand{
				Requester: superAdminRequester,
				TenantId:  tenantID,
			},
			setupSteps: []helper.ServiceMockSetupFunc[tenantServiceMockBundle]{
				getTenantError,
			},
			expectedErr: errMockGetTenant,
		},

		// Step 2: delete tenant
		{
			name: "Fail: delete port returns error",
			input: tenant.DeleteTenantCommand{
				Requester: superAdminRequester,
				TenantId:  tenantID,
			},
			setupSteps: []helper.ServiceMockSetupFunc[tenantServiceMockBundle]{
				getTenantOk,
				deleteTenantError,
			},
			expectedErr: errMock,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := helper.SetupServiceWithOrderedSteps(
				t,
				newTenantServiceMockBundle,
				tc.setupSteps,
				newTenantService,
			)

			out, err := service.DeleteTenant(tc.input)
			assertErr(t, err, tc.expectedErr)
			if tc.expectedErr == nil && out != tc.expectedOut {
				t.Fatalf("expected output %+v, got %+v", tc.expectedOut, out)
			}
		})
	}
}

func TestService_GetTenant(t *testing.T) {
	targetTenantID := uuid.New()
	otherTenantID := uuid.New()

	superAdminRequester := identity.Requester{RequesterUserId: 1, RequesterRole: identity.ROLE_SUPER_ADMIN}
	authorizedTenantAdmin := identity.Requester{RequesterUserId: 2, RequesterTenantId: &targetTenantID, RequesterRole: identity.ROLE_TENANT_ADMIN}
	unauthorizedTenantAdmin := identity.Requester{RequesterUserId: 3, RequesterTenantId: &otherTenantID, RequesterRole: identity.ROLE_TENANT_ADMIN}
	authorizedTenantUser := identity.Requester{RequesterUserId: 4, RequesterTenantId: &targetTenantID, RequesterRole: identity.ROLE_TENANT_USER}
	unauthorizedTenantUser := identity.Requester{RequesterUserId: 5, RequesterTenantId: &otherTenantID, RequesterRole: identity.ROLE_TENANT_USER}

	targetTenant := tenant.Tenant{Id: targetTenantID, Name: "Tenant A", CanImpersonate: true}
	errMock := newMockError(3)

	type testCase struct {
		name        string
		input       tenant.GetTenantCommand
		setupSteps  []helper.ServiceMockSetupFunc[tenantServiceMockBundle]
		expectedOut tenant.Tenant
		expectedErr error
	}

	cases := []testCase{
		{
			name:  "(Super Admin) Success",
			input: tenant.GetTenantCommand{Requester: superAdminRequester, TenantId: targetTenantID},
			setupSteps: []helper.ServiceMockSetupFunc[tenantServiceMockBundle]{
				func(m tenantServiceMockBundle) *gomock.Call {
					return m.getTenantPort.EXPECT().
						GetTenant(targetTenantID).
						Return(targetTenant, nil).
						Times(1)
				},
			},
			expectedOut: targetTenant,
		},
		{
			name:  "(Tenant Admin) Success: same tenant",
			input: tenant.GetTenantCommand{Requester: authorizedTenantAdmin, TenantId: targetTenantID},
			setupSteps: []helper.ServiceMockSetupFunc[tenantServiceMockBundle]{
				func(m tenantServiceMockBundle) *gomock.Call {
					return m.getTenantPort.EXPECT().
						GetTenant(targetTenantID).
						Return(targetTenant, nil).
						Times(1)
				},
			},
			expectedOut: targetTenant,
		},
		{
			name:  "Fail: tenant not found",
			input: tenant.GetTenantCommand{Requester: superAdminRequester, TenantId: targetTenantID},
			setupSteps: []helper.ServiceMockSetupFunc[tenantServiceMockBundle]{
				func(m tenantServiceMockBundle) *gomock.Call {
					return m.getTenantPort.EXPECT().
						GetTenant(targetTenantID).
						Return(tenant.Tenant{}, nil).
						Times(1)
				},
			},
			expectedErr: tenant.ErrTenantNotFound,
		},
		{
			name:  "Fail: get tenant port error",
			input: tenant.GetTenantCommand{Requester: superAdminRequester, TenantId: targetTenantID},
			setupSteps: []helper.ServiceMockSetupFunc[tenantServiceMockBundle]{
				func(m tenantServiceMockBundle) *gomock.Call {
					return m.getTenantPort.EXPECT().
						GetTenant(targetTenantID).
						Return(tenant.Tenant{}, errMock).
						Times(1)
				},
			},
			expectedErr: errMock,
		},
		{
			name:  "(Tenant Admin) Fail: unauthorized",
			input: tenant.GetTenantCommand{Requester: unauthorizedTenantAdmin, TenantId: targetTenantID},
			setupSteps: []helper.ServiceMockSetupFunc[tenantServiceMockBundle]{
				func(m tenantServiceMockBundle) *gomock.Call {
					return m.getTenantPort.EXPECT().
						GetTenant(targetTenantID).
						Return(targetTenant, nil).
						Times(1)
				},
			},
			expectedErr: identity.ErrUnauthorizedAccess,
		},
		{
			name:  "(Tenant User) Success: same tenant",
			input: tenant.GetTenantCommand{Requester: authorizedTenantUser, TenantId: targetTenantID},
			setupSteps: []helper.ServiceMockSetupFunc[tenantServiceMockBundle]{
				func(m tenantServiceMockBundle) *gomock.Call {
					return m.getTenantPort.EXPECT().
						GetTenant(targetTenantID).
						Return(targetTenant, nil).
						Times(1)
				},
			},
			expectedOut: targetTenant,
		},
		{
			name:  "(Tenant User) Fail: unauthorized",
			input: tenant.GetTenantCommand{Requester: unauthorizedTenantUser, TenantId: targetTenantID},
			setupSteps: []helper.ServiceMockSetupFunc[tenantServiceMockBundle]{
				func(m tenantServiceMockBundle) *gomock.Call {
					return m.getTenantPort.EXPECT().
						GetTenant(targetTenantID).
						Return(targetTenant, nil).
						Times(1)
				},
			},
			expectedErr: identity.ErrUnauthorizedAccess,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := helper.SetupServiceWithOrderedSteps(
				t,
				newTenantServiceMockBundle,
				tc.setupSteps,
				newTenantService,
			)

			out, err := service.GetTenant(tc.input)
			assertErr(t, err, tc.expectedErr)
			if tc.expectedErr == nil && out != tc.expectedOut {
				t.Fatalf("expected output %+v, got %+v", tc.expectedOut, out)
			}
		})
	}
}

func TestService_GetAllTenants(t *testing.T) {
	tenantList := []tenant.Tenant{{Id: uuid.New(), Name: "A"}, {Id: uuid.New(), Name: "B"}}
	errMock := newMockError(4)

	type testCase struct {
		name        string
		setupSteps  []helper.ServiceMockSetupFunc[tenantServiceMockBundle]
		expectedOut []tenant.Tenant
		expectedErr error
	}

	cases := []testCase{
		{
			name: "Success",
			setupSteps: []helper.ServiceMockSetupFunc[tenantServiceMockBundle]{
				func(m tenantServiceMockBundle) *gomock.Call {
					return m.getTenantsPort.EXPECT().
						GetAllTenants().
						Return(tenantList, nil).
						Times(1)
				},
			},
			expectedOut: tenantList,
		},
		{
			name: "Fail: port error",
			setupSteps: []helper.ServiceMockSetupFunc[tenantServiceMockBundle]{
				func(m tenantServiceMockBundle) *gomock.Call {
					return m.getTenantsPort.EXPECT().
						GetAllTenants().
						Return(nil, errMock).
						Times(1)
				},
			},
			expectedErr: errMock,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := helper.SetupServiceWithOrderedSteps(
				t,
				newTenantServiceMockBundle,
				tc.setupSteps,
				newTenantService,
			)

			out, err := service.GetAllTenants()
			assertErr(t, err, tc.expectedErr)
			if tc.expectedErr == nil {
				if len(out) != len(tc.expectedOut) {
					t.Fatalf("expected %d tenants, got %d", len(tc.expectedOut), len(out))
				}
			}
		})
	}
}

func TestService_GetTenantList(t *testing.T) {
	superAdminRequester := identity.Requester{RequesterUserId: 1, RequesterRole: identity.ROLE_SUPER_ADMIN}
	tenantAdminRequester := identity.Requester{RequesterUserId: 2, RequesterTenantId: ptrUUID(uuid.New()), RequesterRole: identity.ROLE_TENANT_ADMIN}

	tenantList := []tenant.Tenant{{Id: uuid.New(), Name: "Tenant A"}, {Id: uuid.New(), Name: "Tenant B"}}
	total := uint(22)
	errMock := newMockError(5)

	baseInput := tenant.GetTenantListCommand{Page: 2, Limit: 10}

	inputWith := func(r identity.Requester) tenant.GetTenantListCommand {
		cmd := baseInput
		cmd.Requester = r
		return cmd
	}

	type testCase struct {
		name          string
		input         tenant.GetTenantListCommand
		setupSteps    []helper.ServiceMockSetupFunc[tenantServiceMockBundle]
		expectedOut   []tenant.Tenant
		expectedTotal uint
		expectedErr   error
	}

	getTenantsOk := func(m tenantServiceMockBundle) *gomock.Call {
		return m.getTenantsPort.EXPECT().
			GetTenants(baseInput.Page, baseInput.Limit).
			Return(tenantList, total, nil).
			Times(1)
	}

	getTenantsFail := func(m tenantServiceMockBundle) *gomock.Call {
		return m.getTenantsPort.EXPECT().
			GetTenants(baseInput.Page, baseInput.Limit).
			Return(nil, uint(0), errMock).
			Times(1)
	}

	getTenantsError := func(m tenantServiceMockBundle) *gomock.Call {
		return m.getTenantsPort.EXPECT().
			GetTenants(gomock.Any(), gomock.Any()).
			Times(0)
	}

	cases := []testCase{
		{
			name:  "(Super Admin) Success",
			input: inputWith(superAdminRequester),
			setupSteps: []helper.ServiceMockSetupFunc[tenantServiceMockBundle]{
				getTenantsOk,
			},
			expectedOut:   tenantList,
			expectedTotal: total,
		},
		{
			name:  "(Tenant Admin) Fail: unauthorized",
			input: inputWith(tenantAdminRequester),
			setupSteps: []helper.ServiceMockSetupFunc[tenantServiceMockBundle]{
				getTenantsError,
			},
			expectedErr: identity.ErrUnauthorizedAccess,
		},
		{
			name:  "Fail: get tenants port error",
			input: inputWith(superAdminRequester),
			setupSteps: []helper.ServiceMockSetupFunc[tenantServiceMockBundle]{
				getTenantsFail,
			},
			expectedErr: errMock,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := helper.SetupServiceWithOrderedSteps(
				t,
				newTenantServiceMockBundle,
				tc.setupSteps,
				newTenantService,
			)

			out, gotTotal, err := service.GetTenantList(tc.input)
			assertErr(t, err, tc.expectedErr)
			if tc.expectedErr == nil {
				if len(out) != len(tc.expectedOut) {
					t.Fatalf("expected %d tenants, got %d", len(tc.expectedOut), len(out))
				}
				if gotTotal != tc.expectedTotal {
					t.Fatalf("expected total %d, got %d", tc.expectedTotal, gotTotal)
				}
			}
		})
	}
}

func ptrUUID(target uuid.UUID) *uuid.UUID {
	return &target
}
