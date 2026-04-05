import { FieldDescriptor } from './field-descriptor.model';

export const HEART_RATE_FIELDS: FieldDescriptor[] = [
  { key: 'bpm', label: 'Heart Rate', unit: 'bpm' },
];

export const PULSE_OXIMETER_FIELDS: FieldDescriptor[] = [
  { key: 'spo2', label: 'Blood Oxygen', unit: '%' },
  { key: 'pulseRate', label: 'Pulse Rate', unit: 'bpm' },
];

export const ENVIRONMENTAL_FIELDS: FieldDescriptor[] = [
  { key: 'temperature', label: 'Temperature', unit: '°C' },
  { key: 'humidity', label: 'Humidity', unit: '%' },
];

export const HEALTH_THERMOMETER_FIELDS: FieldDescriptor[] = [
  { key: 'temperature', label: 'Body Temperature', unit: '°C' },
];

export const ECG_FIELDS: FieldDescriptor[] = [
  { key: 'ecg', label: 'ECG Waveform', unit: 'mV' },
];