package sensor_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"backend/internal/gateway"
	"backend/internal/sensor"
	sensorProfile "backend/internal/sensor/profile"
	"backend/internal/shared/identity"
	gatewayMocks "backend/tests/gateway/mocks"
	helper "backend/tests/helper"
	mocks "backend/tests/sensor/mocks"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

type mockSetupFunc_CreateSensorService func(
	createSensorPort *mocks.MockCreateSensorPort,
	sendCreateSensorCmdPort *mocks.MockCreateSensorCmdPort,
	getGatewayPort *gatewayMocks.MockGetGatewayPort,
) *gomock.Call

func setupCreateSensorService(
	t *testing.T,
	setupSteps []mockSetupFunc_CreateSensorService,
) *sensor.CreateSensorService {
	t.Helper()

	ctrl := gomock.NewController(t)
	createSensorPort := mocks.NewMockCreateSensorPort(ctrl)
	sendCreateSensorCmdPort := mocks.NewMockCreateSensorCmdPort(ctrl)
	getGatewayPort := gatewayMocks.NewMockGetGatewayPort(ctrl)

	boundSetupSteps := make([]helper.OrderedMockStep, 0, len(setupSteps))
	for _, step := range setupSteps {
		currentStep := step
		boundSetupSteps = append(boundSetupSteps, func() *gomock.Call {
			return currentStep(createSensorPort, sendCreateSensorCmdPort, getGatewayPort)
		})
	}
	helper.SetupOrderedMockSteps(boundSetupSteps)

	return sensor.NewCreateSensorService(
		createSensorPort,
		sendCreateSensorCmdPort,
		getGatewayPort,
	)
}

func TestService_CreateSensor(t *testing.T) {
	targetTenantId := uuid.New()
	targetGatewayId := uuid.New()

	superAdminRequester := identity.Requester{
		RequesterUserId: uint(1),
		RequesterRole:   identity.ROLE_SUPER_ADMIN,
	}

	tenantAdminRequester := identity.Requester{
		RequesterUserId:   uint(2),
		RequesterTenantId: &targetTenantId,
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	baseCommand := sensor.CreateSensorCommand{
		Name:      "Heart monitor",
		Interval:  1500 * time.Millisecond,
		Profile:   sensorProfile.HEART_RATE,
		GatewayId: targetGatewayId,
	}

	inputWith := func(requester identity.Requester) sensor.CreateSensorCommand {
		cmd := baseCommand
		cmd.Requester = requester
		return cmd
	}

	gatewayFound := gateway.Gateway{
		Id:            targetGatewayId,
		Name:          "Gateway-A",
		TenantId:      &targetTenantId,
		Status:        gateway.GATEWAY_STATUS_ACTIVE,
		IntervalLimit: 3000,
	}

	newStepGatewayOk := func(cmd sensor.CreateSensorCommand) mockSetupFunc_CreateSensorService {
		return func(
			createSensorPort *mocks.MockCreateSensorPort,
			sendCreateSensorCmdPort *mocks.MockCreateSensorCmdPort,
			getGatewayPort *gatewayMocks.MockGetGatewayPort,
		) *gomock.Call {
			return getGatewayPort.EXPECT().
				GetById(cmd.GatewayId).
				Return(gatewayFound, nil).
				Times(1)
		}
	}

	newStepGatewayErr := func(cmd sensor.CreateSensorCommand, expectedErr error) mockSetupFunc_CreateSensorService {
		return func(
			createSensorPort *mocks.MockCreateSensorPort,
			sendCreateSensorCmdPort *mocks.MockCreateSensorCmdPort,
			getGatewayPort *gatewayMocks.MockGetGatewayPort,
		) *gomock.Call {
			return getGatewayPort.EXPECT().
				GetById(cmd.GatewayId).
				Return(gateway.Gateway{}, expectedErr).
				Times(1)
		}
	}

	newStepGatewayNotFound := func(cmd sensor.CreateSensorCommand) mockSetupFunc_CreateSensorService {
		return func(
			createSensorPort *mocks.MockCreateSensorPort,
			sendCreateSensorCmdPort *mocks.MockCreateSensorCmdPort,
			getGatewayPort *gatewayMocks.MockGetGatewayPort,
		) *gomock.Call {
			return getGatewayPort.EXPECT().
				GetById(cmd.GatewayId).
				Return(gateway.Gateway{}, nil).
				Times(1)
		}
	}

	stepSendCreateCmdNeverCalled := func(
		createSensorPort *mocks.MockCreateSensorPort,
		sendCreateSensorCmdPort *mocks.MockCreateSensorCmdPort,
		getGatewayPort *gatewayMocks.MockGetGatewayPort,
	) *gomock.Call {
		return sendCreateSensorCmdPort.EXPECT().SendCreateSensorCmd(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	}

	stepCreateSensorNeverCalled := func(
		createSensorPort *mocks.MockCreateSensorPort,
		sendCreateSensorCmdPort *mocks.MockCreateSensorCmdPort,
		getGatewayPort *gatewayMocks.MockGetGatewayPort,
	) *gomock.Call {
		return createSensorPort.EXPECT().CreateSensor(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	}

	newStepSendCreateCmdOk := func(cmd sensor.CreateSensorCommand) mockSetupFunc_CreateSensorService {
		return func(
			createSensorPort *mocks.MockCreateSensorPort,
			sendCreateSensorCmdPort *mocks.MockCreateSensorCmdPort,
			getGatewayPort *gatewayMocks.MockGetGatewayPort,
		) *gomock.Call {
			return sendCreateSensorCmdPort.EXPECT().
				SendCreateSensorCmd(
					gomock.Not(gomock.Eq(uuid.Nil)),
					cmd.GatewayId,
					cmd.Interval,
					cmd.Profile,
				).
				Return(nil).
				Times(1)
		}
	}

	newStepSendCreateCmdErr := func(cmd sensor.CreateSensorCommand, expectedErr error) mockSetupFunc_CreateSensorService {
		return func(
			createSensorPort *mocks.MockCreateSensorPort,
			sendCreateSensorCmdPort *mocks.MockCreateSensorCmdPort,
			getGatewayPort *gatewayMocks.MockGetGatewayPort,
		) *gomock.Call {
			return sendCreateSensorCmdPort.EXPECT().
				SendCreateSensorCmd(
					gomock.Not(gomock.Eq(uuid.Nil)),
					cmd.GatewayId,
					cmd.Interval,
					cmd.Profile,
				).
				Return(expectedErr).
				Times(1)
		}
	}

	newStepCreateSensorOk := func(cmd sensor.CreateSensorCommand, created sensor.Sensor) mockSetupFunc_CreateSensorService {
		return func(
			createSensorPort *mocks.MockCreateSensorPort,
			sendCreateSensorCmdPort *mocks.MockCreateSensorCmdPort,
			getGatewayPort *gatewayMocks.MockGetGatewayPort,
		) *gomock.Call {
			return createSensorPort.EXPECT().
				CreateSensor(
					gomock.Not(gomock.Eq(uuid.Nil)),
					cmd.GatewayId,
					cmd.Name,
					cmd.Interval,
					cmd.Profile,
				).
				Return(created, nil).
				Times(1)
		}
	}

	newStepCreateSensorErr := func(cmd sensor.CreateSensorCommand, expectedErr error) mockSetupFunc_CreateSensorService {
		return func(
			createSensorPort *mocks.MockCreateSensorPort,
			sendCreateSensorCmdPort *mocks.MockCreateSensorCmdPort,
			getGatewayPort *gatewayMocks.MockGetGatewayPort,
		) *gomock.Call {
			return createSensorPort.EXPECT().
				CreateSensor(
					gomock.Not(gomock.Eq(uuid.Nil)),
					cmd.GatewayId,
					cmd.Name,
					cmd.Interval,
					cmd.Profile,
				).
				Return(sensor.Sensor{}, expectedErr).
				Times(1)
		}
	}

	createdSensor := sensor.Sensor{
		Id:        uuid.New(),
		GatewayId: targetGatewayId,
		Name:      baseCommand.Name,
		Interval:  baseCommand.Interval,
		Profile:   baseCommand.Profile,
		Status:    sensor.Active,
	}

	errGateway := errors.New("gateway repository down")
	errSendCreateCmd := errors.New("cannot send create cmd")
	errSaveSensor := errors.New("cannot save sensor")

	type testCase struct {
		name           string
		input          sensor.CreateSensorCommand
		setupSteps     []mockSetupFunc_CreateSensorService
		expectedSensor sensor.Sensor
		expectedError  error
	}

	superAdminInput := inputWith(superAdminRequester)
	nonSuperAdminInput := inputWith(tenantAdminRequester)

	cases := []testCase{
		{
			name:  "Fail (step 1): errore nel trovare il gateway",
			input: superAdminInput,
			setupSteps: []mockSetupFunc_CreateSensorService{
				newStepGatewayErr(superAdminInput, errGateway),
				stepSendCreateCmdNeverCalled,
				stepCreateSensorNeverCalled,
			},
			expectedError: errGateway,
		},
		{
			name:  "Fail (step 1): gatewayId not found",
			input: superAdminInput,
			setupSteps: []mockSetupFunc_CreateSensorService{
				newStepGatewayNotFound(superAdminInput),
				stepSendCreateCmdNeverCalled,
				stepCreateSensorNeverCalled,
			},
			expectedError: gateway.ErrGatewayNotFound,
		},
		{
			name:  "Fail (step 2): utente non super admin",
			input: nonSuperAdminInput,
			setupSteps: []mockSetupFunc_CreateSensorService{
				newStepGatewayOk(nonSuperAdminInput),
				stepSendCreateCmdNeverCalled,
				stepCreateSensorNeverCalled,
			},
			expectedError: identity.ErrUnauthorizedAccess,
		},
		{
			name:  "Fail (step 3): errore nell'invio del comando di creazione del sensore",
			input: superAdminInput,
			setupSteps: []mockSetupFunc_CreateSensorService{
				newStepGatewayOk(superAdminInput),
				newStepSendCreateCmdErr(superAdminInput, errSendCreateCmd),
				stepCreateSensorNeverCalled,
			},
			expectedError: errSendCreateCmd,
		},
		{
			name:  "Fail (step 4): errore nel salvataggio del nuovo sensore",
			input: superAdminInput,
			setupSteps: []mockSetupFunc_CreateSensorService{
				newStepGatewayOk(superAdminInput),
				newStepSendCreateCmdOk(superAdminInput),
				newStepCreateSensorErr(superAdminInput, errSaveSensor),
			},
			expectedError: errSaveSensor,
		},
		{
			name:  "Success (step 4): salvataggio del nuovo sensore avvenuto correttamente",
			input: superAdminInput,
			setupSteps: []mockSetupFunc_CreateSensorService{
				newStepGatewayOk(superAdminInput),
				newStepSendCreateCmdOk(superAdminInput),
				newStepCreateSensorOk(superAdminInput, createdSensor),
			},
			expectedSensor: createdSensor,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := setupCreateSensorService(t, tc.setupSteps)

			sensorCreated, err := service.CreateSensor(tc.input)

			if tc.expectedError != nil {
				if !errors.Is(err, tc.expectedError) {
					t.Fatalf("expected error %v, got %v", tc.expectedError, err)
				}
				if !sensorCreated.IsZero() {
					t.Fatalf("expected zero sensor on error, got %+v", sensorCreated)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if !reflect.DeepEqual(tc.expectedSensor, sensorCreated) {
				t.Fatalf("unexpected sensor. expected %+v, got %+v", tc.expectedSensor, sensorCreated)
			}
		})
	}
}
