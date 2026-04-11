package gateway_test

import (
	"errors"
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

type gatewayCommandServiceMocks struct {
	createGatewayPort  *gatewayMocks.MockCreateGatewayPort
	deleteGatewayPort  *gatewayMocks.MockDeleteGatewayPort
	getGatewayPort     *gatewayMocks.MockGetGatewayPort
	getTenantPort      *tenantMocks.MockGetTenantPort
	saveGatewayPort    *gatewayMocks.MockSaveGatewayPort
	gatewayCommandPort *gatewayMocks.MockGatewayCommandPort
}

type mockSetupFuncGatewayCommandService = helper.ServiceMockSetupFunc[gatewayCommandServiceMocks]

func setupGatewayCommandService(
	t *testing.T,
	setupSteps []mockSetupFuncGatewayCommandService,
) *gateway.GatewayCommandService {
	t.Helper()

	return helper.SetupServiceWithOrderedSteps(
		t,
		func(ctrl *gomock.Controller) gatewayCommandServiceMocks {
			return gatewayCommandServiceMocks{
				createGatewayPort:  gatewayMocks.NewMockCreateGatewayPort(ctrl),
				deleteGatewayPort:  gatewayMocks.NewMockDeleteGatewayPort(ctrl),
				getGatewayPort:     gatewayMocks.NewMockGetGatewayPort(ctrl),
				getTenantPort:      tenantMocks.NewMockGetTenantPort(ctrl),
				saveGatewayPort:    gatewayMocks.NewMockSaveGatewayPort(ctrl),
				gatewayCommandPort: gatewayMocks.NewMockGatewayCommandPort(ctrl),
			}
		},
		setupSteps,
		func(mockBundle gatewayCommandServiceMocks) *gateway.GatewayCommandService {
			return gateway.NewGatewayCommandService(
				mockBundle.createGatewayPort,
				mockBundle.deleteGatewayPort,
				mockBundle.getGatewayPort,
				mockBundle.getTenantPort,
				mockBundle.saveGatewayPort,
				mockBundle.gatewayCommandPort,
			)
		},
	)
}

func requesterSuperAdmin() identity.Requester {
	return identity.Requester{RequesterUserId: 1, RequesterRole: identity.ROLE_SUPER_ADMIN}
}

func requesterTenantAdmin(tenantID *uuid.UUID) identity.Requester {
	return identity.Requester{RequesterUserId: 2, RequesterRole: identity.ROLE_TENANT_ADMIN, RequesterTenantId: tenantID}
}

func TestService_CommissionGateway(t *testing.T) {
	gatewayID := uuid.New()
	tenantID := uuid.New()

	activeTenant := tenant.Tenant{Id: tenantID, Name: "Tenant-A"}
	baseGateway := gateway.Gateway{Id: gatewayID, Name: "GW-1", Status: gateway.GATEWAY_STATUS_INACTIVE, IntervalLimit: 2 * time.Second}

	t.Run("Fail: utente non autorizzato", func(t *testing.T) {
		cmd := gateway.CommissionGatewayCommand{
			GatewayId:       gatewayID,
			TenantId:        tenantID,
			CommissionToken: "token",
			Requester:       requesterTenantAdmin(&tenantID),
		}

		service := setupGatewayCommandService(t, []mockSetupFuncGatewayCommandService{
			func(m gatewayCommandServiceMocks) *gomock.Call {
				m.getGatewayPort.EXPECT().GetById(gomock.Any()).Times(0)
				m.getTenantPort.EXPECT().GetTenant(gomock.Any()).Times(0)
				m.gatewayCommandPort.EXPECT().SendCommission(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				m.saveGatewayPort.EXPECT().Save(gomock.Any()).Times(0)
				return nil
			},
		})

		_, err := service.CommissionGateway(cmd)
		if !errors.Is(err, identity.ErrUnauthorizedAccess) {
			t.Fatalf("expected %v, got %v", identity.ErrUnauthorizedAccess, err)
		}
	})

	t.Run("Fail: tenant non trovato", func(t *testing.T) {
		cmd := gateway.CommissionGatewayCommand{
			GatewayId:       gatewayID,
			TenantId:        tenantID,
			CommissionToken: "token",
			Requester:       requesterSuperAdmin(),
		}

		service := setupGatewayCommandService(t, []mockSetupFuncGatewayCommandService{
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.getGatewayPort.EXPECT().GetById(gatewayID).Return(baseGateway, nil).Times(1)
			},
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.getTenantPort.EXPECT().GetTenant(tenantID).Return(tenant.Tenant{}, nil).Times(1)
			},
			func(m gatewayCommandServiceMocks) *gomock.Call {
				m.gatewayCommandPort.EXPECT().SendCommission(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				m.saveGatewayPort.EXPECT().Save(gomock.Any()).Times(0)
				return nil
			},
		})

		_, err := service.CommissionGateway(cmd)
		if !errors.Is(err, tenant.ErrTenantNotFound) {
			t.Fatalf("expected %v, got %v", tenant.ErrTenantNotFound, err)
		}
	})

	t.Run("Success: commissiona e salva gateway", func(t *testing.T) {
		cmd := gateway.CommissionGatewayCommand{
			GatewayId:       gatewayID,
			TenantId:        tenantID,
			CommissionToken: "token",
			Requester:       requesterSuperAdmin(),
		}

		service := setupGatewayCommandService(t, []mockSetupFuncGatewayCommandService{
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.getGatewayPort.EXPECT().GetById(gatewayID).Return(baseGateway, nil).Times(1)
			},
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.getTenantPort.EXPECT().GetTenant(tenantID).Return(activeTenant, nil).Times(1)
			},
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.gatewayCommandPort.EXPECT().SendCommission(gatewayID, tenantID, "token").Return(nil).Times(1)
			},
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.saveGatewayPort.EXPECT().Save(gomock.Any()).DoAndReturn(func(g gateway.Gateway) (gateway.Gateway, error) {
					if g.Status != gateway.GATEWAY_STATUS_ACTIVE {
						t.Fatalf("expected gateway status %s, got %s", gateway.GATEWAY_STATUS_ACTIVE, g.Status)
					}
					if g.TenantId == nil || *g.TenantId != tenantID {
						t.Fatalf("expected tenant id to be set to %s", tenantID)
					}
					return g, nil
				}).Times(1)
			},
		})

		updated, err := service.CommissionGateway(cmd)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if updated.TenantId == nil || *updated.TenantId != tenantID {
			t.Fatalf("expected tenant id %s, got %+v", tenantID, updated.TenantId)
		}
	})
}

func TestService_DecommissionGateway(t *testing.T) {
	gatewayID := uuid.New()
	tenantID := uuid.New()

	commissionedGateway := gateway.Gateway{Id: gatewayID, Name: "GW-2", TenantId: &tenantID, Status: gateway.GATEWAY_STATUS_ACTIVE}

	t.Run("Fail: gateway non commissionato", func(t *testing.T) {
		cmd := gateway.DecommissionGatewayCommand{GatewayId: gatewayID, Requester: requesterSuperAdmin()}
		uncommissioned := commissionedGateway
		uncommissioned.TenantId = nil

		service := setupGatewayCommandService(t, []mockSetupFuncGatewayCommandService{
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.getGatewayPort.EXPECT().GetById(gatewayID).Return(uncommissioned, nil).Times(1)
			},
			func(m gatewayCommandServiceMocks) *gomock.Call {
				m.gatewayCommandPort.EXPECT().SendDecommission(gomock.Any()).Times(0)
				m.saveGatewayPort.EXPECT().Save(gomock.Any()).Times(0)
				return nil
			},
		})

		_, err := service.DecommissionGateway(cmd)
		if !errors.Is(err, gateway.ErrGatewayNotCommissioned) {
			t.Fatalf("expected %v, got %v", gateway.ErrGatewayNotCommissioned, err)
		}
	})

	t.Run("Fail: errore comunicazione NATS", func(t *testing.T) {
		cmd := gateway.DecommissionGatewayCommand{GatewayId: gatewayID, Requester: requesterSuperAdmin()}
		natsErr := errors.New("nats down")

		service := setupGatewayCommandService(t, []mockSetupFuncGatewayCommandService{
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.getGatewayPort.EXPECT().GetById(gatewayID).Return(commissionedGateway, nil).Times(1)
			},
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.gatewayCommandPort.EXPECT().SendDecommission(gatewayID).Return(natsErr).Times(1)
			},
		})

		_, err := service.DecommissionGateway(cmd)
		if !errors.Is(err, natsErr) {
			t.Fatalf("expected %v, got %v", natsErr, err)
		}
	})

	t.Run("Success: decommissiona e salva", func(t *testing.T) {
		cmd := gateway.DecommissionGatewayCommand{GatewayId: gatewayID, Requester: requesterSuperAdmin()}

		service := setupGatewayCommandService(t, []mockSetupFuncGatewayCommandService{
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.getGatewayPort.EXPECT().GetById(gatewayID).Return(commissionedGateway, nil).Times(1)
			},
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.gatewayCommandPort.EXPECT().SendDecommission(gatewayID).Return(nil).Times(1)
			},
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.saveGatewayPort.EXPECT().Save(gomock.Any()).DoAndReturn(func(g gateway.Gateway) (gateway.Gateway, error) {
					if g.Status != gateway.GATEWAY_STATUS_DECOMMISSIONED {
						t.Fatalf("expected status %s, got %s", gateway.GATEWAY_STATUS_DECOMMISSIONED, g.Status)
					}
					if g.TenantId != nil {
						t.Fatalf("expected tenant id nil after decommission")
					}
					return g, nil
				}).Times(1)
			},
		})

		decommissioned, err := service.DecommissionGateway(cmd)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if decommissioned.TenantId != nil {
			t.Fatalf("expected nil tenant id, got %+v", decommissioned.TenantId)
		}
	})
}

func TestService_ResetGateway(t *testing.T) {
	gatewayID := uuid.New()
	tenantID := uuid.New()
	otherTenantID := uuid.New()

	baseGateway := gateway.Gateway{
		Id:            gatewayID,
		Name:          "GW-3",
		TenantId:      &tenantID,
		Status:        gateway.GATEWAY_STATUS_ACTIVE,
		IntervalLimit: 30 * time.Second,
	}

	t.Run("Fail: tenant admin non autorizzato", func(t *testing.T) {
		cmd := gateway.ResetGatewayCommand{GatewayId: gatewayID, Requester: requesterTenantAdmin(&otherTenantID)}

		service := setupGatewayCommandService(t, []mockSetupFuncGatewayCommandService{
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.getGatewayPort.EXPECT().GetById(gatewayID).Return(baseGateway, nil).Times(1)
			},
			func(m gatewayCommandServiceMocks) *gomock.Call {
				m.gatewayCommandPort.EXPECT().SendReset(gomock.Any()).Times(0)
				m.saveGatewayPort.EXPECT().Save(gomock.Any()).Times(0)
				return nil
			},
		})

		_, err := service.ResetGateway(cmd)
		if !errors.Is(err, identity.ErrUnauthorizedAccess) {
			t.Fatalf("expected %v, got %v", identity.ErrUnauthorizedAccess, err)
		}
	})

	t.Run("Fail: errore comunicazione NATS", func(t *testing.T) {
		cmd := gateway.ResetGatewayCommand{GatewayId: gatewayID, Requester: requesterSuperAdmin()}
		natsErr := errors.New("nats reset failed")

		service := setupGatewayCommandService(t, []mockSetupFuncGatewayCommandService{
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.getGatewayPort.EXPECT().GetById(gatewayID).Return(baseGateway, nil).Times(1)
			},
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.gatewayCommandPort.EXPECT().SendReset(gatewayID).Return(natsErr).Times(1)
			},
		})

		_, err := service.ResetGateway(cmd)
		if !errors.Is(err, natsErr) {
			t.Fatalf("expected %v, got %v", natsErr, err)
		}
	})

	t.Run("Success: resetta intervallo al default e salva", func(t *testing.T) {
		cmd := gateway.ResetGatewayCommand{GatewayId: gatewayID, Requester: requesterSuperAdmin()}

		service := setupGatewayCommandService(t, []mockSetupFuncGatewayCommandService{
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.getGatewayPort.EXPECT().GetById(gatewayID).Return(baseGateway, nil).Times(1)
			},
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.gatewayCommandPort.EXPECT().SendReset(gatewayID).Return(nil).Times(1)
			},
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.saveGatewayPort.EXPECT().Save(gomock.Any()).DoAndReturn(func(g gateway.Gateway) (gateway.Gateway, error) {
					if g.IntervalLimit != gateway.DEFAULT_INTERVAL_LIMIT {
						t.Fatalf("expected interval %v, got %v", gateway.DEFAULT_INTERVAL_LIMIT, g.IntervalLimit)
					}
					return g, nil
				}).Times(1)
			},
		})

		updated, err := service.ResetGateway(cmd)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if updated.IntervalLimit != gateway.DEFAULT_INTERVAL_LIMIT {
			t.Fatalf("expected interval %v, got %v", gateway.DEFAULT_INTERVAL_LIMIT, updated.IntervalLimit)
		}
	})
}

func TestService_RebootGateway(t *testing.T) {
	gatewayID := uuid.New()
	tenantID := uuid.New()

	baseGateway := gateway.Gateway{Id: gatewayID, Name: "GW-4", TenantId: &tenantID, Status: gateway.GATEWAY_STATUS_ACTIVE}

	t.Run("Fail: errore comunicazione NATS", func(t *testing.T) {
		cmd := gateway.RebootGatewayCommand{GatewayId: gatewayID, Requester: requesterSuperAdmin()}
		natsErr := errors.New("nats reboot failed")
		service := setupGatewayCommandService(t, []mockSetupFuncGatewayCommandService{
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.getGatewayPort.EXPECT().GetById(gatewayID).Return(baseGateway, nil).Times(1)
			},
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.gatewayCommandPort.EXPECT().SendReboot(gatewayID).Return(natsErr).Times(1)
			},
		})

		_, err := service.RebootGateway(cmd)
		if !errors.Is(err, natsErr) {
			t.Fatalf("expected %v, got %v", natsErr, err)
		}
	})

	t.Run("Success: invia reboot senza save", func(t *testing.T) {
		cmd := gateway.RebootGatewayCommand{GatewayId: gatewayID, Requester: requesterSuperAdmin()}
		service := setupGatewayCommandService(t, []mockSetupFuncGatewayCommandService{
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.getGatewayPort.EXPECT().GetById(gatewayID).Return(baseGateway, nil).Times(1)
			},
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.gatewayCommandPort.EXPECT().SendReboot(gatewayID).Return(nil).Times(1)
			},
			func(m gatewayCommandServiceMocks) *gomock.Call {
				m.saveGatewayPort.EXPECT().Save(gomock.Any()).Times(0)
				return nil
			},
		})

		rebooted, err := service.RebootGateway(cmd)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if rebooted.Id != gatewayID {
			t.Fatalf("expected gateway id %s, got %s", gatewayID, rebooted.Id)
		}
	})
}

func TestService_InterruptGateway(t *testing.T) {
	gatewayID := uuid.New()
	tenantID := uuid.New()
	otherTenantID := uuid.New()

	activeGateway := gateway.Gateway{Id: gatewayID, Name: "GW-5", TenantId: &tenantID, Status: gateway.GATEWAY_STATUS_ACTIVE}
	inactiveGateway := gateway.Gateway{Id: gatewayID, Name: "GW-5", TenantId: &tenantID, Status: gateway.GATEWAY_STATUS_INACTIVE}

	t.Run("Fail: gateway non attivo", func(t *testing.T) {
		cmd := gateway.InterruptGatewayCommand{GatewayId: gatewayID, Requester: requesterSuperAdmin()}

		service := setupGatewayCommandService(t, []mockSetupFuncGatewayCommandService{
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.getGatewayPort.EXPECT().GetById(gatewayID).Return(inactiveGateway, nil).Times(1)
			},
			func(m gatewayCommandServiceMocks) *gomock.Call {
				m.gatewayCommandPort.EXPECT().SendInterrupt(gomock.Any()).Times(0)
				m.saveGatewayPort.EXPECT().Save(gomock.Any()).Times(0)
				return nil
			},
		})

		_, err := service.InterruptGateway(cmd)
		if !errors.Is(err, gateway.ErrGatewayNotActive) {
			t.Fatalf("expected %v, got %v", gateway.ErrGatewayNotActive, err)
		}
	})

	t.Run("Fail: tenant admin non autorizzato", func(t *testing.T) {
		cmd := gateway.InterruptGatewayCommand{GatewayId: gatewayID, Requester: requesterTenantAdmin(&otherTenantID)}

		service := setupGatewayCommandService(t, []mockSetupFuncGatewayCommandService{
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.getGatewayPort.EXPECT().GetById(gatewayID).Return(activeGateway, nil).Times(1)
			},
			func(m gatewayCommandServiceMocks) *gomock.Call {
				m.gatewayCommandPort.EXPECT().SendInterrupt(gomock.Any()).Times(0)
				m.saveGatewayPort.EXPECT().Save(gomock.Any()).Times(0)
				return nil
			},
		})

		_, err := service.InterruptGateway(cmd)
		if !errors.Is(err, identity.ErrUnauthorizedAccess) {
			t.Fatalf("expected %v, got %v", identity.ErrUnauthorizedAccess, err)
		}
	})

	t.Run("Fail: errore NATS ritornato dal port", func(t *testing.T) {
		cmd := gateway.InterruptGatewayCommand{GatewayId: gatewayID, Requester: requesterSuperAdmin()}
		natsErr := errors.New("nats interrupt failed")

		service := setupGatewayCommandService(t, []mockSetupFuncGatewayCommandService{
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.getGatewayPort.EXPECT().GetById(gatewayID).Return(activeGateway, nil).Times(1)
			},
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.gatewayCommandPort.EXPECT().SendInterrupt(gatewayID).Return(natsErr).Times(1)
			},
		})

		_, err := service.InterruptGateway(cmd)
		if !errors.Is(err, natsErr) {
			t.Fatalf("expected %v, got %v", natsErr, err)
		}
	})

	t.Run("Success: imposta stato inactive e salva", func(t *testing.T) {
		cmd := gateway.InterruptGatewayCommand{GatewayId: gatewayID, Requester: requesterSuperAdmin()}

		service := setupGatewayCommandService(t, []mockSetupFuncGatewayCommandService{
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.getGatewayPort.EXPECT().GetById(gatewayID).Return(activeGateway, nil).Times(1)
			},
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.gatewayCommandPort.EXPECT().SendInterrupt(gatewayID).Return(nil).Times(1)
			},
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.saveGatewayPort.EXPECT().Save(gomock.Any()).DoAndReturn(func(g gateway.Gateway) (gateway.Gateway, error) {
					if g.Status != gateway.GATEWAY_STATUS_INACTIVE {
						t.Fatalf("expected status %s, got %s", gateway.GATEWAY_STATUS_INACTIVE, g.Status)
					}
					return g, nil
				}).Times(1)
			},
		})

		updated, err := service.InterruptGateway(cmd)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if updated.Status != gateway.GATEWAY_STATUS_INACTIVE {
			t.Fatalf("expected %s, got %s", gateway.GATEWAY_STATUS_INACTIVE, updated.Status)
		}
	})
}

func TestService_ResumeGateway(t *testing.T) {
	gatewayID := uuid.New()
	tenantID := uuid.New()

	inactiveGateway := gateway.Gateway{Id: gatewayID, Name: "GW-6", TenantId: &tenantID, Status: gateway.GATEWAY_STATUS_INACTIVE}
	activeGateway := gateway.Gateway{Id: gatewayID, Name: "GW-6", TenantId: &tenantID, Status: gateway.GATEWAY_STATUS_ACTIVE}

	t.Run("Fail: gateway non inattivo", func(t *testing.T) {
		cmd := gateway.ResumeGatewayCommand{GatewayId: gatewayID, Requester: requesterSuperAdmin()}

		service := setupGatewayCommandService(t, []mockSetupFuncGatewayCommandService{
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.getGatewayPort.EXPECT().GetById(gatewayID).Return(activeGateway, nil).Times(1)
			},
			func(m gatewayCommandServiceMocks) *gomock.Call {
				m.gatewayCommandPort.EXPECT().SendResume(gomock.Any()).Times(0)
				m.saveGatewayPort.EXPECT().Save(gomock.Any()).Times(0)
				return nil
			},
		})

		_, err := service.ResumeGateway(cmd)
		if !errors.Is(err, gateway.ErrGatewayNotInactive) {
			t.Fatalf("expected %v, got %v", gateway.ErrGatewayNotInactive, err)
		}
	})

	t.Run("Fail: errore comunicazione NATS", func(t *testing.T) {
		cmd := gateway.ResumeGatewayCommand{GatewayId: gatewayID, Requester: requesterSuperAdmin()}
		natsErr := errors.New("nats resume failed")

		service := setupGatewayCommandService(t, []mockSetupFuncGatewayCommandService{
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.getGatewayPort.EXPECT().GetById(gatewayID).Return(inactiveGateway, nil).Times(1)
			},
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.gatewayCommandPort.EXPECT().SendResume(gatewayID).Return(natsErr).Times(1)
			},
		})

		_, err := service.ResumeGateway(cmd)
		if !errors.Is(err, natsErr) {
			t.Fatalf("expected %v, got %v", natsErr, err)
		}
	})

	t.Run("Success: imposta stato active e salva", func(t *testing.T) {
		cmd := gateway.ResumeGatewayCommand{GatewayId: gatewayID, Requester: requesterSuperAdmin()}

		service := setupGatewayCommandService(t, []mockSetupFuncGatewayCommandService{
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.getGatewayPort.EXPECT().GetById(gatewayID).Return(inactiveGateway, nil).Times(1)
			},
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.gatewayCommandPort.EXPECT().SendResume(gatewayID).Return(nil).Times(1)
			},
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.saveGatewayPort.EXPECT().Save(gomock.Any()).DoAndReturn(func(g gateway.Gateway) (gateway.Gateway, error) {
					if g.Status != gateway.GATEWAY_STATUS_ACTIVE {
						t.Fatalf("expected status %s, got %s", gateway.GATEWAY_STATUS_ACTIVE, g.Status)
					}
					return g, nil
				}).Times(1)
			},
		})

		updated, err := service.ResumeGateway(cmd)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if updated.Status != gateway.GATEWAY_STATUS_ACTIVE {
			t.Fatalf("expected %s, got %s", gateway.GATEWAY_STATUS_ACTIVE, updated.Status)
		}
	})
}

func TestService_CreateGateway(t *testing.T) {
	t.Run("Fail: utente non autorizzato", func(t *testing.T) {
		cmd := gateway.CreateGatewayCommand{
			Name:      "GW-7",
			Interval:  3 * time.Second,
			Requester: requesterTenantAdmin(nil),
		}

		service := setupGatewayCommandService(t, []mockSetupFuncGatewayCommandService{
			func(m gatewayCommandServiceMocks) *gomock.Call {
				m.gatewayCommandPort.EXPECT().SendCreateGateway(gomock.Any(), gomock.Any()).Times(0)
				m.createGatewayPort.EXPECT().Create(gomock.Any()).Times(0)
				return nil
			},
		})

		_, err := service.CreateGateway(cmd)
		if !errors.Is(err, identity.ErrUnauthorizedAccess) {
			t.Fatalf("expected %v, got %v", identity.ErrUnauthorizedAccess, err)
		}
	})

	t.Run("Fail: errore comunicazione NATS", func(t *testing.T) {
		cmd := gateway.CreateGatewayCommand{
			Name:      "GW-7",
			Interval:  3 * time.Second,
			Requester: requesterSuperAdmin(),
		}
		natsErr := errors.New("nats create failed")

		service := setupGatewayCommandService(t, []mockSetupFuncGatewayCommandService{
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.gatewayCommandPort.EXPECT().SendCreateGateway(gomock.Not(gomock.Eq(uuid.Nil)), int64(3000)).Return(natsErr).Times(1)
			},
			func(m gatewayCommandServiceMocks) *gomock.Call {
				m.createGatewayPort.EXPECT().Create(gomock.Any()).Times(0)
				return nil
			},
		})

		_, err := service.CreateGateway(cmd)
		if !errors.Is(err, natsErr) {
			t.Fatalf("expected %v, got %v", natsErr, err)
		}
	})

	t.Run("Success: un gateway nuovo nasce decommissioned", func(t *testing.T) {
		cmd := gateway.CreateGatewayCommand{
			Name:      "GW-7",
			Interval:  3 * time.Second,
			Requester: requesterSuperAdmin(),
		}

		service := setupGatewayCommandService(t, []mockSetupFuncGatewayCommandService{
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.gatewayCommandPort.EXPECT().SendCreateGateway(gomock.Not(gomock.Eq(uuid.Nil)), int64(3000)).Return(nil).Times(1)
			},
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.createGatewayPort.EXPECT().Create(gomock.Any()).DoAndReturn(func(g gateway.Gateway) (gateway.Gateway, error) {
					if g.Status != gateway.GATEWAY_STATUS_DECOMMISSIONED {
						t.Fatalf("expected status %s for newly created gateway, got %s", gateway.GATEWAY_STATUS_DECOMMISSIONED, g.Status)
					}
					return g, nil
				}).Times(1)
			},
		})

		_, _ = service.CreateGateway(cmd)
	})
}

func TestService_DeleteGateway(t *testing.T) {
	gatewayID := uuid.New()
	oldGateway := gateway.Gateway{Id: gatewayID, Name: "GW-8", Status: gateway.GATEWAY_STATUS_INACTIVE}

	t.Run("Fail: utente non autorizzato", func(t *testing.T) {
		cmd := gateway.DeleteGatewayCommand{GatewayId: gatewayID, Requester: requesterTenantAdmin(nil)}

		service := setupGatewayCommandService(t, []mockSetupFuncGatewayCommandService{
			func(m gatewayCommandServiceMocks) *gomock.Call {
				m.getGatewayPort.EXPECT().GetById(gomock.Any()).Times(0)
				m.gatewayCommandPort.EXPECT().SendDeleteGateway(gomock.Any()).Times(0)
				m.deleteGatewayPort.EXPECT().Delete(gomock.Any()).Times(0)
				return nil
			},
		})

		_, err := service.DeleteGateway(cmd)
		if !errors.Is(err, identity.ErrUnauthorizedAccess) {
			t.Fatalf("expected %v, got %v", identity.ErrUnauthorizedAccess, err)
		}
	})

	t.Run("Fail: gateway non trovato", func(t *testing.T) {
		cmd := gateway.DeleteGatewayCommand{GatewayId: gatewayID, Requester: requesterSuperAdmin()}

		service := setupGatewayCommandService(t, []mockSetupFuncGatewayCommandService{
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.getGatewayPort.EXPECT().GetById(gatewayID).Return(gateway.Gateway{}, nil).Times(1)
			},
			func(m gatewayCommandServiceMocks) *gomock.Call {
				m.gatewayCommandPort.EXPECT().SendDeleteGateway(gomock.Any()).Times(0)
				m.deleteGatewayPort.EXPECT().Delete(gomock.Any()).Times(0)
				return nil
			},
		})

		_, err := service.DeleteGateway(cmd)
		if !errors.Is(err, gateway.ErrGatewayNotFound) {
			t.Fatalf("expected %v, got %v", gateway.ErrGatewayNotFound, err)
		}
	})

	t.Run("Success: invia delete command e rimuove dal DB", func(t *testing.T) {
		cmd := gateway.DeleteGatewayCommand{GatewayId: gatewayID, Requester: requesterSuperAdmin()}

		service := setupGatewayCommandService(t, []mockSetupFuncGatewayCommandService{
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.getGatewayPort.EXPECT().GetById(gatewayID).Return(oldGateway, nil).Times(1)
			},
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.gatewayCommandPort.EXPECT().SendDeleteGateway(gatewayID).Return(nil).Times(1)
			},
			func(m gatewayCommandServiceMocks) *gomock.Call {
				return m.deleteGatewayPort.EXPECT().Delete(gatewayID).Return(oldGateway, nil).Times(1)
			},
		})

		deleted, err := service.DeleteGateway(cmd)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if deleted.Id != gatewayID {
			t.Fatalf("expected deleted id %s, got %s", gatewayID, deleted.Id)
		}
	})
}
