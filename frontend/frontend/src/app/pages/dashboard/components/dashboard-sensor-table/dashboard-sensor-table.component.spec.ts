import { ComponentFixture, TestBed } from '@angular/core/testing';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { MatTableModule } from '@angular/material/table';
import { By } from '@angular/platform-browser';
import { PageEvent } from '@angular/material/paginator';

import { DashboardSensorTableComponent } from './dashboard-sensor-table.component';
import { Sensor } from '../../../../models/sensor/sensor.model';
import { SensorProfiles } from '../../../../models/sensor/sensor-profiles.enum';
import { ChartType } from '../../../../models/chart/chart-type.enum';
import { ActionMode } from '../../../../models/action-mode.model';
import { Status } from '../../../../models/gateway-sensor-status.enum';

describe('DashboardSensorTableComponent', () => {
  let component: DashboardSensorTableComponent;
  let fixture: ComponentFixture<DashboardSensorTableComponent>;

  const mockSensors: Sensor[] = [
    {
      id: '1',
      gatewayId: 'gw-1',
      name: 'Temperature',
      profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
      status: Status.ACTIVE,
      dataInterval: 60,
    },
    {
      id: '2',
      gatewayId: 'gw-2',
      name: 'Humidity',
      profile: SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE,
      status: Status.INACTIVE,
      dataInterval: 120,
    },
  ];

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [DashboardSensorTableComponent, MatTableModule],
      schemas: [NO_ERRORS_SCHEMA],
    }).compileComponents();

    fixture = TestBed.createComponent(DashboardSensorTableComponent);
    component = fixture.componentInstance;

    fixture.componentRef.setInput('sensors', mockSensors);
    fixture.componentRef.setInput('total', mockSensors.length);
    fixture.detectChanges();
  });

  describe('initial state', () => {
    it('should create with correct setup', () => {
      expect(component).toBeTruthy();
      expect(component['displayedColumns']()).toEqual([
        'id',
        'gatewayId',
        'name',
        'profile',
        'actions',
      ]);
      expect(component['ChartType']).toBe(ChartType);
    });

    it('should default pagination inputs', () => {
      const fresh = TestBed.createComponent(DashboardSensorTableComponent);
      fresh.componentRef.setInput('sensors', []);
      fresh.detectChanges();

      expect(fresh.componentInstance.total()).toBe(0);
      expect(fresh.componentInstance.pageIndex()).toBe(0);
      expect(fresh.componentInstance.limit()).toBe(10);
      expect(fresh.componentInstance.loading()).toBeUndefined();
    });

    it('should default actionMode to dashboard', () => {
      expect(component.actionMode()).toBe('dashboard');
    });
  });

  describe('actionMode', () => {
    it('should show actions column in dashboard mode', () => {
      fixture.componentRef.setInput('actionMode', 'dashboard' as ActionMode);
      fixture.detectChanges();

      expect(component['displayedColumns']()).toContain('actions');
      expect(component['displayedColumns']()).not.toContain('delete');
    });

    it('should show delete column in manage mode', () => {
      fixture.componentRef.setInput('actionMode', 'manage' as ActionMode);
      fixture.detectChanges();

      expect(component['displayedColumns']()).toContain('delete');
      expect(component['displayedColumns']()).not.toContain('actions');
    });

    it('should render manager header and create button in manage mode', () => {
      fixture.componentRef.setInput('actionMode', 'manage' as ActionMode);
      fixture.detectChanges();

      expect(fixture.debugElement.query(By.css('.manager-header'))).toBeTruthy();
    });

    it('should not render manager header in dashboard mode', () => {
      fixture.componentRef.setInput('actionMode', 'dashboard' as ActionMode);
      fixture.detectChanges();

      expect(fixture.debugElement.query(By.css('.manager-header'))).toBeFalsy();
    });

    it('should emit deleteRequested when delete button is clicked in manage mode', () => {
      fixture.componentRef.setInput('actionMode', 'manage' as ActionMode);
      fixture.detectChanges();

      const spy = vi.fn();
      component.deleteRequested.subscribe(spy);

      const deleteButton = fixture.debugElement.query(By.css('mat-cell button'));
      deleteButton.triggerEventHandler('click', new MouseEvent('click'));

      expect(spy).toHaveBeenCalledWith(mockSensors[0]);
    });

    it('should emit createRequested when new sensor button is clicked in manage mode', () => {
      fixture.componentRef.setInput('actionMode', 'manage' as ActionMode);
      fixture.detectChanges();

      const spy = vi.fn();
      component.createRequested.subscribe(spy);

      const createButton = fixture.debugElement.query(By.css('.manager-header button'));
      createButton.triggerEventHandler('click', new MouseEvent('click'));

      expect(spy).toHaveBeenCalled();
    });
  });

  describe('loading state', () => {
    beforeEach(() => {
      fixture.componentRef.setInput('loading', true);
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
      fixture.componentRef.setInput('sensors', []);
      fixture.componentRef.setInput('total', 0);
      fixture.componentRef.setInput('loading', false);
      fixture.detectChanges();
    });

    it('should render only empty state', () => {
      const emptyState = fixture.debugElement.query(By.css('.empty-state'));
      expect(emptyState).toBeTruthy();
      expect(emptyState.query(By.css('p')).nativeElement.textContent).toContain(
        'No sensors available',
      );
      expect(emptyState.query(By.css('mat-icon')).nativeElement.textContent).toContain('router');
      expect(fixture.debugElement.query(By.css('mat-table'))).toBeFalsy();
      expect(fixture.debugElement.query(By.css('mat-spinner'))).toBeFalsy();
      expect(fixture.debugElement.query(By.css('mat-paginator'))).toBeFalsy();
    });
  });

  describe('table with data', () => {
    it('should render table with header and correct rows', () => {
      expect(fixture.debugElement.query(By.css('mat-table'))).toBeTruthy();
      expect(fixture.debugElement.query(By.css('mat-header-row'))).toBeTruthy();
      expect(fixture.debugElement.queryAll(By.css('mat-row')).length).toBe(2);
      expect(fixture.debugElement.query(By.css('mat-spinner'))).toBeFalsy();
      expect(fixture.debugElement.query(By.css('.empty-state'))).toBeFalsy();
    });

    it('should render sensor data in cells', () => {
      const cellTexts = fixture.debugElement
        .queryAll(By.css('mat-cell'))
        .map((cell) => cell.nativeElement.textContent.trim());

      expect(cellTexts).toEqual(expect.arrayContaining(['1', '2']));
      expect(cellTexts).toEqual(expect.arrayContaining(['Temperature', 'Humidity']));
      expect(cellTexts).toEqual(expect.arrayContaining(['gw-1', 'gw-2']));
    });

    it('should render paginator', () => {
      expect(fixture.debugElement.query(By.css('mat-paginator'))).toBeTruthy();
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

      const paginator = fixture.debugElement.query(By.css('mat-paginator'));
      const event: PageEvent = { pageIndex: 2, pageSize: 10, length: 50 };

      paginator.triggerEventHandler('page', event);

      expect(spy).toHaveBeenCalledWith(event);
    });
  });

  describe('chart actions', () => {
    it('should render chart buttons for each sensor', () => {
      const buttons = fixture.debugElement.queryAll(By.css('mat-cell button'));
      const historicButtons = buttons.filter((btn) =>
        btn.nativeElement.textContent.includes('query_stats'),
      );
      const realtimeButtons = buttons.filter((btn) =>
        btn.nativeElement.textContent.includes('ssid_chart'),
      );
      expect(historicButtons.length).toBe(2);
      expect(realtimeButtons.length).toBe(2);
    });

    it('should emit chartRequested with HISTORIC when historic button is clicked', () => {
      const spy = vi.fn();
      component.chartRequested.subscribe(spy);

      const historicButton = fixture.debugElement
        .queryAll(By.css('mat-cell button'))
        .find((btn) => btn.nativeElement.textContent.includes('query_stats'));
      historicButton!.triggerEventHandler('click');

      expect(spy).toHaveBeenCalledWith({
        sensor: mockSensors[0],
        chartType: ChartType.HISTORIC,
        timeInterval: null!,
      });
    });

    it('should emit chartRequested with REALTIME when realtime button is clicked', () => {
      const spy = vi.fn();
      component.chartRequested.subscribe(spy);

      const realtimeButton = fixture.debugElement
        .queryAll(By.css('mat-cell button'))
        .find((btn) => btn.nativeElement.textContent.includes('ssid_chart'));
      realtimeButton!.triggerEventHandler('click');

      expect(spy).toHaveBeenCalledWith(
        expect.objectContaining({
          sensor: mockSensors[0],
          chartType: ChartType.REALTIME,
        }),
      );
    });
  });

  describe('inputs', () => {
    it('should accept all standard inputs', () => {
      expect(component.sensors()).toEqual(mockSensors);

      fixture.componentRef.setInput('sensors', []);
      fixture.detectChanges();
      expect(component.sensors()).toEqual([]);
    });

    it('should accept loading input', () => {
      fixture.componentRef.setInput('loading', true);
      fixture.detectChanges();

      expect(component.loading()).toBe(true);
    });
  });
});
