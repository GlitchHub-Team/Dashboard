import { ComponentFixture, TestBed } from '@angular/core/testing';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { MatTableModule } from '@angular/material/table';
import { By } from '@angular/platform-browser';

import { DashboardSensorTableComponent } from './dashboard-sensor-table.component';
import { Sensor } from '../../../../models/sensor.model';
import { SensorProfiles } from '../../../../models/sensor-profiles.enum';
import { ChartType } from '../../../../models/chart-type.enum';
import { ChartRequest } from '../../../../models/chart-request.model';

describe('DashboardSensorTableComponent', () => {
  let component: DashboardSensorTableComponent;
  let fixture: ComponentFixture<DashboardSensorTableComponent>;

  const mockSensors: Sensor[] = [
    {
      id: '1',
      gatewayId: 'gw-1',
      name: 'Temperature',
      profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
    },
    {
      id: '2',
      gatewayId: 'gw-2',
      name: 'Humidity',
      profile: SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE,
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
    fixture.detectChanges();
  });

  describe('initial state', () => {
    it('should create', () => {
      expect(component).toBeTruthy();
    });

    it('should have correct displayed columns', () => {
      expect(component['displayedColumns']).toEqual([
        'id',
        'gatewayId',
        'name',
        'profile',
        'actions',
      ]);
    });

    it('should expose ChartType enum', () => {
      expect(component['ChartType']).toBe(ChartType);
    });
  });

  describe('loading state', () => {
    it('should render spinner when loading', () => {
      fixture.componentRef.setInput('loading', true);
      fixture.detectChanges();

      const spinner = fixture.debugElement.query(By.css('mat-spinner'));
      expect(spinner).toBeTruthy();
    });

    it('should not render table when loading', () => {
      fixture.componentRef.setInput('loading', true);
      fixture.detectChanges();

      const table = fixture.debugElement.query(By.css('mat-table'));
      expect(table).toBeFalsy();
    });

    it('should not render empty state when loading', () => {
      fixture.componentRef.setInput('loading', true);
      fixture.detectChanges();

      const emptyState = fixture.debugElement.query(By.css('.empty-state'));
      expect(emptyState).toBeFalsy();
    });
  });

  describe('empty state', () => {
    beforeEach(() => {
      fixture.componentRef.setInput('sensors', []);
      fixture.componentRef.setInput('loading', false);
      fixture.detectChanges();
    });

    it('should render empty state when no sensors', () => {
      const emptyState = fixture.debugElement.query(By.css('.empty-state'));
      expect(emptyState).toBeTruthy();
    });

    it('should display no sensors message', () => {
      const emptyState = fixture.debugElement.query(By.css('.empty-state p'));
      expect(emptyState.nativeElement.textContent).toContain('No sensors available');
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

    it('should render correct number of rows', () => {
      const rows = fixture.debugElement.queryAll(By.css('mat-row'));
      expect(rows.length).toBe(2);
    });

    it('should render header row', () => {
      const headerRow = fixture.debugElement.query(By.css('mat-header-row'));
      expect(headerRow).toBeTruthy();
    });

    it('should render sensor ids', () => {
      const cells = fixture.debugElement.queryAll(By.css('mat-cell'));
      const idCells = cells.filter(
        (cell) =>
          cell.nativeElement.textContent.trim() === '1' ||
          cell.nativeElement.textContent.trim() === '2',
      );
      expect(idCells.length).toBeGreaterThan(0);
    });

    it('should render sensor names', () => {
      const cells = fixture.debugElement.queryAll(By.css('mat-cell'));
      const nameCells = cells.filter(
        (cell) =>
          cell.nativeElement.textContent.includes('Temperature') ||
          cell.nativeElement.textContent.includes('Humidity'),
      );
      expect(nameCells.length).toBe(2);
    });

    it('should render sensor gateway ids', () => {
      const cells = fixture.debugElement.queryAll(By.css('mat-cell'));
      const gatewayCells = cells.filter(
        (cell) =>
          cell.nativeElement.textContent.includes('gw-1') ||
          cell.nativeElement.textContent.includes('gw-2'),
      );
      expect(gatewayCells.length).toBe(2);
    });

    it('should render profile with titlecase', () => {
      const cells = fixture.debugElement.queryAll(By.css('mat-cell'));
      const profileCells = cells.filter((cell) => {
        const text = cell.nativeElement.textContent.trim();
        return (
          text.charAt(0) === text.charAt(0).toUpperCase() &&
          text.length > 3 &&
          !text.includes('gw-')
        );
      });
      expect(profileCells.length).toBeGreaterThan(0);
    });
  });

  describe('chart actions', () => {
    it('should render historic chart button for each sensor', () => {
      const buttons = fixture.debugElement.queryAll(By.css('mat-cell button'));
      const historicButtons = buttons.filter((btn) =>
        btn.nativeElement.textContent.includes('query_stats'),
      );
      expect(historicButtons.length).toBe(2);
    });

    it('should render realtime chart button for each sensor', () => {
      const buttons = fixture.debugElement.queryAll(By.css('mat-cell button'));
      const realtimeButtons = buttons.filter((btn) =>
        btn.nativeElement.textContent.includes('ssid_chart'),
      );
      expect(realtimeButtons.length).toBe(2);
    });

    it('should emit chartRequested with HISTORIC when historic button is clicked', () => {
      const spy = vi.fn();
      component.chartRequested.subscribe(spy);

      const buttons = fixture.debugElement.queryAll(By.css('mat-cell button'));
      const historicButton = buttons.find((btn) =>
        btn.nativeElement.textContent.includes('query_stats'),
      );
      historicButton!.triggerEventHandler('click');
      fixture.detectChanges();

      const expected: ChartRequest = {
        sensor: mockSensors[0],
        chartType: ChartType.HISTORIC,
        timeInterval: null!,
      };
      expect(spy).toHaveBeenCalledWith(expected);
    });

    it('should emit chartRequested with REALTIME when realtime button is clicked', () => {
      const spy = vi.fn();
      component.chartRequested.subscribe(spy);

      const buttons = fixture.debugElement.queryAll(By.css('mat-cell button'));
      const realtimeButton = buttons.find((btn) =>
        btn.nativeElement.textContent.includes('ssid_chart'),
      );
      realtimeButton!.triggerEventHandler('click');
      fixture.detectChanges();

      expect(spy).toHaveBeenCalledWith(
        expect.objectContaining({
          sensor: mockSensors[0],
          chartType: ChartType.REALTIME,
        }),
      );
    });
  });

  describe('inputs', () => {
    it('should accept sensors', () => {
      expect(component.sensors()).toEqual(mockSensors);
    });

    it('should accept empty sensors array', () => {
      fixture.componentRef.setInput('sensors', []);
      fixture.detectChanges();

      expect(component.sensors()).toEqual([]);
    });

    it('should default loading to undefined', () => {
      expect(component.loading()).toBeUndefined();
    });

    it('should accept loading input', () => {
      fixture.componentRef.setInput('loading', true);
      fixture.detectChanges();

      expect(component.loading()).toBe(true);
    });
  });
});
