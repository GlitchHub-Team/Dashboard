import { EnumMapper } from './enum.utils';
import { Status } from '../models/gateway-sensor-status.enum';

export const statusMapper = new EnumMapper<Status, string>(
  {
    [Status.ACTIVE]: 'attivo',
    [Status.INACTIVE]: 'inattivo',
  },
  Status.INACTIVE,
);
