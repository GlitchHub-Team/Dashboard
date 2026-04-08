import { describe, it, expect } from 'vitest';
import { PulseOximeterHistoricAdapter } from './pulse-oximeter-historic.adapter';
import { PULSE_OXIMETER_FIELDS } from '../../models/sensor-data/sensor-fields.model';
import { HistoricResponse } from '../../models/sensor-data/historic-response.model';

describe('PulseOximeterHistoricAdapter', () => {
  const adapter = new PulseOximeterHistoricAdapter();

  const response: HistoricResponse = {
    count: 2,
    samples: [
      { sensor_id: 's1', gateway_id: 'gw1', tenant_id: 't1', timestamp: '2024-01-01T00:00:00.000Z', profile: 'pulse_oximeter', data: { Spo2Value: 98, PulseRateValue: 72 } },
      { sensor_id: 's1', gateway_id: 'gw1', tenant_id: 't1', timestamp: '2024-01-01T00:00:01.000Z', profile: 'pulse_oximeter', data: { Spo2Value: 97, PulseRateValue: 75 } },
    ],
  };

  it('should expose PULSE_OXIMETER_FIELDS', () => {
    expect(adapter.fields).toBe(PULSE_OXIMETER_FIELDS);
  });

  describe('fromResponse', () => {
    it('should set dataCount equal to number of samples', () => {
      expect(adapter.fromResponse(response).dataCount).toBe(2);
    });

    it('should set fields reference', () => {
      expect(adapter.fromResponse(response).fields).toBe(PULSE_OXIMETER_FIELDS);
    });

    it('should map each sample to a reading with correct values', () => {
      const { readings } = adapter.fromResponse(response);
      expect(readings[0]).toEqual({ timestamp: new Date('2024-01-01T00:00:00.000Z').toISOString(), value: { spo2: 98, pulseRate: 72 } });
      expect(readings[1]).toEqual({ timestamp: new Date('2024-01-01T00:00:01.000Z').toISOString(), value: { spo2: 97, pulseRate: 75 } });
    });

    it('should return empty readings for empty samples', () => {
      const result = adapter.fromResponse({ count: 0, samples: [] });
      expect(result.dataCount).toBe(0);
      expect(result.readings).toEqual([]);
    });
  });
});
