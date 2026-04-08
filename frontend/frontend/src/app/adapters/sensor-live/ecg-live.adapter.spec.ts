import { describe, it, expect } from 'vitest';
import { EcgLiveAdapter } from './ecg-live.adapter';
import { ECG_FIELDS } from '../../models/sensor-data/sensor-fields.model';
import { RealTimeReading } from '../../models/sensor-data/real-time-reading.model';

describe('EcgLiveAdapter', () => {
  const adapter = new EcgLiveAdapter();

  const baseTimestamp = '2024-01-01T00:00:00.000Z';
  const waveform = [0.1, 0.5, 1.2, 0.8, -0.3];

  const dto: RealTimeReading = {
    timestamp: baseTimestamp,
    profile: 'ecg_custom',
    data: { Waveform: waveform },
  };

  it('should expose ECG_FIELDS', () => {
    expect(adapter.fields).toBe(ECG_FIELDS);
  });

  describe('fromDTO', () => {
    it('should return one reading per waveform sample', () => {
      expect(adapter.fromDTO(dto)).toHaveLength(waveform.length);
    });

    it('should assign correct ecg value for each sample', () => {
      adapter.fromDTO(dto).forEach((reading, i) => {
        expect(reading.value).toEqual({ ecg: waveform[i] });
      });
    });

    it('should offset timestamps using variable spacing based on waveform length', () => {
      const baseTime = new Date(baseTimestamp).getTime();
      const sampleIntervalMs = 1000 / waveform.length;
      adapter.fromDTO(dto).forEach((reading, i) => {
        expect(reading.timestamp).toBe(new Date(baseTime + i * sampleIntervalMs).toISOString());
      });
    });

    it('should return empty array for empty waveform', () => {
      expect(adapter.fromDTO({ ...dto, data: { Waveform: [] } })).toEqual([]);
    });
  });
});
