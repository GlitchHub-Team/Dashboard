import { ComponentFixture, TestBed } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { PageEvent } from '@angular/material/paginator';
import { MatDialog } from '@angular/material/dialog';
import { Subject } from 'rxjs';

import { SensorTableComponent } from './sensor-table.component';
import { HistoricChartFiltersDialog } from '../../../dashboard/dialogs/historic-chart-filters/historic-chart-filters.dialog';
import { Sensor } from '../../../../models/sensor/sensor.model';
import { SensorProfiles } from '../../../../models/sensor/sensor-profiles.enum';
import { ChartType } from '../../../../models/chart/chart-type.enum';
import { ChartRequest } from '../../../../models/chart/chart-request.model';
import { ActionMode } from '../../../../models/action-mode.model';
import { SensorStatus } from '../../../../models/sensor-status.enum';

describe('SensorTableComponent (Unit)', () => {
  let component: SensorTableComponent;
  let fixture: ComponentFixture<SensorTableComponent>;

  const mockSensors: Sensor[] = [
    {
      id: '1',
      gatewayId: 'gw-1',
      name: 'Temperature',
      profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
      status: SensorStatus.ACTIVE,
      dataInterval: 60,
    },
    {
      id: '2',
      gatewayId: 'gw-2',
      name: 'Humidity',
      profile: SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE,
      status: SensorStatus.INACTIVE,
      dataInterval: 120,
    },
  ];

  let afterClosedSubject: Subject<ChartRequest | undefined>;
  let dialogMock: { open: ReturnType<typeof vi.fn> };

  const setInput = (key: string, value: unknown) => {
    fixture.componentRef.setInput(key, value);
    fixture.detectChanges();
  };

  beforeEach(async () => {
    afterClosedSubject = new Subject();
    dialogMock = {
      open: vi.fn().mockReturnValue({ afterClosed: () => afterClosedSubject.asObservable() }),
    };

    await TestBed.configureTestingModule({
      imports: [SensorTableComponent],
      providers: [{ provide: MatDialog, useValue: dialogMock }],
    }).compileComponents();

    fixture = TestBed.createComponent(SensorTableComponent);
    component = fixture.componentInstance;
    fixture.componentRef.setInput('sensors', mockSensors);
    fixture.componentRef.setInput('total', mockSensors.length);
    fixture.detectChanges();
  });

  describe('initial state', () => {
    it('should create with correct setup, defaults, and dashboard actionMode', () => {
      expect(component).toBeTruthy();
      expect(component['displayedColumns']()).toEqual(['id', 'name', 'profile', 'status', 'commands', 'actions']);
      expect(component['ChartType']).toBe(ChartType);
      expect(component.actionMode()).toBe('dashboard');

      const fresh = TestBed.createComponent(SensorTableComponent);
      fresh.componentRef.setInput('sensors', []);
      fresh.detectChanges();
      expect(fresh.componentInstance.total()).toBe(0);
      expect(fresh.componentInstance.pageIndex()).toBe(0);
      expect(fresh.componentInstance.limit()).toBe(10);
      expect(fresh.componentInstance.loading()).toBeUndefined();
    });
  });

  describe('actionMode', () => {
    it('should show actions column in dashboard mode and delete column in manage mode', () => {
      setInput('actionMode', 'dashboard' as ActionMode);
      expect(component['displayedColumns']()).toContain('actions');
      expect(component['displayedColumns']()).not.toContain('delete');

      setInput('actionMode', 'manage' as ActionMode);
      expect(component['displayedColumns']()).toContain('delete');
      expect(component['displayedColumns']()).not.toContain('actions');
    });

    it('should show manager header only in manage mode', () => {
      setInput('actionMode', 'dashboard' as ActionMode);
      expect(fixture.debugElement.query(By.css('.manager-header'))).toBeFalsy();

      setInput('actionMode', 'manage' as ActionMode);
      expect(fixture.debugElement.query(By.css('.manager-header'))).toBeTruthy();
    });

    it('should emit deleteRequested when delete button is clicked in manage mode', () => {
      setInput('actionMode', 'manage' as ActionMode);
      const spy = vi.fn();
      component.deleteRequested.subscribe(spy);
      fixture.debugElement
        .queryAll(By.css('mat-cell button'))
        .find((btn) => btn.nativeElement.textContent.includes('delete'))!
        .triggerEventHandler('click', { stopPropagation: vi.fn() });
      expect(spy).toHaveBeenCalledWith(mockSensors[0]);
    });

    it('should emit createRequested when new sensor button is clicked in manage mode', () => {
      setInput('actionMode', 'manage' as ActionMode);
      const spy = vi.fn();
      component.createRequested.subscribe(spy);
      fixture.debugElement
        .query(By.css('.manager-header button'))
        .triggerEventHandler('click', new MouseEvent('click'));
      expect(spy).toHaveBeenCalled();
    });
  });

  describe('loading state', () => {
    it('should render only spinner when loading', () => {
      setInput('loading', true);
      expect(fixture.debugElement.query(By.css('mat-spinner'))).toBeTruthy();
      expect(fixture.debugElement.query(By.css('mat-table'))).toBeFalsy();
      expect(fixture.debugElement.query(By.css('.empty-state'))).toBeFalsy();
      expect(fixture.debugElement.query(By.css('mat-paginator'))).toBeFalsy();
    });
  });

  describe('empty state', () => {
    it('should render only empty state', () => {
      fixture.componentRef.setInput('sensors', []);
      fixture.componentRef.setInput('total', 0);
      fixture.componentRef.setInput('loading', false);
      fixture.detectChanges();
      const emptyState = fixture.debugElement.query(By.css('.empty-state'));
      expect(emptyState).toBeTruthy();
      expect(emptyState.query(By.css('p')).nativeElement.textContent).toContain('Nessun sensore disponibile');
      expect(emptyState.query(By.css('mat-icon')).nativeElement.textContent).toContain('router');
      expect(fixture.debugElement.query(By.css('mat-table'))).toBeFalsy();
      expect(fixture.debugElement.query(By.css('mat-spinner'))).toBeFalsy();
      expect(fixture.debugElement.query(By.css('mat-paginator'))).toBeFalsy();
    });
  });

  describe('table with data', () => {
    it('should render table with header, correct rows, and paginator', () => {
      expect(fixture.debugElement.query(By.css('mat-table'))).toBeTruthy();
      expect(fixture.debugElement.query(By.css('mat-header-row'))).toBeTruthy();
      expect(fixture.debugElement.queryAll(By.css('mat-row')).length).toBe(2);
      expect(fixture.debugElement.query(By.css('mat-spinner'))).toBeFalsy();
      expect(fixture.debugElement.query(By.css('.empty-state'))).toBeFalsy();
      expect(fixture.debugElement.query(By.css('mat-paginator'))).toBeTruthy();
    });

    it('should render sensor data in cells', () => {
      const cellTexts = fixture.debugElement
        .queryAll(By.css('mat-cell'))
        .map((cell) => cell.nativeElement.textContent.trim());
      expect(cellTexts).toEqual(expect.arrayContaining(['1', '2']));
      expect(cellTexts).toEqual(expect.arrayContaining(['Temperature', 'Humidity']));
      expect(cellTexts).toEqual(expect.arrayContaining(['ATTIVO', 'INATTIVO']));
    });
  });

  describe('pagination', () => {
    it('should accept pagination inputs', () => {
      fixture.componentRef.setInput('total', 50);
      fixture.componentRef.setInput('pageIndex', 3);
      fixture.componentRef.setInput('limit', 25);
      fixture.detectChanges();
      expect(component.total()).toBe(50);
      expect(component.pageIndex()).toBe(3);
      expect(component.limit()).toBe(25);
    });

    it('should emit pageChange when paginator emits page event', () => {
      const spy = vi.fn();
      component.pageChange.subscribe(spy);
      const event: PageEvent = { pageIndex: 2, pageSize: 10, length: 50 };
      fixture.debugElement.query(By.css('mat-paginator')).triggerEventHandler('page', event);
      expect(spy).toHaveBeenCalledWith(event);
    });
  });

  describe('chart actions', () => {
    it('should render chart buttons only for active sensors', () => {
      const buttons = fixture.debugElement.queryAll(By.css('mat-cell button'));
      expect(
        buttons.filter((btn) => btn.nativeElement.textContent.includes('query_stats')).length,
      ).toBe(1);
      expect(
        buttons.filter((btn) => btn.nativeElement.textContent.includes('ssid_chart')).length,
      ).toBe(1);
    });

    it('should not render chart buttons for inactive sensors', () => {
      fixture.componentRef.setInput('sensors', [mockSensors[1]]);
      fixture.detectChanges();
      const buttons = fixture.debugElement.queryAll(By.css('mat-cell button'));
      expect(
        buttons.filter((btn) => btn.nativeElement.textContent.includes('query_stats')).length,
      ).toBe(0);
      expect(
        buttons.filter((btn) => btn.nativeElement.textContent.includes('ssid_chart')).length,
      ).toBe(0);
    });

    it('should open HistoricChartFiltersDialog and emit chartRequested with dialog result', () => {
      const spy = vi.fn();
      component.chartRequested.subscribe(spy);

      fixture.debugElement
        .queryAll(By.css('mat-cell button'))
        .find((btn) => btn.nativeElement.textContent.includes('query_stats'))!
        .triggerEventHandler('click');

      expect(dialogMock.open).toHaveBeenCalledWith(HistoricChartFiltersDialog, {
        data: { sensor: mockSensors[0], chartType: ChartType.HISTORIC },
      });
      expect(spy).not.toHaveBeenCalled();

      const result: ChartRequest = {
        sensor: mockSensors[0],
        chartType: ChartType.HISTORIC,
        timeInterval: { from: new Date('2025-01-01'), to: new Date('2025-01-02') },
      };
      afterClosedSubject.next(result);
      expect(spy).toHaveBeenCalledWith(result);
    });

    it('should not emit chartRequested when historic dialog is cancelled', () => {
      const spy = vi.fn();
      component.chartRequested.subscribe(spy);

      fixture.debugElement
        .queryAll(By.css('mat-cell button'))
        .find((btn) => btn.nativeElement.textContent.includes('query_stats'))!
        .triggerEventHandler('click');

      afterClosedSubject.next(undefined);
      expect(spy).not.toHaveBeenCalled();
    });

    it('should emit chartRequested with REALTIME when realtime button is clicked', () => {
      const spy = vi.fn();
      component.chartRequested.subscribe(spy);
      fixture.debugElement
        .queryAll(By.css('mat-cell button'))
        .find((btn) => btn.nativeElement.textContent.includes('ssid_chart'))!
        .triggerEventHandler('click');
      expect(spy).toHaveBeenCalledWith({ sensor: mockSensors[0], chartType: ChartType.REALTIME });
    });
  });

  describe('inputs', () => {
    it('should accept all standard inputs', () => {
      expect(component.sensors()).toEqual(mockSensors);
      setInput('sensors', []);
      expect(component.sensors()).toEqual([]);

      setInput('loading', true);
      expect(component.loading()).toBe(true);
    });
  });
});
