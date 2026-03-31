package sensor_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"backend/internal/gateway"
	"backend/internal/sensor"
	"backend/internal/shared/identity"
	gatewayMocks "backend/tests/gateway/mocks"
	helper "backend/tests/helper"
	mocks "backend/tests/sensor/mocks"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

type getSensorByIdServiceMocks struct {
	getSensorByIdPort *mocks.MockGetSensorByIdPort
	getGatewayPort    *gatewayMocks.MockGetGatewayPort
}

type mockSetupFunc_GetSensorByIdService = helper.ServiceMockSetupFunc[getSensorByIdServiceMocks]

func setupGetSensorByIdService(
	t *testing.T,
	setupSteps []mockSetupFunc_GetSensorByIdService,
) *sensor.GetSensorByIdService {
	t.Helper()

	return helper.SetupServiceWithOrderedSteps(
		t,
		func(ctrl *gomock.Controller) getSensorByIdServiceMocks {
			return getSensorByIdServiceMocks{
				getSensorByIdPort: mocks.NewMockGetSensorByIdPort(ctrl),
				getGatewayPort:    gatewayMocks.NewMockGetGatewayPort(ctrl),
			}
		},
		setupSteps,
		func(mockBundle getSensorByIdServiceMocks) *sensor.GetSensorByIdService {
			return sensor.NewGetSensorByIdService(mockBundle.getSensorByIdPort, mockBundle.getGatewayPort)
		},
	)
}

func TestService_GetSensorById(t *testing.T) {
	targetTenantId := uuid.New()
	otherTenantId := uuid.New()
	targetSensorId := uuid.New()
	targetGatewayId := uuid.New()

	superAdminRequester := identity.Requester{
		RequesterUserId: uint(1),
		RequesterRole:   identity.ROLE_SUPER_ADMIN,
	}

	tenantAdminWithTenant := identity.Requester{
		RequesterUserId:   uint(2),
		RequesterTenantId: &targetTenantId,
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	tenantAdminWithDifferentTenant := identity.Requester{
		RequesterUserId:   uint(3),
		RequesterTenantId: &otherTenantId,
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	tenantAdminWithoutTenant := identity.Requester{
		RequesterUserId: uint(4),
		RequesterRole:   identity.ROLE_TENANT_ADMIN,
	}

	baseCommand := sensor.GetSensorCommand{SensorId: targetSensorId}

	inputWith := func(requester identity.Requester) sensor.GetSensorCommand {
		cmd := baseCommand
		cmd.Requester = requester
		return cmd
	}

	foundSensor := sensor.Sensor{
		Id:        targetSensorId,
		GatewayId: targetGatewayId,
		Name:      "Heart monitor",
		Interval:  1500 * time.Millisecond,
		Status:    sensor.Active,
		Profile:   sensor.HEART_RATE,
	}

	gatewayBelongsToTenant := gateway.Gateway{
		Id:            targetGatewayId,
		Name:          "Gateway-A",
		TenantId:      &targetTenantId,
		Status:        gateway.GATEWAY_STATUS_ACTIVE,
		IntervalLimit: 3000,
	}

	gatewayWithoutTenant := gateway.Gateway{
		Id:            targetGatewayId,
		Name:          "Gateway-B",
		Status:        gateway.GATEWAY_STATUS_ACTIVE,
		IntervalLimit: 3000,
	}

	gatewayWithOtherTenant := gateway.Gateway{
		Id:            targetGatewayId,
		Name:          "Gateway-C",
		TenantId:      &otherTenantId,
		Status:        gateway.GATEWAY_STATUS_ACTIVE,
		IntervalLimit: 3000,
	}

	newStepGetSensorOk := func(cmd sensor.GetSensorCommand) mockSetupFunc_GetSensorByIdService {
		return func(mockBundle getSensorByIdServiceMocks) *gomock.Call {
			return mockBundle.getSensorByIdPort.EXPECT().
				GetSensorById(cmd.SensorId).
				Return(foundSensor, nil).
				Times(1)
		}
	}

	newStepGetSensorErr := func(cmd sensor.GetSensorCommand, expectedErr error) mockSetupFunc_GetSensorByIdService {
		return func(mockBundle getSensorByIdServiceMocks) *gomock.Call {
			return mockBundle.getSensorByIdPort.EXPECT().
				GetSensorById(cmd.SensorId).
				Return(sensor.Sensor{}, expectedErr).
				Times(1)
		}
	}

	newStepGetSensorNotFound := func(cmd sensor.GetSensorCommand) mockSetupFunc_GetSensorByIdService {
		return func(mockBundle getSensorByIdServiceMocks) *gomock.Call {
			return mockBundle.getSensorByIdPort.EXPECT().
				GetSensorById(cmd.SensorId).
				Return(sensor.Sensor{}, nil).
				Times(1)
		}
	}

	newStepGatewayOk := func(gat gateway.Gateway) mockSetupFunc_GetSensorByIdService {
		return func(mockBundle getSensorByIdServiceMocks) *gomock.Call {
			return mockBundle.getGatewayPort.EXPECT().
				GetById(foundSensor.GatewayId.String()).
				Return(gat, nil).
				Times(1)
		}
	}

	newStepGatewayErr := func(expectedErr error) mockSetupFunc_GetSensorByIdService {
		return func(mockBundle getSensorByIdServiceMocks) *gomock.Call {
			return mockBundle.getGatewayPort.EXPECT().
				GetById(foundSensor.GatewayId.String()).
				Return(gateway.Gateway{}, expectedErr).
				Times(1)
		}
	}

	newStepGatewayNotFound := func() mockSetupFunc_GetSensorByIdService {
		return func(mockBundle getSensorByIdServiceMocks) *gomock.Call {
			return mockBundle.getGatewayPort.EXPECT().
				GetById(foundSensor.GatewayId.String()).
				Return(gateway.Gateway{}, nil).
				Times(1)
		}
	}

	stepGetGatewayNeverCalled := func(mockBundle getSensorByIdServiceMocks) *gomock.Call {
		mockBundle.getGatewayPort.EXPECT().GetById(gomock.Any()).Times(0)
		return nil
	}

	errGetSensor := errors.New("cannot fetch sensor")
	errGetGateway := errors.New("cannot fetch gateway")

	type testCase struct {
		name           string
		input          sensor.GetSensorCommand
		setupSteps     []mockSetupFunc_GetSensorByIdService
		expectedSensor sensor.Sensor
		expectedError  error
	}

	superAdminInput := inputWith(superAdminRequester)
	tenantAdminNilInput := inputWith(tenantAdminWithoutTenant)
	tenantAdminAuthorizedInput := inputWith(tenantAdminWithTenant)
	tenantAdminDifferentTenantInput := inputWith(tenantAdminWithDifferentTenant)

	cases := []testCase{
		{
			name:  "Fail (step 1): errore nel trovare il sensore",
			input: superAdminInput,
			setupSteps: []mockSetupFunc_GetSensorByIdService{
				newStepGetSensorErr(superAdminInput, errGetSensor),
				stepGetGatewayNeverCalled,
			},
			expectedError: errGetSensor,
		},
		{
			name:  "Fail (step 1): sensorId not found",
			input: superAdminInput,
			setupSteps: []mockSetupFunc_GetSensorByIdService{
				newStepGetSensorNotFound(superAdminInput),
				stepGetGatewayNeverCalled,
			},
			expectedError: sensor.ErrSensorNotFound,
		},
		{
			name:  "Fail (step 2): errore nel trovare il gateway",
			input: superAdminInput,
			setupSteps: []mockSetupFunc_GetSensorByIdService{
				newStepGetSensorOk(superAdminInput),
				newStepGatewayErr(errGetGateway),
			},
			expectedError: errGetGateway,
		},
		{
			name:  "Fail (step 2): gatewayId not found",
			input: superAdminInput,
			setupSteps: []mockSetupFunc_GetSensorByIdService{
				newStepGetSensorOk(superAdminInput),
				newStepGatewayNotFound(),
			},
			expectedError: gateway.ErrGatewayNotFound,
		},
		{
			name:  "Fail (step 3): utente non super admin con RequesterTenantId nil",
			input: tenantAdminNilInput,
			setupSteps: []mockSetupFunc_GetSensorByIdService{
				newStepGetSensorOk(tenantAdminNilInput),
				newStepGatewayOk(gatewayBelongsToTenant),
			},
			expectedError: identity.ErrUnauthorizedAccess,
		},
		{
			name:  "Fail (step 3): utente non super admin con RequesterTenantId diverso da nil e gateway tenantId nil",
			input: tenantAdminAuthorizedInput,
			setupSteps: []mockSetupFunc_GetSensorByIdService{
				newStepGetSensorOk(tenantAdminAuthorizedInput),
				newStepGatewayOk(gatewayWithoutTenant),
			},
			expectedError: identity.ErrUnauthorizedAccess,
		},
		{
			name:  "Fail (step 3): utente non super admin con RequesterTenantId diverso da nil e tenant del gateway diverso",
			input: tenantAdminAuthorizedInput,
			setupSteps: []mockSetupFunc_GetSensorByIdService{
				newStepGetSensorOk(tenantAdminAuthorizedInput),
				newStepGatewayOk(gatewayWithOtherTenant),
			},
			expectedError: identity.ErrUnauthorizedAccess,
		},
		{
			name:  "Success (step 3): utente non super admin con RequesterTenantId uguale a quello del gateway",
			input: tenantAdminAuthorizedInput,
			setupSteps: []mockSetupFunc_GetSensorByIdService{
				newStepGetSensorOk(tenantAdminAuthorizedInput),
				newStepGatewayOk(gatewayBelongsToTenant),
			},
			expectedSensor: foundSensor,
		},
		{
			name:  "Success (step 3): utente super admin",
			input: superAdminInput,
			setupSteps: []mockSetupFunc_GetSensorByIdService{
				newStepGetSensorOk(superAdminInput),
				newStepGatewayOk(gatewayBelongsToTenant),
			},
			expectedSensor: foundSensor,
		},
		{
			name:  "Fail (step 3): utente non super admin con RequesterTenantId diverso dal tenant del gateway",
			input: tenantAdminDifferentTenantInput,
			setupSteps: []mockSetupFunc_GetSensorByIdService{
				newStepGetSensorOk(tenantAdminDifferentTenantInput),
				newStepGatewayOk(gatewayBelongsToTenant),
			},
			expectedError: identity.ErrUnauthorizedAccess,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := setupGetSensorByIdService(t, tc.setupSteps)

			sensorFound, err := service.GetSensorById(tc.input)

			if tc.expectedError != nil {
				if !errors.Is(err, tc.expectedError) {
					t.Fatalf("expected error %v, got %v", tc.expectedError, err)
				}
				if !sensorFound.IsZero() {
					t.Fatalf("expected zero sensor on error, got %+v", sensorFound)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if !reflect.DeepEqual(tc.expectedSensor, sensorFound) {
				t.Fatalf("unexpected sensor. expected %+v, got %+v", tc.expectedSensor, sensorFound)
			}
		})
	}
}
