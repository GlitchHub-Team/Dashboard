import { describe, it, expect, vi, beforeEach } from 'vitest';
import { signal, WritableSignal, Component, input, output } from '@angular/core';
import { TestBed, ComponentFixture } from '@angular/core/testing';
import { MatDialog } from '@angular/material/dialog';
import { MatSnackBar } from '@angular/material/snack-bar';
import { PageEvent } from '@angular/material/paginator';
import { By } from '@angular/platform-browser';
import { ActivatedRoute, Router } from '@angular/router';
import { of } from 'rxjs';
import { Observable } from 'rxjs';

import { DashboardPage } from './dashboard.page';
import { DashboardService } from '../../services/dashboard/dashboard.service';
import { GatewayTableComponent } from '../shared/components/gateway-table/gateway-table.component';
import { SensorTableComponent } from '../shared/components/sensor-table/sensor-table.component';
import { ChartContainerComponent } from './components/chart-container/chart-container.component';
import { Gateway } from '../../models/gateway/gateway.model';
import { Sensor } from '../../models/sensor/sensor.model';
import { Status } from '../../models/gateway-sensor-status.enum';
import { ChartRequest } from '../../models/chart/chart-request.model';
import { ChartType } from '../../models/chart/chart-type.enum';
import { SensorProfiles } from '../../models/sensor/sensor-profiles.enum';
import { UserSessionService } from '../../services/user-session/user-session.service';
import { UserRole } from '../../models/user/user-role.enum';
import { UserSession } from '../../models/auth/user-session.model';

@Component({ selector: 'app-gateway-table', template: '', standalone: true })
class StubGatewayTable {
  actionMode = input<string>();
  gateways = input<Gateway[]>();
  sensors = input<Sensor[]>();
  expandedGateway = input<Gateway | null>();
  gatewayLoading = input<boolean>();
  sensorLoading = input<boolean>();
  gatewayTotal = input<number>();
  gatewayPageIndex = input<number>();
  gatewayLimit = input<number>();
  sensorTotal = input<number>();
  sensorPageIndex = input<number>();
  sensorLimit = input<number>();
  commandRequested = output<boolean>();
  chartRequested = output<ChartRequest>();
  expandedGatewayChange = output<Gateway>();
  gatewayPageChange = output<PageEvent>();
  sensorPageChange = output<PageEvent>();
}

@Component({ selector: 'app-sensor-table', template: '', standalone: true })
class StubSensorTable {
  sensors = input<Sensor[]>();
  loading = input<boolean>();
  total = input<number>();
  pageIndex = input<number>();
  limit = input<number>();
  chartRequested = output<ChartRequest>();
  pageChange = output<PageEvent>();
}

@Component({ selector: 'app-chart-container', template: '', standalone: true })
class StubChartContainer {
  chartRequest = input<ChartRequest>();
  chartClosed = output<void>();
}

describe('DashboardPage (Unit)', () => {
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
    status: Status.ACTIVE,
    interval: 60,
  };
  const mockSensor: Sensor = {
    id: 'sensor-1',
    gatewayId: 'gw-1',
    name: 'Heart Rate Sensor',
    profile: SensorProfiles.HEART_RATE_SERVICE,
    status: Status.ACTIVE,
    dataInterval: 1000,
  };
  const mockChartRequest: ChartRequest = {
    sensor: mockSensor,
    chartType: ChartType.HISTORIC,
    timeInterval: { from: new Date('2025-01-01'), to: new Date('2025-01-02') },
  };

  let dashboardServiceMock: any;
  let snackBarMock: { open: ReturnType<typeof vi.fn> };
  let routerMock: { navigate: ReturnType<typeof vi.fn> };
  let activatedRouteMock: { queryParams: Observable<Record<string, unknown>> };
  let userSessionMock: { currentUser: () => UserSession | null };

  const getGatewayTable = () =>
    fixture.debugElement.query(By.directive(StubGatewayTable))
      ?.componentInstance as StubGatewayTable;
  const getSensorTable = () =>
    fixture.debugElement.query(By.directive(StubSensorTable))?.componentInstance as StubSensorTable;
  const getChartContainer = () =>
    fixture.debugElement.query(By.directive(StubChartContainer))
      ?.componentInstance as StubChartContainer;

  beforeEach(async () => {
    vi.resetAllMocks();

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
    routerMock = { navigate: vi.fn() };
    activatedRouteMock = { queryParams: of({}) };
    userSessionMock = {
      currentUser: () => ({ userId: '1', role: UserRole.TENANT_ADMIN, tenantId: 'tenant-01' }),
    };

    await TestBed.configureTestingModule({
      imports: [DashboardPage],
      providers: [
        { provide: DashboardService, useValue: dashboardServiceMock },
        { provide: MatSnackBar, useValue: snackBarMock },
        { provide: MatDialog, useValue: {} },
        { provide: Router, useValue: routerMock },
        { provide: ActivatedRoute, useValue: activatedRouteMock },
        { provide: UserSessionService, useValue: userSessionMock },
      ],
    })
      .overrideComponent(DashboardPage, {
        remove: {
          imports: [
            GatewayTableComponent,
            SensorTableComponent,
            ChartContainerComponent,
          ],
        },
        add: {
          imports: [StubGatewayTable, StubSensorTable, StubChartContainer],
        },
      })
      .compileComponents();

    fixture = TestBed.createComponent(DashboardPage);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  describe('lifecycle', () => {
    it('should call loadDashboard with session tenantId on init', () => {
      expect(dashboardServiceMock.loadDashboard).toHaveBeenCalledOnce();
      expect(dashboardServiceMock.loadDashboard).toHaveBeenCalledWith('tenant-01');
    });

    it('should call closeChart on destroy', () => {
      fixture.destroy();
      expect(dashboardServiceMock.closeChart).toHaveBeenCalledOnce();
    });
  });

  describe('error banner', () => {
    const getBanner = () => fixture.debugElement.query(By.css('.error-banner'));

    it('should display gateway error', () => {
      gatewayErrorSig.set('Gateway connection failed');
      fixture.detectChanges();
      expect(getBanner()).toBeTruthy();
      expect(getBanner().nativeElement.textContent).toContain('Gateway connection failed');
    });

    it('should display sensor error', () => {
      sensorErrorSig.set('Sensor timeout');
      fixture.detectChanges();
      expect(getBanner().nativeElement.textContent).toContain('Sensor timeout');
    });

    it('should prefer gatewayError over sensorError', () => {
      gatewayErrorSig.set('Gateway error');
      sensorErrorSig.set('Sensor error');
      fixture.detectChanges();
      expect(getBanner().nativeElement.textContent).toContain('Gateway error');
    });

    it('should show sensorError when gatewayError is null', () => {
      gatewayErrorSig.set(null);
      sensorErrorSig.set('Sensor error');
      fixture.detectChanges();
      expect(getBanner().nativeElement.textContent).toContain('Sensor error');
    });

    it('should not display banner when there are no errors', () => {
      expect(getBanner()).toBeNull();
    });
  });

  describe('when canSendCommands is true', () => {
    it('should render gateway table and not sensor table', () => {
      expect(fixture.debugElement.query(By.directive(StubGatewayTable))).toBeTruthy();
      expect(fixture.debugElement.query(By.directive(StubSensorTable))).toBeNull();
    });

    it('should pass correct inputs to gateway table', () => {
      gatewayListSig.set([mockGateway]);
      sensorListSig.set([mockSensor]);
      gatewayTotalSig.set(25);
      gatewayPageIndexSig.set(2);
      gatewayLimitSig.set(5);
      sensorTotalSig.set(10);
      sensorPageIndexSig.set(1);
      sensorLimitSig.set(15);
      gatewayLoadingSig.set(true);
      sensorLoadingSig.set(true);
      expandedGatewaySig.set(mockGateway);
      fixture.detectChanges();

      const table = getGatewayTable();
      expect(table.actionMode()).toBe('dashboard');
      expect(table.gateways()).toEqual([mockGateway]);
      expect(table.sensors()).toEqual([mockSensor]);
      expect(table.expandedGateway()).toEqual(mockGateway);
      expect(table.gatewayLoading()).toBe(true);
      expect(table.sensorLoading()).toBe(true);
      expect(table.gatewayTotal()).toBe(25);
      expect(table.gatewayPageIndex()).toBe(2);
      expect(table.gatewayLimit()).toBe(5);
      expect(table.sensorTotal()).toBe(10);
      expect(table.sensorPageIndex()).toBe(1);
      expect(table.sensorLimit()).toBe(15);
    });
  });

  describe('when canSendCommands is false', () => {
    beforeEach(() => {
      canSendCommandsSig.set(false);
      fixture.detectChanges();
    });

    it('should render sensor table and not gateway table', () => {
      expect(fixture.debugElement.query(By.directive(StubSensorTable))).toBeTruthy();
      expect(fixture.debugElement.query(By.directive(StubGatewayTable))).toBeNull();
    });

    it('should pass correct inputs to sensor table', () => {
      sensorListSig.set([mockSensor]);
      sensorTotalSig.set(50);
      sensorPageIndexSig.set(3);
      sensorLimitSig.set(20);
      sensorLoadingSig.set(true);
      fixture.detectChanges();

      const table = getSensorTable();
      expect(table.sensors()).toEqual([mockSensor]);
      expect(table.loading()).toBe(true);
      expect(table.total()).toBe(50);
      expect(table.pageIndex()).toBe(3);
      expect(table.limit()).toBe(20);
    });
  });

  describe('chart container', () => {
    it('should show chart container with chartRequest when set, hide when cleared', () => {
      expect(fixture.debugElement.query(By.directive(StubChartContainer))).toBeNull();

      selectedChartSig.set(mockChartRequest);
      fixture.detectChanges();
      expect(fixture.debugElement.query(By.directive(StubChartContainer))).toBeTruthy();
      expect(getChartContainer().chartRequest()).toEqual(mockChartRequest);

      selectedChartSig.set(null);
      fixture.detectChanges();
      expect(fixture.debugElement.query(By.directive(StubChartContainer))).toBeNull();
    });
  });

  describe('output events', () => {
    it('should call toggleExpandedGateway on expandedGatewayChange', () => {
      fixture.debugElement
        .query(By.directive(StubGatewayTable))
        .triggerEventHandler('expandedGatewayChange', mockGateway);
      expect(dashboardServiceMock.toggleExpandedGateway).toHaveBeenCalledWith(mockGateway);
    });

    it('should call changeGatewayPage on gatewayPageChange', () => {
      fixture.debugElement
        .query(By.directive(StubGatewayTable))
        .triggerEventHandler('gatewayPageChange', { pageIndex: 3, pageSize: 25, length: 100 });
      expect(dashboardServiceMock.changeGatewayPage).toHaveBeenCalledWith(3, 25);
    });

    it('should call changeSensorPage on gateway table sensorPageChange', () => {
      fixture.debugElement
        .query(By.directive(StubGatewayTable))
        .triggerEventHandler('sensorPageChange', { pageIndex: 1, pageSize: 10, length: 50 });
      expect(dashboardServiceMock.changeSensorPage).toHaveBeenCalledWith(1, 10);
    });

    it.each([
      [true, true] as const,
      [false, false] as const,
    ])('commandRequested(%s): snackbar opened = %s', (value, shouldCall) => {
      fixture.debugElement
        .query(By.directive(StubGatewayTable))
        .triggerEventHandler('commandRequested', value);
      if (shouldCall) {
        expect(snackBarMock.open).toHaveBeenCalledWith('Comando inviato correttamente', 'Close', {
          duration: 3000,
        });
      } else {
        expect(snackBarMock.open).not.toHaveBeenCalled();
      }
    });

    it('should call openChart on gateway table chartRequested', () => {
      fixture.debugElement
        .query(By.directive(StubGatewayTable))
        .triggerEventHandler('chartRequested', mockChartRequest);
      expect(dashboardServiceMock.openChart).toHaveBeenCalledWith(mockChartRequest);
    });

    it('should call openChart on sensor table chartRequested', () => {
      canSendCommandsSig.set(false);
      fixture.detectChanges();
      fixture.debugElement
        .query(By.directive(StubSensorTable))
        .triggerEventHandler('chartRequested', mockChartRequest);
      expect(dashboardServiceMock.openChart).toHaveBeenCalledWith(mockChartRequest);
    });

    it('should call changeSensorPage on sensor table pageChange', () => {
      canSendCommandsSig.set(false);
      fixture.detectChanges();
      fixture.debugElement
        .query(By.directive(StubSensorTable))
        .triggerEventHandler('pageChange', { pageIndex: 2, pageSize: 15, length: 30 });
      expect(dashboardServiceMock.changeSensorPage).toHaveBeenCalledWith(2, 15);
    });

    it('should call closeChart on chartClosed', () => {
      selectedChartSig.set(mockChartRequest);
      fixture.detectChanges();
      dashboardServiceMock.closeChart.mockClear();
      fixture.debugElement
        .query(By.directive(StubChartContainer))
        .triggerEventHandler('chartClosed');
      expect(dashboardServiceMock.closeChart).toHaveBeenCalledOnce();
    });
  });

  describe('signal reactivity', () => {
    it('should toggle between gateway and sensor table when canSendCommands changes', () => {
      expect(fixture.debugElement.query(By.directive(StubGatewayTable))).toBeTruthy();

      canSendCommandsSig.set(false);
      fixture.detectChanges();
      expect(fixture.debugElement.query(By.directive(StubSensorTable))).toBeTruthy();
      expect(fixture.debugElement.query(By.directive(StubGatewayTable))).toBeNull();
    });

    it('should show and hide error banner reactively', () => {
      expect(fixture.debugElement.query(By.css('.error-banner'))).toBeNull();

      gatewayErrorSig.set('New error');
      fixture.detectChanges();
      expect(fixture.debugElement.query(By.css('.error-banner'))).toBeTruthy();

      gatewayErrorSig.set(null);
      fixture.detectChanges();
      expect(fixture.debugElement.query(By.css('.error-banner'))).toBeNull();
    });
  });

  describe('SUPER_ADMIN ngOnInit', () => {
    const setupAsSuperAdmin = async (queryParams: Record<string, unknown>) => {
      vi.resetAllMocks();
      activatedRouteMock = { queryParams: of(queryParams) };
      userSessionMock = { currentUser: () => ({ userId: 'admin', role: UserRole.SUPER_ADMIN }) };

      TestBed.resetTestingModule();
      await TestBed.configureTestingModule({
        imports: [DashboardPage],
        providers: [
          { provide: DashboardService, useValue: dashboardServiceMock },
          { provide: MatSnackBar, useValue: snackBarMock },
          { provide: MatDialog, useValue: {} },
          { provide: Router, useValue: routerMock },
          { provide: ActivatedRoute, useValue: activatedRouteMock },
          { provide: UserSessionService, useValue: userSessionMock },
        ],
      })
        .overrideComponent(DashboardPage, {
          remove: {
            imports: [
              GatewayTableComponent,
              SensorTableComponent,
              ChartContainerComponent,
            ],
          },
          add: { imports: [StubGatewayTable, StubSensorTable, StubChartContainer] },
        })
        .compileComponents();

      fixture = TestBed.createComponent(DashboardPage);
      component = fixture.componentInstance;
      fixture.detectChanges();
    };

    it('should call loadDashboard with tenantId from queryParams and set activeTenantId', async () => {
      await setupAsSuperAdmin({ tenantId: 'super-tenant' });
      expect(dashboardServiceMock.loadDashboard).toHaveBeenCalledWith('super-tenant');
      expect(component['activeTenantId']()).toBe('super-tenant');
    });

    it('should call loadDashboard with undefined and set activeTenantId to null when no tenantId', async () => {
      await setupAsSuperAdmin({});
      expect(dashboardServiceMock.loadDashboard).toHaveBeenCalledWith(undefined);
      expect(component['activeTenantId']()).toBeNull();
    });
  });
});
