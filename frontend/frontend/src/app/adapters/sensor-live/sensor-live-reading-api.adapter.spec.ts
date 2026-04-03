import { describe, it, expect } from 'vitest';
import { SensorLiveReadingApiAdapter } from './sensor-live-reading-api.adapter';

describe('SensorLiveReadingApiAdapter', () => {
  const adapter = new SensorLiveReadingApiAdapter();

  describe('fromDTO', () => {
    it.each([
      { datum: 72, timestamp: 1704067200000 },
      { datum: 98.6, timestamp: 1704067260000 },
      { datum: 0, timestamp: 0 },
    ])('should map datum=$datum and timestamp correctly', ({ datum, timestamp }) => {
      const result = adapter.fromDTO({ datum, timestamp });

      expect(result.value).toBe(datum);
      expect(result.timestamp).toBe(new Date(timestamp).toISOString());
    });
  });
});
