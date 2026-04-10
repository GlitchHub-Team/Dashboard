import { TestBed } from '@angular/core/testing';
import { Subject } from 'rxjs';

import { environment } from '../../../environments/environment';
import { Sensor } from '../../models/sensor/sensor.model';
import { SensorProfiles } from '../../models/sensor/sensor-profiles.enum';
import { Status } from '../../models/gateway-sensor-status.enum';
import { ChartRequest } from '../../models/chart/chart-request.model';
import { ChartType } from '../../models/chart/chart-type.enum';
import { SensorLiveReadingsApiService } from './sensor-live-readings-api.service';
import { TokenStorageService } from '../token-storage/token-storage.service';

describe('SensorLiveReadingsApiService', () => {
  let service: SensorLiveReadingsApiService;
  let mockTokenService: Partial<TokenStorageService>;
  let createWebSocketSpy: ReturnType<typeof vi.spyOn>;

  const wsUrl = environment.wsUrl;
  const mockToken = 'mock-jwt-token';

  const mockSensor: Sensor = {
    id: 'sensor-1',
    gatewayId: 'gateway-1',
    name: 'Temperature',
    profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
    dataInterval: 60,
    status: Status.ACTIVE,
  };

  const mockRequest: ChartRequest = {
    sensor: mockSensor,
    chartType: ChartType.REALTIME,
    tenantId: 'tenant-1',
  };

  const createMockSocket = () => {
    const subject = new Subject();
    return {
      pipe: vi.fn().mockReturnValue(subject.asObservable()),
      complete: vi.fn(),
    };
  };

  beforeEach(() => {
    vi.clearAllMocks();

    mockTokenService = {
      getToken: vi.fn().mockReturnValue(mockToken),
    };

    TestBed.configureTestingModule({
      providers: [
        SensorLiveReadingsApiService,
        { provide: TokenStorageService, useValue: mockTokenService },
      ],
    });

    service = TestBed.inject(SensorLiveReadingsApiService);
    createWebSocketSpy = vi.spyOn(service as any, 'createWebSocket');
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  describe('connect', () => {
    it('should create a WebSocket with the correct URL and return an observable', () => {
      const mockSocket = createMockSocket();
      createWebSocketSpy.mockReturnValue(mockSocket as any);

      const result = service.connect(mockRequest);

      expect(createWebSocketSpy).toHaveBeenCalledWith(
        `${wsUrl}/tenant/${mockRequest.tenantId}/sensor/${mockSensor.id}/real_time_data?jwt=${mockToken}`,
      );
      expect(mockSocket.pipe).toHaveBeenCalled();
      expect(result).toBeDefined();
    });

    it('should use an empty string as JWT when token is null', () => {
      vi.mocked(mockTokenService.getToken!).mockReturnValue(null);
      const mockSocket = createMockSocket();
      createWebSocketSpy.mockReturnValue(mockSocket as any);

      service.connect(mockRequest);

      expect(createWebSocketSpy).toHaveBeenCalledWith(
        `${wsUrl}/tenant/${mockRequest.tenantId}/sensor/${mockSensor.id}/real_time_data?jwt=`,
      );
    });

    it('should disconnect the previous socket before creating a new one', () => {
      const mockSocket1 = createMockSocket();
      const mockSocket2 = createMockSocket();
      createWebSocketSpy
        .mockReturnValueOnce(mockSocket1 as any)
        .mockReturnValueOnce(mockSocket2 as any);

      service.connect(mockRequest);
      service.connect(mockRequest);

      expect(createWebSocketSpy).toHaveBeenCalledTimes(2);
      expect(mockSocket1.complete).toHaveBeenCalledTimes(1);
    });
  });

  describe('disconnect', () => {
    it('should complete the socket when there is an active connection', () => {
      const mockSocket = createMockSocket();
      createWebSocketSpy.mockReturnValue(mockSocket as any);

      service.connect(mockRequest);
      service.disconnect();

      expect(mockSocket.complete).toHaveBeenCalledTimes(1);
    });

    it('should only complete once when called multiple times', () => {
      const mockSocket = createMockSocket();
      createWebSocketSpy.mockReturnValue(mockSocket as any);

      service.connect(mockRequest);
      service.disconnect();
      service.disconnect();

      expect(mockSocket.complete).toHaveBeenCalledTimes(1);
    });

    it('should not throw when there is no active connection', () => {
      expect(() => service.disconnect()).not.toThrow();
    });
  });
});