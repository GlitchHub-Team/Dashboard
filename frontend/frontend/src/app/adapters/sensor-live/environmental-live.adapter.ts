import { SensorLiveReadingAdapter } from './sensor-live-reading.adapter';
import { RealTimeReading } from '../../models/sensor-data/real-time-reading.model';
import { SensorReading } from '../../models/sensor-data/sensor-reading.model';
import { FieldDescriptor } from '../../models/sensor-data/field-descriptor.model';
import { ENVIRONMENTAL_FIELDS } from '../../models/sensor-data/sensor-fields.model';

export class EnvironmentalLiveAdapter extends SensorLiveReadingAdapter {
  readonly fields: FieldDescriptor[] = ENVIRONMENTAL_FIELDS;

  fromDTO(dto: RealTimeReading): SensorReading[] {
    return [
      {
        timestamp: new Date(dto.timestamp).toISOString(),
        value: {
          temperature: dto.data['TemperatureValue'],
          humidity: dto.data['HumidityValue'],
        },
      },
    ];
  }
}