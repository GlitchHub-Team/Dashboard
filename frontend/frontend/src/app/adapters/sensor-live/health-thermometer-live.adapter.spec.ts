import { describe, it, expect } from 'vitest';
import { HealthThermometerLiveAdapter } from './health-thermometer-live.adapter';
import { HEALTH_THERMOMETER_FIELDS } from '../../models/sensor-data/sensor-fields.model';
import { RealTimeReading } from '../../models/sensor-data/real-time-reading.model';

describe('HealthThermometerLiveAdapter', () => {
  const adapter = new HealthThermometerLiveAdapter();

  const dto: RealTimeReading = {
    sensor_id: 's1',
    gateway_id: 'gw1',
    tenant_id: 't1',
    timestamp: '2024-03-10T08:30:00.000Z',
    profile: 'health_thermometer',
    data: { TemperatureValue: 36.6 },
  };

  it('should expose HEALTH_THERMOMETER_FIELDS', () => {
    expect(adapter.fields).toBe(HEALTH_THERMOMETER_FIELDS);
  });

  describe('fromDTO', () => {
    it('should return a single reading', () => {
      expect(adapter.fromDTO(dto)).toHaveLength(1);
    });

    it.each([
      ['timestamp', new Date(dto.timestamp).toISOString()],
      ['value.temperature', 36.6],
    ] as const)('should map %s correctly', (path, expected) => {
      const [reading] = adapter.fromDTO(dto);
      const value = path.split('.').reduce((o: any, k) => o[k], reading);
      expect(value).toBe(expected);
    });

    it('should reflect updated temperature value', () => {
      const [reading] = adapter.fromDTO({ ...dto, data: { TemperatureValue: 38.2 } });
      expect(reading.value).toEqual({ temperature: 38.2 });
    });
  });
});
