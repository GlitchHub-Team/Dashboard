import { SensorReading } from '../../models/sensor-data/sensor-reading.model';

export abstract class SensorLiveReadingAdapter {
  abstract fromDTO(dto: unknown): SensorReading;
}
