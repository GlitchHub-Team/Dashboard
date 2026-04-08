import { SensorHistoricAdapter } from './sensor-historic.adapter';
import { HistoricResponse } from '../../models/sensor-data/historic-response.model';
import { HistoricReadings } from '../../models/sensor-data/historic-readings.model';
import { SensorReading } from '../../models/sensor-data/sensor-reading.model';
import { FieldDescriptor } from '../../models/sensor-data/field-descriptor.model';
import { HEALTH_THERMOMETER_FIELDS } from '../../models/sensor-data/sensor-fields.model';

export class HealthThermometerHistoricAdapter extends SensorHistoricAdapter {
  readonly fields: FieldDescriptor[] = HEALTH_THERMOMETER_FIELDS;

  fromResponse(response: HistoricResponse): HistoricReadings {
    const readings: SensorReading[] = response.samples.map((point) => ({
      timestamp: new Date(point.timestamp).toISOString(),
      value: {
        temperature: point.data['TemperatureValue'],
      },
    }));

    return {
      dataCount: readings.length,
      readings,
      fields: this.fields,
    };
  }
}