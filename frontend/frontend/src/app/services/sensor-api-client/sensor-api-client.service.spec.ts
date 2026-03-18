import { TestBed } from '@angular/core/testing';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { provideHttpClient } from '@angular/common/http';

import { SensorApiClientService } from './sensor-api-client.service';
import { environment } from '../../../environments/environment';
import { SensorBackend } from '../../models/sensor/sensor-backend.model';
import { SensorConfig } from '../../models/sensor/sensor-config.model';
import { SensorProfiles } from '../../models/sensor/sensor-profiles.enum';
import { PaginatedResponse } from '../../models/paginated-response.model';

describe('SensorApiClientService', () => {
  let service: SensorApiClientService;
  let httpMock: HttpTestingController;

  const apiUrl = `${environment.apiUrl}/sensor`;

  const mockSensors: SensorBackend[] = [
    {
      SensorId: 's-1',
      GatewayId: 'gw-1',
      Name: 'Temperature',
      Profile: 'health thermometer',
    },
    {
      SensorId: 's-2',
      GatewayId: 'gw-1',
      Name: 'Humidity',
      Profile: 'environmental sensing',
    },
  ];

  const mockPaginatedResponse: PaginatedResponse<SensorBackend> = {
    count: 2,
    total: 10,
    data: mockSensors,
  };

  beforeEach(() => {
    vi.resetAllMocks();

    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
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

  describe('getSensorListByGateway', () => {
    it('should send GET request with correct URL and query params', () => {
      service.getSensorListByGateway('gw-1', 1, 20).subscribe((response) => {
        expect(response).toEqual(mockPaginatedResponse);
      });

      const req = httpMock.expectOne(`${apiUrl}/gw-1/list?page=1&limit=20`);
      expect(req.request.method).toBe('GET');
      expect(req.request.params.get('page')).toBe('1');
      expect(req.request.params.get('limit')).toBe('20');
      req.flush(mockPaginatedResponse);
    });

    it('should return a PaginatedResponse of SensorBackend', () => {
      service.getSensorListByGateway('gw-1', 1, 20).subscribe((response) => {
        expect(response.count).toBe(2);
        expect(response.total).toBe(10);
        expect(response.data.length).toBe(2);
        expect(response.data[0].SensorId).toBe('s-1');
        expect(response.data[1].SensorId).toBe('s-2');
      });

      const req = httpMock.expectOne(`${apiUrl}/gw-1/list?page=1&limit=20`);
      req.flush(mockPaginatedResponse);
    });
  });

  describe('getSensorListByTenant', () => {
    it('should send GET request with correct URL and query params', () => {
      service.getSensorListByTenant('tenant-1', 1, 10).subscribe((response) => {
        expect(response).toEqual(mockPaginatedResponse);
      });

      const req = httpMock.expectOne(`${apiUrl}/tenant/tenant-1/list?page=1&limit=10`);
      expect(req.request.method).toBe('GET');
      expect(req.request.params.get('page')).toBe('1');
      expect(req.request.params.get('limit')).toBe('10');
      req.flush(mockPaginatedResponse);
    });

    it('should return a PaginatedResponse of SensorBackend', () => {
      service.getSensorListByTenant('tenant-1', 1, 10).subscribe((response) => {
        expect(response.count).toBe(2);
        expect(response.total).toBe(10);
        expect(response.data.length).toBe(2);
      });

      const req = httpMock.expectOne(`${apiUrl}/tenant/tenant-1/list?page=1&limit=10`);
      req.flush(mockPaginatedResponse);
    });
  });

  describe('addNewSensor', () => {
    const mockConfig: SensorConfig = {
      gatewayId: 'gw-1',
      name: 'New Sensor',
      profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
    };

    const mockResponse: SensorBackend = {
      SensorId: 's-3',
      GatewayId: 'gw-1',
      Name: 'New Sensor',
      Profile: 'health thermometer',
    };

    it('should send POST request with sensor config as body', () => {
      service.addNewSensor(mockConfig).subscribe((sensor) => {
        expect(sensor).toEqual(mockResponse);
      });

      const req = httpMock.expectOne(`${apiUrl}/add`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual(mockConfig);
      req.flush(mockResponse);
    });

    it('should return a SensorBackend', () => {
      service.addNewSensor(mockConfig).subscribe((sensor) => {
        expect(sensor.SensorId).toBe('s-3');
        expect(sensor.Name).toBe('New Sensor');
      });

      const req = httpMock.expectOne(`${apiUrl}/add`);
      req.flush(mockResponse);
    });
  });

  describe('deleteSensor', () => {
    it('should send DELETE request with sensor id in the URL', () => {
      service.deleteSensor('s-1').subscribe();

      const req = httpMock.expectOne(`${apiUrl}/delete/s-1`);
      expect(req.request.method).toBe('DELETE');
      req.flush(null);
    });

    it('should return an observable of void', () => {
      service.deleteSensor('s-1').subscribe((result) => {
        expect(result).toBeNull();
      });

      const req = httpMock.expectOne(`${apiUrl}/delete/s-1`);
      req.flush(null);
    });
  });
});
