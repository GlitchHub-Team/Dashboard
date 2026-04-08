import { describe, it, expect } from 'vitest';
import { SensorApiAdapter } from './sensor-api.adapter';
import { SensorBackend } from '../../models/sensor/sensor-backend.model';
import { Status } from '../../models/gateway-sensor-status.enum';
import { SensorProfiles } from '../../models/sensor/sensor-profiles.enum';

describe('SensorApiAdapter', () => {
  const adapter = new SensorApiAdapter();

  const dto: SensorBackend = {
    sensor_id: 'sensor-1',
    gateway_id: 'gw-1',
    sensor_name: 'Heart Rate Sensor',
    status: 'attivo',
    profile: 'heart_rate_service',
    sensor_interval: 1000,
  };

  describe('fromDTO', () => {
    it.each([
      { field: 'id', expected: 'sensor-1' },
      { field: 'gatewayId', expected: 'gw-1' },
      { field: 'name', expected: 'Heart Rate Sensor' },
      { field: 'status', expected: Status.ACTIVE },
      { field: 'profile', expected: SensorProfiles.HEART_RATE_SERVICE },
      { field: 'dataInterval', expected: 1000 },
    ] as const)('should map $field correctly', ({ field, expected }) => {
      expect(adapter.fromDTO(dto)[field]).toEqual(expected);
    });

    it.each([
      ['heart_rate', SensorProfiles.HEART_RATE_SERVICE],
      ['pulse_oximeter', SensorProfiles.PULSE_OXIMETER_SERVICE],
      ['custom_ecg', SensorProfiles.CUSTOM_ECG_SERVICE],
      ['health_thermometer', SensorProfiles.HEALTH_THERMOMETER_SERVICE],
      ['environmental_sensing', SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE],
    ])('should map profile "%s" correctly', (backendProfile, expected) => {
      expect(adapter.fromDTO({ ...dto, profile: backendProfile }).profile).toBe(expected);
    });
  });

  describe('fromPaginatedDTO', () => {
    it('should map count, total and all sensors', () => {
      const response = {
        count: 2,
        total: 10,
        sensors: [dto, { ...dto, sensor_id: 'sensor-2', status: 'inattivo' }],
      };

      const result = adapter.fromPaginatedDTO(response);

      expect(result.count).toBe(2);
      expect(result.total).toBe(10);
      expect(result.sensors).toHaveLength(2);
      expect(result.sensors[0].id).toBe('sensor-1');
      expect(result.sensors[1].status).toBe(Status.INACTIVE);
    });

    it('should handle empty array', () => {
      const result = adapter.fromPaginatedDTO({ count: 0, total: 0, sensors: [] });
      expect(result.sensors).toEqual([]);
    });
  });
});
