import { SensorReading } from '../../models/sensor-data/sensor-reading.model';
import { RealTimeReading } from '../../models/sensor-data/real-time-reading.model';

export abstract class SensorLiveReadingAdapter {
  abstract fromDTO(dto: RealTimeReading): SensorReading;
}
