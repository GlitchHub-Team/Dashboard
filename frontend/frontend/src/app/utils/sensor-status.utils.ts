import { EnumMapper } from './enum.utils';
import { SensorStatus } from '../models/sensor-status.enum';

export const sensorStatusMapper = new EnumMapper<SensorStatus, string>(
  {
    [SensorStatus.ACTIVE]: 'active',
    [SensorStatus.INACTIVE]: 'inactive',
  },
  SensorStatus.INACTIVE,
);
