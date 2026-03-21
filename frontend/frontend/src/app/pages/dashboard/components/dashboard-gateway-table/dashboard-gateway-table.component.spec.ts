import { ComponentFixture, TestBed } from '@angular/core/testing';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { MatTableModule } from '@angular/material/table';
import { By } from '@angular/platform-browser';

import { DashboardGatewayTableComponent } from './dashboard-gateway-table.component';
import { Gateway } from '../../../../models/gateway/gateway.model';
import { GatewayStatus } from '../../../../models/gateway/gateway-status.enum';
import { Sensor } from '../../../../models/sensor/sensor.model';
import { SensorProfiles } from '../../../../models/sensor/sensor-profiles.enum';
import { ChartRequest } from '../../../../models/chart/chart-request.model';
import { ChartType } from '../../../../models/chart/chart-type.enum';

describe('DashboardGatewayTableComponent', () => {
  let component: DashboardGatewayTableComponent;
  let fixture: ComponentFixture<DashboardGatewayTableComponent>;

  const mockGateways: Gateway[] = [
    { id: 'gw-1', tenantId: 'tenant-1', name: 'Gateway Alpha', status: GatewayStatus.ONLINE },
    { id: 'gw-2', tenantId: 'tenant-1', name: 'Gateway Beta', status: GatewayStatus.OFFLINE },
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

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [DashboardGatewayTableComponent, MatTableModule],
      schemas: [NO_ERRORS_SCHEMA],
    }).compileComponents();

    fixture = TestBed.createComponent(DashboardGatewayTableComponent);
    component = fixture.componentInstance;

    fixture.componentRef.setInput('gateways', mockGateways);
    fixture.componentRef.setInput('sensors', mockSensors);
    fixture.componentRef.setInput('gatewayTotal', mockGateways.length);
    fixture.componentRef.setInput('sensorTotal', mockSensors.length);
    fixture.detectChanges();
  });

  describe('initial state', () => {
    it('should create with correct defaults', () => {
      expect(component).toBeTruthy();
      expect(component.expandedGateway()).toBeNull();
      expect(component.gatewayPageIndex()).toBe(0);
      expect(component.gatewayLimit()).toBe(10);
      expect(component.sensorPageIndex()).toBe(0);
      expect(component.sensorLimit()).toBe(10);
    });
  });

  describe('loading state', () => {
    beforeEach(() => {
      fixture.componentRef.setInput('gatewayLoading', true);
      fixture.detectChanges();
    });

    it('should render only spinner when loading', () => {
      expect(fixture.debugElement.query(By.css('mat-spinner'))).toBeTruthy();
      expect(fixture.debugElement.query(By.css('mat-table'))).toBeFalsy();
      expect(fixture.debugElement.query(By.css('.empty-state'))).toBeFalsy();
      expect(fixture.debugElement.query(By.css('mat-paginator'))).toBeFalsy();
    });
  });

  describe('empty state', () => {
    beforeEach(() => {
      fixture.componentRef.setInput('gateways', []);
      fixture.componentRef.setInput('gatewayTotal', 0);
      fixture.componentRef.setInput('gatewayLoading', false);
      fixture.detectChanges();
    });

    it('should render only empty state', () => {
      const emptyState = fixture.debugElement.query(By.css('.empty-state'));
      expect(emptyState).toBeTruthy();
      expect(emptyState.query(By.css('p')).nativeElement.textContent).toContain(
        'No gateways available',
      );
      expect(emptyState.query(By.css('mat-icon')).nativeElement.textContent).toContain('router');
      expect(fixture.debugElement.query(By.css('mat-table'))).toBeFalsy();
      expect(fixture.debugElement.query(By.css('mat-spinner'))).toBeFalsy();
      expect(fixture.debugElement.query(By.css('mat-paginator'))).toBeFalsy();
    });
  });

  describe('table with data', () => {
    it('should render table, paginator, and no spinner or empty state', () => {
      expect(fixture.debugElement.query(By.css('mat-table'))).toBeTruthy();
      expect(fixture.debugElement.query(By.css('mat-header-row'))).toBeTruthy();
      expect(fixture.debugElement.query(By.css('mat-paginator'))).toBeTruthy();
      expect(fixture.debugElement.query(By.css('mat-spinner'))).toBeFalsy();
      expect(fixture.debugElement.query(By.css('.empty-state'))).toBeFalsy();
    });

    it('should render gateway data in cells', () => {
      const cellTexts = fixture.debugElement
        .queryAll(By.css('mat-cell'))
        .map((cell) => cell.nativeElement.textContent.trim());

      expect(cellTexts).toEqual(expect.arrayContaining(['gw-1', 'gw-2']));
      expect(cellTexts).toEqual(expect.arrayContaining(['Gateway Alpha', 'Gateway Beta']));
      expect(cellTexts).toEqual(expect.arrayContaining(['ONLINE', 'OFFLINE']));
    });
  });

  describe('gateway pagination', () => {
    it('should accept pagination inputs', () => {
      fixture.componentRef.setInput('gatewayTotal', 83);
      fixture.componentRef.setInput('gatewayPageIndex', 2);
      fixture.componentRef.setInput('gatewayLimit', 25);
      fixture.detectChanges();

      expect(component.gatewayTotal()).toBe(83);
      expect(component.gatewayPageIndex()).toBe(2);
      expect(component.gatewayLimit()).toBe(25);
    });

    it('should emit gatewayPageChange when paginator emits page event', () => {
      const spy = vi.fn();
      component.gatewayPageChange.subscribe(spy);

      const pageEvent = { pageIndex: 2, pageSize: 10, length: 83 };
      fixture.debugElement.query(By.css('mat-paginator')).triggerEventHandler('page', pageEvent);

      expect(spy).toHaveBeenCalledWith(pageEvent);
    });
  });

  describe('displayedColumns', () => {
    it('should show commands column by default and delete column in manage mode', () => {
      expect(component['displayedColumns']()).toEqual([
        'id',
        'tenantId',
        'name',
        'status',
        'commands',
      ]);

      fixture.componentRef.setInput('actionMode', 'manage');
      fixture.detectChanges();

      expect(component['displayedColumns']()).toEqual([
        'id',
        'tenantId',
        'name',
        'status',
        'delete',
      ]);
    });
  });

  describe('manage mode', () => {
    beforeEach(() => {
      fixture.componentRef.setInput('actionMode', 'manage');
      fixture.detectChanges();
    });

    it('should render manager header', () => {
      expect(fixture.debugElement.query(By.css('.manager-header'))).toBeTruthy();
    });

    it('should not render manager header in dashboard mode', () => {
      fixture.componentRef.setInput('actionMode', 'dashboard');
      fixture.detectChanges();

      expect(fixture.debugElement.query(By.css('.manager-header'))).toBeFalsy();
    });

    it('should render delete buttons for each gateway', () => {
      const deleteButtons = fixture.debugElement
        .queryAll(By.css('mat-cell button'))
        .filter((btn) => btn.nativeElement.textContent.includes('delete'));
      expect(deleteButtons.length).toBe(2);
    });

    it('should emit gatewayDeleteRequested when delete button is clicked', () => {
      const spy = vi.fn();
      component.gatewayDeleteRequested.subscribe(spy);

      const deleteButton = fixture.debugElement
        .queryAll(By.css('mat-cell button'))
        .find((btn) => btn.nativeElement.textContent.includes('delete'));
      deleteButton!.triggerEventHandler('click', { stopPropagation: vi.fn() });

      expect(spy).toHaveBeenCalledWith(mockGateways[0]);
    });

    it('should emit gatewayCreateRequested when new gateway button is clicked', () => {
      const spy = vi.fn();
      component.gatewayCreateRequested.subscribe(spy);

      const createButton = fixture.debugElement.query(By.css('.manager-header button'));
      createButton.triggerEventHandler('click', new MouseEvent('click'));

      expect(spy).toHaveBeenCalled();
    });
  });

  describe('commands column', () => {
    it('should render command buttons for each gateway', () => {
      const commandButtons = fixture.debugElement
        .queryAll(By.css('mat-cell button'))
        .filter((btn) => btn.nativeElement.textContent.includes('terminal'));
      expect(commandButtons.length).toBe(2);
    });

    it('should emit commandRequested when command button is clicked', () => {
      const spy = vi.fn();
      component.commandRequested.subscribe(spy);

      const commandButton = fixture.debugElement
        .queryAll(By.css('mat-cell button'))
        .find((btn) => btn.nativeElement.textContent.includes('terminal'));
      commandButton!.triggerEventHandler('click', { stopPropagation: vi.fn() });

      expect(spy).toHaveBeenCalledWith(mockGateways[0]);
    });
  });

  describe('row expansion', () => {
    it('should emit expandedGatewayChange when row is clicked', () => {
      const spy = vi.fn();
      component.expandedGatewayChange.subscribe(spy);

      const dataRow = fixture.debugElement
        .queryAll(By.css('mat-row'))
        .find((row) => !row.nativeElement.classList.contains('detail-row'));
      dataRow!.triggerEventHandler('click');

      expect(spy).toHaveBeenCalledWith(mockGateways[0]);
    });

    it('should hide expanded component by default and show it when a gateway is expanded', () => {
      expect(fixture.debugElement.query(By.css('app-dashboard-gateway-expanded'))).toBeFalsy();

      fixture.componentRef.setInput('expandedGateway', mockGateways[0]);
      fixture.detectChanges();

      expect(fixture.debugElement.query(By.css('app-dashboard-gateway-expanded'))).toBeTruthy();
    });

    it('should apply correct CSS classes when expanded', () => {
      fixture.componentRef.setInput('expandedGateway', mockGateways[0]);
      fixture.detectChanges();

      const dataRows = fixture.debugElement
        .queryAll(By.css('mat-row'))
        .filter((row) => !row.nativeElement.classList.contains('detail-row'));
      expect(dataRows[0].nativeElement.classList.contains('expanded')).toBe(true);
      expect(dataRows[1].nativeElement.classList.contains('expanded')).toBe(false);

      const visibleDetailRows = fixture.debugElement.queryAll(By.css('mat-row.detail-row.visible'));
      expect(visibleDetailRows.length).toBe(1);
    });

    it('should emit sensorPageChange from expanded component', () => {
      fixture.componentRef.setInput('expandedGateway', mockGateways[0]);
      fixture.detectChanges();

      const spy = vi.fn();
      component.sensorPageChange.subscribe(spy);

      fixture.debugElement
        .query(By.css('app-dashboard-gateway-expanded'))
        .triggerEventHandler('sensorPageChange', { pageIndex: 1, pageSize: 10, length: 42 });

      expect(spy).toHaveBeenCalledWith({ pageIndex: 1, pageSize: 10, length: 42 });
    });

    it('should emit chartRequested from expanded component', () => {
      fixture.componentRef.setInput('expandedGateway', mockGateways[0]);
      fixture.detectChanges();

      const spy = vi.fn();
      component.chartRequested.subscribe(spy);

      fixture.debugElement
        .query(By.css('app-dashboard-gateway-expanded'))
        .triggerEventHandler('chartRequested', mockChartRequest);

      expect(spy).toHaveBeenCalledWith(mockChartRequest);
    });
  });

  describe('isExpanded', () => {
    it('should return true only for the matching gateway and false otherwise', () => {
      expect(component['isExpanded'](mockGateways[0])).toBe(false);

      fixture.componentRef.setInput('expandedGateway', mockGateways[0]);
      fixture.detectChanges();

      expect(component['isExpanded'](mockGateways[0])).toBe(true);
      expect(component['isExpanded'](mockGateways[1])).toBe(false);
    });
  });

  describe('inputs', () => {
    it('should accept all inputs', () => {
      fixture.componentRef.setInput('expandedGateway', mockGateways[0]);
      fixture.componentRef.setInput('actionMode', 'manage');
      fixture.componentRef.setInput('gatewayLoading', true);
      fixture.componentRef.setInput('sensorLoading', true);
      fixture.componentRef.setInput('gatewayTotal', 100);
      fixture.componentRef.setInput('gatewayPageIndex', 5);
      fixture.componentRef.setInput('gatewayLimit', 25);
      fixture.componentRef.setInput('sensorTotal', 50);
      fixture.componentRef.setInput('sensorPageIndex', 3);
      fixture.componentRef.setInput('sensorLimit', 5);
      fixture.detectChanges();

      expect(component.gateways()).toEqual(mockGateways);
      expect(component.sensors()).toEqual(mockSensors);
      expect(component.expandedGateway()).toEqual(mockGateways[0]);
      expect(component.gatewayLoading()).toBe(true);
      expect(component.sensorLoading()).toBe(true);
      expect(component.gatewayTotal()).toBe(100);
      expect(component.gatewayPageIndex()).toBe(5);
      expect(component.gatewayLimit()).toBe(25);
      expect(component.sensorTotal()).toBe(50);
      expect(component.sensorPageIndex()).toBe(3);
      expect(component.sensorLimit()).toBe(5);
    });
  });
});
