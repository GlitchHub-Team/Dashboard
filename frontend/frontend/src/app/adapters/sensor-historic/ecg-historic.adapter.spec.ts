import { describe, it, expect } from 'vitest';
import { EcgHistoricAdapter } from './ecg-historic.adapter';
import { ECG_FIELDS } from '../../models/sensor-data/sensor-fields.model';
import { HistoricResponse } from '../../models/sensor-data/historic-response.model';

describe('EcgHistoricAdapter', () => {
  const adapter = new EcgHistoricAdapter();
  const SAMPLE_RATE = 250;
  const SAMPLE_INTERVAL_MS = 1000 / SAMPLE_RATE;

  const ts1 = '2024-01-01T00:00:00.000Z';
  const ts2 = '2024-01-01T00:00:01.000Z';
  const waveform1 = [0.1, 0.5, 1.2];
  const waveform2 = [0.3, 0.8];

  const response: HistoricResponse = {
    count: 2,
    samples: [
      { sensor_id: 's1', gateway_id: 'gw1', tenant_id: 't1', timestamp: ts1, profile: 'ecg_custom', data: { Waveform: waveform1 } },
      { sensor_id: 's1', gateway_id: 'gw1', tenant_id: 't1', timestamp: ts2, profile: 'ecg_custom', data: { Waveform: waveform2 } },
    ],
  };

  it('should expose ECG_FIELDS', () => {
    expect(adapter.fields).toBe(ECG_FIELDS);
  });

  describe('fromResponse', () => {
    it('should set dataCount to total number of waveform samples across all packets', () => {
      expect(adapter.fromResponse(response).dataCount).toBe(waveform1.length + waveform2.length);
    });

    it('should set fields reference', () => {
      expect(adapter.fromResponse(response).fields).toBe(ECG_FIELDS);
    });

    it('should expand each waveform sample into a reading with offset timestamp', () => {
      const { readings } = adapter.fromResponse(response);
      const base1 = new Date(ts1).getTime();
      const base2 = new Date(ts2).getTime();

      waveform1.forEach((v, i) => {
        expect(readings[i]).toEqual({ timestamp: new Date(base1 + i * SAMPLE_INTERVAL_MS).toISOString(), value: { ecg: v } });
      });
      waveform2.forEach((v, i) => {
        expect(readings[waveform1.length + i]).toEqual({ timestamp: new Date(base2 + i * SAMPLE_INTERVAL_MS).toISOString(), value: { ecg: v } });
      });
    });

    it('should return empty readings for empty samples', () => {
      const result = adapter.fromResponse({ count: 0, samples: [] });
      expect(result.dataCount).toBe(0);
      expect(result.readings).toEqual([]);
    });

    it('should handle a sample with an empty waveform', () => {
      const result = adapter.fromResponse({ count: 1, samples: [{ sensor_id: 's1', gateway_id: 'gw1', tenant_id: 't1', timestamp: ts1, profile: 'ecg_custom', data: { Waveform: [] } }] });
      expect(result.dataCount).toBe(0);
      expect(result.readings).toEqual([]);
    });
  });
});
