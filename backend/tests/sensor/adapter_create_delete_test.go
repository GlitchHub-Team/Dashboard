package sensor_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"backend/internal/sensor"
	helper "backend/tests/helper"
	mocks "backend/tests/sensor/mocks"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type dbSensorAdapterMocks struct {
	databaseRepo      *mocks.MockDatabaseRepository
	messageBrokerRepo *mocks.MockMessageBrokerRepository
}

type sendCmdAdapterMocks struct {
	databaseRepo      *mocks.MockDatabaseRepository
	messageBrokerRepo *mocks.MockMessageBrokerRepository
}

type (
	mockSetupFunc_DbSensorAdapter = helper.AdapterMockSetupFunc[dbSensorAdapterMocks]
	mockSetupFunc_SendCmdAdapter  = helper.AdapterMockSetupFunc[sendCmdAdapterMocks]
)

func setupDbSensorAdapter(
	t *testing.T,
	setupSteps []mockSetupFunc_DbSensorAdapter,
) *sensor.DbSensorAdapter {
	t.Helper()

	return helper.SetupAdapterWithOrderedSteps(
		t,
		func(ctrl *gomock.Controller) dbSensorAdapterMocks {
			return dbSensorAdapterMocks{
				databaseRepo:      mocks.NewMockDatabaseRepository(ctrl),
				messageBrokerRepo: mocks.NewMockMessageBrokerRepository(ctrl),
			}
		},
		setupSteps,
		func(mockBundle dbSensorAdapterMocks) *sensor.DbSensorAdapter {
			return sensor.NewDbSensorAdapter(zap.NewNop(), mockBundle.databaseRepo)
		},
	)
}

func setupSendCmdAdapter(
	t *testing.T,
	setupSteps []mockSetupFunc_SendCmdAdapter,
) *sensor.SendCmdAdapter {
	t.Helper()

	return helper.SetupAdapterWithOrderedSteps(
		t,
		func(ctrl *gomock.Controller) sendCmdAdapterMocks {
			return sendCmdAdapterMocks{
				databaseRepo:      mocks.NewMockDatabaseRepository(ctrl),
				messageBrokerRepo: mocks.NewMockMessageBrokerRepository(ctrl),
			}
		},
		setupSteps,
		func(mockBundle sendCmdAdapterMocks) *sensor.SendCmdAdapter {
			return sensor.NewSendCmdAdapter(zap.NewNop(), mockBundle.messageBrokerRepo)
		},
	)
}

func TestDbSensorAdapter_CreateSensor(t *testing.T) {
	targetSensorId := uuid.New()
	targetGatewayId := uuid.New()
	targetName := "Heart monitor"
	targetInterval := 1500 * time.Millisecond
	targetProfile := sensor.HEART_RATE

	expectedSensor := sensor.Sensor{
		Id:        targetSensorId,
		GatewayId: targetGatewayId,
		Name:      targetName,
		Interval:  targetInterval,
		Profile:   targetProfile,
		Status:    sensor.Active,
	}

	stepCreateSensorOk := func(mockBundle dbSensorAdapterMocks) *gomock.Call {
		return mockBundle.databaseRepo.EXPECT().
			CreateSensor(gomock.AssignableToTypeOf(&sensor.SensorEntity{})).
			DoAndReturn(func(entity *sensor.SensorEntity) error {
				if entity.ID != targetSensorId.String() {
					t.Fatalf("expected entity Id %s, got %s", targetSensorId.String(), entity.ID)
				}
				if entity.GatewayID != targetGatewayId.String() {
					t.Fatalf("expected entity GatewayId %s, got %s", targetGatewayId.String(), entity.GatewayID)
				}
				if entity.Name != targetName {
					t.Fatalf("expected entity Name %s, got %s", targetName, entity.Name)
				}
				if entity.Interval != targetInterval.Milliseconds() {
					t.Fatalf("expected entity Interval %d, got %d", targetInterval.Milliseconds(), entity.Interval)
				}
				if entity.Profile != string(targetProfile) {
					t.Fatalf("expected entity Profile %s, got %s", string(targetProfile), entity.Profile)
				}

				entity.Status = string(sensor.Active)
				return nil
			}).
			Times(1)
	}

	mockCreateSensorError := errors.New("unexpected create sensor error")
	stepCreateSensorErr := func(mockBundle dbSensorAdapterMocks) *gomock.Call {
		return mockBundle.databaseRepo.EXPECT().
			CreateSensor(gomock.AssignableToTypeOf(&sensor.SensorEntity{})).
			Return(mockCreateSensorError).
			Times(1)
	}

	type testCase struct {
		name           string
		setupSteps     []mockSetupFunc_DbSensorAdapter
		expectedSensor sensor.Sensor
		expectedError  error
	}

	cases := []testCase{
		{
			name: "Success",
			setupSteps: []mockSetupFunc_DbSensorAdapter{
				stepCreateSensorOk,
			},
			expectedSensor: expectedSensor,
		},
		{
			name: "Fail: CreateSensor returns error",
			setupSteps: []mockSetupFunc_DbSensorAdapter{
				stepCreateSensorErr,
			},
			expectedError: mockCreateSensorError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			adapter := setupDbSensorAdapter(t, tc.setupSteps)

			sensorCreated, err := adapter.CreateSensor(
				targetSensorId,
				targetGatewayId,
				targetName,
				targetInterval,
				targetProfile,
			)

			if tc.expectedError != nil {
				if !errors.Is(err, tc.expectedError) {
					t.Fatalf("expected error %v, got %v", tc.expectedError, err)
				}
				if !sensorCreated.IsZero() {
					t.Fatalf("expected zero sensor, got %+v", sensorCreated)
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

func TestDbSensorAdapter_DeleteSensor(t *testing.T) {
	targetSensorId := uuid.New()
	targetGatewayId := uuid.New()
	targetName := "Heart monitor"
	targetInterval := 2 * time.Second
	targetProfile := sensor.PULSE_OXIMETER

	expectedSensor := sensor.Sensor{
		Id:        targetSensorId,
		GatewayId: targetGatewayId,
		Name:      targetName,
		Interval:  targetInterval,
		Profile:   targetProfile,
		Status:    sensor.Inactive,
	}

	stepDeleteSensorOk := func(mockBundle dbSensorAdapterMocks) *gomock.Call {
		return mockBundle.databaseRepo.EXPECT().
			DeleteSensor(gomock.AssignableToTypeOf(&sensor.SensorEntity{})).
			DoAndReturn(func(entity *sensor.SensorEntity) error {
				if entity.ID != targetSensorId.String() {
					t.Fatalf("expected entity Id %s, got %s", targetSensorId.String(), entity.ID)
				}

				entity.GatewayID = targetGatewayId.String()
				entity.Name = targetName
				entity.Interval = targetInterval.Milliseconds()
				entity.Profile = string(targetProfile)
				entity.Status = string(sensor.Inactive)
				return nil
			}).
			Times(1)
	}

	mockDeleteSensorError := errors.New("unexpected delete sensor error")
	stepDeleteSensorErr := func(mockBundle dbSensorAdapterMocks) *gomock.Call {
		return mockBundle.databaseRepo.EXPECT().
			DeleteSensor(gomock.AssignableToTypeOf(&sensor.SensorEntity{})).
			Return(mockDeleteSensorError).
			Times(1)
	}

	type testCase struct {
		name           string
		setupSteps     []mockSetupFunc_DbSensorAdapter
		expectedSensor sensor.Sensor
		expectedError  error
	}

	cases := []testCase{
		{
			name: "Success",
			setupSteps: []mockSetupFunc_DbSensorAdapter{
				stepDeleteSensorOk,
			},
			expectedSensor: expectedSensor,
		},
		{
			name: "Fail: DeleteSensor returns error",
			setupSteps: []mockSetupFunc_DbSensorAdapter{
				stepDeleteSensorErr,
			},
			expectedError: mockDeleteSensorError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			adapter := setupDbSensorAdapter(t, tc.setupSteps)

			sensorDeleted, err := adapter.DeleteSensor(targetSensorId)

			if tc.expectedError != nil {
				if !errors.Is(err, tc.expectedError) {
					t.Fatalf("expected error %v, got %v", tc.expectedError, err)
				}
				if !sensorDeleted.IsZero() {
					t.Fatalf("expected zero sensor, got %+v", sensorDeleted)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if !reflect.DeepEqual(tc.expectedSensor, sensorDeleted) {
				t.Fatalf("unexpected sensor. expected %+v, got %+v", tc.expectedSensor, sensorDeleted)
			}
		})
	}
}

func TestSendCmdAdapter_SendCreateSensorCmd(t *testing.T) {
	targetSensorId := uuid.New()
	targetGatewayId := uuid.New()
	targetInterval := 1750 * time.Millisecond
	targetProfile := sensor.ECG_CUSTOM

	stepSendCreateSensorCmdOk := func(mockBundle sendCmdAdapterMocks) *gomock.Call {
		return mockBundle.messageBrokerRepo.EXPECT().
			SendCreateSensorCmd(gomock.AssignableToTypeOf(&sensor.CreateSensorCmdEntity{})).
			DoAndReturn(func(cmd *sensor.CreateSensorCmdEntity) error {
				if cmd.SensorId != targetSensorId.String() {
					t.Fatalf("expected cmd SensorId %s, got %s", targetSensorId.String(), cmd.SensorId)
				}
				if cmd.GatewayId != targetGatewayId.String() {
					t.Fatalf("expected cmd GatewayId %s, got %s", targetGatewayId.String(), cmd.GatewayId)
				}
				if cmd.Interval != targetInterval.Milliseconds() {
					t.Fatalf("expected cmd Interval %d, got %d", targetInterval.Milliseconds(), cmd.Interval)
				}
				if cmd.Profile != string(targetProfile) {
					t.Fatalf("expected cmd Profile %s, got %s", string(targetProfile), cmd.Profile)
				}
				return nil
			}).
			Times(1)
	}

	mockSendCreateSensorCmdError := errors.New("unexpected send create command error")
	stepSendCreateSensorCmdErr := func(mockBundle sendCmdAdapterMocks) *gomock.Call {
		return mockBundle.messageBrokerRepo.EXPECT().
			SendCreateSensorCmd(gomock.AssignableToTypeOf(&sensor.CreateSensorCmdEntity{})).
			Return(mockSendCreateSensorCmdError).
			Times(1)
	}

	type testCase struct {
		name          string
		setupSteps    []mockSetupFunc_SendCmdAdapter
		expectedError error
	}

	cases := []testCase{
		{
			name: "Success",
			setupSteps: []mockSetupFunc_SendCmdAdapter{
				stepSendCreateSensorCmdOk,
			},
		},
		{
			name: "Fail: SendCreateSensorCmd returns error",
			setupSteps: []mockSetupFunc_SendCmdAdapter{
				stepSendCreateSensorCmdErr,
			},
			expectedError: mockSendCreateSensorCmdError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			adapter := setupSendCmdAdapter(t, tc.setupSteps)

			err := adapter.SendCreateSensorCmd(targetSensorId, targetGatewayId, targetInterval, targetProfile)

			if tc.expectedError != nil {
				if !errors.Is(err, tc.expectedError) {
					t.Fatalf("expected error %v, got %v", tc.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestSendCmdAdapter_SendDeleteSensorCmd(t *testing.T) {
	targetSensorId := uuid.New()
	targetGatewayId := uuid.New()

	stepSendDeleteSensorCmdOk := func(mockBundle sendCmdAdapterMocks) *gomock.Call {
		return mockBundle.messageBrokerRepo.EXPECT().
			SendDeleteSensorCmd(gomock.AssignableToTypeOf(&sensor.DeleteSensorCmdEntity{})).
			DoAndReturn(func(cmd *sensor.DeleteSensorCmdEntity) error {
				if cmd.SensorId != targetSensorId.String() {
					t.Fatalf("expected cmd SensorId %s, got %s", targetSensorId.String(), cmd.SensorId)
				}
				if cmd.GatewayId != targetGatewayId.String() {
					t.Fatalf("expected cmd GatewayId %s, got %s", targetGatewayId.String(), cmd.GatewayId)
				}
				return nil
			}).
			Times(1)
	}

	mockSendDeleteSensorCmdError := errors.New("unexpected send delete command error")
	stepSendDeleteSensorCmdErr := func(mockBundle sendCmdAdapterMocks) *gomock.Call {
		return mockBundle.messageBrokerRepo.EXPECT().
			SendDeleteSensorCmd(gomock.AssignableToTypeOf(&sensor.DeleteSensorCmdEntity{})).
			Return(mockSendDeleteSensorCmdError).
			Times(1)
	}

	type testCase struct {
		name          string
		setupSteps    []mockSetupFunc_SendCmdAdapter
		expectedError error
	}

	cases := []testCase{
		{
			name: "Success",
			setupSteps: []mockSetupFunc_SendCmdAdapter{
				stepSendDeleteSensorCmdOk,
			},
		},
		{
			name: "Fail: SendDeleteSensorCmd returns error",
			setupSteps: []mockSetupFunc_SendCmdAdapter{
				stepSendDeleteSensorCmdErr,
			},
			expectedError: mockSendDeleteSensorCmdError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			adapter := setupSendCmdAdapter(t, tc.setupSteps)

			err := adapter.SendDeleteSensorCmd(targetSensorId, targetGatewayId)

			if tc.expectedError != nil {
				if !errors.Is(err, tc.expectedError) {
					t.Fatalf("expected error %v, got %v", tc.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
