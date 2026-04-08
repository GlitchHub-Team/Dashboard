import { SensorLiveReadingAdapter } from './sensor-live-reading.adapter';
import { RealTimeReading } from '../../models/sensor-data/real-time-reading.model';
import { SensorReading } from '../../models/sensor-data/sensor-reading.model';
import { FieldDescriptor } from '../../models/sensor-data/field-descriptor.model';
import { PULSE_OXIMETER_FIELDS } from '../../models/sensor-data/sensor-fields.model';

export class PulseOximeterLiveAdapter extends SensorLiveReadingAdapter {
  readonly fields: FieldDescriptor[] = PULSE_OXIMETER_FIELDS;

  fromDTO(dto: RealTimeReading): SensorReading[] {
    return [
      {
        timestamp: new Date(dto.timestamp).toISOString(),
        value: {
          spo2: dto.data['Spo2Value'],
          pulseRate: dto.data['PulseRateValue'],
        },
      },
    ];
  }
}