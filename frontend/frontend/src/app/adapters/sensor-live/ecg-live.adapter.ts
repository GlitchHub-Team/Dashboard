import { SensorLiveReadingAdapter } from './sensor-live-reading.adapter';
import { RealTimeReading } from '../../models/sensor-data/real-time-reading.model';
import { SensorReading } from '../../models/sensor-data/sensor-reading.model';
import { FieldDescriptor } from '../../models/sensor-data/field-descriptor.model';
import { ECG_FIELDS } from '../../models/sensor-data/sensor-fields.model';

export class EcgLiveAdapter extends SensorLiveReadingAdapter {
  readonly fields: FieldDescriptor[] = ECG_FIELDS;

  fromDTO(dto: RealTimeReading): SensorReading[] {
    const waveform = dto.data['Waveform'] as number[];
    const baseTime = new Date(dto.timestamp).getTime();

    return waveform.map((value, i) => ({
      timestamp: new Date(baseTime + i * 1000 / waveform.length).toISOString(),
      value: {
        ecg: value,
      },
    }));
  }
}