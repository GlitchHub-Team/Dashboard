import { EnumMapper } from './enum.utils';
import { Status } from '../models/gateway-sensor-status.enum';

export const statusMapper = new EnumMapper<Status, string>(
  {
    [Status.ACTIVE]: 'active',
    [Status.INACTIVE]: 'inactive',
  },
  Status.INACTIVE,
);
