import { describe, it, expect } from 'vitest';
import { PulseOximeterLiveAdapter } from './pulse-oximeter-live.adapter';
import { PULSE_OXIMETER_FIELDS } from '../../models/sensor-data/sensor-fields.model';
import { RealTimeReading } from '../../models/sensor-data/real-time-reading.model';

describe('PulseOximeterLiveAdapter', () => {
  const adapter = new PulseOximeterLiveAdapter();

  const dto: RealTimeReading = {
    sensor_id: 's1',
    gateway_id: 'gw1',
    tenant_id: 't1',
    timestamp: '2024-01-01T00:00:00.000Z',
    profile: 'pulse_oximeter',
    data: { Spo2Value: 98, PulseRateValue: 72 },
  };

  it('should expose PULSE_OXIMETER_FIELDS', () => {
    expect(adapter.fields).toBe(PULSE_OXIMETER_FIELDS);
  });

  describe('fromDTO', () => {
    it('should return a single reading', () => {
      expect(adapter.fromDTO(dto)).toHaveLength(1);
    });

    it.each([
      ['timestamp', new Date(dto.timestamp).toISOString()],
      ['value.spo2', 98],
      ['value.pulseRate', 72],
    ] as const)('should map %s correctly', (path, expected) => {
      const [reading] = adapter.fromDTO(dto);
      const value = path.split('.').reduce((o: any, k) => o[k], reading);
      expect(value).toBe(expected);
    });

    it('should reflect updated data values', () => {
      const [reading] = adapter.fromDTO({ ...dto, data: { Spo2Value: 95, PulseRateValue: 60 } });
      expect(reading.value).toEqual({ spo2: 95, pulseRate: 60 });
    });
  });
});
