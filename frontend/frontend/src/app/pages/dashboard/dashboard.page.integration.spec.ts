import { ComponentFixture, TestBed } from '@angular/core/testing';
import { WritableSignal, signal } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { MatDialog } from '@angular/material/dialog';
import { MatSnackBar } from '@angular/material/snack-bar';
import { By } from '@angular/platform-browser';
import { of, BehaviorSubject } from 'rxjs';
import { describe, expect, it, vi, afterEach } from 'vitest';

import { DashboardPage } from './dashboard.page';
import { GatewayTableComponent } from '../shared/components/gateway-table/gateway-table.component';
import { SensorTableComponent } from '../shared/components/sensor-table/sensor-table.component';
import { ChartContainerComponent } from './components/chart-container/chart-container.component';
import { DashboardService } from '../../services/dashboard/dashboard.service';
import { UserSessionService } from '../../services/user-session/user-session.service';
import { SensorChartService } from '../../services/sensor-chart/sensor-chart.service';
import { UserRole } from '../../models/user/user-role.enum';
import { Gateway } from '../../models/gateway/gateway.model';
import { Sensor } from '../../models/sensor/sensor.model';
import { ChartRequest } from '../../models/chart/chart-request.model';
import { UserSession } from '../../models/auth/user-session.model';
import { Status } from '../../models/gateway-sensor-status.enum';
import { SensorProfiles } from '../../models/sensor/sensor-profiles.enum';
import { ChartType } from '../../models/chart/chart-type.enum';

const mockGateways: Gateway[] = [
  {
    id: 'gw-1',
    tenantId: 'tenant-1',
    name: 'Gateway Alpha',
    status: Status.ACTIVE,
    interval: 60,
  },
  {
    id: 'gw-2',
    tenantId: 'tenant-1',
    name: 'Gateway Beta',
    status: Status.INACTIVE,
    interval: 120,
  },
];

const mockSensors: Sensor[] = [
  {
    id: 'sensor-1',
    gatewayId: 'gw-1',
    name: 'Temperature',
    profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
    status: Status.ACTIVE,
    dataInterval: 30,
  },
  {
    id: 'sensor-2',
    gatewayId: 'gw-1',
    name: 'Humidity',
    profile: SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE,
    status: Status.ACTIVE,
    dataInterval: 30,
  },
];

const mockChartRequest: ChartRequest = { sensor: mockSensors[0], chartType: ChartType.HISTORIC };

const tenantAdminSession: UserSession = {
  userId: '1',
  role: UserRole.TENANT_ADMIN,
  tenantId: 'tenant-1',
};
const superAdminSession: UserSession = { userId: '1', role: UserRole.SUPER_ADMIN };

function createDashboardServiceMock() {
  return {
    gatewayList: signal<Gateway[]>([]),
    gatewayTotal: signal(0),
    gatewayPageIndex: signal(0),
    gatewayLimit: signal(10),
    gatewayLoading: signal(false),
    gatewayError: signal<string | null>(null),

    sensorList: signal<Sensor[]>([]),
    sensorTotal: signal(0),
    sensorPageIndex: signal(0),
    sensorLimit: signal(10),
    sensorLoading: signal(false),
    sensorError: signal<string | null>(null),

    expandedGateway: signal<Gateway | null>(null),
    selectedChart: signal<ChartRequest | null>(null),
    canSendCommands: signal(true),

    loadDashboard: vi.fn(),
    toggleExpandedGateway: vi.fn(),
    changeGatewayPage: vi.fn(),
    changeSensorPage: vi.fn(),
    openChart: vi.fn(),
    closeChart: vi.fn(),
  };
}

function createChartServiceMock() {
  return {
    historicReadings: signal<any[]>([]),
    liveReadings: signal<any[]>([]),
    loading: signal(false),
    connectionStatus: signal('disconnected'),
    error: signal<string | null>(null),
    startChart: vi.fn(),
    stopChart: vi.fn(),
  };
}

function setupTestBed(options: { session: UserSession; queryParams?: Record<string, string> }) {
  const queryParamsSubject = new BehaviorSubject<Record<string, string>>(options.queryParams ?? {});
  const dashboardServiceMock = createDashboardServiceMock();
  const chartServiceMock = createChartServiceMock();
  const dialogMock = { open: vi.fn().mockReturnValue({ afterClosed: () => of(undefined) }) };
  const snackBarMock = { open: vi.fn() };
  const routerMock = { navigate: vi.fn() };

  TestBed.configureTestingModule({
    imports: [
      DashboardPage,
      GatewayTableComponent,
      SensorTableComponent,
      ChartContainerComponent,
    ],
    providers: [
      { provide: DashboardService, useValue: dashboardServiceMock },
      { provide: UserSessionService, useValue: { currentUser: signal(options.session) } },
      { provide: SensorChartService, useValue: chartServiceMock },
      { provide: Router, useValue: routerMock },
      { provide: ActivatedRoute, useValue: { queryParams: queryParamsSubject.asObservable() } },
    ],
  })
    .overrideProvider(MatDialog, { useValue: dialogMock })
    .overrideProvider(MatSnackBar, { useValue: snackBarMock });

  const fixture = TestBed.createComponent(DashboardPage);
  return {
    fixture,
    dashboardServiceMock,
    chartServiceMock,
    dialogMock,
    snackBarMock,
    routerMock,
    queryParamsSubject,
  };
}

function getGatewayTable(fixture: ComponentFixture<DashboardPage>) {
  return fixture.debugElement.query(By.directive(GatewayTableComponent));
}

function getSensorTable(fixture: ComponentFixture<DashboardPage>) {
  return fixture.debugElement.query(By.directive(SensorTableComponent));
}

function getChartContainer(fixture: ComponentFixture<DashboardPage>) {
  return fixture.debugElement.query(By.directive(ChartContainerComponent));
}

function getGatewayRows(fixture: ComponentFixture<DashboardPage>): HTMLElement[] {
  return Array.from(
    fixture.nativeElement.querySelectorAll('.table-container mat-row:not(.detail-row)'),
  );
}

describe('DashboardPage (Integration)', () => {
  afterEach(() => {
    TestBed.resetTestingModule();
    vi.resetAllMocks();
  });

  describe('Initialization', () => {
    it('should render page title and call loadDashboard for TENANT_ADMIN', () => {
      const { fixture, dashboardServiceMock } = setupTestBed({ session: tenantAdminSession });
      fixture.detectChanges();

      expect(fixture.nativeElement.querySelector('h1').textContent).toContain('Dashboard');
      expect(dashboardServiceMock.loadDashboard).toHaveBeenCalledWith('tenant-1');
    });

    it.each([
      ['with tenantId from queryParams', { tenantId: 'tenant-from-url' }, 'tenant-from-url'],
      ['without tenantId', {}, undefined],
    ] as const)('should call loadDashboard for SUPER_ADMIN %s', (_label, queryParams, expected) => {
      const { fixture, dashboardServiceMock } = setupTestBed({
        session: superAdminSession,
        queryParams,
      });
      fixture.detectChanges();

      expect(dashboardServiceMock.loadDashboard).toHaveBeenCalledWith(expected);
    });

    it('should call closeChart on destroy', () => {
      const { fixture, dashboardServiceMock } = setupTestBed({ session: tenantAdminSession });
      fixture.detectChanges();
      fixture.destroy();

      expect(dashboardServiceMock.closeChart).toHaveBeenCalled();
    });
  });

  describe('Banners', () => {
    it('should show tenant banner with content and 2 action buttons for SUPER_ADMIN with tenantId', () => {
      const { fixture } = setupTestBed({
        session: superAdminSession,
        queryParams: { tenantId: 'tenant-xyz' },
      });
      fixture.detectChanges();

      const banner = fixture.nativeElement.querySelector('.tenant-banner');
      expect(banner).toBeTruthy();
      expect(banner.textContent).toContain('tenant-xyz');
      expect(banner.querySelectorAll('.banner-actions button').length).toBe(2);
    });

    it('should show warning banner for SUPER_ADMIN without tenantId', () => {
      const { fixture } = setupTestBed({ session: superAdminSession, queryParams: {} });
      fixture.detectChanges();

      const warning = fixture.nativeElement.querySelector('.warning-banner');
      expect(warning).toBeTruthy();
      expect(warning.textContent).toContain('Per visualizzare i dati, devi selezionare un tenant');
    });

    it('should NOT show any banner for TENANT_ADMIN', () => {
      const { fixture } = setupTestBed({ session: tenantAdminSession });
      fixture.detectChanges();

      expect(fixture.nativeElement.querySelector('.tenant-banner')).toBeFalsy();
      expect(fixture.nativeElement.querySelector('.warning-banner')).toBeFalsy();
    });

    it.each([
      ['gatewayError' as const, 'Gateway load failed'],
      ['sensorError' as const, 'Sensor load failed'],
    ])('should show error banner when %s has value', (errorKey, message) => {
      const { fixture, dashboardServiceMock } = setupTestBed({ session: tenantAdminSession });
      (dashboardServiceMock[errorKey] as WritableSignal<string | null>).set(message);
      fixture.detectChanges();

      const errorBanner = fixture.nativeElement.querySelector('.error-banner');
      expect(errorBanner).toBeTruthy();
      expect(errorBanner.textContent).toContain(message);
    });
  });

  describe('Conditional Table: canSendCommands', () => {
    it.each([
      ['gateway table when canSendCommands is true', true],
      ['sensor table when canSendCommands is false', false],
    ] as const)('should show only %s', (_label, canSendCommands) => {
      const { fixture, dashboardServiceMock } = setupTestBed({ session: tenantAdminSession });
      (dashboardServiceMock.canSendCommands as WritableSignal<boolean>).set(canSendCommands);
      (dashboardServiceMock.gatewayList as WritableSignal<Gateway[]>).set(mockGateways);
      (dashboardServiceMock.sensorList as WritableSignal<Sensor[]>).set(mockSensors);
      fixture.detectChanges();

      if (canSendCommands) {
        expect(getGatewayTable(fixture)).toBeTruthy();
        expect(getSensorTable(fixture)).toBeFalsy();
      } else {
        expect(getGatewayTable(fixture)).toBeFalsy();
        expect(getSensorTable(fixture)).toBeTruthy();
      }
    });

    it('should NOT show tables when SUPER_ADMIN has no tenantId', () => {
      const { fixture } = setupTestBed({ session: superAdminSession, queryParams: {} });
      fixture.detectChanges();

      expect(getGatewayTable(fixture)).toBeFalsy();
      expect(getSensorTable(fixture)).toBeFalsy();
    });
  });

  describe('Page → GatewayTable: Input Bindings', () => {
    it('should render gateways and display correct data in cells', () => {
      const { fixture, dashboardServiceMock } = setupTestBed({ session: tenantAdminSession });
      (dashboardServiceMock.gatewayList as WritableSignal<Gateway[]>).set(mockGateways);
      fixture.detectChanges();

      expect(getGatewayRows(fixture).length).toBe(2);

      (dashboardServiceMock.gatewayList as WritableSignal<Gateway[]>).set([mockGateways[0]]);
      fixture.detectChanges();

      const cellTexts = Array.from<HTMLElement>(
        fixture.nativeElement.querySelectorAll('mat-row:not(.detail-row) mat-cell'),
      ).map((c) => c.textContent?.trim());
      expect(cellTexts).toContain('gw-1');
      expect(cellTexts).toContain('Gateway Alpha');
      expect(cellTexts).toContain('tenant-1');
      expect(cellTexts).toContain('ATTIVO');
    });

    it('should show spinner when loading and empty state when idle with no gateways', () => {
      const { fixture, dashboardServiceMock } = setupTestBed({ session: tenantAdminSession });
      (dashboardServiceMock.gatewayLoading as WritableSignal<boolean>).set(true);
      fixture.detectChanges();

      expect(fixture.nativeElement.querySelector('.table-container mat-spinner')).toBeTruthy();

      (dashboardServiceMock.gatewayLoading as WritableSignal<boolean>).set(false);
      fixture.detectChanges();

      const emptyState = fixture.nativeElement.querySelector('.empty-state');
      expect(emptyState).toBeTruthy();
      expect(emptyState.textContent).toContain('Nessun gateway disponibile');
    });

    it('should pass actionMode as dashboard to gateway table', () => {
      const { fixture, dashboardServiceMock } = setupTestBed({ session: tenantAdminSession });
      (dashboardServiceMock.gatewayList as WritableSignal<Gateway[]>).set(mockGateways);
      fixture.detectChanges();

      expect(getGatewayTable(fixture).componentInstance.actionMode()).toBe('dashboard');
    });
  });

  describe('Page → SensorTable: Input Bindings', () => {
    it('should render sensors and display correct data in cells', () => {
      const { fixture, dashboardServiceMock } = setupTestBed({ session: tenantAdminSession });
      (dashboardServiceMock.canSendCommands as WritableSignal<boolean>).set(false);
      (dashboardServiceMock.sensorList as WritableSignal<Sensor[]>).set(mockSensors);
      fixture.detectChanges();

      expect(fixture.nativeElement.querySelectorAll('mat-row').length).toBe(2);

      (dashboardServiceMock.sensorList as WritableSignal<Sensor[]>).set([mockSensors[0]]);
      fixture.detectChanges();

      const cellTexts = Array.from<HTMLElement>(
        fixture.nativeElement.querySelectorAll('mat-row mat-cell'),
      ).map((c) => c.textContent?.trim());
      expect(cellTexts).toContain('sensor-1');
      expect(cellTexts).toContain('Temperature');
      expect(cellTexts).toContain('ATTIVO');
    });

    it('should show spinner when loading and empty state when idle with no sensors', () => {
      const { fixture, dashboardServiceMock } = setupTestBed({ session: tenantAdminSession });
      (dashboardServiceMock.canSendCommands as WritableSignal<boolean>).set(false);
      (dashboardServiceMock.sensorLoading as WritableSignal<boolean>).set(true);
      fixture.detectChanges();

      expect(fixture.nativeElement.querySelector('.table-container mat-spinner')).toBeTruthy();

      (dashboardServiceMock.sensorLoading as WritableSignal<boolean>).set(false);
      fixture.detectChanges();

      const emptyState = fixture.nativeElement.querySelector('.empty-state');
      expect(emptyState).toBeTruthy();
      expect(emptyState.textContent).toContain('Nessun sensore disponibile');
    });
  });

  describe('GatewayTable → Page: Output Events', () => {
    it('should call toggleExpandedGateway with the correct gateway for each row clicked', () => {
      const { fixture, dashboardServiceMock } = setupTestBed({ session: tenantAdminSession });
      (dashboardServiceMock.gatewayList as WritableSignal<Gateway[]>).set(mockGateways);
      fixture.detectChanges();

      const rows = getGatewayRows(fixture);
      rows[0].click();
      expect(dashboardServiceMock.toggleExpandedGateway).toHaveBeenCalledWith(mockGateways[0]);
      rows[1].click();
      expect(dashboardServiceMock.toggleExpandedGateway).toHaveBeenCalledWith(mockGateways[1]);
    });

    it('should call changeGatewayPage when gateway paginator emits', () => {
      const { fixture, dashboardServiceMock } = setupTestBed({ session: tenantAdminSession });
      (dashboardServiceMock.gatewayList as WritableSignal<Gateway[]>).set(mockGateways);
      fixture.detectChanges();

      getGatewayTable(fixture).componentInstance.gatewayPageChange.emit({
        pageIndex: 1,
        pageSize: 10,
        length: 50,
      });

      expect(dashboardServiceMock.changeGatewayPage).toHaveBeenCalledWith(1, 10);
    });

    it('should show snackbar only when command result is true', () => {
      const { fixture, dashboardServiceMock, snackBarMock } = setupTestBed({
        session: tenantAdminSession,
      });
      (dashboardServiceMock.gatewayList as WritableSignal<Gateway[]>).set(mockGateways);
      fixture.detectChanges();

      const gatewayTable = getGatewayTable(fixture);
      gatewayTable.componentInstance.commandRequested.emit(false);
      expect(snackBarMock.open).not.toHaveBeenCalled();

      gatewayTable.componentInstance.commandRequested.emit(true);
      expect(snackBarMock.open).toHaveBeenCalledWith('Comando inviato correttamente', 'Close', {
        duration: 3000,
      });
    });
  });

  describe('SensorTable → Page: Output Events', () => {
    it('should call openChart when sensor table emits chartRequested', () => {
      const { fixture, dashboardServiceMock } = setupTestBed({ session: tenantAdminSession });
      (dashboardServiceMock.canSendCommands as WritableSignal<boolean>).set(false);
      (dashboardServiceMock.sensorList as WritableSignal<Sensor[]>).set(mockSensors);
      fixture.detectChanges();

      getSensorTable(fixture).componentInstance.chartRequested.emit(mockChartRequest);

      expect(dashboardServiceMock.openChart).toHaveBeenCalledWith(mockChartRequest);
    });

    it('should call changeSensorPage when sensor table emits pageChange', () => {
      const { fixture, dashboardServiceMock } = setupTestBed({ session: tenantAdminSession });
      (dashboardServiceMock.canSendCommands as WritableSignal<boolean>).set(false);
      (dashboardServiceMock.sensorList as WritableSignal<Sensor[]>).set(mockSensors);
      fixture.detectChanges();

      getSensorTable(fixture).componentInstance.pageChange.emit({
        pageIndex: 2,
        pageSize: 25,
        length: 100,
      });

      expect(dashboardServiceMock.changeSensorPage).toHaveBeenCalledWith(2, 25);
    });
  });

  describe('Chart Container', () => {
    it('should show chart container, pass chartRequest, and render sensor name in title', () => {
      const { fixture, dashboardServiceMock } = setupTestBed({ session: tenantAdminSession });
      (dashboardServiceMock.selectedChart as WritableSignal<ChartRequest | null>).set(
        mockChartRequest,
      );
      fixture.detectChanges();

      const chart = getChartContainer(fixture);
      expect(chart).toBeTruthy();
      expect(chart.componentInstance.chartRequest()).toEqual(mockChartRequest);
      expect(fixture.nativeElement.querySelector('mat-card-title').textContent).toContain(
        'Temperature',
      );
    });

    it('should NOT show chart container when selectedChart is null', () => {
      const { fixture, dashboardServiceMock } = setupTestBed({ session: tenantAdminSession });
      (dashboardServiceMock.selectedChart as WritableSignal<ChartRequest | null>).set(null);
      fixture.detectChanges();

      expect(getChartContainer(fixture)).toBeFalsy();
    });

    it('should call closeChart when chart container emits chartClosed', () => {
      const { fixture, dashboardServiceMock } = setupTestBed({ session: tenantAdminSession });
      (dashboardServiceMock.selectedChart as WritableSignal<ChartRequest | null>).set(
        mockChartRequest,
      );
      fixture.detectChanges();

      getChartContainer(fixture).componentInstance.chartClosed.emit();

      expect(dashboardServiceMock.closeChart).toHaveBeenCalled();
    });
  });

  describe('Navigation', () => {
    it.each([
      [
        'tenant-management from tenant banner back',
        { tenantId: 'tenant-1' },
        '.tenant-banner .banner-actions button:last-child',
        [['/tenant-management']],
      ],
      [
        'tenant user management from tenant banner',
        { tenantId: 'tenant-xyz' },
        '.tenant-banner .banner-actions button:first-child',
        [['/user-management/tenant-users'], { queryParams: { tenantId: 'tenant-xyz' } }],
      ],
      [
        'tenant-management from warning banner',
        {},
        '.warning-banner button',
        [['/tenant-management']],
      ],
    ] as const)('should navigate to %s', (_label, queryParams, selector, expectedArgs) => {
      const { fixture, routerMock } = setupTestBed({ session: superAdminSession, queryParams });
      fixture.detectChanges();

      fixture.nativeElement.querySelector(selector)!.click();

      expect(routerMock.navigate).toHaveBeenCalledWith(...expectedArgs);
    });
  });

  describe('Route Changes', () => {
    it('should reload dashboard when queryParams change for SUPER_ADMIN', () => {
      const { fixture, dashboardServiceMock, queryParamsSubject } = setupTestBed({
        session: superAdminSession,
        queryParams: { tenantId: 'tenant-1' },
      });
      fixture.detectChanges();

      dashboardServiceMock.loadDashboard.mockClear();
      queryParamsSubject.next({ tenantId: 'tenant-2' });
      fixture.detectChanges();

      expect(dashboardServiceMock.loadDashboard).toHaveBeenCalledWith('tenant-2');
    });
  });
});
