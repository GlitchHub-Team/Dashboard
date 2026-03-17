import { ComponentFixture, TestBed } from '@angular/core/testing';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { MatTableModule } from '@angular/material/table';
import { By } from '@angular/platform-browser';

import { DashboardGatewayTableComponent } from './dashboard-gateway-table.component';
import { Gateway } from '../../../../models/gateway.model';
import { GatewayStatus } from '../../../../models/gateway-status.enum';
import { Sensor } from '../../../../models/sensor.model';
import { SensorProfiles } from '../../../../models/sensor-profiles.enum';
import { ChartRequest } from '../../../../models/chart-request.model';
import { ChartType } from '../../../../models/chart-type.enum';

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

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [DashboardGatewayTableComponent, MatTableModule],
      schemas: [NO_ERRORS_SCHEMA],
    }).compileComponents();

    fixture = TestBed.createComponent(DashboardGatewayTableComponent);
    component = fixture.componentInstance;

    fixture.componentRef.setInput('gateways', mockGateways);
    fixture.componentRef.setInput('sensors', mockSensors);
    fixture.detectChanges();
  });

  describe('initial state', () => {
    it('should create', () => {
      expect(component).toBeTruthy();
    });

    it('should default expandedGateway to null', () => {
      expect(component.expandedGateway()).toBeNull();
    });

    it('should default canSendCommands to false', () => {
      expect(component.canSendCommands()).toBe(false);
    });
  });

  describe('loading state', () => {
    it('should render spinner when loading', () => {
      fixture.componentRef.setInput('gatewayLoading', true);
      fixture.detectChanges();

      const spinner = fixture.debugElement.query(By.css('mat-spinner'));
      expect(spinner).toBeTruthy();
    });

    it('should not render table when loading', () => {
      fixture.componentRef.setInput('gatewayLoading', true);
      fixture.detectChanges();

      const table = fixture.debugElement.query(By.css('mat-table'));
      expect(table).toBeFalsy();
    });

    it('should not render empty state when loading', () => {
      fixture.componentRef.setInput('gatewayLoading', true);
      fixture.detectChanges();

      const emptyState = fixture.debugElement.query(By.css('.empty-state'));
      expect(emptyState).toBeFalsy();
    });
  });

  describe('empty state', () => {
    beforeEach(() => {
      fixture.componentRef.setInput('gateways', []);
      fixture.componentRef.setInput('gatewayLoading', false);
      fixture.detectChanges();
    });

    it('should render empty state', () => {
      const emptyState = fixture.debugElement.query(By.css('.empty-state'));
      expect(emptyState).toBeTruthy();
    });

    it('should display no gateways message', () => {
      const message = fixture.debugElement.query(By.css('.empty-state p'));
      expect(message.nativeElement.textContent).toContain('No gateways available');
    });

    it('should render router icon', () => {
      const icon = fixture.debugElement.query(By.css('.empty-state mat-icon'));
      expect(icon.nativeElement.textContent).toContain('router');
    });

    it('should not render table', () => {
      const table = fixture.debugElement.query(By.css('mat-table'));
      expect(table).toBeFalsy();
    });

    it('should not render spinner', () => {
      const spinner = fixture.debugElement.query(By.css('mat-spinner'));
      expect(spinner).toBeFalsy();
    });
  });

  describe('table with data', () => {
    it('should render the table', () => {
      const table = fixture.debugElement.query(By.css('mat-table'));
      expect(table).toBeTruthy();
    });

    it('should not render spinner', () => {
      const spinner = fixture.debugElement.query(By.css('mat-spinner'));
      expect(spinner).toBeFalsy();
    });

    it('should not render empty state', () => {
      const emptyState = fixture.debugElement.query(By.css('.empty-state'));
      expect(emptyState).toBeFalsy();
    });

    it('should render header row', () => {
      const headerRow = fixture.debugElement.query(By.css('mat-header-row'));
      expect(headerRow).toBeTruthy();
    });

    it('should render gateway ids', () => {
      const cells = fixture.debugElement.queryAll(By.css('mat-cell'));
      const idCells = cells.filter(
        (cell) =>
          cell.nativeElement.textContent.trim() === 'gw-1' ||
          cell.nativeElement.textContent.trim() === 'gw-2',
      );
      expect(idCells.length).toBe(2);
    });

    it('should render gateway names', () => {
      const cells = fixture.debugElement.queryAll(By.css('mat-cell'));
      const nameCells = cells.filter(
        (cell) =>
          cell.nativeElement.textContent.includes('Gateway Alpha') ||
          cell.nativeElement.textContent.includes('Gateway Beta'),
      );
      expect(nameCells.length).toBe(2);
    });

    it('should render gateway status in uppercase', () => {
      const cells = fixture.debugElement.queryAll(By.css('mat-cell'));
      const statusCells = cells.filter(
        (cell) =>
          cell.nativeElement.textContent.includes('ONLINE') ||
          cell.nativeElement.textContent.includes('OFFLINE'),
      );
      expect(statusCells.length).toBe(2);
    });
  });

  describe('displayedColumns', () => {
    it('should not include commands column when canSendCommands is false', () => {
      expect(component['displayedColumns']()).toEqual(['id', 'tenantId', 'name', 'status']);
    });

    it('should include commands column when canSendCommands is true', () => {
      fixture.componentRef.setInput('canSendCommands', true);
      fixture.detectChanges();

      expect(component['displayedColumns']()).toEqual([
        'id',
        'tenantId',
        'name',
        'status',
        'commands',
      ]);
    });
  });

  describe('commands column', () => {
    beforeEach(() => {
      fixture.componentRef.setInput('canSendCommands', true);
      fixture.detectChanges();
    });

    it('should render command buttons when canSendCommands is true', () => {
      const buttons = fixture.debugElement.queryAll(By.css('mat-cell button'));
      const commandButtons = buttons.filter((btn) =>
        btn.nativeElement.textContent.includes('terminal'),
      );
      expect(commandButtons.length).toBe(2);
    });

    it('should emit commandRequested when command button is clicked', () => {
      const spy = vi.fn();
      component.commandRequested.subscribe(spy);

      const buttons = fixture.debugElement.queryAll(By.css('mat-cell button'));
      const commandButton = buttons.find((btn) =>
        btn.nativeElement.textContent.includes('terminal'),
      );
      commandButton!.triggerEventHandler('click', { stopPropagation: vi.fn() });
      fixture.detectChanges();

      expect(spy).toHaveBeenCalledWith(mockGateways[0]);
    });
  });

  describe('row interactions', () => {
    it('should emit expandedGatewayChange when row is clicked', () => {
      const spy = vi.fn();
      component.expandedGatewayChange.subscribe(spy);

      const rows = fixture.debugElement.queryAll(By.css('mat-row'));
      const dataRow = rows.find((row) => !row.nativeElement.classList.contains('detail-row'));
      dataRow!.triggerEventHandler('click');
      fixture.detectChanges();

      expect(spy).toHaveBeenCalledWith(mockGateways[0]);
    });
  });

  describe('expanded row', () => {
    it('should not render expanded component when no gateway is expanded', () => {
      const expanded = fixture.debugElement.query(By.css('app-dashboard-gateway-expanded'));
      expect(expanded).toBeFalsy();
    });

    it('should render expanded component when gateway is expanded', () => {
      fixture.componentRef.setInput('expandedGateway', mockGateways[0]);
      fixture.detectChanges();

      const expanded = fixture.debugElement.query(By.css('app-dashboard-gateway-expanded'));
      expect(expanded).toBeTruthy();
    });

    it('should add expanded class to expanded row', () => {
      fixture.componentRef.setInput('expandedGateway', mockGateways[0]);
      fixture.detectChanges();

      const rows = fixture.debugElement.queryAll(By.css('mat-row'));
      const dataRows = rows.filter((row) => !row.nativeElement.classList.contains('detail-row'));
      expect(dataRows[0].nativeElement.classList.contains('expanded')).toBe(true);
      expect(dataRows[1].nativeElement.classList.contains('expanded')).toBe(false);
    });

    it('should add visible class to detail row when expanded', () => {
      fixture.componentRef.setInput('expandedGateway', mockGateways[0]);
      fixture.detectChanges();

      const detailRows = fixture.debugElement.queryAll(By.css('mat-row.detail-row'));
      const visibleRows = detailRows.filter((row) =>
        row.nativeElement.classList.contains('visible'),
      );
      expect(visibleRows.length).toBe(1);
    });

    it('should emit chartRequested from expanded component', () => {
      fixture.componentRef.setInput('expandedGateway', mockGateways[0]);
      fixture.detectChanges();

      const spy = vi.fn();
      component.chartRequested.subscribe(spy);

      const expanded = fixture.debugElement.query(By.css('app-dashboard-gateway-expanded'));

      const request: ChartRequest = {
        sensor: mockSensors[0],
        chartType: ChartType.HISTORIC,
        timeInterval: null!,
      };

      expanded.triggerEventHandler('chartRequested', request);
      fixture.detectChanges();

      expect(spy).toHaveBeenCalledWith(request);
    });
  });

  describe('isExpanded', () => {
    it('should return true when gateway matches expandedGateway', () => {
      fixture.componentRef.setInput('expandedGateway', mockGateways[0]);
      fixture.detectChanges();

      expect(component['isExpanded'](mockGateways[0])).toBe(true);
    });

    it('should return false when gateway does not match', () => {
      fixture.componentRef.setInput('expandedGateway', mockGateways[0]);
      fixture.detectChanges();

      expect(component['isExpanded'](mockGateways[1])).toBe(false);
    });

    it('should return false when expandedGateway is null', () => {
      fixture.componentRef.setInput('expandedGateway', null);
      fixture.detectChanges();

      expect(component['isExpanded'](mockGateways[0])).toBe(false);
    });
  });

  describe('inputs', () => {
    it('should accept gateways', () => {
      expect(component.gateways()).toEqual(mockGateways);
    });

    it('should accept sensors', () => {
      expect(component.sensors()).toEqual(mockSensors);
    });

    it('should accept expandedGateway', () => {
      fixture.componentRef.setInput('expandedGateway', mockGateways[0]);
      fixture.detectChanges();

      expect(component.expandedGateway()).toEqual(mockGateways[0]);
    });

    it('should accept canSendCommands', () => {
      fixture.componentRef.setInput('canSendCommands', true);
      fixture.detectChanges();

      expect(component.canSendCommands()).toBe(true);
    });

    it('should accept gatewayLoading', () => {
      fixture.componentRef.setInput('gatewayLoading', true);
      fixture.detectChanges();

      expect(component.gatewayLoading()).toBe(true);
    });

    it('should accept sensorLoading', () => {
      fixture.componentRef.setInput('sensorLoading', true);
      fixture.detectChanges();

      expect(component.sensorLoading()).toBe(true);
    });
  });
});
