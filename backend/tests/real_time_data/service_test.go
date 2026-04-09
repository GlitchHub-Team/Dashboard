package real_time_data_test

import (
    "errors"
    "testing"

    "go.uber.org/mock/gomock"
    "github.com/google/uuid"

    "backend/internal/gateway"
    "backend/internal/real_time_data"
    "backend/internal/sensor"
    "backend/internal/shared/identity"
    "backend/internal/tenant"
    mocks "backend/tests/real_time_data/mocks"
    sensorMocks "backend/tests/sensor/mocks"
)

func TestRealTimeDataService_RetrieveRealTimeData(t *testing.T) {
    targetSensorId := uuid.New()
    targetTenantId := uuid.New()
    targetGatewayId := uuid.New()

    targetSensor := sensor.Sensor{
        Id:        targetSensorId,
        GatewayId: targetGatewayId,
        Name:      "Test Sensor",
        Status:    sensor.Active,
    }

    targetGatewayWithTenant := gateway.Gateway{
        Id:       targetGatewayId,
        TenantId: &targetTenantId,
    }

    targetGatewayWithoutTenant := gateway.Gateway{
        Id:       targetGatewayId,
        TenantId: nil,
    }

    targetTenant := tenant.Tenant{
        Id: targetTenantId,
    }

    requesterSuperAdmin := identity.Requester{
        RequesterUserId:   uint(1),
        RequesterTenantId: nil,
        RequesterRole:     identity.ROLE_SUPER_ADMIN,
    }

    requesterTenantMember := identity.Requester{
        RequesterUserId:   uint(2),
        RequesterTenantId: &targetTenantId,
        RequesterRole:     identity.ROLE_TENANT_USER,
    }

    cmdSuperAdmin := real_time_data.RetrieveRealTimeDataCommand{
        Requester: requesterSuperAdmin,
        SensorId:  targetSensorId,
    }

    cmdTenantMember := real_time_data.RetrieveRealTimeDataCommand{
        Requester: requesterTenantMember,
        SensorId:  targetSensorId,
    }

    type mockSetupFunc func(
        mockSensorWithGateway *sensorMocks.MockGetSensorWithGatewayPort,
        mockSensorByTenant *sensorMocks.MockGetSensorByTenantPort,
        mockRealTimeData *mocks.MockRealTimeDataPort,
    ) *gomock.Call

    type testCase struct {
        name          string
        input         real_time_data.RetrieveRealTimeDataCommand
        setupSteps    []mockSetupFunc
        expectedError error
    }

    // Step: Get Sensor with Gateway (Super Admin) ----------------------------------------------------
    stepGetSensorWithGatewayOk := func(
        mockSensorWithGateway *sensorMocks.MockGetSensorWithGatewayPort, mockSensorByTenant *sensorMocks.MockGetSensorByTenantPort, mockRealTimeData *mocks.MockRealTimeDataPort,
    ) *gomock.Call {
        return mockSensorWithGateway.EXPECT().
            GetSensorWithGateway(targetSensorId).
            Return(targetSensor, targetGatewayWithTenant, nil).
            Times(1)
    }

    stepGetSensorWithGatewayNotActive := func(
        mockSensorWithGateway *sensorMocks.MockGetSensorWithGatewayPort, mockSensorByTenant *sensorMocks.MockGetSensorByTenantPort, mockRealTimeData *mocks.MockRealTimeDataPort,
    ) *gomock.Call {
        return mockSensorWithGateway.EXPECT().
            GetSensorWithGateway(targetSensorId).
            Return(targetSensor, targetGatewayWithoutTenant, nil).
            Times(1)
    }

    // Step: Get Sensor by Tenant (Tenant Member) -----------------------------------------------------
    stepGetSensorByTenantOk := func(
        mockSensorWithGateway *sensorMocks.MockGetSensorWithGatewayPort, mockSensorByTenant *sensorMocks.MockGetSensorByTenantPort, mockRealTimeData *mocks.MockRealTimeDataPort,
    ) *gomock.Call {
        return mockSensorByTenant.EXPECT().
            GetSensorByTenant(targetTenantId, targetSensorId).
            Return(targetSensor, targetTenant, nil).
            Times(1)
    }

    stepGetSensorByTenantNotFound := func(
        mockSensorWithGateway *sensorMocks.MockGetSensorWithGatewayPort, mockSensorByTenant *sensorMocks.MockGetSensorByTenantPort, mockRealTimeData *mocks.MockRealTimeDataPort,
    ) *gomock.Call {
        return mockSensorByTenant.EXPECT().
            GetSensorByTenant(targetTenantId, targetSensorId).
            Return(sensor.Sensor{}, tenant.Tenant{}, sensor.ErrSensorNotFound).
            Times(1)
    }

    // Step: Start Data Retriever ---------------------------------------------------------------------
    stepStartDataRetrieverOk := func(
        mockSensorWithGateway *sensorMocks.MockGetSensorWithGatewayPort, mockSensorByTenant *sensorMocks.MockGetSensorByTenantPort, mockRealTimeData *mocks.MockRealTimeDataPort,
    ) *gomock.Call {
        return mockRealTimeData.EXPECT().
            StartDataRetriever(targetTenantId, targetSensor, gomock.Any(), gomock.Any()).
            Return(nil).
            Times(1)
    }

    errMockRetriever := errors.New("failed to connect to nats")
    stepStartDataRetrieverError := func(
        mockSensorWithGateway *sensorMocks.MockGetSensorWithGatewayPort, mockSensorByTenant *sensorMocks.MockGetSensorByTenantPort, mockRealTimeData *mocks.MockRealTimeDataPort,
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
                stepGetSensorWithGatewayOk,
                stepStartDataRetrieverOk,
            },
            expectedError: nil,
        },
        {
            name:  "Success (Tenant Member): active sensor creates channels",
            input: cmdTenantMember,
            setupSteps: []mockSetupFunc{
                stepGetSensorByTenantOk,
                stepStartDataRetrieverOk,
            },
            expectedError: nil,
        },
        {
            name:  "Fail (Tenant Member): sensor does not exist",
            input: cmdTenantMember,
            setupSteps: []mockSetupFunc{
                stepGetSensorByTenantNotFound,
            },
            expectedError: sensor.ErrSensorNotFound,
        },
		{
            name:  "Fail (Super Admin): sensor does not exist",
            input: cmdTenantMember,
            setupSteps: []mockSetupFunc{
                stepGetSensorByTenantNotFound,
            },
            expectedError: sensor.ErrSensorNotFound,
        },
        {
            name:  "Fail (Super Admin): sensor not active (nil tenant id on gateway)",
            input: cmdSuperAdmin,
            setupSteps: []mockSetupFunc{
                stepGetSensorWithGatewayNotActive,
            },
            expectedError: sensor.ErrSensorNotActive,
        },
        {
            name:  "Fail: unexpected error starting data retriever",
            input: cmdTenantMember,
            setupSteps: []mockSetupFunc{
                stepGetSensorByTenantOk,
                stepStartDataRetrieverError,
            },
            expectedError: errMockRetriever,
        },
    }

    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()

            mockSensorWithGateway := sensorMocks.NewMockGetSensorWithGatewayPort(ctrl)
            mockSensorByTenant := sensorMocks.NewMockGetSensorByTenantPort(ctrl)
            mockRealTimeData := mocks.NewMockRealTimeDataPort(ctrl)

            var expectedCalls []any

            for _, step := range tc.setupSteps {
                call := step(mockSensorWithGateway, mockSensorByTenant, mockRealTimeData)
                if call != nil {
                    expectedCalls = append(expectedCalls, call)
                }
            }

            if len(expectedCalls) > 0 {
                gomock.InOrder(expectedCalls...)
            }

            service := real_time_data.NewRealTimeDataService(
                mockSensorWithGateway,
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