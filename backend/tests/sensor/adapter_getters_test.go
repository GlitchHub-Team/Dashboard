package sensor_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"backend/internal/infra/database/pagination"
	"backend/internal/sensor"
	sensorProfile "backend/internal/sensor/profile"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestDbSensorAdapter_GetSensorsByGatewayId(t *testing.T) {
	targetGatewayId := uuid.New()

	entityA := sensor.SensorEntity{
		ID:        uuid.New().String(),
		GatewayID: targetGatewayId.String(),
		Name:      "Heart monitor",
		Interval:  1200,
		Profile:   string(sensorProfile.HEART_RATE),
		Status:    string(sensor.Active),
	}
	entityB := sensor.SensorEntity{
		ID:        uuid.New().String(),
		GatewayID: targetGatewayId.String(),
		Name:      "Pulse",
		Interval:  2000,
		Profile:   string(sensorProfile.PULSE_OXIMETER),
		Status:    string(sensor.Inactive),
	}

	expectedSensors := []sensor.Sensor{
		{
			Id:        uuid.MustParse(entityA.ID),
			Name:      entityA.Name,
			Interval:  time.Duration(entityA.Interval),
			Profile:   sensorProfile.SensorProfile(entityA.Profile),
			GatewayId: uuid.MustParse(entityA.GatewayID),
			Status:    sensor.SensorStatus(entityA.Status),
		},
		{
			Id:        uuid.MustParse(entityB.ID),
			Name:      entityB.Name,
			Interval:  time.Duration(entityB.Interval),
			Profile:   sensorProfile.SensorProfile(entityB.Profile),
			GatewayId: uuid.MustParse(entityB.GatewayID),
			Status:    sensor.SensorStatus(entityB.Status),
		},
	}
	expectedCount := uint(2)

	stepGetSensorsByGatewayNeverCalled := func(mockBundle dbSensorAdapterMocks) *gomock.Call {
		return mockBundle.databaseRepo.EXPECT().
			GetSensorsByGatewayId(gomock.Any(), gomock.Any(), gomock.Any()).
			Times(0)
	}

	mockGetSensorsByGatewayErr := errors.New("unexpected error getting sensors by gateway")
	stepGetSensorsByGatewayErr := func(mockBundle dbSensorAdapterMocks) *gomock.Call {
		return mockBundle.databaseRepo.EXPECT().
			GetSensorsByGatewayId(targetGatewayId.String(), 10, 5).
			Return(nil, uint(0), mockGetSensorsByGatewayErr).
			Times(1)
	}

	stepGetSensorsByGatewayOk := func(mockBundle dbSensorAdapterMocks) *gomock.Call {
		return mockBundle.databaseRepo.EXPECT().
			GetSensorsByGatewayId(targetGatewayId.String(), 2, 2).
			Return([]sensor.SensorEntity{entityA, entityB}, expectedCount, nil).
			Times(1)
	}

	type testCase struct {
		name            string
		page            int
		limit           int
		setupSteps      []mockSetupFunc_DbSensorAdapter
		expectedSensors []sensor.Sensor
		expectedCount   uint
		expectedError   error
	}

	cases := []testCase{
		{
			name:  "Fail: invalid page (page < 1)",
			page:  0,
			limit: 10,
			setupSteps: []mockSetupFunc_DbSensorAdapter{
				stepGetSensorsByGatewayNeverCalled,
			},
			expectedError: pagination.ErrInvalidPage,
		},
		{
			name:  "Fail: invalid limit (limit < 1)",
			page:  1,
			limit: 0,
			setupSteps: []mockSetupFunc_DbSensorAdapter{
				stepGetSensorsByGatewayNeverCalled,
			},
			expectedError: pagination.ErrInvalidLimit,
		},
		{
			name:  "Fail: repository error",
			page:  3,
			limit: 5,
			setupSteps: []mockSetupFunc_DbSensorAdapter{
				stepGetSensorsByGatewayErr,
			},
			expectedError: mockGetSensorsByGatewayErr,
		},
		{
			name:  "Success",
			page:  2,
			limit: 2,
			setupSteps: []mockSetupFunc_DbSensorAdapter{
				stepGetSensorsByGatewayOk,
			},
			expectedSensors: expectedSensors,
			expectedCount:   expectedCount,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			adapter := setupDbSensorAdapter(t, tc.setupSteps)

			sensors, count, err := adapter.GetSensorsByGatewayId(targetGatewayId, tc.page, tc.limit)

			if tc.expectedError != nil {
				if !errors.Is(err, tc.expectedError) {
					t.Fatalf("expected error %v, got %v", tc.expectedError, err)
				}
				if sensors != nil {
					t.Fatalf("expected nil sensors on error, got %+v", sensors)
				}
				if count != 0 {
					t.Fatalf("expected zero count on error, got %d", count)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if !reflect.DeepEqual(tc.expectedSensors, sensors) {
				t.Fatalf("unexpected sensors. expected %+v, got %+v", tc.expectedSensors, sensors)
			}
			if count != tc.expectedCount {
				t.Fatalf("unexpected count. expected %d, got %d", tc.expectedCount, count)
			}
		})
	}
}

func TestDbSensorAdapter_GetSensorById(t *testing.T) {
	targetSensorId := uuid.New()
	targetGatewayId := uuid.New()

	entity := sensor.SensorEntity{
		ID:        targetSensorId.String(),
		GatewayID: targetGatewayId.String(),
		Name:      "ECG",
		Interval:  1500,
		Profile:   string(sensorProfile.ECG_CUSTOM),
		Status:    string(sensor.Active),
	}

	expectedSensor := sensor.Sensor{
		Id:        uuid.MustParse(entity.ID),
		Name:      entity.Name,
		Interval:  time.Duration(entity.Interval),
		Profile:   sensorProfile.SensorProfile(entity.Profile),
		GatewayId: uuid.MustParse(entity.GatewayID),
		Status:    sensor.SensorStatus(entity.Status),
	}

	mockGetSensorByIdErr := errors.New("unexpected error getting sensor by id")
	stepGetSensorByIdErr := func(mockBundle dbSensorAdapterMocks) *gomock.Call {
		return mockBundle.databaseRepo.EXPECT().
			GetSensorById(targetSensorId.String()).
			Return(sensor.SensorEntity{}, mockGetSensorByIdErr).
			Times(1)
	}

	stepGetSensorByIdOk := func(mockBundle dbSensorAdapterMocks) *gomock.Call {
		return mockBundle.databaseRepo.EXPECT().
			GetSensorById(targetSensorId.String()).
			Return(entity, nil).
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
			name: "Fail: repository error",
			setupSteps: []mockSetupFunc_DbSensorAdapter{
				stepGetSensorByIdErr,
			},
			expectedError: mockGetSensorByIdErr,
		},
		{
			name: "Success",
			setupSteps: []mockSetupFunc_DbSensorAdapter{
				stepGetSensorByIdOk,
			},
			expectedSensor: expectedSensor,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			adapter := setupDbSensorAdapter(t, tc.setupSteps)

			sensorFound, err := adapter.GetSensorById(targetSensorId)

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

func TestDbSensorAdapter_GetSensorsByTenant(t *testing.T) {
	targetTenantId := uuid.New()

	entityA := sensor.SensorEntity{
		ID:        uuid.New().String(),
		GatewayID: uuid.New().String(),
		Name:      "Thermometer",
		Interval:  1000,
		Profile:   string(sensorProfile.HEALTH_THERMOMETER),
		Status:    string(sensor.Active),
	}
	entityB := sensor.SensorEntity{
		ID:        uuid.New().String(),
		GatewayID: uuid.New().String(),
		Name:      "Environment",
		Interval:  3000,
		Profile:   string(sensorProfile.ENVIRONMENTAL_SENSING),
		Status:    string(sensor.Inactive),
	}

	expectedSensors := []sensor.Sensor{
		{
			Id:        uuid.MustParse(entityA.ID),
			Name:      entityA.Name,
			Interval:  time.Duration(entityA.Interval),
			Profile:   sensorProfile.SensorProfile(entityA.Profile),
			GatewayId: uuid.MustParse(entityA.GatewayID),
			Status:    sensor.SensorStatus(entityA.Status),
		},
		{
			Id:        uuid.MustParse(entityB.ID),
			Name:      entityB.Name,
			Interval:  time.Duration(entityB.Interval),
			Profile:   sensorProfile.SensorProfile(entityB.Profile),
			GatewayId: uuid.MustParse(entityB.GatewayID),
			Status:    sensor.SensorStatus(entityB.Status),
		},
	}
	expectedCount := uint(2)

	stepGetSensorsByTenantNeverCalled := func(mockBundle dbSensorAdapterMocks) *gomock.Call {
		return mockBundle.databaseRepo.EXPECT().
			GetSensorsByTenantId(gomock.Any(), gomock.Any(), gomock.Any()).
			Times(0)
	}

	mockGetSensorsByTenantErr := errors.New("unexpected error getting sensors by tenant")
	stepGetSensorsByTenantErr := func(mockBundle dbSensorAdapterMocks) *gomock.Call {
		return mockBundle.databaseRepo.EXPECT().
			GetSensorsByTenantId(targetTenantId.String(), 4, 4).
			Return(nil, uint(0), mockGetSensorsByTenantErr).
			Times(1)
	}

	stepGetSensorsByTenantOk := func(mockBundle dbSensorAdapterMocks) *gomock.Call {
		return mockBundle.databaseRepo.EXPECT().
			GetSensorsByTenantId(targetTenantId.String(), 5, 5).
			Return([]sensor.SensorEntity{entityA, entityB}, expectedCount, nil).
			Times(1)
	}

	type testCase struct {
		name            string
		page            int
		limit           int
		setupSteps      []mockSetupFunc_DbSensorAdapter
		expectedSensors []sensor.Sensor
		expectedCount   uint
		expectedError   error
	}

	cases := []testCase{
		{
			name:  "Fail: invalid page (page < 1)",
			page:  0,
			limit: 10,
			setupSteps: []mockSetupFunc_DbSensorAdapter{
				stepGetSensorsByTenantNeverCalled,
			},
			expectedError: pagination.ErrInvalidPage,
		},
		{
			name:  "Fail: invalid limit (limit < 1)",
			page:  1,
			limit: 0,
			setupSteps: []mockSetupFunc_DbSensorAdapter{
				stepGetSensorsByTenantNeverCalled,
			},
			expectedError: pagination.ErrInvalidLimit,
		},
		{
			name:  "Fail: repository error",
			page:  2,
			limit: 4,
			setupSteps: []mockSetupFunc_DbSensorAdapter{
				stepGetSensorsByTenantErr,
			},
			expectedError: mockGetSensorsByTenantErr,
		},
		{
			name:  "Success",
			page:  2,
			limit: 5,
			setupSteps: []mockSetupFunc_DbSensorAdapter{
				stepGetSensorsByTenantOk,
			},
			expectedSensors: expectedSensors,
			expectedCount:   expectedCount,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			adapter := setupDbSensorAdapter(t, tc.setupSteps)

			sensors, count, err := adapter.GetSensorsByTenant(targetTenantId, tc.page, tc.limit)

			if tc.expectedError != nil {
				if !errors.Is(err, tc.expectedError) {
					t.Fatalf("expected error %v, got %v", tc.expectedError, err)
				}
				if sensors != nil {
					t.Fatalf("expected nil sensors on error, got %+v", sensors)
				}
				if count != 0 {
					t.Fatalf("expected zero count on error, got %d", count)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if !reflect.DeepEqual(tc.expectedSensors, sensors) {
				t.Fatalf("unexpected sensors. expected %+v, got %+v", tc.expectedSensors, sensors)
			}
			if count != tc.expectedCount {
				t.Fatalf("unexpected count. expected %d, got %d", tc.expectedCount, count)
			}
		})
	}
}
