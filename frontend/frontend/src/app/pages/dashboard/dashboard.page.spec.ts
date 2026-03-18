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
  let sensorListSignal: WritableSignal<Sensor[]>;
  let gatewayLoadingSignal: WritableSignal<boolean>;
  let sensorLoadingSignal: WritableSignal<boolean>;
  let expandedGatewaySignal: WritableSignal<Gateway | null>;
  let selectedChartSignal: WritableSignal<ChartRequest | null>;
  let canSendCommandsSignal: WritableSignal<boolean>;
  let gatewayErrorSignal: WritableSignal<string | null>;
  let sensorErrorSignal: WritableSignal<string | null>;

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

  // ESLint whining
  let dashboardServiceMock: any;
  const dialogMock = { open: vi.fn() };
  const snackBarMock = { open: vi.fn() };

  beforeEach(async () => {
    vi.resetAllMocks();

    gatewayListSignal = signal<Gateway[]>([]);
    sensorListSignal = signal<Sensor[]>([]);
    gatewayLoadingSignal = signal(false);
    sensorLoadingSignal = signal(false);
    expandedGatewaySignal = signal<Gateway | null>(null);
    selectedChartSignal = signal<ChartRequest | null>(null);
    canSendCommandsSignal = signal(false);
    gatewayErrorSignal = signal<string | null>(null);
    sensorErrorSignal = signal<string | null>(null);

    dashboardServiceMock = {
      gatewayList: gatewayListSignal.asReadonly(),
      sensorList: sensorListSignal.asReadonly(),
      gatewayLoading: gatewayLoadingSignal.asReadonly(),
      sensorLoading: sensorLoadingSignal.asReadonly(),
      expandedGateway: expandedGatewaySignal.asReadonly(),
      selectedChart: selectedChartSignal.asReadonly(),
      canSendCommands: canSendCommandsSignal.asReadonly(),
      gatewayError: gatewayErrorSignal.asReadonly(),
      sensorError: sensorErrorSignal.asReadonly(),
      loadDashboard: vi.fn(),
      toggleExpandedGateway: vi.fn(),
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

  describe('initial state', () => {
    it('should create', () => {
      expect(component).toBeTruthy();
    });

    it('should call loadDashboard on init', () => {
      expect(dashboardServiceMock.loadDashboard).toHaveBeenCalled();
    });

    it('should render the dashboard layout', () => {
      const layout = fixture.debugElement.query(By.css('.dashboard-layout'));
      expect(layout).toBeTruthy();
    });

    it('should render the left panel', () => {
      const left = fixture.debugElement.query(By.css('.dashboard-left'));
      expect(left).toBeTruthy();
    });

    it('should render the top right panel', () => {
      const topRight = fixture.debugElement.query(By.css('.dashboard-top-right'));
      expect(topRight).toBeTruthy();
    });

    it('should render the bottom right panel', () => {
      const bottomRight = fixture.debugElement.query(By.css('.dashboard-bottom-right'));
      expect(bottomRight).toBeTruthy();
    });
  });

  describe('error state', () => {
    it('should not render error banner by default', () => {
      const errorBanner = fixture.debugElement.query(By.css('.error-banner'));
      expect(errorBanner).toBeFalsy();
    });

    it('should render error banner when gateway error exists', () => {
      gatewayErrorSignal.set('Gateway failed');
      fixture.detectChanges();

      const errorBanner = fixture.debugElement.query(By.css('.error-banner'));
      expect(errorBanner).toBeTruthy();
      expect(errorBanner.nativeElement.textContent).toContain('Gateway failed');
    });

    it('should render error banner when sensor error exists', () => {
      sensorErrorSignal.set('Sensor failed');
      fixture.detectChanges();

      const errorBanner = fixture.debugElement.query(By.css('.error-banner'));
      expect(errorBanner).toBeTruthy();
      expect(errorBanner.nativeElement.textContent).toContain('Sensor failed');
    });

    it('should prefer gateway error over sensor error', () => {
      gatewayErrorSignal.set('Gateway failed');
      sensorErrorSignal.set('Sensor failed');
      fixture.detectChanges();

      const errorBanner = fixture.debugElement.query(By.css('.error-banner'));
      expect(errorBanner.nativeElement.textContent).toContain('Gateway failed');
    });

    it('should render error icon', () => {
      gatewayErrorSignal.set('Some error');
      fixture.detectChanges();

      const icon = fixture.debugElement.query(By.css('.error-banner mat-icon'));
      expect(icon.nativeElement.textContent).toContain('error');
    });
  });

  describe('error computed', () => {
    it('should return null when no errors', () => {
      expect(component['error']()).toBeNull();
    });

    it('should return gateway error when present', () => {
      gatewayErrorSignal.set('Gateway failed');
      expect(component['error']()).toBe('Gateway failed');
    });

    it('should return sensor error when no gateway error', () => {
      sensorErrorSignal.set('Sensor failed');
      expect(component['error']()).toBe('Sensor failed');
    });
  });

  describe('canSendCommands is false (sensor table)', () => {
    it('should render sensor table by default', () => {
      const sensorTable = fixture.debugElement.query(By.css('app-dashboard-sensor-table'));
      expect(sensorTable).toBeTruthy();
    });

    it('should not render gateway table', () => {
      const gatewayTable = fixture.debugElement.query(By.css('app-dashboard-gateway-table'));
      expect(gatewayTable).toBeFalsy();
    });

    it('should emit onChartOpen when sensor table emits chartRequested', () => {
      const sensorTable = fixture.debugElement.query(By.css('app-dashboard-sensor-table'));

      const request: ChartRequest = {
        sensor: mockSensors[0],
        chartType: ChartType.HISTORIC,
        timeInterval: null!,
      };

      sensorTable.triggerEventHandler('chartRequested', request);
      fixture.detectChanges();

      expect(dashboardServiceMock.openChart).toHaveBeenCalledWith(request);
    });
  });

  describe('canSendCommands is true (gateway table)', () => {
    beforeEach(() => {
      canSendCommandsSignal.set(true);
      gatewayListSignal.set(mockGateways);
      sensorListSignal.set(mockSensors);
      fixture.detectChanges();
    });

    it('should render gateway table', () => {
      const gatewayTable = fixture.debugElement.query(By.css('app-dashboard-gateway-table'));
      expect(gatewayTable).toBeTruthy();
    });

    it('should not render sensor table', () => {
      const sensorTable = fixture.debugElement.query(By.css('app-dashboard-sensor-table'));
      expect(sensorTable).toBeFalsy();
    });

    it('should call toggleExpandedGateway when expandedGatewayChange is emitted', () => {
      const gatewayTable = fixture.debugElement.query(By.css('app-dashboard-gateway-table'));

      gatewayTable.triggerEventHandler('expandedGatewayChange', mockGateways[0]);
      fixture.detectChanges();

      expect(dashboardServiceMock.toggleExpandedGateway).toHaveBeenCalledWith(mockGateways[0]);
    });

    it('should open snackbar when commandRequested is emitted', () => {
      const gatewayTable = fixture.debugElement.query(By.css('app-dashboard-gateway-table'));

      gatewayTable.triggerEventHandler('commandRequested', mockGateways[0]);
      fixture.detectChanges();

      expect(snackBarMock.open).toHaveBeenCalledWith('gw-1', 'Close', { duration: 2000 });
    });

    it('should call openChart when gateway table emits chartRequested', () => {
      const gatewayTable = fixture.debugElement.query(By.css('app-dashboard-gateway-table'));

      const request: ChartRequest = {
        sensor: mockSensors[0],
        chartType: ChartType.HISTORIC,
        timeInterval: null!,
      };

      gatewayTable.triggerEventHandler('chartRequested', request);
      fixture.detectChanges();

      expect(dashboardServiceMock.openChart).toHaveBeenCalledWith(request);
    });
  });

  describe('selectedChart', () => {
    it('should not render chart container by default', () => {
      const topRight = fixture.debugElement.query(By.css('.dashboard-top-right'));
      expect(topRight.children.length).toBe(0);
    });

    it('should enter selectedChart block when chart is selected', () => {
      const request: ChartRequest = {
        sensor: mockSensors[0],
        chartType: ChartType.HISTORIC,
        timeInterval: null!,
      };
      selectedChartSignal.set(request);
      fixture.detectChanges();

      expect(component['selectedChart']()).toEqual(request);
    });
  });
});
