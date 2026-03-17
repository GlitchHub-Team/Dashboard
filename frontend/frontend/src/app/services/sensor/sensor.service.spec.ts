import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';

import { SensorService } from './sensor.service';
import { SensorApiClientService } from '../sensor-api-client/sensor-api-client.service';
import { Sensor } from '../../models/sensor.model';
import { SensorConfig } from '../../models/sensor-config.model';
import { SensorProfiles } from '../../models/sensor-profiles.enum';
import { ApiError } from '../../models/api-error.model';

describe('SensorService', () => {
  let service: SensorService;

  const mockSensors: Sensor[] = [
    {
      id: 's-1',
      gatewayId: 'gw-1',
      name: 'Temperature',
      profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
    },
    {
      id: 's-2',
      gatewayId: 'gw-1',
      name: 'Humidity',
      profile: SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE,
    },
  ];

  const mockNewSensor: Sensor = {
    id: 's-3',
    gatewayId: 'gw-1',
    name: 'Pressure',
    profile: SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE,
  };

  const mockConfig: SensorConfig = {
    gatewayId: 'gw-1',
    name: 'Pressure',
    profile: SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE,
  };

  const sensorApiMock = {
    getSensorListByGateway: vi.fn(),
    getSensorListByTenant: vi.fn(),
    addNewSensor: vi.fn(),
    deleteSensor: vi.fn(),
  };

  beforeEach(() => {
    vi.resetAllMocks();

    TestBed.configureTestingModule({
      providers: [SensorService, { provide: SensorApiClientService, useValue: sensorApiMock }],
    });

    service = TestBed.inject(SensorService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  describe('initial state', () => {
    it('should have empty sensor list', () => {
      expect(service.sensorList()).toEqual([]);
    });

    it('should not be loading', () => {
      expect(service.loading()).toBe(false);
    });

    it('should have no error', () => {
      expect(service.error()).toBeNull();
    });
  });

  describe('getSensorsByGateway', () => {
    it('should call api with gatewayId', () => {
      sensorApiMock.getSensorListByGateway.mockReturnValue(of(mockSensors));

      service.getSensorsByGateway('gw-1');

      expect(sensorApiMock.getSensorListByGateway).toHaveBeenCalledWith('gw-1');
    });

    it('should populate sensor list on success', () => {
      sensorApiMock.getSensorListByGateway.mockReturnValue(of(mockSensors));

      service.getSensorsByGateway('gw-1');

      expect(service.sensorList()).toEqual(mockSensors);
      expect(service.loading()).toBe(false);
      expect(service.error()).toBeNull();
    });

    it('should clear previous sensor list before fetching', () => {
      sensorApiMock.getSensorListByGateway.mockReturnValue(of(mockSensors));
      service.getSensorsByGateway('gw-1');

      sensorApiMock.getSensorListByGateway.mockReturnValue(of([]));
      service.getSensorsByGateway('gw-2');

      expect(service.sensorList()).toEqual([]);
    });

    it('should set error on failure with message', () => {
      const apiError: ApiError = { status: 500, message: 'Gateway not found' };
      sensorApiMock.getSensorListByGateway.mockReturnValue(throwError(() => apiError));

      service.getSensorsByGateway('gw-1');

      expect(service.error()).toBe('Gateway not found');
      expect(service.loading()).toBe(false);
      expect(service.sensorList()).toEqual([]);
    });

    it('should set default error message when error has no message', () => {
      const apiError: ApiError = { status: 500 } as ApiError;
      sensorApiMock.getSensorListByGateway.mockReturnValue(throwError(() => apiError));

      service.getSensorsByGateway('gw-1');

      expect(service.error()).toBe('Failed to load sensors');
    });

    it('should clear previous error before fetching', () => {
      const apiError: ApiError = { status: 500, message: 'Error' };
      sensorApiMock.getSensorListByGateway.mockReturnValue(throwError(() => apiError));
      service.getSensorsByGateway('gw-1');
      expect(service.error()).toBe('Error');

      sensorApiMock.getSensorListByGateway.mockReturnValue(of(mockSensors));
      service.getSensorsByGateway('gw-1');
      expect(service.error()).toBeNull();
    });
  });

  describe('getSensorsByTenant', () => {
    it('should call api with tenantId', () => {
      sensorApiMock.getSensorListByTenant.mockReturnValue(of(mockSensors));

      service.getSensorsByTenant('tenant-1');

      expect(sensorApiMock.getSensorListByTenant).toHaveBeenCalledWith('tenant-1');
    });

    it('should populate sensor list on success', () => {
      sensorApiMock.getSensorListByTenant.mockReturnValue(of(mockSensors));

      service.getSensorsByTenant('tenant-1');

      expect(service.sensorList()).toEqual(mockSensors);
      expect(service.loading()).toBe(false);
      expect(service.error()).toBeNull();
    });

    it('should clear previous sensor list before fetching', () => {
      sensorApiMock.getSensorListByTenant.mockReturnValue(of(mockSensors));
      service.getSensorsByTenant('tenant-1');

      sensorApiMock.getSensorListByTenant.mockReturnValue(of([]));
      service.getSensorsByTenant('tenant-1');

      expect(service.sensorList()).toEqual([]);
    });

    it('should set error on failure with message', () => {
      const apiError: ApiError = { status: 500, message: 'Tenant not found' };
      sensorApiMock.getSensorListByTenant.mockReturnValue(throwError(() => apiError));

      service.getSensorsByTenant('tenant-1');

      expect(service.error()).toBe('Tenant not found');
      expect(service.loading()).toBe(false);
      expect(service.sensorList()).toEqual([]);
    });

    it('should set default error message when error has no message', () => {
      const apiError: ApiError = { status: 500 } as ApiError;
      sensorApiMock.getSensorListByTenant.mockReturnValue(throwError(() => apiError));

      service.getSensorsByTenant('tenant-1');

      expect(service.error()).toBe('Failed to load sensors');
    });

    it('should clear previous error before fetching', () => {
      const apiError: ApiError = { status: 500, message: 'Error' };
      sensorApiMock.getSensorListByTenant.mockReturnValue(throwError(() => apiError));
      service.getSensorsByTenant('tenant-1');
      expect(service.error()).toBe('Error');

      sensorApiMock.getSensorListByTenant.mockReturnValue(of(mockSensors));
      service.getSensorsByTenant('tenant-1');
      expect(service.error()).toBeNull();
    });
  });

  describe('addNewSensor', () => {
    it('should call api with sensor config', () => {
      sensorApiMock.addNewSensor.mockReturnValue(of(mockNewSensor));

      service.addNewSensor(mockConfig).subscribe();

      expect(sensorApiMock.addNewSensor).toHaveBeenCalledWith(mockConfig);
    });

    it('should append new sensor to list on success', () => {
      sensorApiMock.getSensorListByGateway.mockReturnValue(of(mockSensors));
      service.getSensorsByGateway('gw-1');

      sensorApiMock.addNewSensor.mockReturnValue(of(mockNewSensor));
      service.addNewSensor(mockConfig).subscribe();

      expect(service.sensorList()).toEqual([...mockSensors, mockNewSensor]);
      expect(service.loading()).toBe(false);
      expect(service.error()).toBeNull();
    });

    it('should set loading to false after success', () => {
      sensorApiMock.addNewSensor.mockReturnValue(of(mockNewSensor));

      service.addNewSensor(mockConfig).subscribe();

      expect(service.loading()).toBe(false);
    });

    it('should set error on failure with message', () => {
      const apiError: ApiError = { status: 500, message: 'Duplicate sensor' };
      sensorApiMock.addNewSensor.mockReturnValue(throwError(() => apiError));

      service.addNewSensor(mockConfig).subscribe();

      expect(service.error()).toBe('Duplicate sensor');
      expect(service.loading()).toBe(false);
    });

    it('should set default error message when error has no message', () => {
      const apiError: ApiError = { status: 500 } as ApiError;
      sensorApiMock.addNewSensor.mockReturnValue(throwError(() => apiError));

      service.addNewSensor(mockConfig).subscribe();

      expect(service.error()).toBe('Failed to add sensor');
    });

    it('should clear previous error before adding', () => {
      const apiError: ApiError = { status: 500, message: 'Error' };
      sensorApiMock.addNewSensor.mockReturnValue(throwError(() => apiError));
      service.addNewSensor(mockConfig).subscribe();
      expect(service.error()).toBe('Error');

      sensorApiMock.addNewSensor.mockReturnValue(of(mockNewSensor));
      service.addNewSensor(mockConfig).subscribe();
      expect(service.error()).toBeNull();
    });

    it('should return the new sensor', () => {
      sensorApiMock.addNewSensor.mockReturnValue(of(mockNewSensor));

      let result: Sensor | undefined;
      service.addNewSensor(mockConfig).subscribe((sensor) => {
        result = sensor;
      });

      expect(result).toEqual(mockNewSensor);
    });

    it('should complete without emitting on error', () => {
      const apiError: ApiError = { status: 500, message: 'Error' };
      sensorApiMock.addNewSensor.mockReturnValue(throwError(() => apiError));

      const nextSpy = vi.fn();
      const errorSpy = vi.fn();
      const completeSpy = vi.fn();

      service.addNewSensor(mockConfig).subscribe({
        next: nextSpy,
        error: errorSpy,
        complete: completeSpy,
      });

      expect(nextSpy).not.toHaveBeenCalled();
      expect(errorSpy).not.toHaveBeenCalled();
      expect(completeSpy).toHaveBeenCalled();
    });
  });

  describe('deleteSensor', () => {
    beforeEach(() => {
      sensorApiMock.getSensorListByGateway.mockReturnValue(of(mockSensors));
      service.getSensorsByGateway('gw-1');
    });

    it('should call api with sensor id', () => {
      sensorApiMock.deleteSensor.mockReturnValue(of(undefined));

      service.deleteSensor('s-1').subscribe();

      expect(sensorApiMock.deleteSensor).toHaveBeenCalledWith('s-1');
    });

    it('should remove sensor from list on success', () => {
      sensorApiMock.deleteSensor.mockReturnValue(of(undefined));

      service.deleteSensor('s-1').subscribe();

      expect(service.sensorList()).toEqual([mockSensors[1]]);
      expect(service.loading()).toBe(false);
      expect(service.error()).toBeNull();
    });

    it('should set loading to false after success', () => {
      sensorApiMock.deleteSensor.mockReturnValue(of(undefined));

      service.deleteSensor('s-1').subscribe();

      expect(service.loading()).toBe(false);
    });

    it('should set error on failure with message', () => {
      const apiError: ApiError = { status: 500, message: 'Sensor in use' };
      sensorApiMock.deleteSensor.mockReturnValue(throwError(() => apiError));

      service.deleteSensor('s-1').subscribe();

      expect(service.error()).toBe('Sensor in use');
      expect(service.loading()).toBe(false);
    });

    it('should set default error message when error has no message', () => {
      const apiError: ApiError = {} as ApiError;
      sensorApiMock.deleteSensor.mockReturnValue(throwError(() => apiError));

      service.deleteSensor('s-1').subscribe();

      expect(service.error()).toBe('Failed to delete sensor');
    });

    it('should not remove sensor from list on failure', () => {
      const apiError: ApiError = { status: 500, message: 'Error' };
      sensorApiMock.deleteSensor.mockReturnValue(throwError(() => apiError));

      service.deleteSensor('s-1').subscribe();

      expect(service.sensorList()).toEqual(mockSensors);
    });

    it('should clear previous error before deleting', () => {
      const apiError: ApiError = { status: 500, message: 'Error' };
      sensorApiMock.deleteSensor.mockReturnValue(throwError(() => apiError));
      service.deleteSensor('s-1').subscribe();
      expect(service.error()).toBe('Error');

      sensorApiMock.deleteSensor.mockReturnValue(of(undefined));
      service.deleteSensor('s-1').subscribe();
      expect(service.error()).toBeNull();
    });

    it('should complete without emitting on error', () => {
      const apiError: ApiError = { status: 500, message: 'Error' };
      sensorApiMock.deleteSensor.mockReturnValue(throwError(() => apiError));

      const nextSpy = vi.fn();
      const errorSpy = vi.fn();
      const completeSpy = vi.fn();

      service.deleteSensor('s-1').subscribe({
        next: nextSpy,
        error: errorSpy,
        complete: completeSpy,
      });

      expect(nextSpy).not.toHaveBeenCalled();
      expect(errorSpy).not.toHaveBeenCalled();
      expect(completeSpy).toHaveBeenCalled();
    });
  });

  describe('clearSensors', () => {
    it('should clear the sensor list', () => {
      sensorApiMock.getSensorListByGateway.mockReturnValue(of(mockSensors));
      service.getSensorsByGateway('gw-1');
      expect(service.sensorList()).toEqual(mockSensors);

      service.clearSensors();

      expect(service.sensorList()).toEqual([]);
    });

    it('should clear even when list is already empty', () => {
      service.clearSensors();

      expect(service.sensorList()).toEqual([]);
    });
  });
});
