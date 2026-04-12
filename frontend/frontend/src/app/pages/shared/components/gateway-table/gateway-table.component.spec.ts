import { ComponentFixture, TestBed } from '@angular/core/testing';
import { Component, input, output } from '@angular/core';
import { PageEvent } from '@angular/material/paginator';
import { By } from '@angular/platform-browser';

import { MatSnackBar } from '@angular/material/snack-bar';

import { GatewayTableComponent } from './gateway-table.component';
import { GatewayExpandedComponent } from '../gateway-expanded/gateway-expanded.component';
import { Gateway } from '../../../../models/gateway/gateway.model';
import { Sensor } from '../../../../models/sensor/sensor.model';
import { SensorStatus } from '../../../../models/sensor-status.enum';
import { GatewayStatus } from '../../../../models/gateway-status.enum';
import { SensorProfiles } from '../../../../models/sensor/sensor-profiles.enum';
import { ChartRequest } from '../../../../models/chart/chart-request.model';
import { ChartType } from '../../../../models/chart/chart-type.enum';
import { ActionMode } from '../../../../models/action-mode.model';

@Component({ selector: 'app-gateway-expanded', template: '', standalone: true })
class StubGatewayExpanded {
  gateway = input<Gateway>();
  sensors = input<Sensor[]>();
  loading = input<boolean>();
  actionMode = input<ActionMode>();
  sensorTotal = input<number>();
  sensorPageIndex = input<number>();
  sensorLimit = input<number>();
  chartRequested = output<ChartRequest>();
  sensorDeleteRequested = output<Sensor>();
  sensorCreateRequested = output<Gateway>();
  sensorPageChange = output<PageEvent>();
}

describe('GatewayTableComponent (Unit)', () => {
  let component: GatewayTableComponent;
  let fixture: ComponentFixture<GatewayTableComponent>;

  const mockGateways: Gateway[] = [
    {
      id: 'gw-1',
      tenantId: 'tenant-1',
      name: 'Gateway Alpha',
      status: GatewayStatus.ACTIVE,
      interval: 60,
      publicIdentifier: 'pk-gw-1',
    },
    {
      id: 'gw-2',
      tenantId: undefined,
      name: 'Gateway Beta',
      status: GatewayStatus.INACTIVE,
      interval: 120,
    },
  ];

  const mockSensors: Sensor[] = [
    {
      id: 's-1',
      gatewayId: 'gw-1',
      name: 'Temperature',
      profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
      status: SensorStatus.ACTIVE,
      dataInterval: 60,
    },
  ];

  const mockChartRequest: ChartRequest = {
    sensor: mockSensors[0],
    chartType: ChartType.HISTORIC,
    tenantId: 'tenant-1',
    timeInterval: null!,
  };

  const setInput = (key: string, value: unknown) => fixture.componentRef.setInput(key, value);

  const getExpanded = () =>
    fixture.debugElement.query(By.directive(StubGatewayExpanded))
      ?.componentInstance as StubGatewayExpanded;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [GatewayTableComponent],
      providers: [{ provide: MatSnackBar, useValue: { open: vi.fn() } }],
    })
      .overrideComponent(GatewayTableComponent, {
        remove: { imports: [GatewayExpandedComponent] },
        add: { imports: [StubGatewayExpanded] },
      })
      .compileComponents();

    fixture = TestBed.createComponent(GatewayTableComponent);
    component = fixture.componentInstance;

    setInput('gateways', mockGateways);
    setInput('sensors', mockSensors);
    setInput('gatewayTotal', mockGateways.length);
    setInput('sensorTotal', mockSensors.length);
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
    it('should render only spinner when loading', () => {
      setInput('gatewayLoading', true);
      fixture.detectChanges();

      expect(fixture.debugElement.query(By.css('mat-spinner'))).toBeTruthy();
      expect(fixture.debugElement.query(By.css('mat-table'))).toBeFalsy();
      expect(fixture.debugElement.query(By.css('.empty-state'))).toBeFalsy();
      expect(fixture.debugElement.query(By.css('mat-paginator'))).toBeFalsy();
    });
  });

  describe('empty state', () => {
    it('should render only empty state', () => {
      setInput('gateways', []);
      setInput('gatewayTotal', 0);
      setInput('gatewayLoading', false);
      fixture.detectChanges();

      const emptyState = fixture.debugElement.query(By.css('.empty-state'));
      expect(emptyState).toBeTruthy();
      expect(emptyState.query(By.css('p')).nativeElement.textContent).toContain('Nessun gateway disponibile');
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
        .map((c) => c.nativeElement.textContent.trim());
      expect(cellTexts).toEqual(expect.arrayContaining(['gw-1', 'gw-2']));
      expect(cellTexts).toEqual(expect.arrayContaining(['Gateway Alpha', 'Gateway Beta']));
      expect(cellTexts).toEqual(expect.arrayContaining(['ATTIVO', 'INATTIVO']));
      expect(cellTexts).toEqual(expect.arrayContaining(['60', '120']));
    });
  });

  describe('gateway pagination', () => {
    it('should accept pagination inputs', () => {
      setInput('gatewayTotal', 83);
      setInput('gatewayPageIndex', 2);
      setInput('gatewayLimit', 25);
      fixture.detectChanges();

      expect(component.gatewayTotal()).toBe(83);
      expect(component.gatewayPageIndex()).toBe(2);
      expect(component.gatewayLimit()).toBe(25);
    });

    it('should emit gatewayPageChange when paginator fires page event', () => {
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
        'interval',
        'commands',
      ]);

      setInput('actionMode', 'manage');
      fixture.detectChanges();

      expect(component['displayedColumns']()).toEqual([
        'id',
        'tenantId',
        'name',
        'status',
        'interval',
        'publicKey',
        'commands',
        'delete',
      ]);
    });
  });

  describe('manage mode', () => {
    beforeEach(() => {
      setInput('actionMode', 'manage');
      fixture.detectChanges();
    });

    it('should render manager header in manage mode and hide it in dashboard mode', () => {
      expect(fixture.debugElement.query(By.css('.manager-header'))).toBeTruthy();

      setInput('actionMode', 'dashboard');
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

      fixture.debugElement
        .query(By.css('.manager-header button'))
        .triggerEventHandler('click', new MouseEvent('click'));

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
  });

  describe('publicKey column', () => {
    beforeEach(() => {
      setInput('actionMode', 'manage');
      fixture.detectChanges();
    });

    it('should render copy button only for gateways with a publicIdentifier', () => {
      const copyButtons = fixture.debugElement
        .queryAll(By.css('mat-cell button'))
        .filter((btn) => btn.nativeElement.textContent.includes('content_copy'));
      expect(copyButtons.length).toBe(1);
    });

    it('should call navigator.clipboard.writeText when copy button is clicked', async () => {
      const writeText = vi.fn().mockResolvedValue(undefined);
      vi.stubGlobal('navigator', { clipboard: { writeText } });

      const copyButton = fixture.debugElement
        .queryAll(By.css('mat-cell button'))
        .find((btn) => btn.nativeElement.textContent.includes('content_copy'));
      copyButton!.triggerEventHandler('click', { stopPropagation: vi.fn() });

      expect(writeText).toHaveBeenCalledWith('pk-gw-1');

      vi.unstubAllGlobals();
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
      expect(fixture.debugElement.query(By.directive(StubGatewayExpanded))).toBeFalsy();

      setInput('expandedGateway', mockGateways[0]);
      fixture.detectChanges();
      expect(fixture.debugElement.query(By.directive(StubGatewayExpanded))).toBeTruthy();
    });

    it('should pass correct inputs to expanded component', () => {
      setInput('expandedGateway', mockGateways[0]);
      setInput('sensorLoading', true);
      setInput('actionMode', 'dashboard');
      setInput('sensorTotal', 42);
      setInput('sensorPageIndex', 3);
      setInput('sensorLimit', 5);
      fixture.detectChanges();

      const expanded = getExpanded();
      expect(expanded.gateway()).toEqual(mockGateways[0]);
      expect(expanded.sensors()).toEqual(mockSensors);
      expect(expanded.loading()).toBe(true);
      expect(expanded.actionMode()).toBe('dashboard');
      expect(expanded.sensorTotal()).toBe(42);
      expect(expanded.sensorPageIndex()).toBe(3);
      expect(expanded.sensorLimit()).toBe(5);
    });

    it('should apply correct CSS classes when expanded', () => {
      setInput('expandedGateway', mockGateways[0]);
      fixture.detectChanges();

      const dataRows = fixture.debugElement
        .queryAll(By.css('mat-row'))
        .filter((row) => !row.nativeElement.classList.contains('detail-row'));
      expect(dataRows[0].nativeElement.classList.contains('expanded')).toBe(true);
      expect(dataRows[1].nativeElement.classList.contains('expanded')).toBe(false);
      expect(fixture.debugElement.queryAll(By.css('mat-row.detail-row.visible'))).toHaveLength(1);
    });

    it('should emit sensorPageChange from expanded component', () => {
      setInput('expandedGateway', mockGateways[0]);
      fixture.detectChanges();

      const spy = vi.fn();
      component.sensorPageChange.subscribe(spy);

      fixture.debugElement
        .query(By.directive(StubGatewayExpanded))
        .triggerEventHandler('sensorPageChange', { pageIndex: 1, pageSize: 10, length: 42 });

      expect(spy).toHaveBeenCalledWith({ pageIndex: 1, pageSize: 10, length: 42 });
    });

    it('should emit chartRequested from expanded component', () => {
      setInput('expandedGateway', mockGateways[0]);
      fixture.detectChanges();

      const spy = vi.fn();
      component.chartRequested.subscribe(spy);

      fixture.debugElement
        .query(By.directive(StubGatewayExpanded))
        .triggerEventHandler('chartRequested', mockChartRequest);

      expect(spy).toHaveBeenCalledWith(mockChartRequest);
    });

    it('should emit sensorDeleteRequested from expanded component', () => {
      setInput('expandedGateway', mockGateways[0]);
      setInput('actionMode', 'manage');
      fixture.detectChanges();

      const spy = vi.fn();
      component.sensorDeleteRequested.subscribe(spy);

      fixture.debugElement
        .query(By.directive(StubGatewayExpanded))
        .triggerEventHandler('sensorDeleteRequested', mockSensors[0]);

      expect(spy).toHaveBeenCalledWith(mockSensors[0]);
    });

    it('should emit sensorCreateRequested from expanded component', () => {
      setInput('expandedGateway', mockGateways[0]);
      setInput('actionMode', 'manage');
      fixture.detectChanges();

      const spy = vi.fn();
      component.sensorCreateRequested.subscribe(spy);

      fixture.debugElement
        .query(By.directive(StubGatewayExpanded))
        .triggerEventHandler('sensorCreateRequested', mockGateways[0]);

      expect(spy).toHaveBeenCalledWith(mockGateways[0]);
    });
  });

  describe('isExpanded', () => {
    it('should return true only for the matching gateway and false otherwise', () => {
      expect(component['isExpanded'](mockGateways[0])).toBe(false);

      setInput('expandedGateway', mockGateways[0]);
      fixture.detectChanges();

      expect(component['isExpanded'](mockGateways[0])).toBe(true);
      expect(component['isExpanded'](mockGateways[1])).toBe(false);
    });
  });

  describe('inputs', () => {
    it('should accept all inputs', () => {
      setInput('expandedGateway', mockGateways[0]);
      setInput('actionMode', 'manage');
      setInput('gatewayLoading', true);
      setInput('sensorLoading', true);
      setInput('gatewayTotal', 100);
      setInput('gatewayPageIndex', 5);
      setInput('gatewayLimit', 25);
      setInput('sensorTotal', 50);
      setInput('sensorPageIndex', 3);
      setInput('sensorLimit', 5);
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
