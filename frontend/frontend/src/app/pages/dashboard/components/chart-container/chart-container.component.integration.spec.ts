import { ComponentFixture, TestBed } from '@angular/core/testing';
import { WritableSignal, signal } from '@angular/core';
import { By } from '@angular/platform-browser';
import { describe, expect, it, vi, afterEach } from 'vitest';

import { ChartContainerComponent } from './chart-container.component';
import { HistoricChartComponent } from './components/historic-chart/historic-chart.component';
import { RealTimeChartComponent } from './components/real-time-chart/real-time-chart.component';
import { SensorChartService } from '../../../../services/sensor-chart/sensor-chart.service';
import { ChartRequest } from '../../../../models/chart/chart-request.model';
import { ChartType } from '../../../../models/chart/chart-type.enum';
import { Sensor } from '../../../../models/sensor/sensor.model';
import { SensorReading } from '../../../../models/sensor-data/sensor-reading.model';
import { FieldDescriptor } from '../../../../models/sensor-data/field-descriptor.model';
import { SensorStatus } from '../../../../models/sensor-status.enum';
import { SensorProfiles } from '../../../../models/sensor/sensor-profiles.enum';

type ConnectionStatus = 'connected' | 'connecting' | 'disconnected' | 'reconnecting';

const mockSensor: Sensor = {
  id: 'sensor-1',
  gatewayId: 'gw-1',
  name: 'Temperature',
  profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
  status: SensorStatus.ACTIVE,
  dataInterval: 30,
};

const mockReadings: SensorReading[] = [
  { timestamp: new Date(1700000000000).toISOString(), value: { temperature: 36.5 } },
  { timestamp: new Date(1700000060000).toISOString(), value: { temperature: 36.7 } },
  { timestamp: new Date(1700000120000).toISOString(), value: { temperature: 36.6 } },
];

const historicRequest: ChartRequest = { sensor: mockSensor, chartType: ChartType.HISTORIC };
const realtimeRequest: ChartRequest = { sensor: mockSensor, chartType: ChartType.REALTIME };

function createChartServiceMock() {
  return {
    historicReadings: signal<SensorReading[]>([]),
    liveReadings: signal<SensorReading[]>([]),
    fields: signal<FieldDescriptor[]>([]),
    loading: signal(false),
    connectionStatus: signal<ConnectionStatus>('disconnected'),
    error: signal<string | null>(null),
    startChart: vi.fn(),
    stopChart: vi.fn(),
  };
}

function setupTestBed(chartRequest: ChartRequest) {
  const chartServiceMock = createChartServiceMock();

  TestBed.configureTestingModule({
    imports: [ChartContainerComponent, HistoricChartComponent, RealTimeChartComponent],
    providers: [{ provide: SensorChartService, useValue: chartServiceMock }],
  });

  const fixture = TestBed.createComponent(ChartContainerComponent);
  fixture.componentRef.setInput('chartRequest', chartRequest);

  return { fixture, chartServiceMock };
}

const el = (f: ComponentFixture<ChartContainerComponent>): HTMLElement => f.nativeElement;
const getHistoricChart = (f: ComponentFixture<ChartContainerComponent>) =>
  f.debugElement.query(By.directive(HistoricChartComponent));
const getRealTimeChart = (f: ComponentFixture<ChartContainerComponent>) =>
  f.debugElement.query(By.directive(RealTimeChartComponent));

describe('ChartContainerComponent (Integration)', () => {
  afterEach(() => {
    TestBed.resetTestingModule();
    vi.resetAllMocks();
  });

  describe('Service Interaction', () => {
    it('should call startChart via effect on init', () => {
      const { fixture, chartServiceMock } = setupTestBed(historicRequest);
      fixture.detectChanges();

      expect(chartServiceMock.startChart).toHaveBeenCalledWith(historicRequest);
    });

    it('should call stopChart and emit chartClosed on close button click', () => {
      const { fixture, chartServiceMock } = setupTestBed(historicRequest);
      fixture.detectChanges();

      const spy = vi.fn();
      fixture.componentInstance.chartClosed.subscribe(spy);

      el(fixture).querySelector<HTMLButtonElement>('button[mat-icon-button]')!.click();

      expect(chartServiceMock.stopChart).toHaveBeenCalled();
      expect(spy).toHaveBeenCalled();
    });

    it('should call stopChart on destroy', () => {
      const { fixture, chartServiceMock } = setupTestBed(historicRequest);
      fixture.detectChanges();
      fixture.destroy();

      expect(chartServiceMock.stopChart).toHaveBeenCalled();
    });
  });

  describe('Card Title', () => {
    it('should display sensor name and profile label', () => {
      const { fixture } = setupTestBed(historicRequest);
      fixture.detectChanges();

      const title = el(fixture).querySelector('mat-card-title')!;
      expect(title.textContent).toContain('Temperature');
      expect(title.textContent).toContain('Thermometer');
    });
  });

  describe('Loading State', () => {
    it('should show spinner and hide chart when loading', () => {
      const { fixture, chartServiceMock } = setupTestBed(historicRequest);
      (chartServiceMock.loading as WritableSignal<boolean>).set(true);
      fixture.detectChanges();

      expect(el(fixture).querySelector('mat-spinner')).toBeTruthy();
      expect(getHistoricChart(fixture)).toBeFalsy();
    });
  });

  describe('Error States', () => {
    it('should show error with chart-error class when not reconnecting', () => {
      const { fixture, chartServiceMock } = setupTestBed(historicRequest);
      (chartServiceMock.error as WritableSignal<string | null>).set('Connection lost');
      (chartServiceMock.connectionStatus as WritableSignal<ConnectionStatus>).set('disconnected');
      fixture.detectChanges();

      const errorDiv = el(fixture).querySelector('.chart-error');
      expect(errorDiv).toBeTruthy();
      expect(errorDiv!.textContent).toContain('Connection lost');
      expect(el(fixture).querySelector('.chart-warning')).toBeFalsy();
    });

    it('should show error with chart-warning class when reconnecting', () => {
      const { fixture, chartServiceMock } = setupTestBed(realtimeRequest);
      (chartServiceMock.error as WritableSignal<string | null>).set('Attempting reconnect...');
      (chartServiceMock.connectionStatus as WritableSignal<ConnectionStatus>).set('reconnecting');
      fixture.detectChanges();

      const warningDiv = el(fixture).querySelector('.chart-warning');
      expect(warningDiv).toBeTruthy();
      expect(warningDiv!.textContent).toContain('Attempting reconnect...');
      expect(el(fixture).querySelector('.chart-error')).toBeFalsy();
    });
  });

  describe('Historic Chart', () => {
    it('should render HistoricChartComponent with correct data when readings exist', () => {
      const { fixture, chartServiceMock } = setupTestBed(historicRequest);
      (chartServiceMock.historicReadings as WritableSignal<SensorReading[]>).set(mockReadings);
      fixture.detectChanges();

      const historic = getHistoricChart(fixture);
      expect(historic).toBeTruthy();
      expect(historic.componentInstance.readings()).toEqual(mockReadings);
      expect(historic.componentInstance.sensor()).toEqual(mockSensor);
      expect(getRealTimeChart(fixture)).toBeFalsy();
    });

    it('should NOT render HistoricChartComponent when readings are empty', () => {
      const { fixture, chartServiceMock } = setupTestBed(historicRequest);
      (chartServiceMock.historicReadings as WritableSignal<SensorReading[]>).set([]);
      fixture.detectChanges();

      expect(getHistoricChart(fixture)).toBeFalsy();
    });
  });

  describe('Real-Time Chart', () => {
    it('should render RealTimeChartComponent with correct data when readings exist', () => {
      const { fixture, chartServiceMock } = setupTestBed(realtimeRequest);
      (chartServiceMock.liveReadings as WritableSignal<SensorReading[]>).set(mockReadings);
      fixture.detectChanges();

      const realtime = getRealTimeChart(fixture);
      expect(realtime).toBeTruthy();
      expect(realtime.componentInstance.readings()).toEqual(mockReadings);
      expect(realtime.componentInstance.sensor()).toEqual(mockSensor);
      expect(getHistoricChart(fixture)).toBeFalsy();
    });

    it('should NOT render RealTimeChartComponent when readings are empty', () => {
      const { fixture, chartServiceMock } = setupTestBed(realtimeRequest);
      (chartServiceMock.liveReadings as WritableSignal<SensorReading[]>).set([]);
      fixture.detectChanges();

      expect(getRealTimeChart(fixture)).toBeFalsy();
    });
  });

  describe('Connection Status (Live Chart Only)', () => {
    it('should NOT show connection status for historic chart', () => {
      const { fixture } = setupTestBed(historicRequest);
      fixture.detectChanges();

      expect(el(fixture).querySelector('.connection-status')).toBeFalsy();
    });

    it.each([
      ['connected', 'Connected', 'status-connected'],
      ['connecting', 'Connecting...', 'status-connecting'],
      ['disconnected', 'Disconnected', 'status-disconnected'],
      ['reconnecting', 'Reconnecting...', 'status-reconnecting'],
    ] as const)('should display "%s" with label "%s" and class "%s"', (status, label, cssClass) => {
      const { fixture, chartServiceMock } = setupTestBed(realtimeRequest);
      (chartServiceMock.connectionStatus as WritableSignal<ConnectionStatus>).set(status);
      fixture.detectChanges();

      const statusEl = el(fixture).querySelector('.connection-status')!;
      expect(statusEl.textContent).toContain(label);
      expect(statusEl.classList).toContain(cssClass);
    });
  });
});
