import { Subject } from 'rxjs';
import { webSocket } from 'rxjs/webSocket';

import { environment } from '../../../environments/environment';
import { Sensor } from '../../models/sensor/sensor.model';
import { SensorProfiles } from '../../models/sensor/sensor-profiles.enum';
import { Status } from '../../models/gateway-sensor-status.enum';

vi.mock('rxjs/webSocket', () => ({
  webSocket: vi.fn(),
  WebSocketSubject: vi.fn(),
}));

describe('SensorLiveReadingsApiService', () => {
  let service: any;

  const apiUrl = environment.apiUrl;

  const mockSensor: Sensor = {
    id: 'sensor-1',
    gatewayId: 'gateway-1',
    name: 'Temperature',
    profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
    dataInterval: 60,
    status: Status.ACTIVE,
  };

  const createMockSocket = () => ({
    asObservable: vi.fn().mockReturnValue(new Subject().asObservable()),
    complete: vi.fn(),
  });

  beforeEach(async () => {
    vi.resetAllMocks();
    vi.resetModules();
    const module = await import('./sensor-live-readings-api.service');
    service = new module.SensorLiveReadingsApiService();
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  describe('connect', () => {
    it('should call webSocket with the correct URL', () => {
      const mockSocket = createMockSocket();
      vi.mocked(webSocket).mockReturnValue(mockSocket as any);

      service.connect(mockSensor);

      expect(webSocket).toHaveBeenCalledWith(
        `${apiUrl}/live?gatewayId=gateway-1&sensorId=sensor-1`,
      );
    });

    it('should return the observable from the socket', () => {
      const mockSocket = createMockSocket();
      vi.mocked(webSocket).mockReturnValue(mockSocket as any);

      const result = service.connect(mockSensor);

      expect(mockSocket.asObservable).toHaveBeenCalled();
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
    it('should complete the socket after connect', () => {
      const mockSocket = createMockSocket();
      vi.mocked(webSocket).mockReturnValue(mockSocket as any);

      service.connect(mockSensor);
      service.disconnect();

      expect(mockSocket.complete).toHaveBeenCalledTimes(1);
    });

    it('should not throw when there is no active connection', () => {
      expect(() => service.disconnect()).not.toThrow();
    });

    it('should set socket to null so a second disconnect is a no-op', () => {
      const mockSocket = createMockSocket();
      vi.mocked(webSocket).mockReturnValue(mockSocket as any);

      service.connect(mockSensor);
      service.disconnect();
      service.disconnect();

      expect(mockSocket.complete).toHaveBeenCalledTimes(1);
    });
  });
});
