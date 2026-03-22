import { ComponentFixture, TestBed } from '@angular/core/testing';
import { Component, input, output } from '@angular/core';
import { PageEvent } from '@angular/material/paginator';
import { By } from '@angular/platform-browser';

import { DashboardGatewayExpandedComponent } from './dashboard-gateway-expanded.component';
import { DashboardSensorTableComponent } from '../dashboard-sensor-table/dashboard-sensor-table.component';
import { Gateway } from '../../../../models/gateway/gateway.model';
import { Sensor } from '../../../../models/sensor/sensor.model';
import { Status } from '../../../../models/gateway-sensor-status.enum';
import { SensorProfiles } from '../../../../models/sensor/sensor-profiles.enum';
import { ChartRequest } from '../../../../models/chart/chart-request.model';
import { ChartType } from '../../../../models/chart/chart-type.enum';
import { ActionMode } from '../../../../models/action-mode.model';

@Component({ selector: 'app-dashboard-sensor-table', template: '', standalone: true })
class StubSensorTable {
  sensors = input<Sensor[]>();
  loading = input<boolean>();
  actionMode = input<ActionMode>();
  total = input<number>();
  pageIndex = input<number>();
  limit = input<number>();
  chartRequested = output<ChartRequest>();
  createRequested = output<void>();
  deleteRequested = output<Sensor>();
  pageChange = output<PageEvent>();
}

describe('DashboardGatewayExpandedComponent (Unit)', () => {
  let component: DashboardGatewayExpandedComponent;
  let fixture: ComponentFixture<DashboardGatewayExpandedComponent>;

  const mockGateway: Gateway = {
    id: 'gw-1',
    tenantId: 'tenant-1',
    name: 'Gateway Alpha',
    status: Status.ACTIVE,
    interval: 60,
  };
  const mockSensors: Sensor[] = [
    {
      id: 's-1',
      gatewayId: 'gw-1',
      name: 'Temperature',
      profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
      status: Status.ACTIVE,
      dataInterval: 60,
    },
    {
      id: 's-2',
      gatewayId: 'gw-1',
      name: 'Humidity',
      profile: SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE,
      status: Status.INACTIVE,
      dataInterval: 120,
    },
  ];

  const setInput = (key: string, value: unknown) => fixture.componentRef.setInput(key, value);

  const getTable = () =>
    fixture.debugElement.query(By.directive(StubSensorTable)).componentInstance as StubSensorTable;

  beforeEach(async () => {
    await TestBed.configureTestingModule({ imports: [DashboardGatewayExpandedComponent] })
      .overrideComponent(DashboardGatewayExpandedComponent, {
        remove: { imports: [DashboardSensorTableComponent] },
        add: { imports: [StubSensorTable] },
      })
      .compileComponents();

    fixture = TestBed.createComponent(DashboardGatewayExpandedComponent);
    component = fixture.componentInstance;

    setInput('sensors', mockSensors);
    setInput('gateway', mockGateway);
    setInput('sensorTotal', mockSensors.length);
    fixture.detectChanges();
  });

  describe('initial state', () => {
    it('should create with correct defaults', () => {
      expect(component).toBeTruthy();
      expect(component.loading()).toBeUndefined();
      expect(component.actionMode()).toBe('dashboard');
    });

    it('should default pagination inputs', () => {
      const fresh = TestBed.createComponent(DashboardGatewayExpandedComponent);
      fresh.componentRef.setInput('sensors', []);
      fresh.componentRef.setInput('gateway', mockGateway);
      fresh.detectChanges();

      expect(fresh.componentInstance.sensorTotal()).toBe(0);
      expect(fresh.componentInstance.sensorPageIndex()).toBe(0);
      expect(fresh.componentInstance.sensorLimit()).toBe(10);
    });
  });

  describe('template', () => {
    it('should render gateway id in heading and sensor table stub', () => {
      expect(fixture.debugElement.query(By.css('h2')).nativeElement.textContent).toContain('gw-1');
      expect(fixture.debugElement.query(By.directive(StubSensorTable))).toBeTruthy();
    });
  });

  describe('input bindings', () => {
    it('should pass sensors, loading, actionMode, and pagination to sensor table', () => {
      setInput('loading', true);
      setInput('actionMode', 'manage');
      setInput('sensorTotal', 50);
      setInput('sensorPageIndex', 3);
      setInput('sensorLimit', 5);
      fixture.detectChanges();

      const table = getTable();
      expect(table.sensors()).toEqual(mockSensors);
      expect(table.loading()).toBe(true);
      expect(table.actionMode()).toBe('manage');
      expect(table.total()).toBe(50);
      expect(table.pageIndex()).toBe(3);
      expect(table.limit()).toBe(5);
    });
  });

  describe('output events', () => {
    it('should emit chartRequested when sensor table emits chartRequested', () => {
      const spy = vi.fn();
      component.chartRequested.subscribe(spy);

      const request: ChartRequest = {
        sensor: mockSensors[0],
        chartType: ChartType.HISTORIC,
        timeInterval: null!,
      };
      fixture.debugElement
        .query(By.directive(StubSensorTable))
        .triggerEventHandler('chartRequested', request);

      expect(spy).toHaveBeenCalledWith(request);
    });

    it('should emit sensorCreateRequested with gateway when sensor table emits createRequested', () => {
      const spy = vi.fn();
      component.sensorCreateRequested.subscribe(spy);

      fixture.debugElement
        .query(By.directive(StubSensorTable))
        .triggerEventHandler('createRequested');

      expect(spy).toHaveBeenCalledWith(mockGateway);
    });

    it('should emit sensorDeleteRequested when sensor table emits deleteRequested', () => {
      const spy = vi.fn();
      component.sensorDeleteRequested.subscribe(spy);

      fixture.debugElement
        .query(By.directive(StubSensorTable))
        .triggerEventHandler('deleteRequested', mockSensors[0]);

      expect(spy).toHaveBeenCalledWith(mockSensors[0]);
    });

    it('should emit sensorPageChange when sensor table emits pageChange', () => {
      const spy = vi.fn();
      component.sensorPageChange.subscribe(spy);

      const event: PageEvent = { pageIndex: 2, pageSize: 10, length: 50 };
      fixture.debugElement
        .query(By.directive(StubSensorTable))
        .triggerEventHandler('pageChange', event);

      expect(spy).toHaveBeenCalledWith(event);
    });
  });

  describe('inputs', () => {
    it('should accept all inputs', () => {
      setInput('actionMode', 'manage');
      setInput('loading', true);
      setInput('sensorTotal', 50);
      setInput('sensorPageIndex', 3);
      setInput('sensorLimit', 5);
      fixture.detectChanges();

      expect(component.sensors()).toEqual(mockSensors);
      expect(component.gateway()).toEqual(mockGateway);
      expect(component.actionMode()).toBe('manage');
      expect(component.loading()).toBe(true);
      expect(component.sensorTotal()).toBe(50);
      expect(component.sensorPageIndex()).toBe(3);
      expect(component.sensorLimit()).toBe(5);
    });
  });
});
