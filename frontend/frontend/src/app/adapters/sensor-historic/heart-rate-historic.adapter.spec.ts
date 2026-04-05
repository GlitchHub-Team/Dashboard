import { describe, it, expect } from 'vitest';
import { HeartRateHistoricAdapter } from './heart-rate-historic.adapter';
import { HEART_RATE_FIELDS } from '../../models/sensor-data/sensor-fields.model';
import { HistoricResponse } from '../../models/sensor-data/historic-response.model';

describe('HeartRateHistoricAdapter', () => {
  const adapter = new HeartRateHistoricAdapter();

  const response: HistoricResponse = {
    count: 2,
    samples: [
      { sensor_id: 's1', gateway_id: 'gw1', tenant_id: 't1', timestamp: '2024-05-20T10:00:00.000Z', profile: 'heart_rate', data: { BpmValue: 75 } },
      { sensor_id: 's1', gateway_id: 'gw1', tenant_id: 't1', timestamp: '2024-05-20T10:00:01.000Z', profile: 'heart_rate', data: { BpmValue: 80 } },
    ],
  };

  it('should expose HEART_RATE_FIELDS', () => {
    expect(adapter.fields).toBe(HEART_RATE_FIELDS);
  });

  describe('fromResponse', () => {
    it('should set dataCount equal to number of samples', () => {
      expect(adapter.fromResponse(response).dataCount).toBe(2);
    });

    it('should set fields reference', () => {
      expect(adapter.fromResponse(response).fields).toBe(HEART_RATE_FIELDS);
    });

    it('should map each sample to a reading with correct values', () => {
      const { readings } = adapter.fromResponse(response);
      expect(readings[0]).toEqual({ timestamp: new Date('2024-05-20T10:00:00.000Z').toISOString(), value: { bpm: 75 } });
      expect(readings[1]).toEqual({ timestamp: new Date('2024-05-20T10:00:01.000Z').toISOString(), value: { bpm: 80 } });
    });

    it('should return empty readings for empty samples', () => {
      const result = adapter.fromResponse({ count: 0, samples: [] });
      expect(result.dataCount).toBe(0);
      expect(result.readings).toEqual([]);
    });
  });
});
