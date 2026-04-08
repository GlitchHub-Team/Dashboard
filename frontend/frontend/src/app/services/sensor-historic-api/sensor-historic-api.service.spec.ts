import { TestBed } from '@angular/core/testing';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { provideHttpClient } from '@angular/common/http';

import { SensorHistoricApiService } from './sensor-historic-api.service';
import { environment } from '../../../environments/environment';
import { HistoricResponse } from '../../models/sensor-data/historic-response.model';
import { Sensor } from '../../models/sensor/sensor.model';
import { SensorProfiles } from '../../models/sensor/sensor-profiles.enum';
import { ChartRequest } from '../../models/chart/chart-request.model';
import { ChartType } from '../../models/chart/chart-type.enum';
import { Status } from '../../models/gateway-sensor-status.enum';

describe('SensorHistoricApiService', () => {
  let service: SensorHistoricApiService;
  let httpMock: HttpTestingController;

  const apiUrl = `${environment.apiUrl}`;

  const mockSensor: Sensor = {
    id: 'sensor-1',
    gatewayId: 'gateway-1',
    name: 'Test Sensor',
    profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
    status: Status.ACTIVE,
    dataInterval: 60,
  };

  const mockChartRequest: ChartRequest = {
    sensor: mockSensor,
    chartType: ChartType.HISTORIC,
    timeInterval: {
      from: new Date('2026-01-01T00:00:00.000Z'),
      to: new Date('2026-01-02T00:00:00.000Z'),
    },
    dataPointsCounter: 250,
  };

  const mockHistoricResponse: HistoricResponse = {
    count: 2,
    samples: [
      {
        sensor_id: 'sensor-1',
        gateway_id: 'gateway-1',
        tenant_id: 'tenant-1',
        timestamp: '2026-01-01T00:00:00.000Z',
        profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
        data: { temperature: 20 },
      },
      {
        sensor_id: 'sensor-1',
        gateway_id: 'gateway-1',
        tenant_id: 'tenant-1',
        timestamp: '2026-01-01T00:01:00.000Z',
        profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
        data: { temperature: 21 },
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
    it('should send GET to correct URL with all query params and return a HistoricResponse', () => {
      service.getHistoricData(mockChartRequest).subscribe((response) => {
        expect(response).toEqual(mockHistoricResponse);
      });

      const req = httpMock.expectOne((r) => r.url === `${apiUrl}/sensor/sensor-1/historical_data`);
      expect(req.request.method).toBe('GET');
      expect(req.request.params.get('from_time')).toBe('2026-01-01T00:00:00.000Z');
      expect(req.request.params.get('to_time')).toBe('2026-01-02T00:00:00.000Z');
      expect(req.request.params.get('max_data_points')).toBe('250');
      req.flush(mockHistoricResponse);
    });
  });
});
