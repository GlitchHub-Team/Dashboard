package sensor_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"backend/internal/sensor"
	"backend/internal/shared/identity"
	helper "backend/tests/helper"
	mocks "backend/tests/sensor/mocks"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

type getSensorByTenantIdServiceMocks struct {
	getSensorsByTenantPort *mocks.MockGetSensorsByTenantIdPort
}

type mockSetupFunc_GetSensorsByTenantService = helper.ServiceMockSetupFunc[getSensorByTenantIdServiceMocks]

func setupGetSensorsByTenantService(
	t *testing.T,
	setupSteps []mockSetupFunc_GetSensorsByTenantService,
) *sensor.GetSensorByTenantIdService {
	t.Helper()

	return helper.SetupServiceWithOrderedSteps(
		t,
		func(ctrl *gomock.Controller) getSensorByTenantIdServiceMocks {
			return getSensorByTenantIdServiceMocks{
				getSensorsByTenantPort: mocks.NewMockGetSensorsByTenantIdPort(ctrl),
			}
		},
		setupSteps,
		func(mockBundle getSensorByTenantIdServiceMocks) *sensor.GetSensorByTenantIdService {
			return sensor.NewGetSensorByTenantIdService(mockBundle.getSensorsByTenantPort)
		},
	)
}

func TestService_GetSensorsByTenant(t *testing.T) {
	targetTenantId := uuid.New()
	otherTenantId := uuid.New()

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

	baseCommand := sensor.GetSensorsByTenantCommand{
		TenantId: targetTenantId,
		Page:     2,
		Limit:    10,
	}

	inputWith := func(requester identity.Requester, tenantId uuid.UUID) sensor.GetSensorsByTenantCommand {
		cmd := baseCommand
		cmd.Requester = requester
		cmd.TenantId = tenantId
		return cmd
	}

	sensors := []sensor.Sensor{
		{
			Id:        uuid.New(),
			GatewayId: uuid.New(),
			Name:      "A",
			Interval:  1200 * time.Millisecond,
			Status:    sensor.Active,
			Profile:   sensor.HEART_RATE,
		},
		{
			Id:        uuid.New(),
			GatewayId: uuid.New(),
			Name:      "B",
			Interval:  2 * time.Second,
			Status:    sensor.Inactive,
			Profile:   sensor.ENVIRONMENTAL_SENSING,
		},
	}

	newStepGetSensorsOk := func(cmd sensor.GetSensorsByTenantCommand, expectedSensors []sensor.Sensor, expectedTotal uint) mockSetupFunc_GetSensorsByTenantService {
		return func(mockBundle getSensorByTenantIdServiceMocks) *gomock.Call {
			return mockBundle.getSensorsByTenantPort.EXPECT().
				GetSensorsByTenant(cmd.TenantId, cmd.Page, cmd.Limit).
				Return(expectedSensors, expectedTotal, nil).
				Times(1)
		}
	}

	newStepGetSensorsErr := func(cmd sensor.GetSensorsByTenantCommand, expectedErr error) mockSetupFunc_GetSensorsByTenantService {
		return func(mockBundle getSensorByTenantIdServiceMocks) *gomock.Call {
			return mockBundle.getSensorsByTenantPort.EXPECT().
				GetSensorsByTenant(cmd.TenantId, cmd.Page, cmd.Limit).
				Return(nil, uint(0), expectedErr).
				Times(1)
		}
	}

	stepGetSensorsNeverCalled := func(mockBundle getSensorByTenantIdServiceMocks) *gomock.Call {
		mockBundle.getSensorsByTenantPort.EXPECT().GetSensorsByTenant(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
		return nil
	}

	errFetchSensors := errors.New("cannot fetch sensors")

	type testCase struct {
		name            string
		input           sensor.GetSensorsByTenantCommand
		setupSteps      []mockSetupFunc_GetSensorsByTenantService
		expectedSensors []sensor.Sensor
		expectedTotal   uint
		expectedError   error
	}

	superAdminInput := inputWith(superAdminRequester, targetTenantId)
	tenantAdminNilInput := inputWith(tenantAdminWithoutTenant, targetTenantId)
	tenantAdminDifferentTenantInput := inputWith(tenantAdminWithDifferentTenant, targetTenantId)
	tenantAdminAuthorizedInput := inputWith(tenantAdminWithTenant, targetTenantId)

	cases := []testCase{
		{
			name:  "Fail (step 1): utente non super admin con requester tenantId nil",
			input: tenantAdminNilInput,
			setupSteps: []mockSetupFunc_GetSensorsByTenantService{
				stepGetSensorsNeverCalled,
			},
			expectedError: identity.ErrUnauthorizedAccess,
		},
		{
			name:  "Fail (step 1): utente non super admin con requester tenantId diverso e tenantId cmd diverso dal requester",
			input: tenantAdminDifferentTenantInput,
			setupSteps: []mockSetupFunc_GetSensorsByTenantService{
				stepGetSensorsNeverCalled,
			},
			expectedError: identity.ErrUnauthorizedAccess,
		},
		{
			name:  "Success (step 2): utente non super admin con requester tenantId uguale al tenantId del cmd",
			input: tenantAdminAuthorizedInput,
			setupSteps: []mockSetupFunc_GetSensorsByTenantService{
				newStepGetSensorsOk(tenantAdminAuthorizedInput, sensors, uint(2)),
			},
			expectedSensors: sensors,
			expectedTotal:   uint(2),
		},
		{
			name:  "Fail (step 2): errore nel fetch dei sensors attraverso la porta",
			input: superAdminInput,
			setupSteps: []mockSetupFunc_GetSensorsByTenantService{
				newStepGetSensorsErr(superAdminInput, errFetchSensors),
			},
			expectedError: errFetchSensors,
		},
		{
			name:  "Success (step 2): utente super admin",
			input: superAdminInput,
			setupSteps: []mockSetupFunc_GetSensorsByTenantService{
				newStepGetSensorsOk(superAdminInput, sensors, uint(2)),
			},
			expectedSensors: sensors,
			expectedTotal:   uint(2),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := setupGetSensorsByTenantService(t, tc.setupSteps)

			sensorsFound, total, err := service.GetSensorsByTenant(tc.input)

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
