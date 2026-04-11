import { SensorReading } from './sensor-reading.model';
import { FieldDescriptor } from './field-descriptor.model';

export interface HistoricReadings {
  dataCount: number;
  readings: SensorReading[];
  fields: FieldDescriptor[];
  samplesPerPacket?: number;
}
