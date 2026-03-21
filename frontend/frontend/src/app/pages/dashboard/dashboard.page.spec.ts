import { describe, it, expect, vi, beforeEach } from 'vitest';
import { signal, WritableSignal, NO_ERRORS_SCHEMA } from '@angular/core';
import { TestBed, ComponentFixture } from '@angular/core/testing';
import { MatDialog } from '@angular/material/dialog';
import { MatSnackBar } from '@angular/material/snack-bar';
import { PageEvent } from '@angular/material/paginator';

import { DashboardPage } from './dashboard.page';
import { DashboardService } from '../../services/dashboard/dashboard.service';
import { SensorChartService } from '../../services/sensor-chart/sensor-chart.service';
import { Gateway } from '../../models/gateway/gateway.model';
import { GatewayStatus } from '../../models/gateway/gateway-status.enum';
import { ChartRequest } from '../../models/chart/chart-request.model';
import { ChartType } from '../../models/chart/chart-type.enum';
import { Sensor } from '../../models/sensor/sensor.model';
import { SensorProfiles } from '../../models/sensor/sensor-profiles.enum';

describe('DashboardPage', () => {
  let fixture: ComponentFixture<DashboardPage>;
  let component: DashboardPage;

  let gatewayListSig: WritableSignal<Gateway[]>;
  let gatewayTotalSig: WritableSignal<number>;
  let gatewayPageIndexSig: WritableSignal<number>;
  let gatewayLimitSig: WritableSignal<number>;
  let gatewayLoadingSig: WritableSignal<boolean>;
  let gatewayErrorSig: WritableSignal<string | null>;

  let sensorListSig: WritableSignal<Sensor[]>;
  let sensorTotalSig: WritableSignal<number>;
  let sensorPageIndexSig: WritableSignal<number>;
  let sensorLimitSig: WritableSignal<number>;
  let sensorLoadingSig: WritableSignal<boolean>;
  let sensorErrorSig: WritableSignal<string | null>;

  let expandedGatewaySig: WritableSignal<Gateway | null>;
  let selectedChartSig: WritableSignal<ChartRequest | null>;
  let canSendCommandsSig: WritableSignal<boolean>;

  const mockGateway: Gateway = {
    id: 'gw-1',
    tenantId: 'tenant-01',
    name: 'Gateway 1',
    status: GatewayStatus.ONLINE,
  };

  const mockSensor: Sensor = {
    id: 'sensor-1',
    gatewayId: 'gw-1',
    name: 'Heart Rate Sensor',
    profile: SensorProfiles.HEART_RATE_SERVICE,
    dataInterval: 1000,
  };

  const mockChartRequest: ChartRequest = {
    sensor: mockSensor,
    chartType: ChartType.HISTORIC,
    timeInterval: { from: new Date('2025-01-01'), to: new Date('2025-01-02') },
  };

  let dashboardServiceMock: any;
  let snackBarMock: any;

  beforeEach(async () => {
    gatewayListSig = signal<Gateway[]>([]);
    gatewayTotalSig = signal(0);
    gatewayPageIndexSig = signal(0);
    gatewayLimitSig = signal(10);
    gatewayLoadingSig = signal(false);
    gatewayErrorSig = signal<string | null>(null);

    sensorListSig = signal<Sensor[]>([]);
    sensorTotalSig = signal(0);
    sensorPageIndexSig = signal(0);
    sensorLimitSig = signal(10);
    sensorLoadingSig = signal(false);
    sensorErrorSig = signal<string | null>(null);

    expandedGatewaySig = signal<Gateway | null>(null);
    selectedChartSig = signal<ChartRequest | null>(null);
    canSendCommandsSig = signal(true);

    dashboardServiceMock = {
      gatewayList: gatewayListSig,
      gatewayTotal: gatewayTotalSig,
      gatewayPageIndex: gatewayPageIndexSig,
      gatewayLimit: gatewayLimitSig,
      gatewayLoading: gatewayLoadingSig,
      gatewayError: gatewayErrorSig,
      sensorList: sensorListSig,
      sensorTotal: sensorTotalSig,
      sensorPageIndex: sensorPageIndexSig,
      sensorLimit: sensorLimitSig,
      sensorLoading: sensorLoadingSig,
      sensorError: sensorErrorSig,
      expandedGateway: expandedGatewaySig,
      selectedChart: selectedChartSig,
      canSendCommands: canSendCommandsSig,
      loadDashboard: vi.fn(),
      closeChart: vi.fn(),
      toggleExpandedGateway: vi.fn(),
      changeGatewayPage: vi.fn(),
      changeSensorPage: vi.fn(),
      openChart: vi.fn(),
    };

    snackBarMock = { open: vi.fn() };

    await TestBed.configureTestingModule({
      imports: [DashboardPage],
      schemas: [NO_ERRORS_SCHEMA],
      providers: [
        { provide: DashboardService, useValue: dashboardServiceMock },
        { provide: MatSnackBar, useValue: snackBarMock },
        { provide: MatDialog, useValue: {} },
        {
          provide: SensorChartService,
          useValue: {
            historicReadings: signal([]),
            liveReadings: signal([]),
            loading: signal(false),
            connectionStatus: signal('disconnected'),
            error: signal(null),
            startChart: vi.fn(),
            stopChart: vi.fn(),
          },
        },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(DashboardPage);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  function query(selector: string): HTMLElement | null {
    return fixture.nativeElement.querySelector(selector);
  }

  describe('ngOnInit', () => {
    it('should call dashboardService.loadDashboard on init', () => {
      expect(dashboardServiceMock.loadDashboard).toHaveBeenCalledOnce();
    });
  });

  describe('ngOnDestroy', () => {
    it('should call dashboardService.closeChart on destroy', () => {
      fixture.destroy();
      expect(dashboardServiceMock.closeChart).toHaveBeenCalledOnce();
    });
  });

  describe('error banner', () => {
    it('should display error banner when gatewayError is set', () => {
      gatewayErrorSig.set('Gateway connection failed');
      fixture.detectChanges();

      const banner = query('.error-banner');
      expect(banner).not.toBeNull();
      expect(banner!.textContent).toContain('Gateway connection failed');
    });

    it('should display error banner when sensorError is set', () => {
      sensorErrorSig.set('Sensor timeout');
      fixture.detectChanges();

      const banner = query('.error-banner');
      expect(banner).not.toBeNull();
      expect(banner!.textContent).toContain('Sensor timeout');
    });

    it('should prefer gatewayError over sensorError (nullish coalescing)', () => {
      gatewayErrorSig.set('Gateway error');
      sensorErrorSig.set('Sensor error');
      fixture.detectChanges();

      const banner = query('.error-banner');
      expect(banner).not.toBeNull();
      expect(banner!.textContent).toContain('Gateway error');
    });

    it('should show sensorError when gatewayError is null', () => {
      gatewayErrorSig.set(null);
      sensorErrorSig.set('Sensor error');
      fixture.detectChanges();

      const banner = query('.error-banner');
      expect(banner).not.toBeNull();
      expect(banner!.textContent).toContain('Sensor error');
    });

    it('should not display error banner when there are no errors', () => {
      gatewayErrorSig.set(null);
      sensorErrorSig.set(null);
      fixture.detectChanges();

      const banner = query('.error-banner');
      expect(banner).toBeNull();
    });
  });

  describe('when canSendCommands is true', () => {
    beforeEach(() => {
      canSendCommandsSig.set(true);
      fixture.detectChanges();
    });

    it('should render app-dashboard-gateway-table', () => {
      expect(query('app-dashboard-gateway-table')).not.toBeNull();
      expect(query('app-dashboard-sensor-table')).toBeNull();
    });
  });

  describe('when canSendCommands is false', () => {
    beforeEach(() => {
      canSendCommandsSig.set(false);
      fixture.detectChanges();
    });

    it('should render app-dashboard-sensor-table', () => {
      expect(query('app-dashboard-sensor-table')).not.toBeNull();
      expect(query('app-dashboard-gateway-table')).toBeNull();
    });
  });

  describe('chart container', () => {
    it('should render app-chart-container when selectedChart is set', () => {
      selectedChartSig.set(mockChartRequest);
      fixture.detectChanges();

      expect(query('app-chart-container')).not.toBeNull();
    });

    it('should not render app-chart-container when selectedChart is null', () => {
      selectedChartSig.set(null);
      fixture.detectChanges();

      expect(query('app-chart-container')).toBeNull();
    });
  });

  describe('onExpandedGatewayChange', () => {
    it('should call toggleExpandedGateway on the service', () => {
      component['onExpandedGatewayChange'](mockGateway);
      expect(dashboardServiceMock.toggleExpandedGateway).toHaveBeenCalledWith(mockGateway);
    });
  });

  describe('onGatewayPageChange', () => {
    it('should call changeGatewayPage with pageIndex and pageSize', () => {
      const event: PageEvent = { pageIndex: 3, pageSize: 25, length: 100 };
      component['onGatewayPageChange'](event);
      expect(dashboardServiceMock.changeGatewayPage).toHaveBeenCalledWith(3, 25);
    });
  });

  describe('onSensorPageChange', () => {
    it('should call changeSensorPage with pageIndex and pageSize', () => {
      const event: PageEvent = { pageIndex: 1, pageSize: 10, length: 50 };
      component['onSensorPageChange'](event);
      expect(dashboardServiceMock.changeSensorPage).toHaveBeenCalledWith(1, 10);
    });
  });

  describe('onCommandRequested', () => {
    it('should open snackbar with gateway id', () => {
      component['onCommandRequested'](mockGateway);
      expect(snackBarMock.open).toHaveBeenCalledWith('gw-1', 'Close', { duration: 2000 });
    });
  });

  describe('onChartOpen', () => {
    it('should call openChart on the service', () => {
      component['onChartOpen'](mockChartRequest);
      expect(dashboardServiceMock.openChart).toHaveBeenCalledWith(mockChartRequest);
    });
  });

  describe('onChartClosed', () => {
    it('should call closeChart on the service', () => {
      dashboardServiceMock.closeChart.mockClear();
      component['onChartClosed']();
      expect(dashboardServiceMock.closeChart).toHaveBeenCalledOnce();
    });
  });

  describe('signal reactivity', () => {
    it('should toggle from gateway table to sensor table when canSendCommands changes', () => {
      canSendCommandsSig.set(true);
      fixture.detectChanges();
      expect(query('app-dashboard-gateway-table')).not.toBeNull();

      canSendCommandsSig.set(false);
      fixture.detectChanges();
      expect(query('app-dashboard-sensor-table')).not.toBeNull();
      expect(query('app-dashboard-gateway-table')).toBeNull();
    });

    it('should show and hide chart container reactively', () => {
      selectedChartSig.set(null);
      fixture.detectChanges();
      expect(query('app-chart-container')).toBeNull();

      selectedChartSig.set(mockChartRequest);
      fixture.detectChanges();
      expect(query('app-chart-container')).not.toBeNull();

      selectedChartSig.set(null);
      fixture.detectChanges();
      expect(query('app-chart-container')).toBeNull();
    });

    it('should show and hide error banner reactively', () => {
      gatewayErrorSig.set(null);
      sensorErrorSig.set(null);
      fixture.detectChanges();
      expect(query('.error-banner')).toBeNull();

      gatewayErrorSig.set('New error');
      fixture.detectChanges();
      expect(query('.error-banner')).not.toBeNull();

      gatewayErrorSig.set(null);
      fixture.detectChanges();
      expect(query('.error-banner')).toBeNull();
    });
  });
});
