package real_time_data

import (
	"encoding/json"
	"time"

	httpDto "backend/internal/infra/transport/http/dto"
	natsDto "backend/internal/infra/transport/nats/dto"
	sensorProfile "backend/internal/sensor/profile"
)

/*
Mappa un messaggio "raw" ricevuto da NATS a un RealTimeSample.
*/
func MapNATSRawToDomain(profile sensorProfile.SensorProfile, raw json.RawMessage) (RealTimeSample, error) {
	obj, err := natsDto.MapRawToDataSampleNATSDto(profile, raw)
	if err != nil {
		return nil, err
	}

	timestamp, err := time.Parse(time.RFC3339Nano, obj.GetTimestamp())
	if err != nil {
		return nil, err
	}

	base := BaseSample{
		Profile:   profile,
		Timestamp: timestamp,
	}

	// NOTA: tutti le type assertion di seguito possono essere eseguite senza il controllo ok poiché
	// natsDto.MapRawToDataSampleNATSDto() assicura che il tipo astratto sia concretizzato nei tipi specificati sotto
	switch profile {
	case sensorProfile.ECG_CUSTOM:
		dto := obj.(*natsDto.ConcreteDataSampleNATSDto[httpDto.ECGData])

		return &ECGSample{
			BaseSample: base,
			Data: EcgSampleData{
				Waveform: dto.Data.Waveform,
			},
		}, nil

	case sensorProfile.ENVIRONMENTAL_SENSING:
		dto := obj.(*natsDto.ConcreteDataSampleNATSDto[httpDto.EnvironmentalSensingData])

		return &EnvironmentalSensingSample{
			BaseSample: base,
			Data: EnvironmentalSensingSampleData{
				Temperature: dto.Data.TemperatureValue,
				Humidity:    dto.Data.HumidityValue,
				Pressure:    dto.Data.PressureValue,
			},
		}, nil

	case sensorProfile.HEALTH_THERMOMETER:
		dto := obj.(*natsDto.ConcreteDataSampleNATSDto[httpDto.HealthThermometerData])

		return &HealthThermometerSample{
			BaseSample: base,
			Data: HealthThermometerSampleData{
				Temperature: dto.Data.TemperatureValue,
			},
		}, nil

	case sensorProfile.HEART_RATE:
		dto := obj.(*natsDto.ConcreteDataSampleNATSDto[httpDto.HeartRateData])

		return &HeartRateSample{
			BaseSample: base,
			Data: HeartRateSampleData{
				BpmValue: dto.Data.BpmValue,
			},
		}, nil

	case sensorProfile.PULSE_OXIMETER:
		dto := obj.(*natsDto.ConcreteDataSampleNATSDto[httpDto.PulseOximeterData])

		return &PulseOximeterSample{
			BaseSample: base,
			Data: PulseOximeterSampleData{
				Spo2:      dto.Data.Spo2Value,
				PulseRate: dto.Data.PulseRateValue,
			},
		}, nil

	default:
		return nil, sensorProfile.ErrUnknownProfile
	}
}

func mapProfileToOutString(profile sensorProfile.SensorProfile) string {
	switch profile {
	case sensorProfile.ECG_CUSTOM:
		return "ECG"
	case sensorProfile.ENVIRONMENTAL_SENSING:
		return "EnvironmentalSensing"
	case sensorProfile.HEALTH_THERMOMETER:
		return "HealthThermometer"
	case sensorProfile.HEART_RATE:
		return "HeartRate"
	case sensorProfile.PULSE_OXIMETER:
		return "PulseOximeter"
	default:
		return ""
	}
}

/*
Mappa un oggetto RealTimeSample a un DTO in output per il client WS.
*/
func MapDomainToWSDto(sample RealTimeSample) RealTimeSampleOutDTO {
	var jsonData any
	sampleData := sample.GetData()

	switch sample.GetProfile() {
	case sensorProfile.ECG_CUSTOM:
		jsonData = httpDto.ECGData(sampleData.(EcgSampleData))

	case sensorProfile.ENVIRONMENTAL_SENSING:
		typedData := sampleData.(EnvironmentalSensingSampleData)

		jsonData = httpDto.EnvironmentalSensingData{
			TemperatureValue: typedData.Temperature,
			HumidityValue:    typedData.Humidity,
			PressureValue:    typedData.Pressure,
		}

	case sensorProfile.HEALTH_THERMOMETER:
		typedData := sampleData.(HealthThermometerSampleData)
		jsonData = httpDto.HealthThermometerData{
			TemperatureValue: typedData.Temperature,
		}

	case sensorProfile.HEART_RATE:
		typedData := sampleData.(HeartRateSampleData)
		jsonData = httpDto.HeartRateData{
			BpmValue: typedData.BpmValue,
		}

	case sensorProfile.PULSE_OXIMETER:
		typedData := sampleData.(PulseOximeterSampleData)
		jsonData = httpDto.PulseOximeterData{
			Spo2Value:      typedData.Spo2,
			PulseRateValue: typedData.PulseRate,
		}
	}

	return RealTimeSampleOutDTO{
		ProfileField:   natsDto.ProfileField{Profile: mapProfileToOutString(sample.GetProfile())},
		TimestampField: natsDto.TimestampField{Timestamp: sample.GetTimestamp().Format(time.RFC3339Nano)},
		Data:           jsonData,
	}
}
