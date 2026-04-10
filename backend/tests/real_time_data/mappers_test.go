package real_time_data_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	httpDto "backend/internal/infra/transport/http/dto"
	"backend/internal/real_time_data"
	sensorProfile "backend/internal/sensor/profile"
)

func TestMapNATSRawToDomain(t *testing.T) {
	validTimestamp := time.Now().UTC()
	validTimeStr := validTimestamp.Format(time.RFC3339Nano)

	type testCase struct {
		name        string
		profile     sensorProfile.SensorProfile
		rawJSON     string
		expectErr   bool
		errContains string
		checkFunc   func(t *testing.T, res real_time_data.RealTimeSample)
	}

	cases := []testCase{
		{
			name:      "Success: ECG_CUSTOM parses to ECGSample",
			profile:   sensorProfile.ECG_CUSTOM,
			rawJSON:   `{"timestamp":"` + validTimeStr + `","data":{"waveform":[1,2,3,4]}}`,
			expectErr: false,
			checkFunc: func(t *testing.T, res real_time_data.RealTimeSample) {
				sample, ok := res.(*real_time_data.ECGSample)
				if !ok {
					t.Fatalf("expected type *ECGSample, got %T", res)
				}
				if sample.Profile != sensorProfile.ECG_CUSTOM {
					t.Errorf("expected profile %v, got %v", sensorProfile.ECG_CUSTOM, sample.Profile)
				}
				if len(sample.Data.Waveform) != 4 {
					t.Errorf("expected waveform length 4, got %d", len(sample.Data.Waveform))
				}
				if sample.Data.Waveform[0] != 1 {
					t.Errorf("expected waveform index 0 to be 1, got %d", sample.Data.Waveform[0])
				}
			},
		},
		{
			name:      "Success: ENVIRONMENTAL_SENSING parses to EnvironmentalSensingSample",
			profile:   sensorProfile.ENVIRONMENTAL_SENSING,
			rawJSON:   `{"timestamp":"` + validTimeStr + `","data":{"temperatureValue":25.5,"humidityValue":60.0,"pressureValue":1013.25}}`,
			expectErr: false,
			checkFunc: func(t *testing.T, res real_time_data.RealTimeSample) {
				sample, ok := res.(*real_time_data.EnvironmentalSensingSample)
				if !ok {
					t.Fatalf("expected type *EnvironmentalSensingSample, got %T", res)
				}
				if sample.Data.Temperature != 25.5 {
					t.Errorf("expected temperature 25.5, got %f", sample.Data.Temperature)
				}
				if sample.Data.Humidity != 60.0 {
					t.Errorf("expected humidity 60.0, got %f", sample.Data.Humidity)
				}
				if sample.Data.Pressure != 1013.25 {
					t.Errorf("expected pressure 1013.25, got %f", sample.Data.Pressure)
				}
			},
		},
		{
			name:      "Success: HEALTH_THERMOMETER parses to HealthThermometerSample",
			profile:   sensorProfile.HEALTH_THERMOMETER,
			rawJSON:   `{"timestamp":"` + validTimeStr + `","data":{"temperatureValue":36.6}}`,
			expectErr: false,
			checkFunc: func(t *testing.T, res real_time_data.RealTimeSample) {
				sample, ok := res.(*real_time_data.HealthThermometerSample)
				if !ok {
					t.Fatalf("expected type *HealthThermometerSample, got %T", res)
				}
				if sample.Data.Temperature != 36.6 {
					t.Errorf("expected temperature 36.6, got %f", sample.Data.Temperature)
				}
			},
		},
		{
			name:      "Success: HEART_RATE parses to HeartRateSample",
			profile:   sensorProfile.HEART_RATE,
			rawJSON:   `{"timestamp":"` + validTimeStr + `","data":{"bpmValue":75}}`,
			expectErr: false,
			checkFunc: func(t *testing.T, res real_time_data.RealTimeSample) {
				sample, ok := res.(*real_time_data.HeartRateSample)
				if !ok {
					t.Fatalf("expected type *HeartRateSample, got %T", res)
				}
				if sample.Data.BpmValue != 75 {
					t.Errorf("expected BPM 75, got %d", sample.Data.BpmValue)
				}
			},
		},
		{
			name:      "Success: PULSE_OXIMETER parses to PulseOximeterSample",
			profile:   sensorProfile.PULSE_OXIMETER,
			rawJSON:   `{"timestamp":"` + validTimeStr + `","data":{"spo2Value":98.5,"pulseRateValue":70}}`,
			expectErr: false,
			checkFunc: func(t *testing.T, res real_time_data.RealTimeSample) {
				sample, ok := res.(*real_time_data.PulseOximeterSample)
				if !ok {
					t.Fatalf("expected type *PulseOximeterSample, got %T", res)
				}
				if sample.Data.Spo2 != 98.5 {
					t.Errorf("expected SpO2 98.5, got %f", sample.Data.Spo2)
				}
				if sample.Data.PulseRate != 70 {
					t.Errorf("expected PulseRate 70, got %d", sample.Data.PulseRate)
				}
			},
		},
		{
			name:        "Fail: Invalid JSON breaks MapRawToDataSampleNATSDto mapping",
			profile:     sensorProfile.HEART_RATE,
			rawJSON:     `{"timestamp":"` + validTimeStr + `", data: malformed}`,
			expectErr:   true,
			errContains: "",
		},
		{
			name:        "Fail: Timestamp parsing fails due to invalid RFC3339Nano format",
			profile:     sensorProfile.HEART_RATE,
			rawJSON:     `{"timestamp":"2026-04-09_INVALID_TIME","data":{"bpmValue":75}}`,
			expectErr:   true,
			errContains: "cannot parse",
		},
		{
			name:        "Fail: Unknown profile returns ErrUnknownProfile",
			profile:     sensorProfile.SensorProfile("UNKNOWN_PROFILE"),
			rawJSON:     `{"timestamp":"` + validTimeStr + `","data":{}}`,
			expectErr:   true,
			errContains: sensorProfile.ErrUnknownProfile.Error(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rawMsg := json.RawMessage(tc.rawJSON)
			res, err := real_time_data.MapNATSRawToDomain(tc.profile, rawMsg)

			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected an error, got nil")
				}
				if tc.errContains != "" && !strings.Contains(err.Error(), tc.errContains) {
					t.Errorf("expected error to contain %q, got %q", tc.errContains, err.Error())
				}
				if res != nil {
					t.Errorf("expected nil result on error, got %v", res)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if res == nil {
					t.Fatalf("expected valid RealTimeSample result, got nil")
				}
				if tc.checkFunc != nil {
					tc.checkFunc(t, res)
				}
			}
		})
	}
}

func TestMapDomainToWSDto(t *testing.T) {
	targetTime := time.Now().UTC()
	expectedTimeStr := targetTime.Format(time.RFC3339Nano)

	baseSample := func(profile sensorProfile.SensorProfile) real_time_data.BaseSample {
		return real_time_data.BaseSample{
			Profile:   profile,
			Timestamp: targetTime,
		}
	}

	type testCase struct {
		name            string
		inputSample     real_time_data.RealTimeSample
		expectedProfile string
		checkDataFunc   func(t *testing.T, data any)
	}

	cases := []testCase{
		{
			name: "Success: ECGSample maps to ECGData",
			inputSample: &real_time_data.ECGSample{
				BaseSample: baseSample(sensorProfile.ECG_CUSTOM),
				Data: real_time_data.EcgSampleData{
					Waveform: []int{10, 20, 30, 40},
				},
			},
			expectedProfile: "ECG",
			checkDataFunc: func(t *testing.T, data any) {
				dtoData, ok := data.(httpDto.ECGData)
				if !ok {
					t.Fatalf("expected type httpDto.ECGData, got %T", data)
				}

				// In mapper.go, la funzione fa casting di EcgSampleData verso httpDto.ECGData,
				// per cui è possibile rieseguire il cast al contrario (ECGData -> EcgSampleData) senza causare errori
				originalData := real_time_data.EcgSampleData(dtoData)
				if !reflect.DeepEqual(originalData.Waveform, []int{10, 20, 30, 40}) {
					t.Errorf("expected waveform to be []int{10, 20, 30, 40}, got %#v", originalData.Waveform)
				}
			},
		},
		{
			name: "Success: EnvironmentalSensingSample maps to EnvironmentalSensingData",
			inputSample: &real_time_data.EnvironmentalSensingSample{
				BaseSample: baseSample(sensorProfile.ENVIRONMENTAL_SENSING),
				Data: real_time_data.EnvironmentalSensingSampleData{
					Temperature: 22.5,
					Humidity:    50.0,
					Pressure:    1013.25,
				},
			},
			expectedProfile: "EnvironmentalSensing",
			checkDataFunc: func(t *testing.T, data any) {
				dtoData, ok := data.(httpDto.EnvironmentalSensingData)
				if !ok {
					t.Fatalf("expected type httpDto.EnvironmentalSensingData, got %T", data)
				}
				if dtoData.TemperatureValue != 22.5 {
					t.Errorf("expected temperature 22.5, got %f", dtoData.TemperatureValue)
				}
				if dtoData.HumidityValue != 50.0 {
					t.Errorf("expected humidity 50.0, got %f", dtoData.HumidityValue)
				}
				if dtoData.PressureValue != 1013.25 {
					t.Errorf("expected pressure 1013.25, got %f", dtoData.PressureValue)
				}
			},
		},
		{
			name: "Success: HealthThermometerSample maps to HealthThermometerData",
			inputSample: &real_time_data.HealthThermometerSample{
				BaseSample: baseSample(sensorProfile.HEALTH_THERMOMETER),
				Data: real_time_data.HealthThermometerSampleData{
					Temperature: 36.6,
				},
			},
			expectedProfile: "HealthThermometer",
			checkDataFunc: func(t *testing.T, data any) {
				dtoData, ok := data.(httpDto.HealthThermometerData)
				if !ok {
					t.Fatalf("expected type httpDto.HealthThermometerData, got %T", data)
				}
				if dtoData.TemperatureValue != 36.6 {
					t.Errorf("expected temperature 36.6, got %f", dtoData.TemperatureValue)
				}
			},
		},
		{
			name: "Success: HeartRateSample maps to HeartRateData",
			inputSample: &real_time_data.HeartRateSample{
				BaseSample: baseSample(sensorProfile.HEART_RATE),
				Data: real_time_data.HeartRateSampleData{
					BpmValue: 75,
				},
			},
			expectedProfile: "HeartRate",
			checkDataFunc: func(t *testing.T, data any) {
				dtoData, ok := data.(httpDto.HeartRateData)
				if !ok {
					t.Fatalf("expected type httpDto.HeartRateData, got %T", data)
				}
				if dtoData.BpmValue != 75 {
					t.Errorf("expected BPM 75, got %d", dtoData.BpmValue)
				}
			},
		},
		{
			name: "Success: PulseOximeterSample maps to PulseOximeterData",
			inputSample: &real_time_data.PulseOximeterSample{
				BaseSample: baseSample(sensorProfile.PULSE_OXIMETER),
				Data: real_time_data.PulseOximeterSampleData{
					Spo2:      98.5,
					PulseRate: 70,
				},
			},
			expectedProfile: "PulseOximeter",
			checkDataFunc: func(t *testing.T, data any) {
				dtoData, ok := data.(httpDto.PulseOximeterData)
				if !ok {
					t.Fatalf("expected type httpDto.PulseOximeterData, got %T", data)
				}
				if dtoData.Spo2Value != 98.5 {
					t.Errorf("expected SpO2 98.5, got %f", dtoData.Spo2Value)
				}
				if dtoData.PulseRateValue != 70 {
					t.Errorf("expected PulseRate 70, got %d", dtoData.PulseRateValue)
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resultDTO := real_time_data.MapDomainToWSDto(tc.inputSample)

			if resultDTO.Profile != tc.expectedProfile {
				t.Errorf("expected profile %q, got %q", tc.expectedProfile, resultDTO.Profile)
			}

			if resultDTO.Timestamp != expectedTimeStr {
				t.Errorf("expected timestamp %q, got %q", expectedTimeStr, resultDTO.Timestamp)
			}

			if tc.checkDataFunc != nil {
				tc.checkDataFunc(t, resultDTO.Data)
			}
		})
	}
}
