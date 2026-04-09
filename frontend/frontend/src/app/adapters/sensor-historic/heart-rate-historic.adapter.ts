import { SensorHistoricAdapter } from './sensor-historic.adapter';
import { HistoricResponse } from '../../models/sensor-data/historic-response.model';
import { HistoricReadings } from '../../models/sensor-data/historic-readings.model';
import { SensorReading } from '../../models/sensor-data/sensor-reading.model';
import { FieldDescriptor } from '../../models/sensor-data/field-descriptor.model';
import { HEART_RATE_FIELDS } from '../../models/sensor-data/sensor-fields.model';

export class HeartRateHistoricAdapter extends SensorHistoricAdapter {
  readonly fields: FieldDescriptor[] = HEART_RATE_FIELDS;

  fromResponse(response: HistoricResponse): HistoricReadings {
    const readings: SensorReading[] = response.samples.map((point) => ({
      timestamp: new Date(point.timestamp).toISOString(),
      value: {
        bpm: point.data['BpmValue'],
      },
    }));

    return {
      dataCount: readings.length,
      readings,
      fields: this.fields,
    };
  }
}