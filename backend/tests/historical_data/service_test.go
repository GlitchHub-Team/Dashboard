package historical_data_test

import (
	"errors"
	"testing"
	"time"

	"backend/internal/historical_data"
	"backend/internal/shared/identity"
	"backend/internal/tenant"
	"backend/tests/historical_data/mocks"
	tenantmocks "backend/tests/tenant/mocks"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

type mockSetupFunc_GetHistoricalDataService func(
	getHistoricalDataPort *mocks.MockGetHistoricalDataPort,
	getTenantPort *tenantmocks.MockGetTenantPort,
) *gomock.Call

func newStepTenantOk_GetHistoricalDataService(
	targetTenantId uuid.UUID, canImpersonate bool,
) mockSetupFunc_GetHistoricalDataService {
	return func(
		getHistoricalDataPort *mocks.MockGetHistoricalDataPort,
		getTenantPort *tenantmocks.MockGetTenantPort,
	) *gomock.Call {
		return getTenantPort.EXPECT().
			GetTenant(targetTenantId).
			Return(tenant.Tenant{
				Id:             targetTenantId,
				CanImpersonate: canImpersonate,
			}, nil).
			Times(1)
	}
}

func newStepTenantNotFound_GetHistoricalDataService(
	targetTenantId uuid.UUID,
) mockSetupFunc_GetHistoricalDataService {
	return func(
		getHistoricalDataPort *mocks.MockGetHistoricalDataPort,
		getTenantPort *tenantmocks.MockGetTenantPort,
	) *gomock.Call {
		return getTenantPort.EXPECT().
			GetTenant(targetTenantId).
			Return(tenant.Tenant{}, nil).
			Times(1)
	}
}

func newStepTenantError_GetHistoricalDataService(
	targetTenantId uuid.UUID,
	mockError error,
) mockSetupFunc_GetHistoricalDataService {
	return func(
		getHistoricalDataPort *mocks.MockGetHistoricalDataPort,
		getTenantPort *tenantmocks.MockGetTenantPort,
	) *gomock.Call {
		return getTenantPort.EXPECT().
			GetTenant(targetTenantId).
			Return(tenant.Tenant{}, mockError).
			Times(1)
	}
}

func TestService_GetSensorHistoricalData(t *testing.T) {
	targetTenantId := uuid.New()
	otherTenantId := uuid.New()
	targetSensorId := uuid.New()
	expectedSamples := []historical_data.HistoricalSample{
		{
			SensorId: targetSensorId,
			TenantId: targetTenantId,
			Profile:  "HeartRate",
		},
	}

	type testCase struct {
		name          string
		input         historical_data.GetSensorHistoricalDataCommand
		setupSteps    []mockSetupFunc_GetHistoricalDataService
		expectedData  []historical_data.HistoricalSample
		expectedError error
	}

	step1TenantOk_CanImpersonate := newStepTenantOk_GetHistoricalDataService(targetTenantId, true)
	step1TenantOk_CannotImpersonate := newStepTenantOk_GetHistoricalDataService(targetTenantId, false)
	step1TenantNotFound := newStepTenantNotFound_GetHistoricalDataService(targetTenantId)

	errMockStep1 := newMockError(1)
	step1TenantError := newStepTenantError_GetHistoricalDataService(targetTenantId, errMockStep1)

	expectedFilter := historical_data.HistoricalDataFilter{
		Limit: historical_data.DefaultHistoricalDataLimit,
	}

	step2GetHistoricalDataOk := func(
		getHistoricalDataPort *mocks.MockGetHistoricalDataPort,
		getTenantPort *tenantmocks.MockGetTenantPort,
	) *gomock.Call {
		return getHistoricalDataPort.EXPECT().
			GetSensorHistoricalData(targetTenantId, targetSensorId, expectedFilter).
			Return(expectedSamples, nil).
			Times(1)
	}
	errMockStep2 := newMockError(2)
	step2GetHistoricalDataError := func(
		getHistoricalDataPort *mocks.MockGetHistoricalDataPort,
		getTenantPort *tenantmocks.MockGetTenantPort,
	) *gomock.Call {
		return getHistoricalDataPort.EXPECT().
			GetSensorHistoricalData(targetTenantId, targetSensorId, expectedFilter).
			Return(nil, errMockStep2).
			Times(1)
	}
	step2NeverCalled := func(
		getHistoricalDataPort *mocks.MockGetHistoricalDataPort,
		getTenantPort *tenantmocks.MockGetTenantPort,
	) *gomock.Call {
		return getHistoricalDataPort.EXPECT().
			GetSensorHistoricalData(gomock.Any(), gomock.Any(), gomock.Any()).
			Times(0)
	}

	superAdminRequester := identity.Requester{
		RequesterUserId: uint(1),
		RequesterRole:   identity.ROLE_SUPER_ADMIN,
	}
	authorizedTenantAdminRequester := identity.Requester{
		RequesterUserId:   uint(2),
		RequesterTenantId: &targetTenantId,
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}
	unauthorizedTenantAdminRequester := identity.Requester{
		RequesterUserId:   uint(2),
		RequesterTenantId: &otherTenantId,
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}
	authorizedTenantUserRequester := identity.Requester{
		RequesterUserId:   uint(3),
		RequesterTenantId: &targetTenantId,
		RequesterRole:     identity.ROLE_TENANT_USER,
	}
	unauthorizedTenantUserRequester := identity.Requester{
		RequesterUserId:   uint(3),
		RequesterTenantId: &otherTenantId,
		RequesterRole:     identity.ROLE_TENANT_USER,
	}

	baseInput := historical_data.GetSensorHistoricalDataCommand{
		TenantId: targetTenantId,
		SensorId: targetSensorId,
	}

	inputWith := func(requester identity.Requester) historical_data.GetSensorHistoricalDataCommand {
		return historical_data.GetSensorHistoricalDataCommand{
			Requester: requester,
			TenantId:  baseInput.TenantId,
			SensorId:  baseInput.SensorId,
		}
	}

	from := time.Now().UTC()
	to := from.Add(-time.Minute)

	cases := []testCase{
		{
			name:  "(Super Admin) Success: impersonation ok",
			input: inputWith(superAdminRequester),
			setupSteps: []mockSetupFunc_GetHistoricalDataService{
				step1TenantOk_CanImpersonate,
				step2GetHistoricalDataOk,
			},
			expectedData:  expectedSamples,
			expectedError: nil,
		},
		{
			name:  "(Tenant Admin) Success: authorization ok",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_GetHistoricalDataService{
				step1TenantOk_CanImpersonate,
				step2GetHistoricalDataOk,
			},
			expectedData:  expectedSamples,
			expectedError: nil,
		},
		{
			name:  "(Tenant User) Success: authorization ok",
			input: inputWith(authorizedTenantUserRequester),
			setupSteps: []mockSetupFunc_GetHistoricalDataService{
				step1TenantOk_CanImpersonate,
				step2GetHistoricalDataOk,
			},
			expectedData:  expectedSamples,
			expectedError: nil,
		},
		{
			name: "Fail: invalid date range",
			input: historical_data.GetSensorHistoricalDataCommand{
				Requester: authorizedTenantAdminRequester,
				TenantId:  targetTenantId,
				SensorId:  targetSensorId,
				From:      &from,
				To:        &to,
			},
			setupSteps: []mockSetupFunc_GetHistoricalDataService{
				step2NeverCalled,
			},
			expectedData:  nil,
			expectedError: historical_data.ErrInvalidDateRange,
		},
		{
			name:  "Fail (step 1): tenant not found",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_GetHistoricalDataService{
				step1TenantNotFound,
				step2NeverCalled,
			},
			expectedData:  nil,
			expectedError: tenant.ErrTenantNotFound,
		},
		{
			name:  "Fail (step 1): unexpected tenant error",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_GetHistoricalDataService{
				step1TenantError,
				step2NeverCalled,
			},
			expectedData:  nil,
			expectedError: errMockStep1,
		},
		{
			name:  "(Super Admin) Fail: impersonation forbidden",
			input: inputWith(superAdminRequester),
			setupSteps: []mockSetupFunc_GetHistoricalDataService{
				step1TenantOk_CannotImpersonate,
				step2NeverCalled,
			},
			expectedData:  nil,
			expectedError: identity.ErrUnauthorizedAccess,
		},
		{
			name:  "(Tenant Admin) Fail: wrong tenant",
			input: inputWith(unauthorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_GetHistoricalDataService{
				step1TenantOk_CanImpersonate,
				step2NeverCalled,
			},
			expectedData:  nil,
			expectedError: identity.ErrUnauthorizedAccess,
		},
		{
			name:  "(Tenant User) Fail: wrong tenant",
			input: inputWith(unauthorizedTenantUserRequester),
			setupSteps: []mockSetupFunc_GetHistoricalDataService{
				step1TenantOk_CanImpersonate,
				step2NeverCalled,
			},
			expectedData:  nil,
			expectedError: identity.ErrUnauthorizedAccess,
		},
		{
			name:  "Fail (step 2): historical data port error",
			input: inputWith(authorizedTenantAdminRequester),
			setupSteps: []mockSetupFunc_GetHistoricalDataService{
				step1TenantOk_CanImpersonate,
				step2GetHistoricalDataError,
			},
			expectedData:  nil,
			expectedError: errMockStep2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			getHistoricalDataPort := mocks.NewMockGetHistoricalDataPort(ctrl)
			getTenantPort := tenantmocks.NewMockGetTenantPort(ctrl)

			var expectedCalls []any
			for _, step := range tc.setupSteps {
				if call := step(getHistoricalDataPort, getTenantPort); call != nil {
					expectedCalls = append(expectedCalls, call)
				}
			}
			if len(expectedCalls) > 0 {
				gomock.InOrder(expectedCalls...)
			}

			service := historical_data.NewGetHistoricalDataService(getHistoricalDataPort, getTenantPort)

			data, err := service.GetSensorHistoricalData(tc.input)
			if !errors.Is(err, tc.expectedError) {
				t.Fatalf("expected error %v, got %v", tc.expectedError, err)
			}
			if tc.expectedError == nil && len(data) != len(tc.expectedData) {
				t.Fatalf("expected %d samples, got %d", len(tc.expectedData), len(data))
			}
		})
	}
}
