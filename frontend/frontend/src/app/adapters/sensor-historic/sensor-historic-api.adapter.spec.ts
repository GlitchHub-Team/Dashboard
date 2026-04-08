import { describe, it, expect } from 'vitest';
import { SensorHistoricApiAdapter } from './sensor-historic-api.adapter';
import { HistoricResponse } from '../../models/sensor-data/historic-response.model';

describe('SensorHistoricApiAdapter', () => {
  const adapter = new SensorHistoricApiAdapter();

  describe('fromDTO', () => {
    it('should map value and timestamp to ISO string', () => {
      const result = adapter.fromDTO(72, 1704067200000); // 2024-01-01T00:00:00.000Z

      expect(result.value).toBe(72);
      expect(result.timestamp).toBe(new Date(1704067200000).toISOString());
    });
  });

  describe('fromResponse', () => {
    it('should map count and pair timestamps with values', () => {
      const response: HistoricResponse = {
        count: { current: 3, real: 3, total: 100 },
        duration: 5000,
        dataset: {
          timestamps: [1704067200000, 1704067260000, 1704067320000],
          values: [72, 75, 73],
        },
        unit: 'bpm',
      };

      const result = adapter.fromResponse(response);

      expect(result.dataCount).toBe(3);
      expect(result.readings).toHaveLength(3);
      expect(result.readings[0]).toEqual({
        value: 72,
        timestamp: new Date(1704067200000).toISOString(),
      });
      expect(result.readings[2]).toEqual({
        value: 73,
        timestamp: new Date(1704067320000).toISOString(),
      });
    });

    it('should handle empty dataset', () => {
      const response: HistoricResponse = {
        count: { current: 0, real: 0, total: 0 },
        duration: 0,
        dataset: { timestamps: [], values: [] },
        unit: 'bpm',
      };

      const result = adapter.fromResponse(response);

      expect(result.dataCount).toBe(0);
      expect(result.readings).toEqual([]);
    });
  });
});
