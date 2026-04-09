package real_time_data_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"backend/internal/real_time_data"
	"backend/internal/sensor"
	"backend/internal/shared/identity"
	"backend/internal/tenant"
	mocks "backend/tests/real_time_data/mocks"
	sensorMocks "backend/tests/sensor/mocks"
	tenantMocks "backend/tests/tenant/mocks"
)

func TestRealTimeDataService_RetrieveRealTimeData(t *testing.T) {
	targetSensorId := uuid.New()
	targetTenantId := uuid.New()
	otherTenantId := uuid.New()
	targetGatewayId := uuid.New()

	targetSensor := sensor.Sensor{
		Id:        targetSensorId,
		GatewayId: targetGatewayId,
		Name:      "Test Sensor",
		Status:    sensor.Active,
	}

    expectedTenant_CanImpersonate := tenant.Tenant{
        Id: targetTenantId,
        CanImpersonate: true,
    }

    expectedTenant_CannotImpersonate := tenant.Tenant{
        Id: targetTenantId,
        CanImpersonate: false,
    }

    // Requester 

	requesterSuperAdmin := identity.Requester{
		RequesterUserId:   uint(1),
		RequesterTenantId: nil,
		RequesterRole:     identity.ROLE_SUPER_ADMIN,
	}

	requesterTenantMember_Auth := identity.Requester{
		RequesterUserId:   uint(2),
		RequesterTenantId: &targetTenantId,
		RequesterRole:     identity.ROLE_TENANT_USER,
	}

    requesterTenantMember_Unauth := identity.Requester{
		RequesterUserId:   uint(2),
		RequesterTenantId: &otherTenantId,
		RequesterRole:     identity.ROLE_TENANT_USER,
	}

    // Command

	cmdSuperAdmin := real_time_data.RetrieveRealTimeDataCommand{
		Requester: requesterSuperAdmin,
		SensorId:  targetSensorId,
		TenantId:  targetTenantId,
	}

	cmdTenantMember_Authorized := real_time_data.RetrieveRealTimeDataCommand{
		Requester: requesterTenantMember_Auth,
		SensorId:  targetSensorId,
		TenantId:  targetTenantId,
	}

    cmdTenantMember_Unauthorized := real_time_data.RetrieveRealTimeDataCommand{
		Requester: requesterTenantMember_Unauth,
		SensorId:  targetSensorId,
		TenantId:  targetTenantId,
	}

	type mockSetupFunc func(
		mockTenantPort *tenantMocks.MockGetTenantPort,
		mockSensorByTenant *sensorMocks.MockGetSensorByTenantPort,
		mockRealTimeData *mocks.MockRealTimeDataPort,
	) *gomock.Call

	type testCase struct {
		name          string
		input         real_time_data.RetrieveRealTimeDataCommand
		setupSteps    []mockSetupFunc
		expectedError error
	}

	// Step: Get Sensor by Tenant --------------------------------------------------------------------------------------
	stepGetSensorByTenantOk := func(
		mockTenantPort *tenantMocks.MockGetTenantPort, mockSensorByTenant *sensorMocks.MockGetSensorByTenantPort, mockRealTimeData *mocks.MockRealTimeDataPort,
	) *gomock.Call {
		return mockSensorByTenant.EXPECT().
			GetSensorByTenant(targetTenantId, targetSensorId).
			Return(targetSensor, &targetTenantId, nil).
			Times(1)
	}

	stepGetSensorByTenantNotFound := func(
		mockTenantPort *tenantMocks.MockGetTenantPort, mockSensorByTenant *sensorMocks.MockGetSensorByTenantPort, mockRealTimeData *mocks.MockRealTimeDataPort,
	) *gomock.Call {
		return mockSensorByTenant.EXPECT().
			GetSensorByTenant(targetTenantId, targetSensorId).
			Return(sensor.Sensor{}, nil, sensor.ErrSensorNotFound).
			Times(1)
	}

	// Step: Get Tenant ------------------------------------------------------------------------------------------------
	stepGetTenantOk_CanImpersonate := func(
		mockTenantPort *tenantMocks.MockGetTenantPort, mockSensorByTenant *sensorMocks.MockGetSensorByTenantPort, mockRealTimeData *mocks.MockRealTimeDataPort,
	) *gomock.Call {
        return mockTenantPort.EXPECT().
            GetTenant(targetTenantId).
            Return(expectedTenant_CanImpersonate, nil).
            Times(1)
    }

    stepGetTenantOk_CannotImpersonate := func(
		mockTenantPort *tenantMocks.MockGetTenantPort, mockSensorByTenant *sensorMocks.MockGetSensorByTenantPort, mockRealTimeData *mocks.MockRealTimeDataPort,
	) *gomock.Call {
        return mockTenantPort.EXPECT().
            GetTenant(targetTenantId).
            Return(expectedTenant_CannotImpersonate, nil).
            Times(1)
    }

    errMockStepGetTenant := errors.New("unexpected error getting tenant")
    stepGetTenantFail := func(
		mockTenantPort *tenantMocks.MockGetTenantPort, mockSensorByTenant *sensorMocks.MockGetSensorByTenantPort, mockRealTimeData *mocks.MockRealTimeDataPort,
	) *gomock.Call {
        return mockTenantPort.EXPECT().
            GetTenant(targetTenantId).
            Return(tenant.Tenant{}, errMockStepGetTenant).
            Times(1)
    }

	// Step: Start Data Retriever ---------------------------------------------------------------------
	stepStartDataRetrieverOk := func(
		mockTenantPort *tenantMocks.MockGetTenantPort, mockSensorByTenant *sensorMocks.MockGetSensorByTenantPort, mockRealTimeData *mocks.MockRealTimeDataPort,
	) *gomock.Call {
		return mockRealTimeData.EXPECT().
			StartDataRetriever(targetTenantId, targetSensor, gomock.Any(), gomock.Any()).
			Return(nil).
			Times(1)
	}

    stepStartDataRetrieverNeverCalled := func(
		mockTenantPort *tenantMocks.MockGetTenantPort, mockSensorByTenant *sensorMocks.MockGetSensorByTenantPort, mockRealTimeData *mocks.MockRealTimeDataPort,
	) *gomock.Call {
        return mockRealTimeData.EXPECT().
            StartDataRetriever(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
            Times(0)
    }

	errMockRetriever := errors.New("failed to connect to nats")
	stepStartDataRetrieverError := func(
		mockTenantPort *tenantMocks.MockGetTenantPort, mockSensorByTenant *sensorMocks.MockGetSensorByTenantPort, mockRealTimeData *mocks.MockRealTimeDataPort,
	) *gomock.Call {
		return mockRealTimeData.EXPECT().
			StartDataRetriever(targetTenantId, targetSensor, gomock.Any(), gomock.Any()).
			Return(errMockRetriever).
			Times(1)
	}

	cases := []testCase{
		{
			name:  "Success (Super Admin): active sensor creates channels",
			input: cmdSuperAdmin,
			setupSteps: []mockSetupFunc{
				stepGetSensorByTenantOk,
                stepGetTenantOk_CanImpersonate,
				stepStartDataRetrieverOk,
			},
			expectedError: nil,
		},
		{
			name:  "Success (Tenant Member): active sensor creates channels",
			input: cmdTenantMember_Authorized,
			setupSteps: []mockSetupFunc{
				stepGetSensorByTenantOk,
                stepGetTenantOk_CanImpersonate,  // NOTA: impersonazione irrilevante
				stepStartDataRetrieverOk,
			},
			expectedError: nil,
		},

        // Step 1: get sensor
		{
			name:  "Fail: sensor not found",
			input: cmdTenantMember_Authorized, // NOTA: non importa requester role, basta che sia autorizzato
			setupSteps: []mockSetupFunc{
				stepGetSensorByTenantNotFound,
			},
			expectedError: sensor.ErrSensorNotFound,
		},
        
        // Step 2: get tenant
        {
			name:  "Fail: unexpected error getting tenant",
			input: cmdTenantMember_Authorized, // NOTA: non importa requester
			setupSteps: []mockSetupFunc{
				stepGetSensorByTenantOk,
                stepGetTenantFail,
			},
			expectedError: errMockStepGetTenant,
		},

        // Check accesso
        {
			name:  "(Super Admin) Fail: cannot impersonate tenant",
			input: cmdSuperAdmin,
			setupSteps: []mockSetupFunc{
				stepGetSensorByTenantOk,
                stepGetTenantOk_CannotImpersonate,
				stepStartDataRetrieverNeverCalled,
			},
			expectedError: tenant.ErrImpersonationFailed,
		},
        {
			name:  "(Tenant Admin) Fail: accesso non autorizzato",
			input: cmdTenantMember_Unauthorized,
			setupSteps: []mockSetupFunc{
				stepGetSensorByTenantOk,
                stepGetTenantOk_CanImpersonate,
				stepStartDataRetrieverNeverCalled,
			},
			expectedError: sensor.ErrSensorNotFound,
		},

        // Step 3: creazione canali
		{
			name:  "Fail: unexpected error starting data retriever",
			input: cmdTenantMember_Authorized, // NOTA: non importa requester role
			setupSteps: []mockSetupFunc{
				stepGetSensorByTenantOk,
                stepGetTenantOk_CanImpersonate,
				stepStartDataRetrieverError,
			},
			expectedError: errMockRetriever,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockTenantPort := tenantMocks.NewMockGetTenantPort(ctrl)
			mockSensorByTenant := sensorMocks.NewMockGetSensorByTenantPort(ctrl)
			mockRealTimeData := mocks.NewMockRealTimeDataPort(ctrl)

			var expectedCalls []any

			for _, step := range tc.setupSteps {
				call := step(mockTenantPort, mockSensorByTenant, mockRealTimeData)
				if call != nil {
					expectedCalls = append(expectedCalls, call)
				}
			}

			if len(expectedCalls) > 0 {
				gomock.InOrder(expectedCalls...)
			}

			service := real_time_data.NewRealTimeDataService(
                mockTenantPort,
				mockSensorByTenant,
				mockRealTimeData,
			)

			dataChan, errChan, err := service.RetrieveRealTimeData(tc.input)

			if err != tc.expectedError {
				t.Errorf("expected error %v, got %v", tc.expectedError, err)
			}

			if tc.expectedError != nil {
				if dataChan != nil || errChan != nil {
					t.Errorf("expected nil channels on error, got dataChan: %v, errChan: %v", dataChan, errChan)
				}
			} else {
				if dataChan == nil || errChan == nil {
					t.Errorf("expected initialized channels on success, got nil")
				}
			}
		})
	}
}
