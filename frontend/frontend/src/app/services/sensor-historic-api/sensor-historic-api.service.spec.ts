import { TestBed } from '@angular/core/testing';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { provideHttpClient } from '@angular/common/http';

import { SensorHistoricApiService } from './sensor-historic-api.service';
import { environment } from '../../../environments/environment';
import { HistoricResponse } from '../../models/sensor-data/historic-response.model';
import { Sensor } from '../../models/sensor/sensor.model';
import { SensorProfiles } from '../../models/sensor/sensor-profiles.enum';
import { TimeInterval } from '../../models/time-interval.model';
import { Status } from '../../models/gateway-sensor-status.enum';

describe('SensorHistoricApiService', () => {
  let service: SensorHistoricApiService;
  let httpMock: HttpTestingController;

  const apiUrl = `${environment.apiUrl}/sensor-historic`;

  const mockSensor: Sensor = {
    id: 'sensor-1',
    gatewayId: 'gateway-1',
    name: 'Test Sensor',
    profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
    status: Status.ACTIVE,
    dataInterval: 60,
  };

  const mockTimeInterval: TimeInterval = {
    from: new Date('2026-01-01T00:00:00.000Z'),
    to: new Date('2026-01-02T00:00:00.000Z'),
  };

  const mockHistoricResponse: HistoricResponse = {
    count: 2,
    resolution: 60,
    data: [
      {
        sensorId: 'sensor-1',
        timestamp: '2026-01-01T00:00:00.000Z',
        profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
        value: 36.6,
      },
      {
        sensorId: 'sensor-1',
        timestamp: '2026-01-01T01:00:00.000Z',
        profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
        value: 37.0,
      },
    ],
  };

  beforeEach(() => {
    vi.resetAllMocks();

    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });

    service = TestBed.inject(SensorHistoricApiService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  describe('getHistoricData', () => {
    it('should send GET request with correct URL and sensorId', () => {
      service.getHistoricData(mockSensor, mockTimeInterval).subscribe((response) => {
        expect(response).toEqual(mockHistoricResponse);
      });

      const req = httpMock.expectOne((r) => r.url === `${apiUrl}/data?sensorId=sensor-1`);
      expect(req.request.method).toBe('GET');
      req.flush(mockHistoricResponse);
    });

    it('should send GET request with from and to as query params', () => {
      service.getHistoricData(mockSensor, mockTimeInterval).subscribe();

      const req = httpMock.expectOne((r) => r.url === `${apiUrl}/data?sensorId=sensor-1`);
      expect(req.request.params.get('from')).toBe('2026-01-01T00:00:00.000Z');
      expect(req.request.params.get('to')).toBe('2026-01-02T00:00:00.000Z');
      req.flush(mockHistoricResponse);
    });

    it('should return a HistoricResponse', () => {
      service.getHistoricData(mockSensor, mockTimeInterval).subscribe((response) => {
        expect(response.count).toBe(2);
        expect(response.resolution).toBe(60);
        expect(response.data.length).toBe(2);
        expect(response.data[0].sensorId).toBe('sensor-1');
        expect(response.data[0].value).toBe(36.6);
      });

      const req = httpMock.expectOne((r) => r.url === `${apiUrl}/data?sensorId=sensor-1`);
      req.flush(mockHistoricResponse);
    });
  });
});
