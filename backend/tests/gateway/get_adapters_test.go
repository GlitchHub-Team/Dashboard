package gateway_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"backend/internal/gateway"
	"backend/internal/infra/database/pagination"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestGatewayPostgreAdapter_GetById(t *testing.T) {
	gatewayID := uuid.New()
	tenantID := uuid.New()
	publicID := "public-gw-1"

	entity := gateway.GatewayEntity{
		ID:               gatewayID.String(),
		Name:             "GW-GET-BY-ID",
		TenantId:         ptrString(tenantID.String()),
		Interval:         4500,
		Status:           string(gateway.GATEWAY_STATUS_ACTIVE),
		PublicIdentifier: &publicID,
	}

	repoErr := errors.New("get by id failed")

	stepGetByIdOk := func(m gatewayAdapterMocks) *gomock.Call {
		return m.repo.EXPECT().GetGatewayById(gatewayID.String()).Return(entity, nil).Times(1)
	}
	stepGetByIdErr := func(m gatewayAdapterMocks) *gomock.Call {
		return m.repo.EXPECT().GetGatewayById(gatewayID.String()).Return(gateway.GatewayEntity{}, repoErr).Times(1)
	}
	stepGetByIdInvalidEntity := func(m gatewayAdapterMocks) *gomock.Call {
		invalid := gateway.GatewayEntity{ID: "invalid-uuid"}
		return m.repo.EXPECT().GetGatewayById(gatewayID.String()).Return(invalid, nil).Times(1)
	}

	type testCase struct {
		name           string
		setupSteps     []mockSetupFuncGatewayAdapter
		expectedValue  gateway.Gateway
		expectedError  error
		expectAnyError bool
	}

	cases := []testCase{
		{
			name:       "Success",
			setupSteps: []mockSetupFuncGatewayAdapter{stepGetByIdOk},
			expectedValue: gateway.Gateway{
				Id:               gatewayID,
				Name:             entity.Name,
				TenantId:         &tenantID,
				IntervalLimit:    4500 * time.Millisecond,
				Status:           gateway.GATEWAY_STATUS_ACTIVE,
				PublicIdentifier: &publicID,
			},
		},
		{
			name:          "Fail: repository error",
			setupSteps:    []mockSetupFuncGatewayAdapter{stepGetByIdErr},
			expectedError: repoErr,
		},
		{
			name:           "Fail: mapping error",
			setupSteps:     []mockSetupFuncGatewayAdapter{stepGetByIdInvalidEntity},
			expectAnyError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			adapter := setupGatewayPostgreAdapter(t, tc.setupSteps)

			found, err := adapter.GetById(gatewayID)

			if tc.expectAnyError {
				if err == nil {
					t.Fatalf("expected an error, got nil")
				}
				if !found.IsZero() {
					t.Fatalf("expected zero gateway on error, got %+v", found)
				}
				return
			}

			if tc.expectedError != nil {
				if !errors.Is(err, tc.expectedError) {
					t.Fatalf("expected error %v, got %v", tc.expectedError, err)
				}
				if !found.IsZero() {
					t.Fatalf("expected zero gateway on error, got %+v", found)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if !reflect.DeepEqual(tc.expectedValue, found) {
				t.Fatalf("unexpected gateway. expected %+v, got %+v", tc.expectedValue, found)
			}
		})
	}
}

func TestGatewayPostgreAdapter_GetGatewayByTenantID(t *testing.T) {
	tenantID := uuid.New()
	gatewayID := uuid.New()
	publicID := "public-gw-2"

	entity := gateway.GatewayEntity{
		ID:               gatewayID.String(),
		Name:             "GW-GET-BY-TENANT",
		TenantId:         ptrString(tenantID.String()),
		Interval:         2000,
		Status:           string(gateway.GATEWAY_STATUS_DECOMMISSIONED),
		PublicIdentifier: &publicID,
	}

	repoErr := errors.New("get by tenant and id failed")

	stepGetByTenantAndIdOk := func(m gatewayAdapterMocks) *gomock.Call {
		return m.repo.EXPECT().GetGatewayByTenantID(tenantID.String(), gatewayID.String()).Return(entity, nil).Times(1)
	}
	stepGetByTenantAndIdErr := func(m gatewayAdapterMocks) *gomock.Call {
		return m.repo.EXPECT().GetGatewayByTenantID(tenantID.String(), gatewayID.String()).Return(gateway.GatewayEntity{}, repoErr).Times(1)
	}
	stepGetByTenantAndIdInvalid := func(m gatewayAdapterMocks) *gomock.Call {
		invalid := gateway.GatewayEntity{ID: "invalid-uuid"}
		return m.repo.EXPECT().GetGatewayByTenantID(tenantID.String(), gatewayID.String()).Return(invalid, nil).Times(1)
	}

	type testCase struct {
		name           string
		setupSteps     []mockSetupFuncGatewayAdapter
		expectedValue  gateway.Gateway
		expectedError  error
		expectAnyError bool
	}

	cases := []testCase{
		{
			name:       "Success",
			setupSteps: []mockSetupFuncGatewayAdapter{stepGetByTenantAndIdOk},
			expectedValue: gateway.Gateway{
				Id:               gatewayID,
				Name:             entity.Name,
				TenantId:         &tenantID,
				IntervalLimit:    2 * time.Second,
				Status:           gateway.GATEWAY_STATUS_DECOMMISSIONED,
				PublicIdentifier: &publicID,
			},
		},
		{
			name:          "Fail: repository error",
			setupSteps:    []mockSetupFuncGatewayAdapter{stepGetByTenantAndIdErr},
			expectedError: repoErr,
		},
		{
			name:           "Fail: mapping error",
			setupSteps:     []mockSetupFuncGatewayAdapter{stepGetByTenantAndIdInvalid},
			expectAnyError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			adapter := setupGatewayPostgreAdapter(t, tc.setupSteps)

			found, err := adapter.GetGatewayByTenantID(tenantID, gatewayID)

			if tc.expectAnyError {
				if err == nil {
					t.Fatalf("expected an error, got nil")
				}
				if !found.IsZero() {
					t.Fatalf("expected zero gateway on error, got %+v", found)
				}
				return
			}

			if tc.expectedError != nil {
				if !errors.Is(err, tc.expectedError) {
					t.Fatalf("expected error %v, got %v", tc.expectedError, err)
				}
				if !found.IsZero() {
					t.Fatalf("expected zero gateway on error, got %+v", found)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if !reflect.DeepEqual(tc.expectedValue, found) {
				t.Fatalf("unexpected gateway. expected %+v, got %+v", tc.expectedValue, found)
			}
		})
	}
}

func TestGatewayPostgreAdapter_GetByTenantId(t *testing.T) {
	tenantID := uuid.New()
	gatewayID1 := uuid.New()
	gatewayID2 := uuid.New()

	entityList := []gateway.GatewayEntity{
		{
			ID:       gatewayID1.String(),
			Name:     "GW-A",
			TenantId: ptrString(tenantID.String()),
			Interval: 1000,
			Status:   string(gateway.GATEWAY_STATUS_ACTIVE),
		},
		{
			ID:       gatewayID2.String(),
			Name:     "GW-B",
			TenantId: ptrString(tenantID.String()),
			Interval: 5000,
			Status:   string(gateway.GATEWAY_STATUS_INACTIVE),
		},
	}

	expectedList := []gateway.Gateway{
		{Id: gatewayID1, Name: "GW-A", TenantId: &tenantID, IntervalLimit: 1 * time.Second, Status: gateway.GATEWAY_STATUS_ACTIVE},
		{Id: gatewayID2, Name: "GW-B", TenantId: &tenantID, IntervalLimit: 5 * time.Second, Status: gateway.GATEWAY_STATUS_INACTIVE},
	}

	repoErr := errors.New("get by tenant failed")

	stepRepoNeverCalled := func(m gatewayAdapterMocks) *gomock.Call {
		return m.repo.EXPECT().GetGatewaysByTenantId(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	}
	stepGetByTenantErr := func(m gatewayAdapterMocks) *gomock.Call {
		return m.repo.EXPECT().GetGatewaysByTenantId(tenantID.String(), 10, 5).Return(nil, uint(0), repoErr).Times(1)
	}
	stepGetByTenantOk := func(m gatewayAdapterMocks) *gomock.Call {
		return m.repo.EXPECT().GetGatewaysByTenantId(tenantID.String(), 5, 5).Return(entityList, uint(2), nil).Times(1)
	}
	stepGetByTenantMappingErr := func(m gatewayAdapterMocks) *gomock.Call {
		invalid := []gateway.GatewayEntity{{ID: "invalid-uuid"}}
		return m.repo.EXPECT().GetGatewaysByTenantId(tenantID.String(), 5, 5).Return(invalid, uint(1), nil).Times(1)
	}

	type testCase struct {
		name           string
		page           int
		limit          int
		setupSteps     []mockSetupFuncGatewayAdapter
		expectedList   []gateway.Gateway
		expectedTotal  uint
		expectedError  error
		expectAnyError bool
	}

	cases := []testCase{
		{
			name:  "Fail: invalid page",
			page:  0,
			limit: 5,
			setupSteps: []mockSetupFuncGatewayAdapter{
				stepRepoNeverCalled,
			},
			expectedError: pagination.ErrInvalidPage,
		},
		{
			name:  "Fail: invalid limit",
			page:  1,
			limit: 0,
			setupSteps: []mockSetupFuncGatewayAdapter{
				stepRepoNeverCalled,
			},
			expectedError: pagination.ErrInvalidLimit,
		},
		{
			name:  "Fail: repository error",
			page:  3,
			limit: 5,
			setupSteps: []mockSetupFuncGatewayAdapter{
				stepGetByTenantErr,
			},
			expectedError: repoErr,
		},
		{
			name:  "Success",
			page:  2,
			limit: 5,
			setupSteps: []mockSetupFuncGatewayAdapter{
				stepGetByTenantOk,
			},
			expectedList:  expectedList,
			expectedTotal: uint(2),
		},
		{
			name:  "Fail: mapping error",
			page:  2,
			limit: 5,
			setupSteps: []mockSetupFuncGatewayAdapter{
				stepGetByTenantMappingErr,
			},
			expectAnyError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			adapter := setupGatewayPostgreAdapter(t, tc.setupSteps)

			found, total, err := adapter.GetByTenantId(tenantID, tc.page, tc.limit)

			if tc.expectAnyError {
				if err == nil {
					t.Fatalf("expected an error, got nil")
				}
				if len(found) != 0 {
					t.Fatalf("expected empty list on error, got %+v", found)
				}
				if total != 0 {
					t.Fatalf("expected total 0 on error, got %d", total)
				}
				return
			}

			if tc.expectedError != nil {
				if !errors.Is(err, tc.expectedError) {
					t.Fatalf("expected error %v, got %v", tc.expectedError, err)
				}
				if found != nil {
					t.Fatalf("expected nil list on error, got %+v", found)
				}
				if total != 0 {
					t.Fatalf("expected total 0 on error, got %d", total)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if !reflect.DeepEqual(tc.expectedList, found) {
				t.Fatalf("unexpected list. expected %+v, got %+v", tc.expectedList, found)
			}
			if tc.expectedTotal != total {
				t.Fatalf("unexpected total. expected %d, got %d", tc.expectedTotal, total)
			}
		})
	}
}

func TestGatewayPostgreAdapter_GetAll(t *testing.T) {
	tenantID := uuid.New()
	gatewayID := uuid.New()

	entityList := []gateway.GatewayEntity{
		{
			ID:       gatewayID.String(),
			Name:     "GW-ALL",
			TenantId: ptrString(tenantID.String()),
			Interval: 3000,
			Status:   string(gateway.GATEWAY_STATUS_ACTIVE),
		},
	}

	expectedList := []gateway.Gateway{
		{Id: gatewayID, Name: "GW-ALL", TenantId: &tenantID, IntervalLimit: 3 * time.Second, Status: gateway.GATEWAY_STATUS_ACTIVE},
	}

	repoErr := errors.New("get all failed")

	stepRepoNeverCalled := func(m gatewayAdapterMocks) *gomock.Call {
		return m.repo.EXPECT().GetAllGateways(gomock.Any(), gomock.Any()).Times(0)
	}
	stepGetAllErr := func(m gatewayAdapterMocks) *gomock.Call {
		return m.repo.EXPECT().GetAllGateways(10, 5).Return(nil, uint(0), repoErr).Times(1)
	}
	stepGetAllOk := func(m gatewayAdapterMocks) *gomock.Call {
		return m.repo.EXPECT().GetAllGateways(5, 5).Return(entityList, uint(1), nil).Times(1)
	}
	stepGetAllMappingErr := func(m gatewayAdapterMocks) *gomock.Call {
		invalid := []gateway.GatewayEntity{{ID: "invalid-uuid"}}
		return m.repo.EXPECT().GetAllGateways(5, 5).Return(invalid, uint(1), nil).Times(1)
	}

	type testCase struct {
		name           string
		page           int
		limit          int
		setupSteps     []mockSetupFuncGatewayAdapter
		expectedList   []gateway.Gateway
		expectedTotal  uint
		expectedError  error
		expectAnyError bool
	}

	cases := []testCase{
		{
			name:  "Fail: invalid page",
			page:  0,
			limit: 5,
			setupSteps: []mockSetupFuncGatewayAdapter{
				stepRepoNeverCalled,
			},
			expectedError: pagination.ErrInvalidPage,
		},
		{
			name:  "Fail: invalid limit",
			page:  1,
			limit: 0,
			setupSteps: []mockSetupFuncGatewayAdapter{
				stepRepoNeverCalled,
			},
			expectedError: pagination.ErrInvalidLimit,
		},
		{
			name:  "Fail: repository error",
			page:  3,
			limit: 5,
			setupSteps: []mockSetupFuncGatewayAdapter{
				stepGetAllErr,
			},
			expectedError: repoErr,
		},
		{
			name:  "Success",
			page:  2,
			limit: 5,
			setupSteps: []mockSetupFuncGatewayAdapter{
				stepGetAllOk,
			},
			expectedList:  expectedList,
			expectedTotal: uint(1),
		},
		{
			name:  "Fail: mapping error",
			page:  2,
			limit: 5,
			setupSteps: []mockSetupFuncGatewayAdapter{
				stepGetAllMappingErr,
			},
			expectAnyError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			adapter := setupGatewayPostgreAdapter(t, tc.setupSteps)

			found, total, err := adapter.GetAll(tc.page, tc.limit)

			if tc.expectAnyError {
				if err == nil {
					t.Fatalf("expected an error, got nil")
				}
				if len(found) != 0 {
					t.Fatalf("expected empty list on error, got %+v", found)
				}
				if total != 0 {
					t.Fatalf("expected total 0 on error, got %d", total)
				}
				return
			}

			if tc.expectedError != nil {
				if !errors.Is(err, tc.expectedError) {
					t.Fatalf("expected error %v, got %v", tc.expectedError, err)
				}
				if found != nil {
					t.Fatalf("expected nil list on error, got %+v", found)
				}
				if total != 0 {
					t.Fatalf("expected total 0 on error, got %d", total)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if !reflect.DeepEqual(tc.expectedList, found) {
				t.Fatalf("unexpected list. expected %+v, got %+v", tc.expectedList, found)
			}
			if tc.expectedTotal != total {
				t.Fatalf("unexpected total. expected %d, got %d", tc.expectedTotal, total)
			}
		})
	}
}

func ptrString(v string) *string {
	return &v
}
