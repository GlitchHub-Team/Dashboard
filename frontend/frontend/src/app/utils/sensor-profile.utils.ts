import { EnumMapper } from './enum.utils';
import { SensorProfiles } from '../models/sensor/sensor-profiles.enum';

export const sensorProfilesMapper = new EnumMapper<SensorProfiles, string>(
  {
    [SensorProfiles.HEART_RATE_SERVICE]: 'heart_rate_service',
    [SensorProfiles.PULSE_OXIMETER_SERVICE]: 'pulse_oximeter_service',
    [SensorProfiles.CUSTOM_ECG_SERVICE]: 'custom_ecg_service',
    [SensorProfiles.HEALTH_THERMOMETER_SERVICE]: 'health_thermometer_service',
    [SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE]: 'environmental_sensing_service',
  },
  SensorProfiles.HEART_RATE_SERVICE,
);
