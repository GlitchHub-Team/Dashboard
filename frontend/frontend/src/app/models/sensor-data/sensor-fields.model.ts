import { FieldDescriptor } from './field-descriptor.model';

export const HEART_RATE_FIELDS: FieldDescriptor[] = [
  { key: 'bpm', label: 'Battito cardiaco', unit: 'bpm' },
];

export const PULSE_OXIMETER_FIELDS: FieldDescriptor[] = [
  { key: 'spo2', label: 'Ossigeno nel sangue', unit: '%' },
  { key: 'pulseRate', label: 'Frequenza cardiaca', unit: 'bpm' },
];

export const ENVIRONMENTAL_FIELDS: FieldDescriptor[] = [
  { key: 'temperature', label: 'Temperatura', unit: '°C' },
  { key: 'humidity', label: 'Umidità', unit: '%' },
];

export const HEALTH_THERMOMETER_FIELDS: FieldDescriptor[] = [
  { key: 'temperature', label: 'Temperatura corporea', unit: '°C' },
];

export const ECG_FIELDS: FieldDescriptor[] = [
  { key: 'ecg', label: "Forma d'onda ECG", unit: 'mV' },
];
