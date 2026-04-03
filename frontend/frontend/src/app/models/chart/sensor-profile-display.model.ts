import { SensorProfiles } from '../../models/sensor/sensor-profiles.enum';

interface SensorProfileDisplay {
  label: string;
  unit: string;
}

const SENSOR_PROFILE_MAP: Record<SensorProfiles, SensorProfileDisplay> = {
  [SensorProfiles.HEART_RATE_SERVICE]: { label: 'Heart Rate', unit: 'bpm' },
  [SensorProfiles.PULSE_OXIMETER_SERVICE]: { label: 'Pulse Oximeter', unit: '%SpO₂' },
  [SensorProfiles.CUSTOM_ECG_SERVICE]: { label: 'ECG', unit: 'mV' },
  [SensorProfiles.HEALTH_THERMOMETER_SERVICE]: { label: 'Thermometer', unit: '°C' },
  [SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE]: { label: 'Environmental', unit: '%' },
};

export function getSensorProfileDisplay(profile: SensorProfiles): SensorProfileDisplay {
  return (
    SENSOR_PROFILE_MAP[profile] ?? {
      label: profile,
      unit: '',
    }
  );
}
