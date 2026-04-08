import { TestBed } from '@angular/core/testing';
import { Subject } from 'rxjs';
import { webSocket } from 'rxjs/webSocket';

import { environment } from '../../../environments/environment';
import { Sensor } from '../../models/sensor/sensor.model';
import { SensorProfiles } from '../../models/sensor/sensor-profiles.enum';
import { Status } from '../../models/gateway-sensor-status.enum';
import { SensorLiveReadingsApiService } from './sensor-live-readings-api.service';
import { TokenStorageService } from '../token-storage/token-storage.service';

vi.mock('rxjs/webSocket', () => ({
  webSocket: vi.fn(),
  WebSocketSubject: vi.fn(),
}));

describe('SensorLiveReadingsApiService', () => {
  let service: SensorLiveReadingsApiService;
  let mockTokenService: Partial<TokenStorageService>;

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

  const createMockSocket = () => {
    const subject = new Subject();
    return {
      pipe: vi.fn().mockReturnValue(subject.asObservable()),
      complete: vi.fn(),
    };
  };

  beforeEach(() => {
    vi.resetAllMocks();

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
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  describe('connect', () => {
    it('should call webSocket with the correct URL including the token and return the socket observable', () => {
      const mockSocket = createMockSocket();
      vi.mocked(webSocket).mockReturnValue(mockSocket as any);

      const result = service.connect(mockSensor);

      expect(webSocket).toHaveBeenCalledWith(
        `${wsUrl}/sensor/${mockSensor.id}/real_time_data?jwt=${mockToken}`,
      );
      expect(mockSocket.pipe).toHaveBeenCalled();
      expect(result).toBeDefined();
    });

    it('should create a new socket on each call', () => {
      const mockSocket1 = createMockSocket();
      const mockSocket2 = createMockSocket();
      vi.mocked(webSocket)
        .mockReturnValueOnce(mockSocket1 as any)
        .mockReturnValueOnce(mockSocket2 as any);

      service.connect(mockSensor);
      service.connect(mockSensor);

      expect(webSocket).toHaveBeenCalledTimes(2);
    });
  });

  describe('disconnect', () => {
    it('should complete the socket and be a no-op on subsequent calls', () => {
      const mockSocket = createMockSocket();
      vi.mocked(webSocket).mockReturnValue(mockSocket as any);

      service.connect(mockSensor);
      service.disconnect();
      expect(mockSocket.complete).toHaveBeenCalledTimes(1);

      service.disconnect();
      expect(mockSocket.complete).toHaveBeenCalledTimes(1);
    });

    it('should not throw when there is no active connection', () => {
      expect(() => service.disconnect()).not.toThrow();
    });
  });
});
