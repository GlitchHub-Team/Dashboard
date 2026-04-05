import { describe, it, expect } from 'vitest';
import { HeartRateLiveAdapter } from './heart-rate-live.adapter';
import { HEART_RATE_FIELDS } from '../../models/sensor-data/sensor-fields.model';
import { RealTimeReading } from '../../models/sensor-data/real-time-reading.model';

describe('HeartRateLiveAdapter', () => {
  const adapter = new HeartRateLiveAdapter();

  const dto: RealTimeReading = {
    sensor_id: 's1',
    gateway_id: 'gw1',
    tenant_id: 't1',
    timestamp: '2024-05-20T10:00:00.000Z',
    profile: 'heart_rate',
    data: { BpmValue: 75 },
  };

  it('should expose HEART_RATE_FIELDS', () => {
    expect(adapter.fields).toBe(HEART_RATE_FIELDS);
  });

  describe('fromDTO', () => {
    it('should return a single reading', () => {
      expect(adapter.fromDTO(dto)).toHaveLength(1);
    });

    it.each([
      ['timestamp', new Date(dto.timestamp).toISOString()],
      ['value.bpm', 75],
    ] as const)('should map %s correctly', (path, expected) => {
      const [reading] = adapter.fromDTO(dto);
      const value = path.split('.').reduce((o: any, k) => o[k], reading);
      expect(value).toBe(expected);
    });

    it('should reflect updated bpm value', () => {
      const [reading] = adapter.fromDTO({ ...dto, data: { BpmValue: 110 } });
      expect(reading.value).toEqual({ bpm: 110 });
    });
  });
});
