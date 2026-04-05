import { SensorHistoricAdapter } from './sensor-historic.adapter';
import { HistoricResponse } from '../../models/sensor-data/historic-response.model';
import { HistoricReadings } from '../../models/sensor-data/historic-readings.model';
import { SensorReading } from '../../models/sensor-data/sensor-reading.model';
import { FieldDescriptor } from '../../models/sensor-data/field-descriptor.model';
import { ECG_FIELDS } from '../../models/sensor-data/sensor-fields.model';

export class EcgHistoricAdapter extends SensorHistoricAdapter {
  readonly fields: FieldDescriptor[] = ECG_FIELDS;

  // Sample rate pescato dai docs di APIDOG
  private readonly SAMPLE_RATE = 250; 
  private readonly SAMPLE_INTERVAL_MS = 1000 / this.SAMPLE_RATE;

  fromResponse(response: HistoricResponse): HistoricReadings {
    const readings: SensorReading[] = [];

    for (const sample of response.samples) {
      const waveform = sample.data['Waveform'] as number[];
      const baseTime = new Date(sample.timestamp).getTime();

      for (let i = 0; i < waveform.length; i++) {
        readings.push({
          timestamp: new Date(baseTime + i * this.SAMPLE_INTERVAL_MS).toISOString(),
          value: {
            ecg: waveform[i],
          },
        });
      }
    }

    return {
      dataCount: readings.length,
      readings,
      fields: this.fields,
    };
  }
}