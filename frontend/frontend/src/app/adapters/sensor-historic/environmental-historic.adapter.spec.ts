import { describe, it, expect } from 'vitest';
import { EnvironmentalHistoricAdapter } from './environmental-historic.adapter';
import { ENVIRONMENTAL_FIELDS } from '../../models/sensor-data/sensor-fields.model';
import { HistoricResponse } from '../../models/sensor-data/historic-response.model';

describe('EnvironmentalHistoricAdapter', () => {
  const adapter = new EnvironmentalHistoricAdapter();

  const response: HistoricResponse = {
    count: 2,
    samples: [
      { sensor_id: 's1', gateway_id: 'gw1', tenant_id: 't1', timestamp: '2024-06-15T12:00:00.000Z', profile: 'environmental_sensing', data: { TemperatureValue: 22.5, HumidityValue: 60 } },
      { sensor_id: 's1', gateway_id: 'gw1', tenant_id: 't1', timestamp: '2024-06-15T12:00:01.000Z', profile: 'environmental_sensing', data: { TemperatureValue: 23.0, HumidityValue: 62 } },
    ],
  };

  it('should expose ENVIRONMENTAL_FIELDS', () => {
    expect(adapter.fields).toBe(ENVIRONMENTAL_FIELDS);
  });

  describe('fromResponse', () => {
    it('should set dataCount equal to number of samples', () => {
      expect(adapter.fromResponse(response).dataCount).toBe(2);
    });

    it('should set fields reference', () => {
      expect(adapter.fromResponse(response).fields).toBe(ENVIRONMENTAL_FIELDS);
    });

    it('should map each sample to a reading with correct values', () => {
      const { readings } = adapter.fromResponse(response);
      expect(readings[0]).toEqual({ timestamp: new Date('2024-06-15T12:00:00.000Z').toISOString(), value: { temperature: 22.5, humidity: 60 } });
      expect(readings[1]).toEqual({ timestamp: new Date('2024-06-15T12:00:01.000Z').toISOString(), value: { temperature: 23.0, humidity: 62 } });
    });

    it('should return empty readings for empty samples', () => {
      const result = adapter.fromResponse({ count: 0, samples: [] });
      expect(result.dataCount).toBe(0);
      expect(result.readings).toEqual([]);
    });
  });
});
