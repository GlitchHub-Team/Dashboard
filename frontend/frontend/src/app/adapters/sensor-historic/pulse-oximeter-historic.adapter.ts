import { SensorHistoricAdapter } from './sensor-historic.adapter';
import { HistoricResponse } from '../../models/sensor-data/historic-response.model';
import { HistoricReadings } from '../../models/sensor-data/historic-readings.model';
import { SensorReading } from '../../models/sensor-data/sensor-reading.model';
import { FieldDescriptor } from '../../models/sensor-data/field-descriptor.model';
import { PULSE_OXIMETER_FIELDS } from '../../models/sensor-data/sensor-fields.model';

export class PulseOximeterHistoricAdapter extends SensorHistoricAdapter {
  readonly fields: FieldDescriptor[] = PULSE_OXIMETER_FIELDS;

  fromResponse(response: HistoricResponse): HistoricReadings {
    const readings: SensorReading[] = response.samples.map((point) => ({
      timestamp: new Date(point.timestamp).toISOString(),
      value: {
        spo2: point.data['Spo2Value'],
        pulseRate: point.data['PulseRateValue'],
      },
    }));

    return {
      dataCount: readings.length,
      readings,
      fields: this.fields,
    };
  }
}