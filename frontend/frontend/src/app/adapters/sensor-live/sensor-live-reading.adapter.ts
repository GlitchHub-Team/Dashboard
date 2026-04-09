import { SensorReading } from '../../models/sensor-data/sensor-reading.model';
import { RealTimeReading } from '../../models/sensor-data/real-time-reading.model';
import { FieldDescriptor } from '../../models/sensor-data/field-descriptor.model';

export abstract class SensorLiveReadingAdapter {
  abstract readonly fields: FieldDescriptor[];
  abstract fromDTO(dto: RealTimeReading): SensorReading[];
}
