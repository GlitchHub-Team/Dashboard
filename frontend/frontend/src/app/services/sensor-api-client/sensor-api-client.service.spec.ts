import { TestBed } from '@angular/core/testing';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { provideHttpClient } from '@angular/common/http';

import { SensorApiClientService } from './sensor-api-client.service';
import { SensorApiAdapter } from '../../adapters/sensor/sensor-api.adapter';
import { environment } from '../../../environments/environment';
import { SensorBackend } from '../../models/sensor/sensor-backend.model';
import { SensorConfig } from '../../models/sensor/sensor-config.model';
import { SensorProfiles } from '../../models/sensor/sensor-profiles.enum';
import { PaginatedSensorResponse } from '../../models/sensor/paginated-sensor-response.model';
import { Sensor } from '../../models/sensor/sensor.model';
import { SensorStatus } from '../../models/sensor-status.enum';

describe('SensorApiClientService', () => {
  let service: SensorApiClientService;
  let httpMock: HttpTestingController;

  const apiUrl = `${environment.apiUrl}`;

  const mockBackendSensors: SensorBackend[] = [
    {
      sensor_id: 's-1',
      gateway_id: 'gw-1',
      sensor_name: 'Temperature',
      profile: 'health thermometer',
      data_interval: 60,
      status: 'attivo',
    },
    {
      sensor_id: 's-2',
      gateway_id: 'gw-1',
      sensor_name: 'Humidity',
      profile: 'environmental sensing',
      data_interval: 60,
      status: 'inattivo',
    },
  ];

  const mockBackendPaginatedResponse: PaginatedSensorResponse<SensorBackend> = {
    count: 2,
    total: 10,
    sensors: mockBackendSensors,
  };

  const mockMappedSensors: Sensor[] = [
    {
      id: 's-1',
      gatewayId: 'gw-1',
      name: 'Temperature',
      profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
      dataInterval: 60,
      status: SensorStatus.ACTIVE,
    },
    {
      id: 's-2',
      gatewayId: 'gw-1',
      name: 'Humidity',
      profile: SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE,
      dataInterval: 60,
      status: SensorStatus.INACTIVE,
    },
  ];

  const mockMappedPaginatedResponse: PaginatedSensorResponse<Sensor> = {
    count: 2,
    total: 10,
    sensors: mockMappedSensors,
  };

  const mapperMock = {
    fromPaginatedDTO: vi.fn(),
    fromDTO: vi.fn(),
  };

  beforeEach(() => {
    vi.resetAllMocks();

    TestBed.configureTestingModule({
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        SensorApiClientService,
        { provide: SensorApiAdapter, useValue: mapperMock },
      ],
    });

    service = TestBed.inject(SensorApiClientService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  describe.each([
    {
      label: 'getSensorListByGateway',
      invoke: (s: SensorApiClientService) => s.getSensorListByGateway('gw-1', 1, 20),
      url: `${apiUrl}/gateway/gw-1/sensors?page=1&limit=20`,
      page: '1',
      limit: '20',
    },
    {
      label: 'getSensorListByTenant',
      invoke: (s: SensorApiClientService) => s.getSensorListByTenant('tenant-1', 1, 10),
      url: `${apiUrl}/tenant/tenant-1/sensors?page=1&limit=10`,
      page: '1',
      limit: '10',
    },
  ])('$label', ({ invoke, url, page, limit }) => {
    it('should send GET with correct URL and params, map through adapter, and return mapped response', () => {
      mapperMock.fromPaginatedDTO.mockReturnValue(mockMappedPaginatedResponse);

      invoke(service).subscribe((response) => {
        expect(response).toEqual(mockMappedPaginatedResponse);
        expect(response.count).toBe(2);
        expect(response.total).toBe(10);
        expect(response.sensors[0].id).toBe('s-1');
        expect(response.sensors[1].id).toBe('s-2');
      });

      const req = httpMock.expectOne(url);
      expect(req.request.method).toBe('GET');
      expect(req.request.params.get('page')).toBe(page);
      expect(req.request.params.get('limit')).toBe(limit);
      req.flush(mockBackendPaginatedResponse);

      expect(mapperMock.fromPaginatedDTO).toHaveBeenCalledWith(mockBackendPaginatedResponse);
    });
  });

  describe('addNewSensor', () => {
    const mockConfig: SensorConfig = {
      gatewayId: 'gw-1',
      name: 'New Sensor',
      profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
      dataInterval: 60,
    };

    const mockBackendResponse: SensorBackend = {
      sensor_id: 's-3',
      gateway_id: 'gw-1',
      sensor_name: 'New Sensor',
      profile: 'health thermometer',
      data_interval: 60,
      status: 'active',
    };

    const mockMappedSensor: Sensor = {
      id: 's-3',
      gatewayId: 'gw-1',
      name: 'New Sensor',
      profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
      dataInterval: 60,
      status: SensorStatus.ACTIVE,
    };

    it('should send POST with mapped body, map response through adapter, and return domain model', () => {
      mapperMock.fromDTO.mockReturnValue(mockMappedSensor);

      service.addNewSensor(mockConfig).subscribe((sensor) => {
        expect(sensor).toEqual(mockMappedSensor);
        expect(sensor.id).toBe('s-3');
        expect(sensor.name).toBe('New Sensor');
        expect(sensor.dataInterval).toBe(60);
      });

      const req = httpMock.expectOne(`${apiUrl}/sensor`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual({
        gateway_id: mockConfig.gatewayId,
        sensor_name: mockConfig.name,
        profile: mockConfig.profile,
        data_interval: mockConfig.dataInterval,
      });
      req.flush(mockBackendResponse);

      expect(mapperMock.fromDTO).toHaveBeenCalledWith(mockBackendResponse);
    });
  });

  describe('deleteSensor', () => {
    it('should send DELETE with sensor id in URL and return void', () => {
      service.deleteSensor('s-1').subscribe((result) => {
        expect(result).toBeNull();
      });

      const req = httpMock.expectOne(`${apiUrl}/sensor/s-1`);
      expect(req.request.method).toBe('DELETE');
      req.flush(null);
    });
  });
});