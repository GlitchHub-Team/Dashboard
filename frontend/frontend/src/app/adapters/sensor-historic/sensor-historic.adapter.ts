import { HistoricReadings } from '../../models/sensor-data/historic-readings.model';
import { HistoricResponse } from '../../models/sensor-data/historic-response.model';
import { FieldDescriptor } from '../../models/sensor-data/field-descriptor.model';

export abstract class SensorHistoricAdapter {
  abstract readonly fields: FieldDescriptor[];
  abstract fromResponse(response: HistoricResponse): HistoricReadings;
}
