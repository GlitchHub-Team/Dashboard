package sensor_test

import (
	"errors"
	"testing"

	"backend/internal/sensor"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestSendCmdAdapter_SendResume(t *testing.T) {
	targetSensorId := uuid.New()
	targetGatewayId := uuid.New()
	expectedErr := errors.New("unexpected send resume error")

	stepSendResumeErr := func(mockBundle sendCmdAdapterMocks) *gomock.Call {
		return mockBundle.messageBrokerRepo.EXPECT().
			SendResumeSensorCmd(gomock.AssignableToTypeOf(&sensor.ResumeSensorCmdEntity{})).
			DoAndReturn(func(cmd *sensor.ResumeSensorCmdEntity) error {
				if cmd.SensorId != targetSensorId.String() {
					t.Fatalf("expected cmd SensorId %s, got %s", targetSensorId.String(), cmd.SensorId)
				}
				if cmd.GatewayId != targetGatewayId.String() {
					t.Fatalf("expected cmd GatewayId %s, got %s", targetGatewayId.String(), cmd.GatewayId)
				}
				return expectedErr
			}).
			Times(1)
	}

	adapter := setupSendCmdAdapter(t, []mockSetupFunc_SendCmdAdapter{stepSendResumeErr})

	err := adapter.SendResume(targetSensorId, targetGatewayId)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestSendCmdAdapter_SendInterrupt(t *testing.T) {
	targetSensorId := uuid.New()
	targetGatewayId := uuid.New()
	expectedErr := errors.New("unexpected send interrupt error")

	stepSendInterruptErr := func(mockBundle sendCmdAdapterMocks) *gomock.Call {
		return mockBundle.messageBrokerRepo.EXPECT().
			SendInterruptSensorCmd(gomock.AssignableToTypeOf(&sensor.InterruptSensorCmdEntity{})).
			DoAndReturn(func(cmd *sensor.InterruptSensorCmdEntity) error {
				if cmd.SensorId != targetSensorId.String() {
					t.Fatalf("expected cmd SensorId %s, got %s", targetSensorId.String(), cmd.SensorId)
				}
				if cmd.GatewayId != targetGatewayId.String() {
					t.Fatalf("expected cmd GatewayId %s, got %s", targetGatewayId.String(), cmd.GatewayId)
				}
				return expectedErr
			}).
			Times(1)
	}

	adapter := setupSendCmdAdapter(t, []mockSetupFunc_SendCmdAdapter{stepSendInterruptErr})

	err := adapter.SendInterrupt(targetSensorId, targetGatewayId)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestDbSensorAdapter_UpdateSensorStatus(t *testing.T) {
	targetSensor := sensor.Sensor{Id: uuid.New()}
	targetStatus := sensor.Inactive
	expectedErr := errors.New("unexpected update status error")

	stepUpdateStatusErr := func(mockBundle dbSensorAdapterMocks) *gomock.Call {
		return mockBundle.databaseRepo.EXPECT().
			UpdateSensor(targetSensor.Id.String(), string(targetStatus)).
			Return(expectedErr).
			Times(1)
	}

	adapter := setupDbSensorAdapter(t, []mockSetupFunc_DbSensorAdapter{stepUpdateStatusErr})

	err := adapter.UpdateSensorStatus(targetSensor, targetStatus)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}
