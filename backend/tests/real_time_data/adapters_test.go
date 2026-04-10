package real_time_data_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"backend/internal/real_time_data"
	"backend/internal/sensor"
	sensorProfile "backend/internal/sensor/profile"
	"backend/tests/real_time_data/mocks"
)

func TestRealTimeDataNATSAdapter_StartDataRetriever(t *testing.T) {
	targetTenantId := uuid.New()
	targetGatewayId := uuid.New()
	targetSensorId := uuid.New()
	targetProfile := sensorProfile.HEART_RATE

	targetSensor := sensor.Sensor{
		Id:        targetSensorId,
		GatewayId: targetGatewayId,
		Profile:   targetProfile,
	}

	expectedSubject := fmt.Sprintf("sensor.%s.%s.%s", targetTenantId.String(), targetGatewayId.String(), targetSensorId.String())

	dataChan := make(chan real_time_data.RealTimeSample, 1)
	errChan := make(chan real_time_data.RealTimeError, 1)

	type testCase struct {
		name           string
		tenantId       uuid.UUID
		sensorObj      sensor.Sensor
		setupFunc      func(mockReader *mocks.MockRealTimeDataNATSReader, wg *sync.WaitGroup)
		expectedErrStr string
	}

	cases := []testCase{
		{
			name:      "Success: Computes subject and starts subscriber asynchronously",
			tenantId:  targetTenantId,
			sensorObj: targetSensor,
			setupFunc: func(mockReader *mocks.MockRealTimeDataNATSReader, wg *sync.WaitGroup) {
				wg.Add(1)
				mockReader.EXPECT().
					StartSubscriber(expectedSubject, targetProfile, dataChan, errChan).
					DoAndReturn(func(subject string, profile sensorProfile.SensorProfile, dChan chan real_time_data.RealTimeSample, eChan chan real_time_data.RealTimeError) error {
						defer wg.Done()
						return nil
					}).
					Times(1)
			},
			expectedErrStr: "",
		},
		{
			name:      "Fail: Nil tenant ID returns ErrSensorNotFound immediately",
			tenantId:  uuid.Nil,
			sensorObj: targetSensor,
			setupFunc: func(mockReader *mocks.MockRealTimeDataNATSReader, wg *sync.WaitGroup) {
				// Se il tenantId è nil, allora StartSubscriber non viene mai chiamata perché getSubject dà errore
				mockReader.EXPECT().
					StartSubscriber(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			expectedErrStr: sensor.ErrSensorNotFound.Error(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockReader := mocks.NewMockRealTimeDataNATSReader(ctrl)

			var wg sync.WaitGroup
			if tc.setupFunc != nil {
				tc.setupFunc(mockReader, &wg)
			}

			adapter := real_time_data.NewRealTimeDataNATSAdapter(mockReader)
			err := adapter.StartDataRetriever(tc.tenantId, tc.sensorObj, dataChan, errChan)

			if tc.expectedErrStr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.expectedErrStr)
				}
				if err.Error() != tc.expectedErrStr {
					t.Errorf("expected error %q, got %q", tc.expectedErrStr, err.Error())
				}
			} else if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			// Attendi finché la goroutine non ha finito di eseguire il mock. Assicura che la chiamata asincrona
			// sia chiusa prima della fine del test.
			wg.Wait()
		})
	}
}
