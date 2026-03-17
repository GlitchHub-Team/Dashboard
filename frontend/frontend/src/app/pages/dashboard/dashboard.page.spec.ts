import { ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA, signal } from '@angular/core';
import { MatDialog } from '@angular/material/dialog';
import { MatSnackBar } from '@angular/material/snack-bar';

import { DashboardPage } from './dashboard.page';
import { DashboardService } from '../../services/dashboard/dashboard.service';
import { DashboardGatewayTableComponent } from './components/dashboard-gateway-table/dashboard-gateway-table.component';
import { DashboardSensorTableComponent } from './components/dashboard-sensor-table/dashboard-sensor-table.component';
import { Gateway } from '../../models/gateway.model';
import { GatewayStatus } from '../../models/gateway-status.enum';
import { ChartRequest } from '../../models/chart-request.model';
import { ChartType } from '../../models/chart-type.enum';
import { SensorProfiles } from '../../models/sensor-profiles.enum';

describe('DashboardPage', () => {
  let component: DashboardPage;
  let fixture: ComponentFixture<DashboardPage>;

  const mockGateways: Gateway[] = [
    { id: 'gw-1', name: 'Gateway Alpha', status: GatewayStatus.ONLINE },
  ];

  const dashboardServiceMock = {
    gatewayList: signal(mockGateways).asReadonly(),
    sensorList: signal([]).asReadonly(),
    gatewayLoading: signal(false).asReadonly(),
    sensorLoading: signal(false).asReadonly(),
    expandedGateway: signal<Gateway | null>(null).asReadonly(),
    selectedChart: signal<ChartRequest | null>(null).asReadonly(),
    canSendCommands: signal(false).asReadonly(),
    gatewayError: signal<string | null>(null).asReadonly(),
    sensorError: signal<string | null>(null).asReadonly(),
    loadDashboard: vi.fn(),
    toggleExpandedGateway: vi.fn(),
    openChart: vi.fn(),
    closeChart: vi.fn(),
  };

  const dialogMock = {
    open: vi.fn(),
  };

  const snackBarMock = {
    open: vi.fn(),
  };

  beforeEach(async () => {
    vi.resetAllMocks();

    await TestBed.configureTestingModule({
      imports: [DashboardPage],
      providers: [
        { provide: DashboardService, useValue: dashboardServiceMock },
        { provide: MatDialog, useValue: dialogMock },
        { provide: MatSnackBar, useValue: snackBarMock },
      ],
    })
      .overrideComponent(DashboardPage, {
        remove: { imports: [DashboardGatewayTableComponent, DashboardSensorTableComponent] },
        add: { schemas: [CUSTOM_ELEMENTS_SCHEMA] },
      })
      .compileComponents();

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
  });

  describe('error', () => {
    it('should return gateway error when present', () => {
      dashboardServiceMock.gatewayError = signal('Gateway failed').asReadonly();
      dashboardServiceMock.sensorError = signal(null).asReadonly();

      // Rebuild component to pick up new signals
      fixture = TestBed.createComponent(DashboardPage);
      component = fixture.componentInstance;
      fixture.detectChanges();

      expect(component['error']()).toBe('Gateway failed');
    });

    it('should return sensor error when gateway error is null', () => {
      dashboardServiceMock.gatewayError = signal(null).asReadonly();
      dashboardServiceMock.sensorError = signal('Sensor failed').asReadonly();

      // Rebuild component to pick up new signals
      fixture = TestBed.createComponent(DashboardPage);
      component = fixture.componentInstance;
      fixture.detectChanges();

      expect(component['error']()).toBe('Sensor failed');
    });

    it('should return null when no errors', () => {
      dashboardServiceMock.gatewayError = signal(null).asReadonly();
      dashboardServiceMock.sensorError = signal(null).asReadonly();

      // Rebuild component to pick up new signals
      fixture = TestBed.createComponent(DashboardPage);
      component = fixture.componentInstance;
      fixture.detectChanges();

      expect(component['error']()).toBeNull();
    });
  });

  describe('onExpandedGatewayChange', () => {
    it('should call toggleExpandedGateway with gateway', () => {
      component['onExpandedGatewayChange'](mockGateways[0]);

      expect(dashboardServiceMock.toggleExpandedGateway).toHaveBeenCalledWith(mockGateways[0]);
    });
  });

  describe('onCommandRequested', () => {
    it('should open snackbar with gateway id', () => {
      component['onCommandRequested'](mockGateways[0]);

      expect(snackBarMock.open).toHaveBeenCalledWith('gw-1', 'Close', { duration: 2000 });
    });
  });

  describe('onChartOpen', () => {
    it('should call openChart with request', () => {
      const mockRequest: ChartRequest = {
        sensor: {
          id: 's-1',
          gatewayId: 'gw-1',
          name: 'Temperature',
          profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
        },
        chartType: ChartType.HISTORIC,
        timeInterval: null!,
      };

      component['onChartOpen'](mockRequest);

      expect(dashboardServiceMock.openChart).toHaveBeenCalledWith(mockRequest);
    });
  });

  describe('onChartClosed', () => {
    it('should call closeChart', () => {
      component['onChartClosed']();

      expect(dashboardServiceMock.closeChart).toHaveBeenCalled();
    });
  });
});
