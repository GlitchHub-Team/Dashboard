import { TestBed } from '@angular/core/testing';
import { of, Subject, throwError } from 'rxjs';

import { SensorChartService } from './sensor-chart.service';
import { SensorHistoricApiService } from '../sensor-historic-api/sensor-historic-api.service';
import { SensorLiveReadingsApiService } from '../sensor-live-api/sensor-live-readings-api.service';
import { SensorHistoricAdapter } from '../../adapters/sensor-historic.adapter';
import { SensorLiveReadingAdapter } from '../../adapters/sensor-live-reading.adapter';
import { ChartRequest } from '../../models/chart/chart-request.model';
import { ChartType } from '../../models/chart/chart-type.enum';
import { Sensor } from '../../models/sensor/sensor.model';
import { SensorProfiles } from '../../models/sensor/sensor-profiles.enum';
import { HistoricResponse } from '../../models/sensor-data/historic-response.model';
import { RealTimeReading } from '../../models/sensor-data/real-time-reading.model';
import { SensorReading } from '../../models/sensor-data/sensor-reading.model';
import { Status } from '../../models/gateway-sensor-status.enum';

describe('SensorChartService', () => {
  let service: SensorChartService;

  const mockSensor: Sensor = {
    id: 'sensor-1',
    gatewayId: 'gateway-1',
    name: 'Test Sensor',
    profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
    status: Status.ACTIVE,
    dataInterval: 60,
  };

  const mockTimeInterval = {
    from: new Date('2026-01-01T00:00:00.000Z'),
    to: new Date('2026-01-02T00:00:00.000Z'),
  };

  const mockHistoricResponse: HistoricResponse = {
    count: 2,
    resolution: 60,
    data: [],
  };

  const mockAdaptedReadings: SensorReading[] = [
    { timestamp: '2026-01-01T00:00:00.000Z', value: 36.6 },
    { timestamp: '2026-01-01T01:00:00.000Z', value: 37.0 },
  ];

  const mockRawReading: RealTimeReading = {
    sensorId: 'sensor-1',
    profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
    value: 36.6,
  };

  const mockAdaptedReading: SensorReading = {
    timestamp: '2026-01-01T00:00:00.000Z',
    value: 36.6,
  };

  const historicApiMock = {
    getHistoricData: vi.fn(),
  };

  const liveReadingsApiMock = {
    connect: vi.fn(),
    disconnect: vi.fn(),
  };

  const historicAdapterMock = {
    fromResponse: vi.fn(),
    fromDTO: vi.fn(),
  };

  const liveReadingsAdapterMock = {
    fromDTO: vi.fn(),
  };

  const historicRequest: ChartRequest = {
    sensor: mockSensor,
    chartType: ChartType.HISTORIC,
    timeInterval: mockTimeInterval,
  };

  const liveRequest: ChartRequest = {
    sensor: mockSensor,
    chartType: ChartType.REALTIME,
  };

  beforeEach(() => {
    vi.resetAllMocks();

    TestBed.configureTestingModule({
      providers: [
        SensorChartService,
        { provide: SensorHistoricApiService, useValue: historicApiMock },
        { provide: SensorLiveReadingsApiService, useValue: liveReadingsApiMock },
        { provide: SensorHistoricAdapter, useValue: historicAdapterMock },
        { provide: SensorLiveReadingAdapter, useValue: liveReadingsAdapterMock },
      ],
    });

    service = TestBed.inject(SensorChartService);
  });

  it('should be created with default state', () => {
    expect(service).toBeTruthy();
    expect(service.historicReadings()).toEqual([]);
    expect(service.liveReadings()).toEqual([]);
    expect(service.loading()).toBe(false);
    expect(service.connectionStatus()).toBe('disconnected');
    expect(service.error()).toBeNull();
    expect(service.resolution()).toBeNull();
  });

  describe('startChart - historic', () => {
    it('should call getHistoricData with the correct args and set state on success', () => {
      historicApiMock.getHistoricData.mockReturnValue(of(mockHistoricResponse));
      historicAdapterMock.fromResponse.mockReturnValue({
        readings: mockAdaptedReadings,
        resolution: 60,
      });

      service.startChart(historicRequest);

      expect(historicApiMock.getHistoricData).toHaveBeenCalledWith(mockSensor, mockTimeInterval);
      expect(service.historicReadings()).toEqual(mockAdaptedReadings);
      expect(service.resolution()).toBe(60);
      expect(service.loading()).toBe(false);
    });

    it('should set loading to true while fetching and false on complete', () => {
      const subject = new Subject<HistoricResponse>();
      historicApiMock.getHistoricData.mockReturnValue(subject.asObservable());
      historicAdapterMock.fromResponse.mockReturnValue({ readings: [], resolution: 0 });

      service.startChart(historicRequest);
      expect(service.loading()).toBe(true);

      subject.next(mockHistoricResponse);
      subject.complete();
      expect(service.loading()).toBe(false);
    });

    it.each([
      { error: { status: 500, message: 'Server error' }, expected: 'Server error' },
      { error: { status: 500 }, expected: 'Failed to load historic data' },
    ])('should set error "$expected" and loading false on failure', ({ error, expected }) => {
      historicApiMock.getHistoricData.mockReturnValue(throwError(() => error));

      service.startChart(historicRequest);

      expect(service.error()).toBe(expected);
      expect(service.loading()).toBe(false);
    });
  });

  describe('startChart - live', () => {
    it('should call connect with the correct sensor and transition connectionStatus', () => {
      const subject = new Subject<RealTimeReading>();
      liveReadingsApiMock.connect.mockReturnValue(subject.asObservable());
      liveReadingsAdapterMock.fromDTO.mockReturnValue(mockAdaptedReading);

      service.startChart(liveRequest);
      expect(liveReadingsApiMock.connect).toHaveBeenCalledWith(mockSensor);
      expect(service.connectionStatus()).toBe('connecting');

      subject.next(mockRawReading);
      expect(service.connectionStatus()).toBe('connected');
    });

    it('should append live readings as they arrive', () => {
      const subject = new Subject<RealTimeReading>();
      liveReadingsApiMock.connect.mockReturnValue(subject.asObservable());
      liveReadingsAdapterMock.fromDTO.mockReturnValue(mockAdaptedReading);

      service.startChart(liveRequest);

      subject.next(mockRawReading);
      subject.next(mockRawReading);
      expect(service.liveReadings().length).toBe(2);
    });

    it('should trim live readings buffer to MAX_LIVE_READINGS (50)', () => {
      const subject = new Subject<RealTimeReading>();
      liveReadingsApiMock.connect.mockReturnValue(subject.asObservable());
      liveReadingsAdapterMock.fromDTO.mockImplementation((_, i) => mockAdaptedReading);

      service.startChart(liveRequest);

      for (let i = 0; i < 55; i++) {
        subject.next(mockRawReading);
      }

      expect(service.liveReadings().length).toBe(50);
    });

    it('should set connectionStatus to disconnected on complete', () => {
      const subject = new Subject<RealTimeReading>();
      liveReadingsApiMock.connect.mockReturnValue(subject.asObservable());

      service.startChart(liveRequest);
      subject.complete();

      expect(service.connectionStatus()).toBe('disconnected');
    });

    it.each([
      { error: { status: 500, message: 'Connection lost' }, expected: 'Connection lost' },
      { error: { status: 500 }, expected: 'Failed to load live readings' },
    ])(
      'should set error "$expected" and connectionStatus disconnected on failure',
      ({ error, expected }) => {
        liveReadingsApiMock.connect.mockReturnValue(throwError(() => error));

        service.startChart(liveRequest);

        expect(service.error()).toBe(expected);
        expect(service.connectionStatus()).toBe('disconnected');
      },
    );
  });

  describe('stopChart', () => {
    it('should call disconnect and set connectionStatus to disconnected', () => {
      const subject = new Subject<RealTimeReading>();
      liveReadingsApiMock.connect.mockReturnValue(subject.asObservable());
      liveReadingsAdapterMock.fromDTO.mockReturnValue(mockAdaptedReading);

      service.startChart(liveRequest);
      subject.next(mockRawReading);
      expect(service.connectionStatus()).toBe('connected');

      service.stopChart();

      expect(liveReadingsApiMock.disconnect).toHaveBeenCalled();
      expect(service.connectionStatus()).toBe('disconnected');
    });

    it('should not throw when called with no active chart', () => {
      expect(() => service.stopChart()).not.toThrow();
    });
  });

  describe('startChart - reset', () => {
    it('should clear state from a previous chart before starting a new one', () => {
      historicApiMock.getHistoricData.mockReturnValue(of(mockHistoricResponse));
      historicAdapterMock.fromResponse.mockReturnValue({
        readings: mockAdaptedReadings,
        resolution: 60,
      });
      service.startChart(historicRequest);
      expect(service.historicReadings().length).toBe(2);

      historicApiMock.getHistoricData.mockReturnValue(of(mockHistoricResponse));
      historicAdapterMock.fromResponse.mockReturnValue({ readings: [], resolution: 0 });
      service.startChart(historicRequest);

      expect(service.historicReadings()).toEqual([]);
      expect(service.error()).toBeNull();
      expect(service.resolution()).toBe(0);
    });

    it('should unsubscribe from previous live subscription when starting a new chart', () => {
      const subject = new Subject<RealTimeReading>();
      liveReadingsApiMock.connect.mockReturnValue(subject.asObservable());

      service.startChart(liveRequest);

      historicApiMock.getHistoricData.mockReturnValue(of(mockHistoricResponse));
      historicAdapterMock.fromResponse.mockReturnValue({ readings: [], resolution: 0 });
      service.startChart(historicRequest);

      liveReadingsAdapterMock.fromDTO.mockReturnValue(mockAdaptedReading);
      subject.next(mockRawReading);

      expect(service.liveReadings()).toEqual([]);
    });
  });
});
