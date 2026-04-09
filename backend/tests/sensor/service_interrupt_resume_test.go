package sensor_test

import (
	"errors"
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

type interruptResumeServiceMocks struct {
	sendInterruptCmdPort    *mocks.MockSendInterruptCmdPort
	sendResumeCmdPort       *mocks.MockSendResumeCmdPort
	getSensorByIdPort       *mocks.MockGetSensorByIdPort
	getGatewayPort          *gatewayMocks.MockGetGatewayPort
	updatedSensorStatusPort *mocks.MockUpdateSensorStatusPort
}

type mockSetupFunc_InterruptResumeService = helper.ServiceMockSetupFunc[interruptResumeServiceMocks]

func setupInterruptSensorService(
	t *testing.T,
	setupSteps []mockSetupFunc_InterruptResumeService,
) *sensor.InterruptSensorService {
	t.Helper()

	return helper.SetupServiceWithOrderedSteps(
		t,
		func(ctrl *gomock.Controller) interruptResumeServiceMocks {
			return interruptResumeServiceMocks{
				sendInterruptCmdPort:    mocks.NewMockSendInterruptCmdPort(ctrl),
				sendResumeCmdPort:       mocks.NewMockSendResumeCmdPort(ctrl),
				getSensorByIdPort:       mocks.NewMockGetSensorByIdPort(ctrl),
				getGatewayPort:          gatewayMocks.NewMockGetGatewayPort(ctrl),
				updatedSensorStatusPort: mocks.NewMockUpdateSensorStatusPort(ctrl),
			}
		},
		setupSteps,
		func(mockBundle interruptResumeServiceMocks) *sensor.InterruptSensorService {
			return sensor.NewInterruptSensorService(
				mockBundle.sendInterruptCmdPort,
				mockBundle.getSensorByIdPort,
				mockBundle.getGatewayPort,
				mockBundle.updatedSensorStatusPort,
			)
		},
	)
}

func setupResumeSensorService(
	t *testing.T,
	setupSteps []mockSetupFunc_InterruptResumeService,
) *sensor.ResumeSensorService {
	t.Helper()

	return helper.SetupServiceWithOrderedSteps(
		t,
		func(ctrl *gomock.Controller) interruptResumeServiceMocks {
			return interruptResumeServiceMocks{
				sendInterruptCmdPort:    mocks.NewMockSendInterruptCmdPort(ctrl),
				sendResumeCmdPort:       mocks.NewMockSendResumeCmdPort(ctrl),
				getSensorByIdPort:       mocks.NewMockGetSensorByIdPort(ctrl),
				getGatewayPort:          gatewayMocks.NewMockGetGatewayPort(ctrl),
				updatedSensorStatusPort: mocks.NewMockUpdateSensorStatusPort(ctrl),
			}
		},
		setupSteps,
		func(mockBundle interruptResumeServiceMocks) *sensor.ResumeSensorService {
			return sensor.NewResumeSensorService(
				mockBundle.sendResumeCmdPort,
				mockBundle.getSensorByIdPort,
				mockBundle.getGatewayPort,
				mockBundle.updatedSensorStatusPort,
			)
		},
	)
}

func newStepGetSensorByIdOk(sensorId uuid.UUID, foundSensor sensor.Sensor) mockSetupFunc_InterruptResumeService {
	return func(mockBundle interruptResumeServiceMocks) *gomock.Call {
		return mockBundle.getSensorByIdPort.EXPECT().
			GetSensorById(sensorId).
			Return(foundSensor, nil).
			Times(1)
	}
}

func newStepGetSensorByIdErr(sensorId uuid.UUID, expectedErr error) mockSetupFunc_InterruptResumeService {
	return func(mockBundle interruptResumeServiceMocks) *gomock.Call {
		return mockBundle.getSensorByIdPort.EXPECT().
			GetSensorById(sensorId).
			Return(sensor.Sensor{}, expectedErr).
			Times(1)
	}
}

func newStepGetSensorByIdNotFound(sensorId uuid.UUID) mockSetupFunc_InterruptResumeService {
	return func(mockBundle interruptResumeServiceMocks) *gomock.Call {
		return mockBundle.getSensorByIdPort.EXPECT().
			GetSensorById(sensorId).
			Return(sensor.Sensor{}, nil).
			Times(1)
	}
}

func newStepGetGatewayOk(foundSensor sensor.Sensor, foundGateway gateway.Gateway) mockSetupFunc_InterruptResumeService {
	return func(mockBundle interruptResumeServiceMocks) *gomock.Call {
		return mockBundle.getGatewayPort.EXPECT().
			GetById(foundSensor.GatewayId.String()).
			Return(foundGateway, nil).
			Times(1)
	}
}

func newStepGetGatewayErr(foundSensor sensor.Sensor, expectedErr error) mockSetupFunc_InterruptResumeService {
	return func(mockBundle interruptResumeServiceMocks) *gomock.Call {
		return mockBundle.getGatewayPort.EXPECT().
			GetById(foundSensor.GatewayId.String()).
			Return(gateway.Gateway{}, expectedErr).
			Times(1)
	}
}

func newStepGetGatewayNotFound(foundSensor sensor.Sensor) mockSetupFunc_InterruptResumeService {
	return func(mockBundle interruptResumeServiceMocks) *gomock.Call {
		return mockBundle.getGatewayPort.EXPECT().
			GetById(foundSensor.GatewayId.String()).
			Return(gateway.Gateway{}, nil).
			Times(1)
	}
}

func newStepSendInterruptOk(sensorId uuid.UUID, gatewayId uuid.UUID) mockSetupFunc_InterruptResumeService {
	return func(mockBundle interruptResumeServiceMocks) *gomock.Call {
		return mockBundle.sendInterruptCmdPort.EXPECT().
			SendInterrupt(sensorId, gatewayId).
			Return(nil).
			Times(1)
	}
}

func newStepSendInterruptErr(sensorId uuid.UUID, gatewayId uuid.UUID, expectedErr error) mockSetupFunc_InterruptResumeService {
	return func(mockBundle interruptResumeServiceMocks) *gomock.Call {
		return mockBundle.sendInterruptCmdPort.EXPECT().
			SendInterrupt(sensorId, gatewayId).
			Return(expectedErr).
			Times(1)
	}
}

func newStepSendResumeOk(sensorId uuid.UUID, gatewayId uuid.UUID) mockSetupFunc_InterruptResumeService {
	return func(mockBundle interruptResumeServiceMocks) *gomock.Call {
		return mockBundle.sendResumeCmdPort.EXPECT().
			SendResume(sensorId, gatewayId).
			Return(nil).
			Times(1)
	}
}

func newStepSendResumeErr(sensorId uuid.UUID, gatewayId uuid.UUID, expectedErr error) mockSetupFunc_InterruptResumeService {
	return func(mockBundle interruptResumeServiceMocks) *gomock.Call {
		return mockBundle.sendResumeCmdPort.EXPECT().
			SendResume(sensorId, gatewayId).
			Return(expectedErr).
			Times(1)
	}
}

func newStepUpdateSensorStatusOk(expectedSensor sensor.Sensor, status sensor.SensorStatus) mockSetupFunc_InterruptResumeService {
	return func(mockBundle interruptResumeServiceMocks) *gomock.Call {
		return mockBundle.updatedSensorStatusPort.EXPECT().
			UpdateSensorStatus(expectedSensor, status).
			Return(nil).
			Times(1)
	}
}

func newStepUpdateSensorStatusErr(expectedSensor sensor.Sensor, status sensor.SensorStatus, expectedErr error) mockSetupFunc_InterruptResumeService {
	return func(mockBundle interruptResumeServiceMocks) *gomock.Call {
		return mockBundle.updatedSensorStatusPort.EXPECT().
			UpdateSensorStatus(expectedSensor, status).
			Return(expectedErr).
			Times(1)
	}
}

func stepGetGatewayNeverCalled(mockBundle interruptResumeServiceMocks) *gomock.Call {
	mockBundle.getGatewayPort.EXPECT().GetById(gomock.Any()).Times(0)
	return nil
}

func stepSendInterruptNeverCalled(mockBundle interruptResumeServiceMocks) *gomock.Call {
	mockBundle.sendInterruptCmdPort.EXPECT().SendInterrupt(gomock.Any(), gomock.Any()).Times(0)
	return nil
}

func stepSendResumeNeverCalled(mockBundle interruptResumeServiceMocks) *gomock.Call {
	mockBundle.sendResumeCmdPort.EXPECT().SendResume(gomock.Any(), gomock.Any()).Times(0)
	return nil
}

func stepUpdateStatusNeverCalled(mockBundle interruptResumeServiceMocks) *gomock.Call {
	mockBundle.updatedSensorStatusPort.EXPECT().UpdateSensorStatus(gomock.Any(), gomock.Any()).Times(0)
	return nil
}

func TestService_InterruptSensor(t *testing.T) {
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

	tenantAdminWithoutTenant := identity.Requester{
		RequesterUserId: uint(3),
		RequesterRole:   identity.ROLE_TENANT_ADMIN,
	}

	tenantAdminWithDifferentTenant := identity.Requester{
		RequesterUserId:   uint(4),
		RequesterTenantId: &otherTenantId,
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	baseCommand := sensor.InterruptSensorCommand{SensorId: targetSensorId}

	inputWith := func(requester identity.Requester) sensor.InterruptSensorCommand {
		cmd := baseCommand
		cmd.Requester = requester
		return cmd
	}

	activeSensor := sensor.Sensor{
		Id:        targetSensorId,
		GatewayId: targetGatewayId,
		Name:      "Heart monitor",
		Interval:  1500 * time.Millisecond,
		Status:    sensor.Active,
		Profile:   sensorProfile.HEART_RATE,
	}

	inactiveSensor := sensor.Sensor{
		Id:        targetSensorId,
		GatewayId: targetGatewayId,
		Name:      "Heart monitor",
		Interval:  1500 * time.Millisecond,
		Status:    sensor.Inactive,
		Profile:   sensorProfile.HEART_RATE,
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

	errGetSensor := errors.New("cannot fetch sensor")
	errGetGateway := errors.New("cannot fetch gateway")
	errSendInterrupt := errors.New("cannot send interrupt command")
	errUpdateStatus := errors.New("cannot update sensor status")

	type testCase struct {
		name          string
		input         sensor.InterruptSensorCommand
		setupSteps    []mockSetupFunc_InterruptResumeService
		expectedError error
	}

	superAdminInput := inputWith(superAdminRequester)
	tenantAdminAuthorizedInput := inputWith(tenantAdminWithTenant)
	tenantAdminNilInput := inputWith(tenantAdminWithoutTenant)
	tenantAdminDifferentTenantInput := inputWith(tenantAdminWithDifferentTenant)

	cases := []testCase{
		{
			name:  "Fail (step 1): errore nel trovare il sensor",
			input: superAdminInput,
			setupSteps: []mockSetupFunc_InterruptResumeService{
				newStepGetSensorByIdErr(superAdminInput.SensorId, errGetSensor),
				stepGetGatewayNeverCalled,
				stepSendInterruptNeverCalled,
				stepSendResumeNeverCalled,
				stepUpdateStatusNeverCalled,
			},
			expectedError: errGetSensor,
		},
		{
			name:  "Fail (step 1): sensorId not found",
			input: superAdminInput,
			setupSteps: []mockSetupFunc_InterruptResumeService{
				newStepGetSensorByIdNotFound(superAdminInput.SensorId),
				stepGetGatewayNeverCalled,
				stepSendInterruptNeverCalled,
				stepSendResumeNeverCalled,
				stepUpdateStatusNeverCalled,
			},
			expectedError: sensor.ErrSensorNotFound,
		},
		{
			name:  "Fail (step 1): sensor gia interrotto",
			input: superAdminInput,
			setupSteps: []mockSetupFunc_InterruptResumeService{
				newStepGetSensorByIdOk(superAdminInput.SensorId, inactiveSensor),
				stepGetGatewayNeverCalled,
				stepSendInterruptNeverCalled,
				stepSendResumeNeverCalled,
				stepUpdateStatusNeverCalled,
			},
			expectedError: sensor.ErrSensorNotActive,
		},
		{
			name:  "Fail (step 2): errore nel trovare il gateway",
			input: superAdminInput,
			setupSteps: []mockSetupFunc_InterruptResumeService{
				newStepGetSensorByIdOk(superAdminInput.SensorId, activeSensor),
				newStepGetGatewayErr(activeSensor, errGetGateway),
				stepSendInterruptNeverCalled,
				stepSendResumeNeverCalled,
				stepUpdateStatusNeverCalled,
			},
			expectedError: errGetGateway,
		},
		{
			name:  "Fail (step 2): gatewayId not found",
			input: superAdminInput,
			setupSteps: []mockSetupFunc_InterruptResumeService{
				newStepGetSensorByIdOk(superAdminInput.SensorId, activeSensor),
				newStepGetGatewayNotFound(activeSensor),
				stepSendInterruptNeverCalled,
				stepSendResumeNeverCalled,
				stepUpdateStatusNeverCalled,
			},
			expectedError: gateway.ErrGatewayNotFound,
		},
		{
			name:  "Fail (step 3): utente non super admin con RequesterTenantId nil",
			input: tenantAdminNilInput,
			setupSteps: []mockSetupFunc_InterruptResumeService{
				newStepGetSensorByIdOk(tenantAdminNilInput.SensorId, activeSensor),
				newStepGetGatewayOk(activeSensor, gatewayBelongsToTenant),
				stepSendInterruptNeverCalled,
				stepSendResumeNeverCalled,
				stepUpdateStatusNeverCalled,
			},
			expectedError: identity.ErrUnauthorizedAccess,
		},
		{
			name:  "Fail (step 3): utente non super admin con RequesterTenantId diverso da nil con gateway tenantId nil",
			input: tenantAdminAuthorizedInput,
			setupSteps: []mockSetupFunc_InterruptResumeService{
				newStepGetSensorByIdOk(tenantAdminAuthorizedInput.SensorId, activeSensor),
				newStepGetGatewayOk(activeSensor, gatewayWithoutTenant),
				stepSendInterruptNeverCalled,
				stepSendResumeNeverCalled,
				stepUpdateStatusNeverCalled,
			},
			expectedError: identity.ErrUnauthorizedAccess,
		},
		{
			name:  "Fail (step 3): utente non super admin con RequesterTenantId diverso da nil con gateway tenantId diverso",
			input: tenantAdminAuthorizedInput,
			setupSteps: []mockSetupFunc_InterruptResumeService{
				newStepGetSensorByIdOk(tenantAdminAuthorizedInput.SensorId, activeSensor),
				newStepGetGatewayOk(activeSensor, gatewayWithOtherTenant),
				stepSendInterruptNeverCalled,
				stepSendResumeNeverCalled,
				stepUpdateStatusNeverCalled,
			},
			expectedError: identity.ErrUnauthorizedAccess,
		},
		{
			name:  "Success (step 4): utente super admin ok",
			input: superAdminInput,
			setupSteps: []mockSetupFunc_InterruptResumeService{
				newStepGetSensorByIdOk(superAdminInput.SensorId, activeSensor),
				newStepGetGatewayOk(activeSensor, gatewayWithoutTenant),
				newStepSendInterruptOk(superAdminInput.SensorId, activeSensor.GatewayId),
				newStepUpdateSensorStatusOk(activeSensor, sensor.Inactive),
				stepSendResumeNeverCalled,
			},
		},
		{
			name:  "Success (step 4): utente non super admin con RequesterTenantId uguale a quello del gateway",
			input: tenantAdminAuthorizedInput,
			setupSteps: []mockSetupFunc_InterruptResumeService{
				newStepGetSensorByIdOk(tenantAdminAuthorizedInput.SensorId, activeSensor),
				newStepGetGatewayOk(activeSensor, gatewayBelongsToTenant),
				newStepSendInterruptOk(tenantAdminAuthorizedInput.SensorId, activeSensor.GatewayId),
				newStepUpdateSensorStatusOk(activeSensor, sensor.Inactive),
				stepSendResumeNeverCalled,
			},
		},
		{
			name:  "Fail (step 4): errore nell'invio dell'interrupt",
			input: superAdminInput,
			setupSteps: []mockSetupFunc_InterruptResumeService{
				newStepGetSensorByIdOk(superAdminInput.SensorId, activeSensor),
				newStepGetGatewayOk(activeSensor, gatewayBelongsToTenant),
				newStepSendInterruptErr(superAdminInput.SensorId, activeSensor.GatewayId, errSendInterrupt),
				stepSendResumeNeverCalled,
				stepUpdateStatusNeverCalled,
			},
			expectedError: errSendInterrupt,
		},
		{
			name:  "Fail (step 5): errore nel salvataggio dello stato del sensore",
			input: superAdminInput,
			setupSteps: []mockSetupFunc_InterruptResumeService{
				newStepGetSensorByIdOk(superAdminInput.SensorId, activeSensor),
				newStepGetGatewayOk(activeSensor, gatewayBelongsToTenant),
				newStepSendInterruptOk(superAdminInput.SensorId, activeSensor.GatewayId),
				newStepUpdateSensorStatusErr(activeSensor, sensor.Inactive, errUpdateStatus),
				stepSendResumeNeverCalled,
			},
			expectedError: errUpdateStatus,
		},
		{
			name:  "Fail (step 3): utente non super admin con RequesterTenantId diverso dal tenant del gateway",
			input: tenantAdminDifferentTenantInput,
			setupSteps: []mockSetupFunc_InterruptResumeService{
				newStepGetSensorByIdOk(tenantAdminDifferentTenantInput.SensorId, activeSensor),
				newStepGetGatewayOk(activeSensor, gatewayBelongsToTenant),
				stepSendInterruptNeverCalled,
				stepSendResumeNeverCalled,
				stepUpdateStatusNeverCalled,
			},
			expectedError: identity.ErrUnauthorizedAccess,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := setupInterruptSensorService(t, tc.setupSteps)

			err := service.InterruptSensor(tc.input)

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

func TestService_ResumeSensor(t *testing.T) {
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

	tenantAdminWithoutTenant := identity.Requester{
		RequesterUserId: uint(3),
		RequesterRole:   identity.ROLE_TENANT_ADMIN,
	}

	tenantAdminWithDifferentTenant := identity.Requester{
		RequesterUserId:   uint(4),
		RequesterTenantId: &otherTenantId,
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	baseCommand := sensor.ResumeSensorCommand{SensorId: targetSensorId}

	inputWith := func(requester identity.Requester) sensor.ResumeSensorCommand {
		cmd := baseCommand
		cmd.Requester = requester
		return cmd
	}

	activeSensor := sensor.Sensor{
		Id:        targetSensorId,
		GatewayId: targetGatewayId,
		Name:      "Heart monitor",
		Interval:  1500 * time.Millisecond,
		Status:    sensor.Active,
		Profile:   sensorProfile.HEART_RATE,
	}

	inactiveSensor := sensor.Sensor{
		Id:        targetSensorId,
		GatewayId: targetGatewayId,
		Name:      "Heart monitor",
		Interval:  1500 * time.Millisecond,
		Status:    sensor.Inactive,
		Profile:   sensorProfile.HEART_RATE,
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

	errGetSensor := errors.New("cannot fetch sensor")
	errGetGateway := errors.New("cannot fetch gateway")
	errSendResume := errors.New("cannot send resume command")
	errUpdateStatus := errors.New("cannot update sensor status")

	type testCase struct {
		name          string
		input         sensor.ResumeSensorCommand
		setupSteps    []mockSetupFunc_InterruptResumeService
		expectedError error
	}

	superAdminInput := inputWith(superAdminRequester)
	tenantAdminAuthorizedInput := inputWith(tenantAdminWithTenant)
	tenantAdminNilInput := inputWith(tenantAdminWithoutTenant)
	tenantAdminDifferentTenantInput := inputWith(tenantAdminWithDifferentTenant)

	cases := []testCase{
		{
			name:  "Fail (step 1): errore nel trovare il sensor",
			input: superAdminInput,
			setupSteps: []mockSetupFunc_InterruptResumeService{
				newStepGetSensorByIdErr(superAdminInput.SensorId, errGetSensor),
				stepGetGatewayNeverCalled,
				stepSendInterruptNeverCalled,
				stepSendResumeNeverCalled,
				stepUpdateStatusNeverCalled,
			},
			expectedError: errGetSensor,
		},
		{
			name:  "Fail (step 1): sensorId not found",
			input: superAdminInput,
			setupSteps: []mockSetupFunc_InterruptResumeService{
				newStepGetSensorByIdNotFound(superAdminInput.SensorId),
				stepGetGatewayNeverCalled,
				stepSendInterruptNeverCalled,
				stepSendResumeNeverCalled,
				stepUpdateStatusNeverCalled,
			},
			expectedError: sensor.ErrSensorNotFound,
		},
		{
			name:  "Fail (step 1): sensor gia attivo",
			input: superAdminInput,
			setupSteps: []mockSetupFunc_InterruptResumeService{
				newStepGetSensorByIdOk(superAdminInput.SensorId, activeSensor),
				stepGetGatewayNeverCalled,
				stepSendInterruptNeverCalled,
				stepSendResumeNeverCalled,
				stepUpdateStatusNeverCalled,
			},
			expectedError: sensor.ErrSensorNotInactive,
		},
		{
			name:  "Fail (step 2): errore nel trovare il gateway",
			input: superAdminInput,
			setupSteps: []mockSetupFunc_InterruptResumeService{
				newStepGetSensorByIdOk(superAdminInput.SensorId, inactiveSensor),
				newStepGetGatewayErr(inactiveSensor, errGetGateway),
				stepSendInterruptNeverCalled,
				stepSendResumeNeverCalled,
				stepUpdateStatusNeverCalled,
			},
			expectedError: errGetGateway,
		},
		{
			name:  "Fail (step 2): gatewayId not found",
			input: superAdminInput,
			setupSteps: []mockSetupFunc_InterruptResumeService{
				newStepGetSensorByIdOk(superAdminInput.SensorId, inactiveSensor),
				newStepGetGatewayNotFound(inactiveSensor),
				stepSendInterruptNeverCalled,
				stepSendResumeNeverCalled,
				stepUpdateStatusNeverCalled,
			},
			expectedError: gateway.ErrGatewayNotFound,
		},
		{
			name:  "Fail (step 3): utente non super admin con RequesterTenantId nil",
			input: tenantAdminNilInput,
			setupSteps: []mockSetupFunc_InterruptResumeService{
				newStepGetSensorByIdOk(tenantAdminNilInput.SensorId, inactiveSensor),
				newStepGetGatewayOk(inactiveSensor, gatewayBelongsToTenant),
				stepSendInterruptNeverCalled,
				stepSendResumeNeverCalled,
				stepUpdateStatusNeverCalled,
			},
			expectedError: identity.ErrUnauthorizedAccess,
		},
		{
			name:  "Fail (step 3): utente non super admin con RequesterTenantId diverso da nil con gateway tenantId nil",
			input: tenantAdminAuthorizedInput,
			setupSteps: []mockSetupFunc_InterruptResumeService{
				newStepGetSensorByIdOk(tenantAdminAuthorizedInput.SensorId, inactiveSensor),
				newStepGetGatewayOk(inactiveSensor, gatewayWithoutTenant),
				stepSendInterruptNeverCalled,
				stepSendResumeNeverCalled,
				stepUpdateStatusNeverCalled,
			},
			expectedError: identity.ErrUnauthorizedAccess,
		},
		{
			name:  "Fail (step 3): utente non super admin con RequesterTenantId diverso da nil con gateway tenantId diverso",
			input: tenantAdminAuthorizedInput,
			setupSteps: []mockSetupFunc_InterruptResumeService{
				newStepGetSensorByIdOk(tenantAdminAuthorizedInput.SensorId, inactiveSensor),
				newStepGetGatewayOk(inactiveSensor, gatewayWithOtherTenant),
				stepSendInterruptNeverCalled,
				stepSendResumeNeverCalled,
				stepUpdateStatusNeverCalled,
			},
			expectedError: identity.ErrUnauthorizedAccess,
		},
		{
			name:  "Success (step 4): utente super admin ok",
			input: superAdminInput,
			setupSteps: []mockSetupFunc_InterruptResumeService{
				newStepGetSensorByIdOk(superAdminInput.SensorId, inactiveSensor),
				newStepGetGatewayOk(inactiveSensor, gatewayWithoutTenant),
				newStepSendResumeOk(superAdminInput.SensorId, inactiveSensor.GatewayId),
				newStepUpdateSensorStatusOk(inactiveSensor, sensor.Active),
				stepSendInterruptNeverCalled,
			},
		},
		{
			name:  "Success (step 4): utente non super admin con RequesterTenantId uguale a quello del gateway",
			input: tenantAdminAuthorizedInput,
			setupSteps: []mockSetupFunc_InterruptResumeService{
				newStepGetSensorByIdOk(tenantAdminAuthorizedInput.SensorId, inactiveSensor),
				newStepGetGatewayOk(inactiveSensor, gatewayBelongsToTenant),
				newStepSendResumeOk(tenantAdminAuthorizedInput.SensorId, inactiveSensor.GatewayId),
				newStepUpdateSensorStatusOk(inactiveSensor, sensor.Active),
				stepSendInterruptNeverCalled,
			},
		},
		{
			name:  "Fail (step 4): errore nell'invio del resume",
			input: superAdminInput,
			setupSteps: []mockSetupFunc_InterruptResumeService{
				newStepGetSensorByIdOk(superAdminInput.SensorId, inactiveSensor),
				newStepGetGatewayOk(inactiveSensor, gatewayBelongsToTenant),
				newStepSendResumeErr(superAdminInput.SensorId, inactiveSensor.GatewayId, errSendResume),
				stepSendInterruptNeverCalled,
				stepUpdateStatusNeverCalled,
			},
			expectedError: errSendResume,
		},
		{
			name:  "Fail (step 5): errore nel salvataggio dello stato del sensore",
			input: superAdminInput,
			setupSteps: []mockSetupFunc_InterruptResumeService{
				newStepGetSensorByIdOk(superAdminInput.SensorId, inactiveSensor),
				newStepGetGatewayOk(inactiveSensor, gatewayBelongsToTenant),
				newStepSendResumeOk(superAdminInput.SensorId, inactiveSensor.GatewayId),
				newStepUpdateSensorStatusErr(inactiveSensor, sensor.Active, errUpdateStatus),
				stepSendInterruptNeverCalled,
			},
			expectedError: errUpdateStatus,
		},
		{
			name:  "Fail (step 3): utente non super admin con RequesterTenantId diverso dal tenant del gateway",
			input: tenantAdminDifferentTenantInput,
			setupSteps: []mockSetupFunc_InterruptResumeService{
				newStepGetSensorByIdOk(tenantAdminDifferentTenantInput.SensorId, inactiveSensor),
				newStepGetGatewayOk(inactiveSensor, gatewayBelongsToTenant),
				stepSendInterruptNeverCalled,
				stepSendResumeNeverCalled,
				stepUpdateStatusNeverCalled,
			},
			expectedError: identity.ErrUnauthorizedAccess,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := setupResumeSensorService(t, tc.setupSteps)

			err := service.ResumeSensor(tc.input)

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
