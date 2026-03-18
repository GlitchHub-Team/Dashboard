import { ComponentFixture, TestBed } from '@angular/core/testing';
import { NO_ERRORS_SCHEMA, signal, WritableSignal } from '@angular/core';
import { MatDialog } from '@angular/material/dialog';
import { MatSnackBar } from '@angular/material/snack-bar';
import { By } from '@angular/platform-browser';

import { DashboardPage } from './dashboard.page';
import { DashboardService } from '../../services/dashboard/dashboard.service';
import { Gateway } from '../../models/gateway/gateway.model';
import { GatewayStatus } from '../../models/gateway/gateway-status.enum';
import { Sensor } from '../../models/sensor/sensor.model';
import { ChartRequest } from '../../models/chart-request.model';
import { ChartType } from '../../models/chart-type.enum';
import { SensorProfiles } from '../../models/sensor/sensor-profiles.enum';

describe('DashboardPage', () => {
  let component: DashboardPage;
  let fixture: ComponentFixture<DashboardPage>;

  let gatewayListSignal: WritableSignal<Gateway[]>;
  let gatewayTotalSignal: WritableSignal<number>;
  let gatewayPageIndexSignal: WritableSignal<number>;
  let gatewayLimitSignal: WritableSignal<number>;
  let gatewayLoadingSignal: WritableSignal<boolean>;
  let gatewayErrorSignal: WritableSignal<string | null>;

  let sensorListSignal: WritableSignal<Sensor[]>;
  let sensorTotalSignal: WritableSignal<number>;
  let sensorPageIndexSignal: WritableSignal<number>;
  let sensorLimitSignal: WritableSignal<number>;
  let sensorLoadingSignal: WritableSignal<boolean>;
  let sensorErrorSignal: WritableSignal<string | null>;

  let expandedGatewaySignal: WritableSignal<Gateway | null>;
  let selectedChartSignal: WritableSignal<ChartRequest | null>;
  let canSendCommandsSignal: WritableSignal<boolean>;

  const mockGateways: Gateway[] = [
    { id: 'gw-1', name: 'Gateway 1', tenantId: 'tenant-1', status: GatewayStatus.ONLINE },
    { id: 'gw-2', name: 'Gateway 2', tenantId: 'tenant-1', status: GatewayStatus.OFFLINE },
  ];

  const mockSensors: Sensor[] = [
    {
      id: 's-1',
      gatewayId: 'gw-1',
      name: 'Temperature',
      profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
    },
  ];

  const mockChartRequest: ChartRequest = {
    sensor: mockSensors[0],
    chartType: ChartType.HISTORIC,
    timeInterval: null!,
  };

  let dashboardServiceMock: any;
  const dialogMock = { open: vi.fn() };
  const snackBarMock = { open: vi.fn() };

  beforeEach(async () => {
    vi.resetAllMocks();

    gatewayListSignal = signal<Gateway[]>([]);
    gatewayTotalSignal = signal(0);
    gatewayPageIndexSignal = signal(0);
    gatewayLimitSignal = signal(10);
    gatewayLoadingSignal = signal(false);
    gatewayErrorSignal = signal<string | null>(null);

    sensorListSignal = signal<Sensor[]>([]);
    sensorTotalSignal = signal(0);
    sensorPageIndexSignal = signal(0);
    sensorLimitSignal = signal(10);
    sensorLoadingSignal = signal(false);
    sensorErrorSignal = signal<string | null>(null);

    expandedGatewaySignal = signal<Gateway | null>(null);
    selectedChartSignal = signal<ChartRequest | null>(null);
    canSendCommandsSignal = signal(false);

    dashboardServiceMock = {
      gatewayList: gatewayListSignal.asReadonly(),
      gatewayTotal: gatewayTotalSignal.asReadonly(),
      gatewayPageIndex: gatewayPageIndexSignal.asReadonly(),
      gatewayLimit: gatewayLimitSignal.asReadonly(),
      gatewayLoading: gatewayLoadingSignal.asReadonly(),
      gatewayError: gatewayErrorSignal.asReadonly(),

      sensorList: sensorListSignal.asReadonly(),
      sensorTotal: sensorTotalSignal.asReadonly(),
      sensorPageIndex: sensorPageIndexSignal.asReadonly(),
      sensorLimit: sensorLimitSignal.asReadonly(),
      sensorLoading: sensorLoadingSignal.asReadonly(),
      sensorError: sensorErrorSignal.asReadonly(),

      expandedGateway: expandedGatewaySignal.asReadonly(),
      selectedChart: selectedChartSignal.asReadonly(),
      canSendCommands: canSendCommandsSignal.asReadonly(),

      loadDashboard: vi.fn(),
      toggleExpandedGateway: vi.fn(),
      changeGatewayPage: vi.fn(),
      changeSensorPage: vi.fn(),
      openChart: vi.fn(),
      closeChart: vi.fn(),
    };

    await TestBed.configureTestingModule({
      imports: [DashboardPage],
      providers: [
        { provide: DashboardService, useValue: dashboardServiceMock },
        { provide: MatDialog, useValue: dialogMock },
        { provide: MatSnackBar, useValue: snackBarMock },
      ],
      schemas: [NO_ERRORS_SCHEMA],
    }).compileComponents();

    fixture = TestBed.createComponent(DashboardPage);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create and call loadDashboard on init', () => {
    expect(component).toBeTruthy();
    expect(dashboardServiceMock.loadDashboard).toHaveBeenCalled();
  });

  it('should render all dashboard layout panels', () => {
    expect(fixture.debugElement.query(By.css('.dashboard-layout'))).toBeTruthy();
    expect(fixture.debugElement.query(By.css('.dashboard-left'))).toBeTruthy();
    expect(fixture.debugElement.query(By.css('.dashboard-top-right'))).toBeTruthy();
    expect(fixture.debugElement.query(By.css('.dashboard-bottom-right'))).toBeTruthy();
  });

  describe('error', () => {
    it('should not render error banner and error computed returns null by default', () => {
      expect(fixture.debugElement.query(By.css('.error-banner'))).toBeFalsy();
      expect(component['error']()).toBeNull();
    });

    it('should show gateway error banner with icon and update computed', () => {
      gatewayErrorSignal.set('Gateway failed');
      fixture.detectChanges();

      const errorBanner = fixture.debugElement.query(By.css('.error-banner'));
      expect(errorBanner).toBeTruthy();
      expect(errorBanner.nativeElement.textContent).toContain('Gateway failed');
      expect(errorBanner.query(By.css('mat-icon')).nativeElement.textContent).toContain('error');
      expect(component['error']()).toBe('Gateway failed');
    });

    it('should show sensor error banner when no gateway error', () => {
      sensorErrorSignal.set('Sensor failed');
      fixture.detectChanges();

      const errorBanner = fixture.debugElement.query(By.css('.error-banner'));
      expect(errorBanner.nativeElement.textContent).toContain('Sensor failed');
      expect(component['error']()).toBe('Sensor failed');
    });

    it('should prefer gateway error over sensor error', () => {
      gatewayErrorSignal.set('Gateway failed');
      sensorErrorSignal.set('Sensor failed');
      fixture.detectChanges();

      expect(
        fixture.debugElement.query(By.css('.error-banner')).nativeElement.textContent,
      ).toContain('Gateway failed');
    });
  });

  describe('canSendCommands is false (sensor table)', () => {
    it('should show sensor table and hide gateway table', () => {
      expect(fixture.debugElement.query(By.css('app-dashboard-sensor-table'))).toBeTruthy();
      expect(fixture.debugElement.query(By.css('app-dashboard-gateway-table'))).toBeFalsy();
    });

    it('should call changeSensorPage when sensor table emits pageChange', () => {
      fixture.debugElement
        .query(By.css('app-dashboard-sensor-table'))
        .triggerEventHandler('pageChange', { pageIndex: 3, pageSize: 10, length: 50 });

      expect(dashboardServiceMock.changeSensorPage).toHaveBeenCalledWith(3, 10);
    });

    it('should call openChart when sensor table emits chartRequested', () => {
      fixture.debugElement
        .query(By.css('app-dashboard-sensor-table'))
        .triggerEventHandler('chartRequested', mockChartRequest);

      expect(dashboardServiceMock.openChart).toHaveBeenCalledWith(mockChartRequest);
    });
  });

  describe('canSendCommands is true (gateway table)', () => {
    beforeEach(() => {
      canSendCommandsSignal.set(true);
      gatewayListSignal.set(mockGateways);
      sensorListSignal.set(mockSensors);
      fixture.detectChanges();
    });

    it('should show gateway table and hide sensor table', () => {
      expect(fixture.debugElement.query(By.css('app-dashboard-gateway-table'))).toBeTruthy();
      expect(fixture.debugElement.query(By.css('app-dashboard-sensor-table'))).toBeFalsy();
    });

    it('should call changeGatewayPage on gatewayPageChange', () => {
      fixture.debugElement
        .query(By.css('app-dashboard-gateway-table'))
        .triggerEventHandler('gatewayPageChange', { pageIndex: 2, pageSize: 25, length: 100 });

      expect(dashboardServiceMock.changeGatewayPage).toHaveBeenCalledWith(2, 25);
    });

    it('should call changeSensorPage on sensorPageChange', () => {
      fixture.debugElement
        .query(By.css('app-dashboard-gateway-table'))
        .triggerEventHandler('sensorPageChange', { pageIndex: 1, pageSize: 10, length: 50 });

      expect(dashboardServiceMock.changeSensorPage).toHaveBeenCalledWith(1, 10);
    });

    it('should call toggleExpandedGateway on expandedGatewayChange', () => {
      fixture.debugElement
        .query(By.css('app-dashboard-gateway-table'))
        .triggerEventHandler('expandedGatewayChange', mockGateways[0]);

      expect(dashboardServiceMock.toggleExpandedGateway).toHaveBeenCalledWith(mockGateways[0]);
    });

    it('should open snackbar on commandRequested', () => {
      fixture.debugElement
        .query(By.css('app-dashboard-gateway-table'))
        .triggerEventHandler('commandRequested', mockGateways[0]);

      expect(snackBarMock.open).toHaveBeenCalledWith('gw-1', 'Close', { duration: 2000 });
    });

    it('should call openChart on chartRequested', () => {
      fixture.debugElement
        .query(By.css('app-dashboard-gateway-table'))
        .triggerEventHandler('chartRequested', mockChartRequest);

      expect(dashboardServiceMock.openChart).toHaveBeenCalledWith(mockChartRequest);
    });
  });

  describe('selectedChart', () => {
    it('should not render chart container by default', () => {
      expect(fixture.debugElement.query(By.css('.dashboard-top-right')).children.length).toBe(0);
    });

    it('should reflect selectedChart signal when chart is selected', () => {
      selectedChartSignal.set(mockChartRequest);
      fixture.detectChanges();

      expect(component['selectedChart']()).toEqual(mockChartRequest);
    });

    it('should call closeChart when onChartClosed is invoked', () => {
      component['onChartClosed']();

      expect(dashboardServiceMock.closeChart).toHaveBeenCalled();
    });
  });
});
