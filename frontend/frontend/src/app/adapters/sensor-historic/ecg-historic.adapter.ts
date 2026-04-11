import { SensorHistoricAdapter } from './sensor-historic.adapter';
import { HistoricResponse } from '../../models/sensor-data/historic-response.model';
import { HistoricReadings } from '../../models/sensor-data/historic-readings.model';
import { SensorReading } from '../../models/sensor-data/sensor-reading.model';
import { FieldDescriptor } from '../../models/sensor-data/field-descriptor.model';
import { ECG_FIELDS } from '../../models/sensor-data/sensor-fields.model';

export class EcgHistoricAdapter extends SensorHistoricAdapter {
  readonly fields: FieldDescriptor[] = ECG_FIELDS;

  fromResponse(response: HistoricResponse): HistoricReadings {
    const readings: SensorReading[] = [];

    for (const sample of response.samples) {
      const waveform = sample.data['Waveform'] as number[];
      const baseTime = new Date(sample.timestamp).getTime();

      for (let i = 0; i < waveform.length; i++) {
        readings.push({
          timestamp: new Date(baseTime + i * 1000 / waveform.length).toISOString(),
          value: {
            ecg: waveform[i],
          },
        });
      }
    }

    const firstWaveform = response.samples[0]?.data['Waveform'] as number[] | undefined;

    return {
      dataCount: readings.length,
      readings,
      fields: this.fields,
      samplesPerPacket: firstWaveform?.length,
    };
  }
}