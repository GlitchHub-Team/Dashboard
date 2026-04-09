import { EnumMapper } from './enum.utils';
import { SensorProfiles } from '../models/sensor/sensor-profiles.enum';

export const sensorProfilesMapper = new EnumMapper<SensorProfiles, string>(
  {
    [SensorProfiles.HEART_RATE_SERVICE]: 'heart_rate',
    [SensorProfiles.PULSE_OXIMETER_SERVICE]: 'pulse_oximeter',
    [SensorProfiles.CUSTOM_ECG_SERVICE]: 'ecg_custom',
    [SensorProfiles.HEALTH_THERMOMETER_SERVICE]: 'health_thermometer',
    [SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE]: 'environmental_sensing',
  },
  SensorProfiles.HEART_RATE_SERVICE,
);
