import { HistoricReadings } from '../../models/sensor-data/historic-readings.model';
import { SensorReading } from '../../models/sensor-data/sensor-reading.model';
import { HistoricResponse } from '../../models/sensor-data/historic-response.model';

export abstract class SensorHistoricAdapter {
  abstract fromDTO(value: number, timestamp: number): SensorReading;
  abstract fromResponse(response: HistoricResponse): HistoricReadings;
}
