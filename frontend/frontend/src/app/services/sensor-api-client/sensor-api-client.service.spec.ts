import { TestBed } from '@angular/core/testing';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { provideHttpClient } from '@angular/common/http';

import { SensorApiClientService } from './sensor-api-client.service';
import { environment } from '../../../environments/environment';
import { Sensor } from '../../models/sensor/sensor.model';
import { SensorConfig } from '../../models/sensor/sensor-config.model';
import { SensorProfiles } from '../../models/sensor/sensor-profiles.enum';

describe('SensorApiClientService', () => {
  let service: SensorApiClientService;
  let httpMock: HttpTestingController;

  const apiUrl = `${environment.apiUrl}/sensor`;

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
    it('should send GET request with gatewayId param', () => {
      service.getSensorListByGateway('gw-1').subscribe((sensors) => {
        expect(sensors).toEqual(mockSensors);
      });

      const req = httpMock.expectOne(`${apiUrl}/list?gatewayId=gw-1`);
      expect(req.request.method).toBe('GET');
      expect(req.request.params.get('gatewayId')).toBe('gw-1');
      req.flush(mockSensors);
    });

    it('should return an observable of Sensor[]', () => {
      service.getSensorListByGateway('gw-1').subscribe((sensors) => {
        expect(sensors.length).toBe(2);
        expect(sensors[0].id).toBe('s-1');
        expect(sensors[1].id).toBe('s-2');
      });

      const req = httpMock.expectOne(`${apiUrl}/list?gatewayId=gw-1`);
      req.flush(mockSensors);
    });
  });

  describe('getSensorListByTenant', () => {
    it('should send GET request with tenantId param', () => {
      service.getSensorListByTenant('tenant-1').subscribe((sensors) => {
        expect(sensors).toEqual(mockSensors);
      });

      const req = httpMock.expectOne(`${apiUrl}/list?tenantId=tenant-1`);
      expect(req.request.method).toBe('GET');
      expect(req.request.params.get('tenantId')).toBe('tenant-1');
      req.flush(mockSensors);
    });

    it('should return an observable of Sensor[]', () => {
      service.getSensorListByTenant('tenant-1').subscribe((sensors) => {
        expect(sensors.length).toBe(2);
      });

      const req = httpMock.expectOne(`${apiUrl}/list?tenantId=tenant-1`);
      req.flush(mockSensors);
    });
  });

  describe('addNewSensor', () => {
    const mockConfig: SensorConfig = {
      gatewayId: 'gw-1',
      name: 'New Sensor',
      profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
    };

    const mockResponse: Sensor = {
      id: 's-3',
      gatewayId: 'gw-1',
      name: 'New Sensor',
      profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
    };

    it('should send POST request with sensor config', () => {
      service.addNewSensor(mockConfig).subscribe((sensor) => {
        expect(sensor).toEqual(mockResponse);
      });

      const req = httpMock.expectOne(`${apiUrl}/add`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual(mockConfig);
      req.flush(mockResponse);
    });

    it('should return an observable of Sensor', () => {
      service.addNewSensor(mockConfig).subscribe((sensor) => {
        expect(sensor.id).toBe('s-3');
        expect(sensor.name).toBe('New Sensor');
      });

      const req = httpMock.expectOne(`${apiUrl}/add`);
      req.flush(mockResponse);
    });
  });

  describe('deleteSensor', () => {
    it('should send DELETE request with sensor id', () => {
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
