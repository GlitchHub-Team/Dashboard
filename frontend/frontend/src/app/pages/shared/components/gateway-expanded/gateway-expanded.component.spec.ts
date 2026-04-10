import { ComponentFixture, TestBed } from '@angular/core/testing';
import { Component, input, output } from '@angular/core';
import { PageEvent } from '@angular/material/paginator';
import { By } from '@angular/platform-browser';

import { GatewayExpandedComponent } from './gateway-expanded.component';
import { SensorTableComponent } from '../sensor-table/sensor-table.component';
import { Gateway } from '../../../../models/gateway/gateway.model';
import { Sensor } from '../../../../models/sensor/sensor.model';
import { SensorStatus } from '../../../../models/sensor-status.enum';
import { GatewayStatus } from '../../../../models/gateway-status.enum';
import { SensorProfiles } from '../../../../models/sensor/sensor-profiles.enum';
import { ChartRequest } from '../../../../models/chart/chart-request.model';
import { ChartType } from '../../../../models/chart/chart-type.enum';
import { ActionMode } from '../../../../models/action-mode.model';

@Component({ selector: 'app-sensor-table', template: '', standalone: true })
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

describe('GatewayExpandedComponent (Unit)', () => {
  let component: GatewayExpandedComponent;
  let fixture: ComponentFixture<GatewayExpandedComponent>;

  const mockGateway: Gateway = {
    id: 'gw-1',
    tenantId: 'tenant-1',
    name: 'Gateway Alpha',
    status: GatewayStatus.ACTIVE,
    interval: 60,
  };
  const mockSensors: Sensor[] = [
    {
      id: 's-1',
      gatewayId: 'gw-1',
      name: 'Temperature',
      profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
      status: SensorStatus.ACTIVE,
      dataInterval: 60,
    },
    {
      id: 's-2',
      gatewayId: 'gw-1',
      name: 'Humidity',
      profile: SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE,
      status: SensorStatus.INACTIVE,
      dataInterval: 120,
    },
  ];
  const mockChartRequest: ChartRequest = {
    sensor: mockSensors[0],
    chartType: ChartType.HISTORIC,
    timeInterval: null!,
  };

  const setInput = (key: string, value: unknown) => {
    fixture.componentRef.setInput(key, value);
    fixture.detectChanges();
  };
  const getTable = () => fixture.debugElement.query(By.directive(StubSensorTable));
  const getTableInstance = () => getTable().componentInstance as StubSensorTable;

  beforeEach(async () => {
    await TestBed.configureTestingModule({ imports: [GatewayExpandedComponent] })
      .overrideComponent(GatewayExpandedComponent, {
        remove: { imports: [SensorTableComponent] },
        add: { imports: [StubSensorTable] },
      })
      .compileComponents();

    fixture = TestBed.createComponent(GatewayExpandedComponent);
    component = fixture.componentInstance;
    fixture.componentRef.setInput('sensors', mockSensors);
    fixture.componentRef.setInput('gateway', mockGateway);
    fixture.componentRef.setInput('sensorTotal', mockSensors.length);
    fixture.detectChanges();
  });

  describe('initial state', () => {
    it('should create with correct inputs/defaults and verify fresh instance pagination defaults', () => {
      expect(component).toBeTruthy();
      expect(component.sensors()).toEqual(mockSensors);
      expect(component.gateway()).toEqual(mockGateway);
      expect(component.loading()).toBeUndefined();
      expect(component.actionMode()).toBe('dashboard');
      expect(component.sensorTotal()).toBe(2);
      expect(component.sensorPageIndex()).toBe(0);
      expect(component.sensorLimit()).toBe(10);

      const fresh = TestBed.createComponent(GatewayExpandedComponent);
      fresh.componentRef.setInput('sensors', []);
      fresh.componentRef.setInput('gateway', mockGateway);
      fresh.detectChanges();
      expect(fresh.componentInstance.sensorTotal()).toBe(0);
      expect(fresh.componentInstance.sensorPageIndex()).toBe(0);
      expect(fresh.componentInstance.sensorLimit()).toBe(10);
    });
  });

  describe('inputs', () => {
    it('should pass default inputs to child table', () => {
      const table = getTableInstance();
      expect(table.sensors()).toEqual(mockSensors);
      expect(table.loading()).toBeUndefined();
      expect(table.actionMode()).toBe('dashboard');
      expect(table.total()).toBe(2);
      expect(table.pageIndex()).toBe(0);
      expect(table.limit()).toBe(10);
    });

    it('should accept all inputs and reflect them on component and child', () => {
      setInput('loading', true);
      setInput('actionMode', 'manage');
      setInput('sensorTotal', 50);
      setInput('sensorPageIndex', 3);
      setInput('sensorLimit', 5);
      setInput('sensors', [mockSensors[0]]);
      setInput('gateway', { ...mockGateway, id: 'gw-99', name: 'New Gateway' });

      expect(component.loading()).toBe(true);
      expect(component.actionMode()).toBe('manage');
      expect(component.sensorTotal()).toBe(50);
      expect(component.sensorPageIndex()).toBe(3);
      expect(component.sensorLimit()).toBe(5);
      expect(component.sensors()).toEqual([mockSensors[0]]);
      expect(component.gateway()).toEqual({ ...mockGateway, id: 'gw-99', name: 'New Gateway' });

      const table = getTableInstance();
      expect(table.loading()).toBe(true);
      expect(table.actionMode()).toBe('manage');
      expect(table.total()).toBe(50);
      expect(table.pageIndex()).toBe(3);
      expect(table.limit()).toBe(5);
      expect(table.sensors()).toEqual([mockSensors[0]]);
    });
  });

  describe('template', () => {
    it('should render gateway id in heading, update on change, and render sensor table', () => {
      expect(fixture.debugElement.query(By.css('h2')).nativeElement.textContent).toContain('gw-1');
      expect(getTable()).toBeTruthy();

      setInput('gateway', { ...mockGateway, id: 'gw-99' });
      expect(fixture.debugElement.query(By.css('h2')).nativeElement.textContent).toContain('gw-99');
    });
  });

  describe('output events', () => {
    it('should forward chartRequested from sensor table', () => {
      const spy = vi.fn();
      component.chartRequested.subscribe(spy);
      getTable().triggerEventHandler('chartRequested', mockChartRequest);
      expect(spy).toHaveBeenCalledWith(mockChartRequest);
    });

    it('should emit sensorCreateRequested with current gateway on createRequested', () => {
      const spy = vi.fn();
      component.sensorCreateRequested.subscribe(spy);
      getTable().triggerEventHandler('createRequested');
      expect(spy).toHaveBeenCalledWith(mockGateway);

      const newGateway: Gateway = { ...mockGateway, id: 'gw-99', name: 'Updated' };
      setInput('gateway', newGateway);
      const spy2 = vi.fn();
      component.sensorCreateRequested.subscribe(spy2);
      getTable().triggerEventHandler('createRequested');
      expect(spy2).toHaveBeenCalledWith(newGateway);
    });

    it('should forward sensorDeleteRequested from sensor table', () => {
      const spy = vi.fn();
      component.sensorDeleteRequested.subscribe(spy);
      getTable().triggerEventHandler('deleteRequested', mockSensors[0]);
      expect(spy).toHaveBeenCalledWith(mockSensors[0]);
    });

    it('should forward sensorPageChange from sensor table', () => {
      const spy = vi.fn();
      component.sensorPageChange.subscribe(spy);
      const event: PageEvent = { pageIndex: 2, pageSize: 10, length: 50 };
      getTable().triggerEventHandler('pageChange', event);
      expect(spy).toHaveBeenCalledWith(event);
    });
  });
});
