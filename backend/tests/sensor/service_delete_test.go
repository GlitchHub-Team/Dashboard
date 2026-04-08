package sensor_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"backend/internal/sensor"
	sensorProfile "backend/internal/sensor/profile"
	"backend/internal/shared/identity"
	helper "backend/tests/helper"
	mocks "backend/tests/sensor/mocks"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

type mockSetupFunc_DeleteSensorService func(
	deleteSensorPort *mocks.MockDeleteSensorPort,
	getSensorByIdPort *mocks.MockGetSensorByIdPort,
	deleteSensorCmdPort *mocks.MockDeleteSensorCmdPort,
) *gomock.Call

func setupDeleteSensorService(
	t *testing.T,
	setupSteps []mockSetupFunc_DeleteSensorService,
) *sensor.DeleteSensorService {
	t.Helper()

	ctrl := gomock.NewController(t)
	deleteSensorPort := mocks.NewMockDeleteSensorPort(ctrl)
	getSensorByIdPort := mocks.NewMockGetSensorByIdPort(ctrl)
	deleteSensorCmdPort := mocks.NewMockDeleteSensorCmdPort(ctrl)

	boundSetupSteps := make([]helper.OrderedMockStep, 0, len(setupSteps))
	for _, step := range setupSteps {
		currentStep := step
		boundSetupSteps = append(boundSetupSteps, func() *gomock.Call {
			return currentStep(deleteSensorPort, getSensorByIdPort, deleteSensorCmdPort)
		})
	}
	helper.SetupOrderedMockSteps(boundSetupSteps)

	return sensor.NewDeleteSensorService(
		deleteSensorPort,
		getSensorByIdPort,
		deleteSensorCmdPort,
	)
}

func TestService_DeleteSensor(t *testing.T) {
	targetTenantId := uuid.New()
	targetSensorId := uuid.New()
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

	baseCommand := sensor.DeleteSensorCommand{
		SensorId: targetSensorId,
	}

	inputWith := func(requester identity.Requester) sensor.DeleteSensorCommand {
		cmd := baseCommand
		cmd.Requester = requester
		return cmd
	}

	foundSensor := sensor.Sensor{
		Id:        targetSensorId,
		GatewayId: targetGatewayId,
		Name:      "Heart monitor",
		Interval:  1200 * time.Millisecond,
		Status:    sensor.Active,
		Profile:   sensorProfile.HEART_RATE,
	}

	newStepGetSensorByIdOk := func(cmd sensor.DeleteSensorCommand) mockSetupFunc_DeleteSensorService {
		return func(
			deleteSensorPort *mocks.MockDeleteSensorPort,
			getSensorByIdPort *mocks.MockGetSensorByIdPort,
			deleteSensorCmdPort *mocks.MockDeleteSensorCmdPort,
		) *gomock.Call {
			return getSensorByIdPort.EXPECT().
				GetSensorById(cmd.SensorId).
				Return(foundSensor, nil).
				Times(1)
		}
	}

	newStepGetSensorByIdErr := func(cmd sensor.DeleteSensorCommand, expectedErr error) mockSetupFunc_DeleteSensorService {
		return func(
			deleteSensorPort *mocks.MockDeleteSensorPort,
			getSensorByIdPort *mocks.MockGetSensorByIdPort,
			deleteSensorCmdPort *mocks.MockDeleteSensorCmdPort,
		) *gomock.Call {
			return getSensorByIdPort.EXPECT().
				GetSensorById(cmd.SensorId).
				Return(sensor.Sensor{}, expectedErr).
				Times(1)
		}
	}

	newStepGetSensorByIdNotFound := func(cmd sensor.DeleteSensorCommand) mockSetupFunc_DeleteSensorService {
		return func(
			deleteSensorPort *mocks.MockDeleteSensorPort,
			getSensorByIdPort *mocks.MockGetSensorByIdPort,
			deleteSensorCmdPort *mocks.MockDeleteSensorCmdPort,
		) *gomock.Call {
			return getSensorByIdPort.EXPECT().
				GetSensorById(cmd.SensorId).
				Return(sensor.Sensor{}, nil).
				Times(1)
		}
	}

	stepSendDeleteCmdNeverCalled := func(
		deleteSensorPort *mocks.MockDeleteSensorPort,
		getSensorByIdPort *mocks.MockGetSensorByIdPort,
		deleteSensorCmdPort *mocks.MockDeleteSensorCmdPort,
	) *gomock.Call {
		return deleteSensorCmdPort.EXPECT().SendDeleteSensorCmd(gomock.Any(), gomock.Any()).Times(0)
	}

	stepDeleteSensorNeverCalled := func(
		deleteSensorPort *mocks.MockDeleteSensorPort,
		getSensorByIdPort *mocks.MockGetSensorByIdPort,
		deleteSensorCmdPort *mocks.MockDeleteSensorCmdPort,
	) *gomock.Call {
		return deleteSensorPort.EXPECT().DeleteSensor(gomock.Any()).Times(0)
	}

	newStepSendDeleteCmdOk := func(cmd sensor.DeleteSensorCommand) mockSetupFunc_DeleteSensorService {
		return func(
			deleteSensorPort *mocks.MockDeleteSensorPort,
			getSensorByIdPort *mocks.MockGetSensorByIdPort,
			deleteSensorCmdPort *mocks.MockDeleteSensorCmdPort,
		) *gomock.Call {
			return deleteSensorCmdPort.EXPECT().
				SendDeleteSensorCmd(cmd.SensorId, foundSensor.GatewayId).
				Return(nil).
				Times(1)
		}
	}

	newStepSendDeleteCmdErr := func(cmd sensor.DeleteSensorCommand, expectedErr error) mockSetupFunc_DeleteSensorService {
		return func(
			deleteSensorPort *mocks.MockDeleteSensorPort,
			getSensorByIdPort *mocks.MockGetSensorByIdPort,
			deleteSensorCmdPort *mocks.MockDeleteSensorCmdPort,
		) *gomock.Call {
			return deleteSensorCmdPort.EXPECT().
				SendDeleteSensorCmd(cmd.SensorId, foundSensor.GatewayId).
				Return(expectedErr).
				Times(1)
		}
	}

	newStepDeleteSensorOk := func(cmd sensor.DeleteSensorCommand, deleted sensor.Sensor) mockSetupFunc_DeleteSensorService {
		return func(
			deleteSensorPort *mocks.MockDeleteSensorPort,
			getSensorByIdPort *mocks.MockGetSensorByIdPort,
			deleteSensorCmdPort *mocks.MockDeleteSensorCmdPort,
		) *gomock.Call {
			return deleteSensorPort.EXPECT().
				DeleteSensor(cmd.SensorId).
				Return(deleted, nil).
				Times(1)
		}
	}

	newStepDeleteSensorErr := func(cmd sensor.DeleteSensorCommand, expectedErr error) mockSetupFunc_DeleteSensorService {
		return func(
			deleteSensorPort *mocks.MockDeleteSensorPort,
			getSensorByIdPort *mocks.MockGetSensorByIdPort,
			deleteSensorCmdPort *mocks.MockDeleteSensorCmdPort,
		) *gomock.Call {
			return deleteSensorPort.EXPECT().
				DeleteSensor(cmd.SensorId).
				Return(sensor.Sensor{}, expectedErr).
				Times(1)
		}
	}

	errGetSensor := errors.New("get sensor failed")
	errSendDeleteCmd := errors.New("cannot send delete command")
	errDeleteAtDb := errors.New("cannot delete at database")

	deletedSensor := sensor.Sensor{
		Id:        targetSensorId,
		GatewayId: targetGatewayId,
		Name:      "Heart monitor",
		Interval:  1200 * time.Millisecond,
		Status:    sensor.Inactive,
		Profile:   sensorProfile.HEART_RATE,
	}

	type testCase struct {
		name           string
		input          sensor.DeleteSensorCommand
		setupSteps     []mockSetupFunc_DeleteSensorService
		expectedSensor sensor.Sensor
		expectedError  error
	}

	superAdminInput := inputWith(superAdminRequester)
	nonSuperAdminInput := inputWith(tenantAdminRequester)

	cases := []testCase{
		{
			name:  "Fail (step 1): errore nel trovare il sensore",
			input: superAdminInput,
			setupSteps: []mockSetupFunc_DeleteSensorService{
				newStepGetSensorByIdErr(superAdminInput, errGetSensor),
				stepSendDeleteCmdNeverCalled,
				stepDeleteSensorNeverCalled,
			},
			expectedError: errGetSensor,
		},
		{
			name:  "Fail (step 1): sensorId non trovato",
			input: superAdminInput,
			setupSteps: []mockSetupFunc_DeleteSensorService{
				newStepGetSensorByIdNotFound(superAdminInput),
				stepSendDeleteCmdNeverCalled,
				stepDeleteSensorNeverCalled,
			},
			expectedError: sensor.ErrSensorNotFound,
		},
		{
			name:  "Fail (step 2): utente non super admin",
			input: nonSuperAdminInput,
			setupSteps: []mockSetupFunc_DeleteSensorService{
				newStepGetSensorByIdOk(nonSuperAdminInput),
				stepSendDeleteCmdNeverCalled,
				stepDeleteSensorNeverCalled,
			},
			expectedError: identity.ErrUnauthorizedAccess,
		},
		{
			name:  "Fail (step 3): errore nell'invio del comando di eliminazione del sensore",
			input: superAdminInput,
			setupSteps: []mockSetupFunc_DeleteSensorService{
				newStepGetSensorByIdOk(superAdminInput),
				newStepSendDeleteCmdErr(superAdminInput, errSendDeleteCmd),
				stepDeleteSensorNeverCalled,
			},
			expectedError: errSendDeleteCmd,
		},
		{
			name:  "Fail (step 4): errore nell'eliminazione del sensore a database",
			input: superAdminInput,
			setupSteps: []mockSetupFunc_DeleteSensorService{
				newStepGetSensorByIdOk(superAdminInput),
				newStepSendDeleteCmdOk(superAdminInput),
				newStepDeleteSensorErr(superAdminInput, errDeleteAtDb),
			},
			expectedError: errDeleteAtDb,
		},
		{
			name:  "Success (step 4): elimina del sensore avvenuta correttamente",
			input: superAdminInput,
			setupSteps: []mockSetupFunc_DeleteSensorService{
				newStepGetSensorByIdOk(superAdminInput),
				newStepSendDeleteCmdOk(superAdminInput),
				newStepDeleteSensorOk(superAdminInput, deletedSensor),
			},
			expectedSensor: deletedSensor,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := setupDeleteSensorService(t, tc.setupSteps)

			deleted, err := service.DeleteSensor(tc.input)

			if tc.expectedError != nil {
				if !errors.Is(err, tc.expectedError) {
					t.Fatalf("expected error %v, got %v", tc.expectedError, err)
				}
				if !deleted.IsZero() {
					t.Fatalf("expected zero sensor on error, got %+v", deleted)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if !reflect.DeepEqual(tc.expectedSensor, deleted) {
				t.Fatalf("unexpected sensor. expected %+v, got %+v", tc.expectedSensor, deleted)
			}
		})
	}
}
