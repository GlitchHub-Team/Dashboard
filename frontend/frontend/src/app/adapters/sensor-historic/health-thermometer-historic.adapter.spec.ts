import { describe, it, expect } from 'vitest';
import { HealthThermometerHistoricAdapter } from './health-thermometer-historic.adapter';
import { HEALTH_THERMOMETER_FIELDS } from '../../models/sensor-data/sensor-fields.model';
import { HistoricResponse } from '../../models/sensor-data/historic-response.model';

describe('HealthThermometerHistoricAdapter', () => {
  const adapter = new HealthThermometerHistoricAdapter();

  const response: HistoricResponse = {
    count: 2,
    samples: [
      { sensor_id: 's1', gateway_id: 'gw1', tenant_id: 't1', timestamp: '2024-03-10T08:00:00.000Z', profile: 'health_thermometer', data: { TemperatureValue: 36.6 } },
      { sensor_id: 's1', gateway_id: 'gw1', tenant_id: 't1', timestamp: '2024-03-10T08:01:00.000Z', profile: 'health_thermometer', data: { TemperatureValue: 37.1 } },
    ],
  };

  it('should expose HEALTH_THERMOMETER_FIELDS', () => {
    expect(adapter.fields).toBe(HEALTH_THERMOMETER_FIELDS);
  });

  describe('fromResponse', () => {
    it('should set dataCount equal to number of samples', () => {
      expect(adapter.fromResponse(response).dataCount).toBe(2);
    });

    it('should set fields reference', () => {
      expect(adapter.fromResponse(response).fields).toBe(HEALTH_THERMOMETER_FIELDS);
    });

    it('should map each sample to a reading with correct values', () => {
      const { readings } = adapter.fromResponse(response);
      expect(readings[0]).toEqual({ timestamp: new Date('2024-03-10T08:00:00.000Z').toISOString(), value: { temperature: 36.6 } });
      expect(readings[1]).toEqual({ timestamp: new Date('2024-03-10T08:01:00.000Z').toISOString(), value: { temperature: 37.1 } });
    });

    it('should return empty readings for empty samples', () => {
      const result = adapter.fromResponse({ count: 0, samples: [] });
      expect(result.dataCount).toBe(0);
      expect(result.readings).toEqual([]);
    });
  });
});
