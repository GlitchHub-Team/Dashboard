package gateway_test

import (
	"errors"
	"testing"
	"time"

	"backend/internal/gateway"
	mocks "backend/tests/gateway/mocks"
	helper "backend/tests/helper"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type gatewayAdapterMocks struct {
	repo *mocks.MockGatewayRepository
}

type mockSetupFuncGatewayAdapter = helper.AdapterMockSetupFunc[gatewayAdapterMocks]

func setupGatewayPostgreAdapter(
	t *testing.T,
	setupSteps []mockSetupFuncGatewayAdapter,
) *gateway.GatewayPostgreAdapter {
	t.Helper()

	return helper.SetupAdapterWithOrderedSteps(
		t,
		func(ctrl *gomock.Controller) gatewayAdapterMocks {
			return gatewayAdapterMocks{repo: mocks.NewMockGatewayRepository(ctrl)}
		},
		setupSteps,
		func(mockBundle gatewayAdapterMocks) *gateway.GatewayPostgreAdapter {
			return gateway.NewGatewayPostgreAdapter(mockBundle.repo, zap.NewNop())
		},
	)
}

func TestGatewayPostgreAdapter_Save(t *testing.T) {
	gatewayID := uuid.New()
	tenantID := uuid.New()
	publicID := "pub-gw"

	input := gateway.Gateway{
		Id:               gatewayID,
		Name:             "GW-SAVE",
		TenantId:         &tenantID,
		Status:           gateway.GATEWAY_STATUS_ACTIVE,
		IntervalLimit:    4 * time.Second,
		PublicIdentifier: &publicID,
	}

	t.Run("Success", func(t *testing.T) {
		adapter := setupGatewayPostgreAdapter(t, []mockSetupFuncGatewayAdapter{
			func(m gatewayAdapterMocks) *gomock.Call {
				return m.repo.EXPECT().SaveGateway(gomock.AssignableToTypeOf(&gateway.GatewayEntity{})).DoAndReturn(func(entity *gateway.GatewayEntity) error {
					if entity.ID != gatewayID.String() {
						t.Fatalf("expected id %s, got %s", gatewayID, entity.ID)
					}
					if entity.Name != input.Name {
						t.Fatalf("expected name %s, got %s", input.Name, entity.Name)
					}
					if entity.Status != string(input.Status) {
						t.Fatalf("expected status %s, got %s", input.Status, entity.Status)
					}
					if entity.Interval != input.IntervalLimit.Milliseconds() {
						t.Fatalf("expected interval %d, got %d", input.IntervalLimit.Milliseconds(), entity.Interval)
					}
					if entity.TenantId == nil || *entity.TenantId != tenantID.String() {
						t.Fatalf("expected tenant id %s", tenantID)
					}
					if entity.PublicIdentifier == nil || *entity.PublicIdentifier != publicID {
						t.Fatalf("expected public identifier %s", publicID)
					}
					return nil
				}).Times(1)
			},
		})

		saved, err := adapter.Save(input)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if saved.Id != input.Id {
			t.Fatalf("expected saved id %s, got %s", input.Id, saved.Id)
		}
	})

	t.Run("Fail: repository error", func(t *testing.T) {
		repoErr := errors.New("save failed")
		adapter := setupGatewayPostgreAdapter(t, []mockSetupFuncGatewayAdapter{
			func(m gatewayAdapterMocks) *gomock.Call {
				return m.repo.EXPECT().SaveGateway(gomock.AssignableToTypeOf(&gateway.GatewayEntity{})).Return(repoErr).Times(1)
			},
		})

		_, err := adapter.Save(input)
		if !errors.Is(err, repoErr) {
			t.Fatalf("expected %v, got %v", repoErr, err)
		}
	})
}

func TestGatewayPostgreAdapter_Create(t *testing.T) {
	gatewayID := uuid.New()
	tenantID := uuid.New()
	publicID := "pub-create"

	input := gateway.Gateway{
		Id:               gatewayID,
		Name:             "GW-CREATE",
		TenantId:         &tenantID,
		Status:           gateway.GATEWAY_STATUS_DECOMMISSIONED,
		IntervalLimit:    3 * time.Second,
		PublicIdentifier: &publicID,
	}

	t.Run("Success", func(t *testing.T) {
		adapter := setupGatewayPostgreAdapter(t, []mockSetupFuncGatewayAdapter{
			func(m gatewayAdapterMocks) *gomock.Call {
				return m.repo.EXPECT().CreateGateway(gomock.AssignableToTypeOf(&gateway.GatewayEntity{})).DoAndReturn(func(entity *gateway.GatewayEntity) (gateway.Gateway, error) {
					if entity.ID != gatewayID.String() {
						t.Fatalf("expected id %s, got %s", gatewayID, entity.ID)
					}
					if entity.Interval != input.IntervalLimit.Milliseconds() {
						t.Fatalf("expected interval %d, got %d", input.IntervalLimit.Milliseconds(), entity.Interval)
					}
					if entity.TenantId == nil || *entity.TenantId != tenantID.String() {
						t.Fatalf("expected tenant id %s", tenantID)
					}
					return input, nil
				}).Times(1)
			},
		})

		created, err := adapter.Create(input)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if created.Id != input.Id {
			t.Fatalf("expected id %s, got %s", input.Id, created.Id)
		}
	})

	t.Run("Fail: repository error", func(t *testing.T) {
		repoErr := errors.New("create failed")
		adapter := setupGatewayPostgreAdapter(t, []mockSetupFuncGatewayAdapter{
			func(m gatewayAdapterMocks) *gomock.Call {
				return m.repo.EXPECT().CreateGateway(gomock.AssignableToTypeOf(&gateway.GatewayEntity{})).Return(gateway.Gateway{}, repoErr).Times(1)
			},
		})

		_, err := adapter.Create(input)
		if !errors.Is(err, repoErr) {
			t.Fatalf("expected %v, got %v", repoErr, err)
		}
	})
}

func TestGatewayPostgreAdapter_Delete(t *testing.T) {
	gatewayID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		adapter := setupGatewayPostgreAdapter(t, []mockSetupFuncGatewayAdapter{
			func(m gatewayAdapterMocks) *gomock.Call {
				return m.repo.EXPECT().DeleteGateway(gomock.AssignableToTypeOf(&gateway.GatewayEntity{})).DoAndReturn(func(entity *gateway.GatewayEntity) error {
					if entity.ID != gatewayID.String() {
						t.Fatalf("expected id %s, got %s", gatewayID, entity.ID)
					}
					entity.Name = "GW-DELETED"
					entity.Status = string(gateway.GATEWAY_STATUS_DECOMMISSIONED)
					entity.Interval = 1000
					return nil
				}).Times(1)
			},
		})

		deleted, err := adapter.Delete(gatewayID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if deleted.Id != gatewayID {
			t.Fatalf("expected deleted id %s, got %s", gatewayID, deleted.Id)
		}
		if deleted.Name != "GW-DELETED" {
			t.Fatalf("expected deleted name GW-DELETED, got %s", deleted.Name)
		}
	})

	t.Run("Fail: repository error", func(t *testing.T) {
		repoErr := errors.New("delete failed")
		adapter := setupGatewayPostgreAdapter(t, []mockSetupFuncGatewayAdapter{
			func(m gatewayAdapterMocks) *gomock.Call {
				return m.repo.EXPECT().DeleteGateway(gomock.AssignableToTypeOf(&gateway.GatewayEntity{})).Return(repoErr).Times(1)
			},
		})

		_, err := adapter.Delete(gatewayID)
		if !errors.Is(err, repoErr) {
			t.Fatalf("expected %v, got %v", repoErr, err)
		}
	})
}
