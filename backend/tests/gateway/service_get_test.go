package gateway_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"backend/internal/gateway"
	"backend/internal/shared/identity"
	"backend/internal/tenant"
	gatewayMocks "backend/tests/gateway/mocks"
	helper "backend/tests/helper"
	tenantMocks "backend/tests/tenant/mocks"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

type gatewayGetServiceMocks struct {
	getGatewayPort  *gatewayMocks.MockGetGatewayPort
	getGatewaysPort *gatewayMocks.MockGetGatewaysPort
	getTenantPort   *tenantMocks.MockGetTenantPort
}

type mockSetupFuncGatewayGetService = helper.ServiceMockSetupFunc[gatewayGetServiceMocks]

func setupGatewayGetService(
	t *testing.T,
	setupSteps []mockSetupFuncGatewayGetService,
) *gateway.GatewayManagementService {
	t.Helper()

	return helper.SetupServiceWithOrderedSteps(
		t,
		func(ctrl *gomock.Controller) gatewayGetServiceMocks {
			return gatewayGetServiceMocks{
				getGatewayPort:  gatewayMocks.NewMockGetGatewayPort(ctrl),
				getGatewaysPort: gatewayMocks.NewMockGetGatewaysPort(ctrl),
				getTenantPort:   tenantMocks.NewMockGetTenantPort(ctrl),
			}
		},
		setupSteps,
		func(mockBundle gatewayGetServiceMocks) *gateway.GatewayManagementService {
			return gateway.NewGatewayManagementService(
				nil,
				mockBundle.getGatewayPort,
				mockBundle.getGatewaysPort,
				mockBundle.getTenantPort,
			)
		},
	)
}

func TestService_GetGateway(t *testing.T) {
	gatewayID := uuid.New()
	requesterSuperAdmin := requesterSuperAdmin()
	requesterTenantAdmin := requesterTenantAdmin(nil)
	cannotFetchGatewayErr := errors.New("cannot fetch gateway")

	expectedGateway := gateway.Gateway{
		Id:            gatewayID,
		Name:          "Gateway A",
		Status:        gateway.GATEWAY_STATUS_ACTIVE,
		IntervalLimit: 3 * time.Second,
	}

	baseCommand := gateway.GetGatewayByIdCommand{GatewayId: gatewayID}

	stepGatewayOk := func(cmd gateway.GetGatewayByIdCommand, result gateway.Gateway) mockSetupFuncGatewayGetService {
		return func(m gatewayGetServiceMocks) *gomock.Call {
			return m.getGatewayPort.EXPECT().GetById(cmd.GatewayId).Return(result, nil).Times(1)
		}
	}
	stepGatewayErr := func(cmd gateway.GetGatewayByIdCommand, expectedErr error) mockSetupFuncGatewayGetService {
		return func(m gatewayGetServiceMocks) *gomock.Call {
			return m.getGatewayPort.EXPECT().GetById(cmd.GatewayId).Return(gateway.Gateway{}, expectedErr).Times(1)
		}
	}
	type testCase struct {
		name          string
		input         gateway.GetGatewayByIdCommand
		setupSteps    []mockSetupFuncGatewayGetService
		expectedValue gateway.Gateway
		expectedError error
	}

	cases := []testCase{
		{
			name: "Success: super admin",
			input: func() gateway.GetGatewayByIdCommand {
				cmd := baseCommand
				cmd.Requester = requesterSuperAdmin
				return cmd
			}(),
			setupSteps: []mockSetupFuncGatewayGetService{
				stepGatewayOk(baseCommand, expectedGateway),
			},
			expectedValue: expectedGateway,
		},
		{
			name: "Fail: gateway port error",
			input: func() gateway.GetGatewayByIdCommand {
				cmd := baseCommand
				cmd.Requester = requesterSuperAdmin
				return cmd
			}(),
			setupSteps: []mockSetupFuncGatewayGetService{
				stepGatewayErr(baseCommand, cannotFetchGatewayErr),
			},
			expectedError: cannotFetchGatewayErr,
		},
		{
			name: "Success: zero value returned by port",
			input: func() gateway.GetGatewayByIdCommand {
				cmd := baseCommand
				cmd.Requester = requesterSuperAdmin
				return cmd
			}(),
			setupSteps: []mockSetupFuncGatewayGetService{
				stepGatewayOk(baseCommand, gateway.Gateway{}),
			},
			expectedValue: gateway.Gateway{},
		},
		{
			name: "Fail: non super admin",
			input: func() gateway.GetGatewayByIdCommand {
				cmd := baseCommand
				cmd.Requester = requesterTenantAdmin
				return cmd
			}(),
			setupSteps: []mockSetupFuncGatewayGetService{
				stepGatewayOk(baseCommand, expectedGateway),
			},
			expectedError: gateway.ErrUnauthorizedAccess,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := setupGatewayGetService(t, tc.setupSteps)

			out, err := service.GetGateway(tc.input)
			if tc.expectedError != nil {
				if !errors.Is(err, tc.expectedError) {
					t.Fatalf("expected error %v, got %v", tc.expectedError, err)
				}
				if out != (gateway.Gateway{}) {
					t.Fatalf("expected zero gateway, got %+v", out)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if !reflect.DeepEqual(tc.expectedValue, out) {
				t.Fatalf("unexpected gateway. expected %+v, got %+v", tc.expectedValue, out)
			}
		})
	}
}

func TestService_GetAllGateways(t *testing.T) {
	gatewayID1 := uuid.New()
	gatewayID2 := uuid.New()
	requesterSuperAdmin := requesterSuperAdmin()
	requesterTenantAdmin := requesterTenantAdmin(nil)
	cannotFetchGatewaysErr := errors.New("cannot fetch gateways")

	gateways := []gateway.Gateway{
		{Id: gatewayID1, Name: "Gateway A", Status: gateway.GATEWAY_STATUS_ACTIVE, IntervalLimit: 3 * time.Second},
		{Id: gatewayID2, Name: "Gateway B", Status: gateway.GATEWAY_STATUS_DECOMMISSIONED, IntervalLimit: 6 * time.Second},
	}

	baseCommand := gateway.GetAllGatewaysCommand{Requester: requesterSuperAdmin, Page: 2, Limit: 10}

	stepGetAllOk := func(cmd gateway.GetAllGatewaysCommand, result []gateway.Gateway, total uint) mockSetupFuncGatewayGetService {
		return func(m gatewayGetServiceMocks) *gomock.Call {
			return m.getGatewaysPort.EXPECT().GetAll(cmd.Page, cmd.Limit).Return(result, total, nil).Times(1)
		}
	}
	stepGetAllErr := func(cmd gateway.GetAllGatewaysCommand, expectedErr error) mockSetupFuncGatewayGetService {
		return func(m gatewayGetServiceMocks) *gomock.Call {
			return m.getGatewaysPort.EXPECT().GetAll(cmd.Page, cmd.Limit).Return(nil, uint(0), expectedErr).Times(1)
		}
	}

	type testCase struct {
		name          string
		input         gateway.GetAllGatewaysCommand
		setupSteps    []mockSetupFuncGatewayGetService
		expectedValue []gateway.Gateway
		expectedTotal uint
		expectedError error
	}

	cases := []testCase{
		{
			name:  "Success: super admin",
			input: baseCommand,
			setupSteps: []mockSetupFuncGatewayGetService{
				stepGetAllOk(baseCommand, gateways, uint(7)),
			},
			expectedValue: gateways,
			expectedTotal: uint(7),
		},
		{
			name:  "Fail: get all error",
			input: baseCommand,
			setupSteps: []mockSetupFuncGatewayGetService{
				stepGetAllErr(baseCommand, cannotFetchGatewaysErr),
			},
			expectedError: cannotFetchGatewaysErr,
		},
		{
			name:  "Fail: nil slice returned",
			input: baseCommand,
			setupSteps: []mockSetupFuncGatewayGetService{
				stepGetAllOk(baseCommand, nil, uint(0)),
			},
			expectedError: gateway.ErrGatewayNotFound,
		},
		{
			name: "Fail: non super admin",
			input: func() gateway.GetAllGatewaysCommand {
				cmd := baseCommand
				cmd.Requester = requesterTenantAdmin
				return cmd
			}(),
			setupSteps: []mockSetupFuncGatewayGetService{
				stepGetAllOk(baseCommand, gateways, uint(7)),
			},
			expectedError: gateway.ErrUnauthorizedAccess,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := setupGatewayGetService(t, tc.setupSteps)

			out, total, err := service.GetAllGateways(tc.input)
			if tc.expectedError != nil {
				if !errors.Is(err, tc.expectedError) {
					t.Fatalf("expected error %v, got %v", tc.expectedError, err)
				}
				if out != nil {
					t.Fatalf("expected nil gateways, got %+v", out)
				}
				if total != 0 {
					t.Fatalf("expected total 0, got %d", total)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if !reflect.DeepEqual(tc.expectedValue, out) {
				t.Fatalf("unexpected gateways. expected %+v, got %+v", tc.expectedValue, out)
			}
			if tc.expectedTotal != total {
				t.Fatalf("unexpected total. expected %d, got %d", tc.expectedTotal, total)
			}
		})
	}
}

func TestService_GetGatewaysByTenant(t *testing.T) {
	targetTenantId := uuid.New()
	otherTenantId := uuid.New()
	gatewayID := uuid.New()
	cannotFetchTenantErr := errors.New("cannot fetch tenant")
	cannotFetchGatewaysErr := errors.New("cannot fetch gateways")

	canImpersonateTenant := tenant.Tenant{Id: targetTenantId, Name: "Tenant A", CanImpersonate: true}
	cannotImpersonateTenant := tenant.Tenant{Id: targetTenantId, Name: "Tenant A", CanImpersonate: false}

	gateways := []gateway.Gateway{
		{Id: gatewayID, Name: "Gateway A", TenantId: &targetTenantId, Status: gateway.GATEWAY_STATUS_ACTIVE, IntervalLimit: 3 * time.Second},
	}

	baseCommand := gateway.GetGatewaysByTenantCommand{TenantId: targetTenantId, Page: 3, Limit: 8}

	superAdmin := requesterSuperAdmin()
	tenantAdmin := requesterTenantAdmin(&targetTenantId)
	tenantUser := identity.Requester{RequesterUserId: 3, RequesterTenantId: &targetTenantId, RequesterRole: identity.ROLE_TENANT_USER}
	tenantAdminOtherTenant := requesterTenantAdmin(&otherTenantId)

	stepTenantOk := func(cmd gateway.GetGatewaysByTenantCommand, foundTenant tenant.Tenant) mockSetupFuncGatewayGetService {
		return func(m gatewayGetServiceMocks) *gomock.Call {
			return m.getTenantPort.EXPECT().GetTenant(cmd.TenantId).Return(foundTenant, nil).Times(1)
		}
	}
	stepTenantErr := func(cmd gateway.GetGatewaysByTenantCommand, expectedErr error) mockSetupFuncGatewayGetService {
		return func(m gatewayGetServiceMocks) *gomock.Call {
			return m.getTenantPort.EXPECT().GetTenant(cmd.TenantId).Return(tenant.Tenant{}, expectedErr).Times(1)
		}
	}
	stepTenantNeverCalled := func() mockSetupFuncGatewayGetService {
		return func(m gatewayGetServiceMocks) *gomock.Call {
			return m.getTenantPort.EXPECT().GetTenant(gomock.Any()).Times(0)
		}
	}
	stepGatewaysOk := func(cmd gateway.GetGatewaysByTenantCommand, result []gateway.Gateway, total uint) mockSetupFuncGatewayGetService {
		return func(m gatewayGetServiceMocks) *gomock.Call {
			return m.getGatewaysPort.EXPECT().GetByTenantId(cmd.TenantId, cmd.Page, cmd.Limit).Return(result, total, nil).Times(1)
		}
	}
	stepGatewaysErr := func(cmd gateway.GetGatewaysByTenantCommand, expectedErr error) mockSetupFuncGatewayGetService {
		return func(m gatewayGetServiceMocks) *gomock.Call {
			return m.getGatewaysPort.EXPECT().GetByTenantId(cmd.TenantId, cmd.Page, cmd.Limit).Return(nil, uint(0), expectedErr).Times(1)
		}
	}
	stepGatewaysNeverCalled := func() mockSetupFuncGatewayGetService {
		return func(m gatewayGetServiceMocks) *gomock.Call {
			return m.getGatewaysPort.EXPECT().GetByTenantId(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
		}
	}

	type testCase struct {
		name          string
		input         gateway.GetGatewaysByTenantCommand
		setupSteps    []mockSetupFuncGatewayGetService
		expectedValue []gateway.Gateway
		expectedTotal uint
		expectedError error
	}

	cases := []testCase{
		{
			name:  "Success: super admin with impersonation",
			input: func() gateway.GetGatewaysByTenantCommand { cmd := baseCommand; cmd.Requester = superAdmin; return cmd }(),
			setupSteps: []mockSetupFuncGatewayGetService{
				stepTenantOk(baseCommand, canImpersonateTenant),
				stepGatewaysOk(baseCommand, gateways, uint(1)),
			},
			expectedValue: gateways,
			expectedTotal: uint(1),
		},
		{
			name:  "Success: tenant admin same tenant",
			input: func() gateway.GetGatewaysByTenantCommand { cmd := baseCommand; cmd.Requester = tenantAdmin; return cmd }(),
			setupSteps: []mockSetupFuncGatewayGetService{
				stepTenantOk(baseCommand, cannotImpersonateTenant),
				stepGatewaysOk(baseCommand, gateways, uint(1)),
			},
			expectedValue: gateways,
			expectedTotal: uint(1),
		},
		{
			name:  "Fail: tenant user unauthorized",
			input: func() gateway.GetGatewaysByTenantCommand { cmd := baseCommand; cmd.Requester = tenantUser; return cmd }(),
			setupSteps: []mockSetupFuncGatewayGetService{
				stepTenantOk(baseCommand, cannotImpersonateTenant),
				stepGatewaysNeverCalled(),
			},
			expectedError: gateway.ErrUnauthorizedAccess,
		},
		{
			name:  "Fail: super admin without impersonation",
			input: func() gateway.GetGatewaysByTenantCommand { cmd := baseCommand; cmd.Requester = superAdmin; return cmd }(),
			setupSteps: []mockSetupFuncGatewayGetService{
				stepTenantOk(baseCommand, cannotImpersonateTenant),
				stepGatewaysNeverCalled(),
			},
			expectedError: gateway.ErrUnauthorizedAccess,
		},
		{
			name:  "Fail: tenant not found",
			input: func() gateway.GetGatewaysByTenantCommand { cmd := baseCommand; cmd.Requester = superAdmin; return cmd }(),
			setupSteps: []mockSetupFuncGatewayGetService{
				stepTenantOk(baseCommand, tenant.Tenant{}),
				stepGatewaysNeverCalled(),
			},
			expectedError: tenant.ErrTenantNotFound,
		},
		{
			name:  "Fail: tenant port error",
			input: func() gateway.GetGatewaysByTenantCommand { cmd := baseCommand; cmd.Requester = superAdmin; return cmd }(),
			setupSteps: []mockSetupFuncGatewayGetService{
				stepTenantErr(baseCommand, cannotFetchTenantErr),
				stepGatewaysNeverCalled(),
			},
			expectedError: cannotFetchTenantErr,
		},
		{
			name:  "Fail: gateways port error",
			input: func() gateway.GetGatewaysByTenantCommand { cmd := baseCommand; cmd.Requester = superAdmin; return cmd }(),
			setupSteps: []mockSetupFuncGatewayGetService{
				stepTenantOk(baseCommand, canImpersonateTenant),
				stepGatewaysErr(baseCommand, cannotFetchGatewaysErr),
			},
			expectedError: cannotFetchGatewaysErr,
		},
		{
			name: "Fail: nil tenant id",
			input: func() gateway.GetGatewaysByTenantCommand {
				cmd := baseCommand
				cmd.Requester = superAdmin
				cmd.TenantId = uuid.Nil
				return cmd
			}(),
			setupSteps: []mockSetupFuncGatewayGetService{
				stepTenantNeverCalled(),
				stepGatewaysNeverCalled(),
			},
			expectedError: gateway.ErrGatewayNotFound,
		},
		{
			name:  "Fail: empty gateway list",
			input: func() gateway.GetGatewaysByTenantCommand { cmd := baseCommand; cmd.Requester = superAdmin; return cmd }(),
			setupSteps: []mockSetupFuncGatewayGetService{
				stepTenantOk(baseCommand, canImpersonateTenant),
				stepGatewaysOk(baseCommand, nil, uint(0)),
			},
			expectedError: gateway.ErrGatewayNotFound,
		},
		{
			name: "Fail: tenant admin on other tenant",
			input: func() gateway.GetGatewaysByTenantCommand {
				cmd := baseCommand
				cmd.Requester = tenantAdminOtherTenant
				return cmd
			}(),
			setupSteps: []mockSetupFuncGatewayGetService{
				stepTenantOk(baseCommand, cannotImpersonateTenant),
				stepGatewaysNeverCalled(),
			},
			expectedError: gateway.ErrUnauthorizedAccess,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := setupGatewayGetService(t, tc.setupSteps)

			out, total, err := service.GetGatewaysByTenant(tc.input)
			if tc.expectedError != nil {
				if !errors.Is(err, tc.expectedError) {
					t.Fatalf("expected error %v, got %v", tc.expectedError, err)
				}
				if out != nil {
					t.Fatalf("expected nil gateways, got %+v", out)
				}
				if total != 0 {
					t.Fatalf("expected total 0, got %d", total)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if !reflect.DeepEqual(tc.expectedValue, out) {
				t.Fatalf("unexpected gateways. expected %+v, got %+v", tc.expectedValue, out)
			}
			if tc.expectedTotal != total {
				t.Fatalf("unexpected total. expected %d, got %d", tc.expectedTotal, total)
			}
		})
	}
}

func TestService_GetGatewayByTenantID(t *testing.T) {
	targetTenantId := uuid.New()
	otherTenantId := uuid.New()
	gatewayID := uuid.New()
	cannotFetchTenantErr := errors.New("cannot fetch tenant")
	cannotFetchGatewayErr := errors.New("cannot fetch gateway")

	canImpersonateTenant := tenant.Tenant{Id: targetTenantId, Name: "Tenant A", CanImpersonate: true}
	cannotImpersonateTenant := tenant.Tenant{Id: targetTenantId, Name: "Tenant A", CanImpersonate: false}

	resultGateway := gateway.Gateway{
		Id:            gatewayID,
		Name:          "Gateway A",
		TenantId:      &targetTenantId,
		Status:        gateway.GATEWAY_STATUS_ACTIVE,
		IntervalLimit: 5 * time.Second,
	}

	baseCommand := gateway.GetGatewayByTenantIDCommand{TenantId: targetTenantId, GatewayId: gatewayID}

	superAdmin := requesterSuperAdmin()
	tenantAdmin := requesterTenantAdmin(&targetTenantId)
	tenantUser := identity.Requester{RequesterUserId: 3, RequesterTenantId: &targetTenantId, RequesterRole: identity.ROLE_TENANT_USER}
	tenantAdminOtherTenant := requesterTenantAdmin(&otherTenantId)

	stepGatewayOk := func(cmd gateway.GetGatewayByTenantIDCommand, result gateway.Gateway) mockSetupFuncGatewayGetService {
		return func(m gatewayGetServiceMocks) *gomock.Call {
			return m.getGatewayPort.EXPECT().GetGatewayByTenantID(cmd.TenantId, cmd.GatewayId).Return(result, nil).Times(1)
		}
	}
	stepGatewayErr := func(cmd gateway.GetGatewayByTenantIDCommand, expectedErr error) mockSetupFuncGatewayGetService {
		return func(m gatewayGetServiceMocks) *gomock.Call {
			return m.getGatewayPort.EXPECT().GetGatewayByTenantID(cmd.TenantId, cmd.GatewayId).Return(gateway.Gateway{}, expectedErr).Times(1)
		}
	}
	stepTenantOk := func(cmd gateway.GetGatewayByTenantIDCommand, foundTenant tenant.Tenant) mockSetupFuncGatewayGetService {
		return func(m gatewayGetServiceMocks) *gomock.Call {
			return m.getTenantPort.EXPECT().GetTenant(cmd.TenantId).Return(foundTenant, nil).Times(1)
		}
	}
	stepTenantErr := func(cmd gateway.GetGatewayByTenantIDCommand, expectedErr error) mockSetupFuncGatewayGetService {
		return func(m gatewayGetServiceMocks) *gomock.Call {
			return m.getTenantPort.EXPECT().GetTenant(cmd.TenantId).Return(tenant.Tenant{}, expectedErr).Times(1)
		}
	}
	stepTenantNeverCalled := func() mockSetupFuncGatewayGetService {
		return func(m gatewayGetServiceMocks) *gomock.Call {
			return m.getTenantPort.EXPECT().GetTenant(gomock.Any()).Times(0)
		}
	}

	type testCase struct {
		name          string
		input         gateway.GetGatewayByTenantIDCommand
		setupSteps    []mockSetupFuncGatewayGetService
		expectedValue gateway.Gateway
		expectedError error
	}

	cases := []testCase{
		{
			name:  "Success: super admin with impersonation",
			input: func() gateway.GetGatewayByTenantIDCommand { cmd := baseCommand; cmd.Requester = superAdmin; return cmd }(),
			setupSteps: []mockSetupFuncGatewayGetService{
				stepGatewayOk(baseCommand, resultGateway),
				stepTenantOk(baseCommand, canImpersonateTenant),
			},
			expectedValue: resultGateway,
		},
		{
			name: "Success: tenant admin same tenant",
			input: func() gateway.GetGatewayByTenantIDCommand {
				cmd := baseCommand
				cmd.Requester = tenantAdmin
				return cmd
			}(),
			setupSteps: []mockSetupFuncGatewayGetService{
				stepGatewayOk(baseCommand, resultGateway),
				stepTenantOk(baseCommand, cannotImpersonateTenant),
			},
			expectedValue: resultGateway,
		},
		{
			name:  "Fail: tenant user unauthorized",
			input: func() gateway.GetGatewayByTenantIDCommand { cmd := baseCommand; cmd.Requester = tenantUser; return cmd }(),
			setupSteps: []mockSetupFuncGatewayGetService{
				stepGatewayOk(baseCommand, resultGateway),
				stepTenantOk(baseCommand, cannotImpersonateTenant),
			},
			expectedError: gateway.ErrUnauthorizedAccess,
		},
		{
			name:  "Fail: super admin without impersonation",
			input: func() gateway.GetGatewayByTenantIDCommand { cmd := baseCommand; cmd.Requester = superAdmin; return cmd }(),
			setupSteps: []mockSetupFuncGatewayGetService{
				stepGatewayOk(baseCommand, resultGateway),
				stepTenantOk(baseCommand, cannotImpersonateTenant),
			},
			expectedError: gateway.ErrUnauthorizedAccess,
		},
		{
			name:  "Fail: gateway not found",
			input: func() gateway.GetGatewayByTenantIDCommand { cmd := baseCommand; cmd.Requester = superAdmin; return cmd }(),
			setupSteps: []mockSetupFuncGatewayGetService{
				stepGatewayOk(baseCommand, gateway.Gateway{}),
				stepTenantNeverCalled(),
			},
			expectedError: gateway.ErrGatewayNotFound,
		},
		{
			name:  "Fail: gateway port error",
			input: func() gateway.GetGatewayByTenantIDCommand { cmd := baseCommand; cmd.Requester = superAdmin; return cmd }(),
			setupSteps: []mockSetupFuncGatewayGetService{
				stepGatewayErr(baseCommand, cannotFetchGatewayErr),
				stepTenantNeverCalled(),
			},
			expectedError: cannotFetchGatewayErr,
		},
		{
			name:  "Fail: tenant not found",
			input: func() gateway.GetGatewayByTenantIDCommand { cmd := baseCommand; cmd.Requester = superAdmin; return cmd }(),
			setupSteps: []mockSetupFuncGatewayGetService{
				stepGatewayOk(baseCommand, resultGateway),
				stepTenantOk(baseCommand, tenant.Tenant{}),
			},
			expectedError: tenant.ErrTenantNotFound,
		},
		{
			name:  "Fail: tenant port error",
			input: func() gateway.GetGatewayByTenantIDCommand { cmd := baseCommand; cmd.Requester = superAdmin; return cmd }(),
			setupSteps: []mockSetupFuncGatewayGetService{
				stepGatewayOk(baseCommand, resultGateway),
				stepTenantErr(baseCommand, cannotFetchTenantErr),
			},
			expectedError: cannotFetchTenantErr,
		},
		{
			name: "Fail: tenant admin on other tenant",
			input: func() gateway.GetGatewayByTenantIDCommand {
				cmd := baseCommand
				cmd.Requester = tenantAdminOtherTenant
				return cmd
			}(),
			setupSteps: []mockSetupFuncGatewayGetService{
				stepGatewayOk(baseCommand, resultGateway),
				stepTenantOk(baseCommand, cannotImpersonateTenant),
			},
			expectedError: gateway.ErrUnauthorizedAccess,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := setupGatewayGetService(t, tc.setupSteps)

			out, err := service.GetGatewayByTenantID(tc.input)
			if tc.expectedError != nil {
				if !errors.Is(err, tc.expectedError) {
					t.Fatalf("expected error %v, got %v", tc.expectedError, err)
				}
				if out != (gateway.Gateway{}) {
					t.Fatalf("expected zero gateway, got %+v", out)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if !reflect.DeepEqual(tc.expectedValue, out) {
				t.Fatalf("unexpected gateway. expected %+v, got %+v", tc.expectedValue, out)
			}
		})
	}
}
