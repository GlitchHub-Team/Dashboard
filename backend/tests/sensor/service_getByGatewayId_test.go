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

type mockSetupFunc_GetSensorsByGatewayIdService func(
	getSensorsByGatewayIdPort *mocks.MockGetSensorsByGatewayIdPort,
	getGatewayPort *gatewayMocks.MockGetGatewayPort,
) *gomock.Call

func setupGetSensorsByGatewayIdService(
	t *testing.T,
	setupSteps []mockSetupFunc_GetSensorsByGatewayIdService,
) *sensor.GetSensorsByGatewayIdService {
	t.Helper()

	ctrl := gomock.NewController(t)
	getSensorsByGatewayIdPort := mocks.NewMockGetSensorsByGatewayIdPort(ctrl)
	getGatewayPort := gatewayMocks.NewMockGetGatewayPort(ctrl)

	boundSetupSteps := make([]helper.OrderedMockStep, 0, len(setupSteps))
	for _, step := range setupSteps {
		currentStep := step
		boundSetupSteps = append(boundSetupSteps, func() *gomock.Call {
			return currentStep(getSensorsByGatewayIdPort, getGatewayPort)
		})
	}
	helper.SetupOrderedMockSteps(boundSetupSteps)

	return sensor.NewGetSensorsByGatewayIdService(
		getSensorsByGatewayIdPort,
		getGatewayPort,
	)
}

func TestService_GetSensorsByGateway(t *testing.T) {
	targetTenantId := uuid.New()
	otherTenantId := uuid.New()
	targetGatewayId := uuid.New()

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
		RequesterUserId:   uint(3),
		RequesterTenantId: &otherTenantId,
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	}

	tenantAdminWithoutTenantId := identity.Requester{
		RequesterUserId: uint(4),
		RequesterRole:   identity.ROLE_TENANT_ADMIN,
	}

	baseCommand := sensor.GetSensorsByGatewayCommand{
		Page:      2,
		Limit:     10,
		GatewayId: targetGatewayId,
	}

	inputWith := func(requester identity.Requester) sensor.GetSensorsByGatewayCommand {
		cmd := baseCommand
		cmd.Requester = requester
		return cmd
	}

	gatewayBelongsToTenant := gateway.Gateway{
		Id:       targetGatewayId,
		Name:     "Gateway-A",
		TenantId: &targetTenantId,
		Status:   gateway.GATEWAY_STATUS_ACTIVE,
	}

	gatewayWithoutTenant := gateway.Gateway{
		Id:     targetGatewayId,
		Name:   "Gateway-B",
		Status: gateway.GATEWAY_STATUS_ACTIVE,
	}

	sensors := []sensor.Sensor{
		{
			Id:        uuid.New(),
			GatewayId: targetGatewayId,
			Name:      "A",
			Interval:  1500 * time.Millisecond,
			Status:    sensor.Active,
			Profile:   sensorProfile.HEART_RATE,
		},
		{
			Id:        uuid.New(),
			GatewayId: targetGatewayId,
			Name:      "B",
			Interval:  2 * time.Second,
			Status:    sensor.Inactive,
			Profile:   sensorProfile.ENVIRONMENTAL_SENSING,
		},
	}

	newStepGatewayOk := func(cmd sensor.GetSensorsByGatewayCommand, foundGateway gateway.Gateway) mockSetupFunc_GetSensorsByGatewayIdService {
		return func(
			getSensorsByGatewayIdPort *mocks.MockGetSensorsByGatewayIdPort,
			getGatewayPort *gatewayMocks.MockGetGatewayPort,
		) *gomock.Call {
			return getGatewayPort.EXPECT().
				GetById(cmd.GatewayId).
				Return(foundGateway, nil).
				Times(1)
		}
	}

	newStepGatewayErr := func(cmd sensor.GetSensorsByGatewayCommand, expectedErr error) mockSetupFunc_GetSensorsByGatewayIdService {
		return func(
			getSensorsByGatewayIdPort *mocks.MockGetSensorsByGatewayIdPort,
			getGatewayPort *gatewayMocks.MockGetGatewayPort,
		) *gomock.Call {
			return getGatewayPort.EXPECT().
				GetById(cmd.GatewayId).
				Return(gateway.Gateway{}, expectedErr).
				Times(1)
		}
	}

	newStepGatewayNotFound := func(cmd sensor.GetSensorsByGatewayCommand) mockSetupFunc_GetSensorsByGatewayIdService {
		return func(
			getSensorsByGatewayIdPort *mocks.MockGetSensorsByGatewayIdPort,
			getGatewayPort *gatewayMocks.MockGetGatewayPort,
		) *gomock.Call {
			return getGatewayPort.EXPECT().
				GetById(cmd.GatewayId).
				Return(gateway.Gateway{}, nil).
				Times(1)
		}
	}

	stepFetchSensorsNeverCalled := func(
		getSensorsByGatewayIdPort *mocks.MockGetSensorsByGatewayIdPort,
		getGatewayPort *gatewayMocks.MockGetGatewayPort,
	) *gomock.Call {
		return getSensorsByGatewayIdPort.EXPECT().GetSensorsByGatewayId(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	}

	newStepFetchSensorsErr := func(cmd sensor.GetSensorsByGatewayCommand, expectedErr error) mockSetupFunc_GetSensorsByGatewayIdService {
		return func(
			getSensorsByGatewayIdPort *mocks.MockGetSensorsByGatewayIdPort,
			getGatewayPort *gatewayMocks.MockGetGatewayPort,
		) *gomock.Call {
			return getSensorsByGatewayIdPort.EXPECT().
				GetSensorsByGatewayId(cmd.GatewayId, cmd.Page, cmd.Limit).
				Return(nil, uint(0), expectedErr).
				Times(1)
		}
	}

	newStepFetchSensorsOk := func(cmd sensor.GetSensorsByGatewayCommand, expectedSensors []sensor.Sensor, expectedTotal uint) mockSetupFunc_GetSensorsByGatewayIdService {
		return func(
			getSensorsByGatewayIdPort *mocks.MockGetSensorsByGatewayIdPort,
			getGatewayPort *gatewayMocks.MockGetGatewayPort,
		) *gomock.Call {
			return getSensorsByGatewayIdPort.EXPECT().
				GetSensorsByGatewayId(cmd.GatewayId, cmd.Page, cmd.Limit).
				Return(expectedSensors, expectedTotal, nil).
				Times(1)
		}
	}

	errGateway := errors.New("cannot fetch gateway")
	errFetchSensors := errors.New("cannot fetch sensors")

	type testCase struct {
		name            string
		input           sensor.GetSensorsByGatewayCommand
		setupSteps      []mockSetupFunc_GetSensorsByGatewayIdService
		expectedSensors []sensor.Sensor
		expectedTotal   uint
		expectedError   error
	}

	superAdminInput := inputWith(superAdminRequester)
	tenantAdminNilInput := inputWith(tenantAdminWithoutTenantId)
	tenantAdminDiffTenantInput := inputWith(unauthorizedTenantAdminRequester)
	authorizedTenantAdminInput := inputWith(authorizedTenantAdminRequester)

	cases := []testCase{
		{
			name:  "Fail (step 1): errore nel trovare il gateway",
			input: superAdminInput,
			setupSteps: []mockSetupFunc_GetSensorsByGatewayIdService{
				newStepGatewayErr(superAdminInput, errGateway),
				stepFetchSensorsNeverCalled,
			},
			expectedError: errGateway,
		},
		{
			name:  "Fail (step 1): gatewayId non trovato",
			input: superAdminInput,
			setupSteps: []mockSetupFunc_GetSensorsByGatewayIdService{
				newStepGatewayNotFound(superAdminInput),
				stepFetchSensorsNeverCalled,
			},
			expectedError: gateway.ErrGatewayNotFound,
		},
		{
			name:  "Fail (step 2): utente non super admin con requesterTenantId nil",
			input: tenantAdminNilInput,
			setupSteps: []mockSetupFunc_GetSensorsByGatewayIdService{
				newStepGatewayOk(tenantAdminNilInput, gatewayBelongsToTenant),
				stepFetchSensorsNeverCalled,
			},
			expectedError: identity.ErrUnauthorizedAccess,
		},
		{
			name:  "Fail (step 2): utente non super admin con RequesterTenantId diverso dal tenantId del gateway",
			input: tenantAdminDiffTenantInput,
			setupSteps: []mockSetupFunc_GetSensorsByGatewayIdService{
				newStepGatewayOk(tenantAdminDiffTenantInput, gatewayBelongsToTenant),
				stepFetchSensorsNeverCalled,
			},
			expectedError: identity.ErrUnauthorizedAccess,
		},
		{
			name:  "Fail (step 2): utente non super admin con tenantId del gateway nil",
			input: authorizedTenantAdminInput,
			setupSteps: []mockSetupFunc_GetSensorsByGatewayIdService{
				newStepGatewayOk(authorizedTenantAdminInput, gatewayWithoutTenant),
				stepFetchSensorsNeverCalled,
			},
			expectedError: identity.ErrUnauthorizedAccess,
		},
		{
			name:  "Fail (step 3): errore nel fetching dei sensors dalla porta",
			input: superAdminInput,
			setupSteps: []mockSetupFunc_GetSensorsByGatewayIdService{
				newStepGatewayOk(superAdminInput, gatewayBelongsToTenant),
				newStepFetchSensorsErr(superAdminInput, errFetchSensors),
			},
			expectedError: errFetchSensors,
		},
		{
			name:  "Success (step 3): sensori fetchati correttamente",
			input: superAdminInput,
			setupSteps: []mockSetupFunc_GetSensorsByGatewayIdService{
				newStepGatewayOk(superAdminInput, gatewayBelongsToTenant),
				newStepFetchSensorsOk(superAdminInput, sensors, uint(12)),
			},
			expectedSensors: sensors,
			expectedTotal:   uint(12),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := setupGetSensorsByGatewayIdService(t, tc.setupSteps)

			sensorsFound, total, err := service.GetSensorsByGateway(tc.input)

			if tc.expectedError != nil {
				if !errors.Is(err, tc.expectedError) {
					t.Fatalf("expected error %v, got %v", tc.expectedError, err)
				}
				if sensorsFound != nil {
					t.Fatalf("expected nil sensors on error, got %+v", sensorsFound)
				}
				if total != 0 {
					t.Fatalf("expected total 0 on error, got %d", total)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if !reflect.DeepEqual(tc.expectedSensors, sensorsFound) {
				t.Fatalf("unexpected sensors. expected %+v, got %+v", tc.expectedSensors, sensorsFound)
			}

			if tc.expectedTotal != total {
				t.Fatalf("unexpected total. expected %d, got %d", tc.expectedTotal, total)
			}
		})
	}
}
