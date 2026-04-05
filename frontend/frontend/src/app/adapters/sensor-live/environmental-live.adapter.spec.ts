import { describe, it, expect } from 'vitest';
import { EnvironmentalLiveAdapter } from './environmental-live.adapter';
import { ENVIRONMENTAL_FIELDS } from '../../models/sensor-data/sensor-fields.model';
import { RealTimeReading } from '../../models/sensor-data/real-time-reading.model';

describe('EnvironmentalLiveAdapter', () => {
  const adapter = new EnvironmentalLiveAdapter();

  const dto: RealTimeReading = {
    sensor_id: 's1',
    gateway_id: 'gw1',
    tenant_id: 't1',
    timestamp: '2024-06-15T12:00:00.000Z',
    profile: 'environmental_sensing',
    data: { TemperatureValue: 22.5, HumidityValue: 60 },
  };

  it('should expose ENVIRONMENTAL_FIELDS', () => {
    expect(adapter.fields).toBe(ENVIRONMENTAL_FIELDS);
  });

  describe('fromDTO', () => {
    it('should return a single reading', () => {
      expect(adapter.fromDTO(dto)).toHaveLength(1);
    });

    it.each([
      ['timestamp', new Date(dto.timestamp).toISOString()],
      ['value.temperature', 22.5],
      ['value.humidity', 60],
    ] as const)('should map %s correctly', (path, expected) => {
      const [reading] = adapter.fromDTO(dto);
      const value = path.split('.').reduce((o: any, k) => o[k], reading);
      expect(value).toBe(expected);
    });

    it('should reflect updated data values', () => {
      const [reading] = adapter.fromDTO({ ...dto, data: { TemperatureValue: 30, HumidityValue: 80 } });
      expect(reading.value).toEqual({ temperature: 30, humidity: 80 });
    });
  });
});
