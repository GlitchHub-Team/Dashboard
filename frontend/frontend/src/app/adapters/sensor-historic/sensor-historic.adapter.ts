import { HistoricReadings } from '../../models/sensor-data/historic-readings.model';
import { SensorReading } from '../../models/sensor-data/sensor-reading.model';

export abstract class SensorHistoricAdapter {
  abstract fromDTO(value: number, timestamp: number): SensorReading;
  abstract fromResponse(response: unknown): HistoricReadings;
}
