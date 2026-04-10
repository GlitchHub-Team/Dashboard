import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';

import { SensorService } from './sensor.service';
import { SensorApiClientService } from '../sensor-api-client/sensor-api-client.service';
import { SensorCommandApiClientService } from '../sensor-command-api-client/sensor-command-api-client.service';
import { SensorAdapter } from '../../adapters/sensor/sensor.adapter';
import { Sensor } from '../../models/sensor/sensor.model';
import { SensorBackend } from '../../models/sensor/sensor-backend.model';
import { SensorConfig } from '../../models/sensor/sensor-config.model';
import { SensorProfiles } from '../../models/sensor/sensor-profiles.enum';
import { PaginatedSensorResponse } from '../../models/sensor/paginated-sensor-response.model';
import { ApiError } from '../../models/api-error.model';
import { SensorStatus } from '../../models/sensor-status.enum';

describe('SensorService', () => {
  let service: SensorService;

  const mockSensors: Sensor[] = [
    {
      id: 's-1',
      gatewayId: 'gw-1',
      name: 'Temperature',
      profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
      status: SensorStatus.ACTIVE,
      dataInterval: 60,
    },
    {
      id: 's-2',
      gatewayId: 'gw-1',
      name: 'Humidity',
      profile: SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE,
      status: SensorStatus.INACTIVE,
      dataInterval: 60,
    },
  ];

  const mockBackendResponse: PaginatedSensorResponse<SensorBackend> = {
    count: 2,
    total: 10,
    sensors: [
      {
        sensor_id: 's-1',
        gateway_id: 'gw-1',
        sensor_name: 'Temperature',
        profile: 'health thermometer',
        data_interval: 60,
        status: SensorStatus.ACTIVE,
      },
      {
        sensor_id: 's-2',
        gateway_id: 'gw-1',
        sensor_name: 'Humidity',
        profile: 'environmental sensing',
        data_interval: 60,
        status: SensorStatus.INACTIVE,
      },
    ],
  };

  const mockAdaptedResponse: PaginatedSensorResponse<Sensor> = {
    count: 2,
    total: 10,
    sensors: mockSensors,
  };
  const emptyBackend: PaginatedSensorResponse<SensorBackend> = { count: 0, total: 0, sensors: [] };
  const emptyAdapted: PaginatedSensorResponse<Sensor> = { count: 0, total: 0, sensors: [] };

  const mockNewBackend: SensorBackend = {
    sensor_id: 's-3',
    gateway_id: 'gw-1',
    sensor_name: 'Pressure',
    profile: 'environmental sensing',
    data_interval: 60,
    status: SensorStatus.ACTIVE,
  };
  const mockNewSensor: Sensor = {
    id: 's-3',
    gatewayId: 'gw-1',
    name: 'Pressure',
    status: SensorStatus.ACTIVE,
    profile: SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE,
    dataInterval: 60,
  };
  const mockConfig: SensorConfig = {
    gatewayId: 'gw-1',
    name: 'Pressure',
    dataInterval: 60,
    profile: SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE,
  };

  const sensorApiMock = {
    getSensorListByGateway: vi.fn(),
    getSensorListByTenant: vi.fn(),
    addNewSensor: vi.fn(),
    deleteSensor: vi.fn(),
  };

  const sensorCommandApiMock = {
    interruptSensor: vi.fn(),
    resumeSensor: vi.fn(),
  };

  const adapterMock = {
    fromPaginatedDTO: vi.fn(),
    fromDTO: vi.fn(),
  };

  type ListApiKey = 'getSensorListByGateway' | 'getSensorListByTenant';

  function mockListSuccess(
    apiKey: ListApiKey,
    backendRes = mockBackendResponse,
    adaptedRes = mockAdaptedResponse,
  ): void {
    sensorApiMock[apiKey].mockReturnValue(of(backendRes));
    adapterMock.fromPaginatedDTO.mockReturnValue(adaptedRes);
  }

  beforeEach(() => {
    vi.resetAllMocks();

    TestBed.configureTestingModule({
      providers: [
        SensorService,
        { provide: SensorApiClientService, useValue: sensorApiMock },
        { provide: SensorCommandApiClientService, useValue: sensorCommandApiMock },
        { provide: SensorAdapter, useValue: adapterMock },
      ],
    });

    service = TestBed.inject(SensorService);
  });

  it('should be created with correct initial state', () => {
    expect(service).toBeTruthy();
    expect(service.sensorList()).toEqual([]);
    expect(service.loading()).toBe(false);
    expect(service.error()).toBeNull();
    expect(service.pageIndex()).toBe(0);
    expect(service.limit()).toBe(10);
    expect(service.total()).toBe(0);
  });

  describe.each([
    {
      label: 'getSensorsByGateway',
      id: 'gw-1',
      apiKey: 'getSensorListByGateway' as ListApiKey,
      invoke: (s: SensorService, page: number, limit: number) =>
        s.getSensorsByGateway('gw-1', page, limit),
    },
    {
      label: 'getSensorsByTenant',
      id: 'tenant-1',
      apiKey: 'getSensorListByTenant' as ListApiKey,
      invoke: (s: SensorService, page: number, limit: number) =>
        s.getSensorsByTenant('tenant-1', page, limit),
    },
  ])('$label', ({ id, apiKey, invoke }) => {
    it('should call api, map through adapter, and populate state on success', () => {
      mockListSuccess(apiKey);
      invoke(service, 0, 10);

      expect(sensorApiMock[apiKey]).toHaveBeenCalledWith(id, 1, 10);
      expect(adapterMock.fromPaginatedDTO).toHaveBeenCalledWith(mockBackendResponse);
      expect(service.sensorList()).toEqual(mockSensors);
      expect(service.total()).toBe(10);
      expect(service.loading()).toBe(false);
      expect(service.error()).toBeNull();
    });

    it('should update pagination signals', () => {
      mockListSuccess(apiKey);
      invoke(service, 2, 25);

      expect(service.pageIndex()).toBe(2);
      expect(service.limit()).toBe(25);
    });

    it('should clear previous sensor list before fetching', () => {
      mockListSuccess(apiKey);
      invoke(service, 0, 10);

      mockListSuccess(apiKey, emptyBackend, emptyAdapted);
      invoke(service, 0, 10);

      expect(service.sensorList()).toEqual([]);
    });

    it('should set error on failure and clear it on next success', () => {
      sensorApiMock[apiKey].mockReturnValue(
        throwError(() => ({ status: 500, message: 'Error' }) as ApiError),
      );
      invoke(service, 0, 10);

      expect(service.error()).toBe('Error');
      expect(service.loading()).toBe(false);

      mockListSuccess(apiKey);
      invoke(service, 0, 10);

      expect(service.error()).toBeNull();
    });

    it('should use default error message when error has no message', () => {
      sensorApiMock[apiKey].mockReturnValue(throwError(() => ({ status: 500 }) as ApiError));
      invoke(service, 0, 10);

      expect(service.error()).toBe('Failed to load sensors');
    });
  });

  describe('addNewSensor', () => {
    it('should call api, map through adapter, and return adapted sensor', () => {
      sensorApiMock.addNewSensor.mockReturnValue(of(mockNewBackend));
      adapterMock.fromDTO.mockReturnValue(mockNewSensor);

      let result: Sensor | undefined;
      service.addNewSensor(mockConfig).subscribe((s) => (result = s));

      expect(sensorApiMock.addNewSensor).toHaveBeenCalledWith(mockConfig);
      expect(adapterMock.fromDTO).toHaveBeenCalledWith(mockNewBackend);
      expect(result).toEqual(mockNewSensor);
    });

    it.each([
      ['gateway', 'getSensorListByGateway' as ListApiKey, () => service.getSensorsByGateway('gw-1', 0, 10)] as const,
      ['tenant', 'getSensorListByTenant' as ListApiKey, () => service.getSensorsByTenant('tenant-1', 0, 10)] as const,
    ])('should not refetch after success (%s context)', (_, apiKey, setup) => {
      mockListSuccess(apiKey);
      setup();
      sensorApiMock[apiKey].mockClear();

      sensorApiMock.addNewSensor.mockReturnValue(of(mockNewBackend));
      adapterMock.fromDTO.mockReturnValue(mockNewSensor);
      service.addNewSensor(mockConfig).subscribe();

      expect(sensorApiMock[apiKey]).not.toHaveBeenCalled();
    });

    it('should propagate error to subscriber and not refetch', () => {
      mockListSuccess('getSensorListByGateway');
      service.getSensorsByGateway('gw-1', 0, 10);
      sensorApiMock.getSensorListByGateway.mockClear();

      const error = { status: 500, message: 'Duplicate sensor' } as ApiError;
      sensorApiMock.addNewSensor.mockReturnValue(throwError(() => error));
      const errorSpy = vi.fn();
      const completeSpy = vi.fn();
      service.addNewSensor(mockConfig).subscribe({ error: errorSpy, complete: completeSpy });

      expect(errorSpy).toHaveBeenCalledWith(error);
      expect(completeSpy).not.toHaveBeenCalled();
      expect(sensorApiMock.getSensorListByGateway).not.toHaveBeenCalled();
    });

    it('should propagate error to subscriber when error has no message', () => {
      const error = { status: 500 } as ApiError;
      sensorApiMock.addNewSensor.mockReturnValue(throwError(() => error));
      const errorSpy = vi.fn();
      service.addNewSensor(mockConfig).subscribe({ error: errorSpy });

      expect(errorSpy).toHaveBeenCalledWith(error);
    });
  });

  describe('deleteSensor', () => {
    beforeEach(() => {
      mockListSuccess('getSensorListByGateway');
      service.getSensorsByGateway('gw-1', 0, 10);
    });

    it('should call api, refetch current page, and set loading to false on success', () => {
      sensorApiMock.getSensorListByGateway.mockClear();
      sensorApiMock.deleteSensor.mockReturnValue(of(undefined));
      mockListSuccess('getSensorListByGateway');

      service.deleteSensor('s-1').subscribe();

      expect(sensorApiMock.deleteSensor).toHaveBeenCalledWith('s-1');
      expect(sensorApiMock.getSensorListByGateway).toHaveBeenCalledWith('gw-1', 1, 10);
      expect(service.loading()).toBe(false);
    });

    it('should set error on failure, not refetch, and complete without emitting', () => {
      sensorApiMock.getSensorListByGateway.mockClear();
      sensorApiMock.deleteSensor.mockReturnValue(
        throwError(() => ({ status: 500, message: 'Sensor in use' }) as ApiError),
      );
      const errorSpy = vi.fn();
      const completeSpy = vi.fn();
      service.deleteSensor('s-1').subscribe({ error: errorSpy, complete: completeSpy });

      expect(service.error()).toBe('Sensor in use');
      expect(service.loading()).toBe(false);
      expect(sensorApiMock.getSensorListByGateway).not.toHaveBeenCalled();
      expect(errorSpy).not.toHaveBeenCalled();
      expect(completeSpy).toHaveBeenCalled();
    });

    it('should use default error message when error has no message', () => {
      sensorApiMock.deleteSensor.mockReturnValue(throwError(() => ({}) as ApiError));
      service.deleteSensor('s-1').subscribe();

      expect(service.error()).toBe('Failed to delete sensor');
    });

    it('should clear previous error before deleting', () => {
      sensorApiMock.deleteSensor.mockReturnValue(
        throwError(() => ({ status: 500, message: 'Error' })),
      );
      service.deleteSensor('s-1').subscribe();

      sensorApiMock.deleteSensor.mockReturnValue(of(undefined));
      mockListSuccess('getSensorListByGateway');
      service.deleteSensor('s-1').subscribe();

      expect(service.error()).toBeNull();
    });
  });

  describe('interruptSensor and resumeSensor', () => {
    beforeEach(() => {
      mockListSuccess('getSensorListByGateway');
      service.getSensorsByGateway('gw-1', 0, 10);
      sensorApiMock.getSensorListByGateway.mockClear();
    });

    it.each([
      ['interruptSensor'] as const,
      ['resumeSensor'] as const,
    ])('%s should call command api and refetch on success', (method) => {
      sensorCommandApiMock[method].mockReturnValue(of(undefined));
      mockListSuccess('getSensorListByGateway');

      service[method]('s-1').subscribe();

      expect(sensorCommandApiMock[method]).toHaveBeenCalledWith('s-1');
      expect(sensorApiMock.getSensorListByGateway).toHaveBeenCalledWith('gw-1', 1, 10);
      expect(service.loading()).toBe(false);
    });

    it.each([
      ['interruptSensor'] as const,
      ['resumeSensor'] as const,
    ])('%s should propagate error to subscriber and not refetch', (method) => {
      const error = { status: 500, message: 'Command failed' } as ApiError;
      sensorCommandApiMock[method].mockReturnValue(throwError(() => error));
      const errorSpy = vi.fn();
      const completeSpy = vi.fn();
      service[method]('s-1').subscribe({ error: errorSpy, complete: completeSpy });

      expect(service.error()).toBeNull();
      expect(sensorApiMock.getSensorListByGateway).not.toHaveBeenCalled();
      expect(errorSpy).toHaveBeenCalledWith(error);
      expect(completeSpy).not.toHaveBeenCalled();
    });
  });

  describe('changePage', () => {
    it('should refetch by gateway when gateway context is active', () => {
      mockListSuccess('getSensorListByGateway');
      service.getSensorsByGateway('gw-1', 0, 10);
      sensorApiMock.getSensorListByGateway.mockClear();
      mockListSuccess('getSensorListByGateway');

      service.changePage(2, 20);

      expect(sensorApiMock.getSensorListByGateway).toHaveBeenCalledWith('gw-1', 3, 20);
    });

    it('should refetch by tenant when tenant context is active', () => {
      mockListSuccess('getSensorListByTenant');
      service.getSensorsByTenant('tenant-1', 0, 10);
      sensorApiMock.getSensorListByTenant.mockClear();
      mockListSuccess('getSensorListByTenant');

      service.changePage(3, 15);

      expect(sensorApiMock.getSensorListByTenant).toHaveBeenCalledWith('tenant-1', 4, 15);
    });

    it('should do nothing if no context is set', () => {
      service.changePage(1, 10);

      expect(sensorApiMock.getSensorListByGateway).not.toHaveBeenCalled();
      expect(sensorApiMock.getSensorListByTenant).not.toHaveBeenCalled();
    });
  });

  describe('clearSensors', () => {
    it('should reset all state and clear context so changePage does nothing', () => {
      mockListSuccess('getSensorListByGateway');
      service.getSensorsByGateway('gw-1', 2, 20);
      sensorApiMock.getSensorListByGateway.mockClear();

      service.clearSensors();

      expect(service.sensorList()).toEqual([]);
      expect(service.total()).toBe(0);
      expect(service.pageIndex()).toBe(0);
      expect(service.error()).toBeNull();

      service.changePage(1, 10);
      expect(sensorApiMock.getSensorListByGateway).not.toHaveBeenCalled();
    });

    it('should clear even when state is already empty', () => {
      service.clearSensors();

      expect(service.sensorList()).toEqual([]);
      expect(service.total()).toBe(0);
    });
  });
});
