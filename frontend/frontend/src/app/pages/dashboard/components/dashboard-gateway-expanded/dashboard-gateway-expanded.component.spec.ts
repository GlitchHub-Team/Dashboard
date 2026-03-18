import { ComponentFixture, TestBed } from '@angular/core/testing';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { PageEvent } from '@angular/material/paginator';

import { DashboardGatewayExpandedComponent } from './dashboard-gateway-expanded.component';
import { Gateway } from '../../../../models/gateway/gateway.model';
import { GatewayStatus } from '../../../../models/gateway/gateway-status.enum';
import { Sensor } from '../../../../models/sensor/sensor.model';
import { SensorProfiles } from '../../../../models/sensor/sensor-profiles.enum';
import { ChartRequest } from '../../../../models/chart-request.model';
import { ChartType } from '../../../../models/chart-type.enum';

describe('DashboardGatewayExpandedComponent', () => {
  let component: DashboardGatewayExpandedComponent;
  let fixture: ComponentFixture<DashboardGatewayExpandedComponent>;

  const mockGateway: Gateway = {
    id: 'gw-1',
    tenantId: 'tenant-1',
    name: 'Gateway Alpha',
    status: GatewayStatus.ONLINE,
  };

  const mockSensors: Sensor[] = [
    {
      id: 's-1',
      gatewayId: 'gw-1',
      name: 'Temperature',
      profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
    },
    {
      id: 's-2',
      gatewayId: 'gw-1',
      name: 'Humidity',
      profile: SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE,
    },
  ];

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [DashboardGatewayExpandedComponent],
      schemas: [NO_ERRORS_SCHEMA],
    }).compileComponents();

    fixture = TestBed.createComponent(DashboardGatewayExpandedComponent);
    component = fixture.componentInstance;

    fixture.componentRef.setInput('sensors', mockSensors);
    fixture.componentRef.setInput('gateway', mockGateway);
    fixture.componentRef.setInput('sensorTotal', mockSensors.length);
    fixture.detectChanges();
  });

  describe('initial state', () => {
    it('should create with correct defaults', () => {
      expect(component).toBeTruthy();
      expect(component.loading()).toBeUndefined();
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

  describe('inputs', () => {
    it('should accept all standard inputs', () => {
      expect(component.sensors()).toEqual(mockSensors);
      expect(component.gateway()).toEqual(mockGateway);
    });

    it('should accept loading', () => {
      fixture.componentRef.setInput('loading', true);
      fixture.detectChanges();

      expect(component.loading()).toBe(true);
    });

    it('should accept all pagination inputs', () => {
      fixture.componentRef.setInput('sensorTotal', 50);
      fixture.componentRef.setInput('sensorPageIndex', 3);
      fixture.componentRef.setInput('sensorLimit', 5);
      fixture.detectChanges();

      expect(component.sensorTotal()).toBe(50);
      expect(component.sensorPageIndex()).toBe(3);
      expect(component.sensorLimit()).toBe(5);
    });
  });

  describe('outputs', () => {
    it('should emit chartRequested', () => {
      const spy = vi.fn();
      component.chartRequested.subscribe(spy);

      const request: ChartRequest = {
        sensor: mockSensors[0],
        chartType: ChartType.HISTORIC,
        timeInterval: null!,
      };
      component.chartRequested.emit(request);

      expect(spy).toHaveBeenCalledWith(request);
    });

    it('should emit sensorPageChange', () => {
      const spy = vi.fn();
      component.sensorPageChange.subscribe(spy);

      const event: PageEvent = { pageIndex: 2, pageSize: 10, length: 50 };
      component.sensorPageChange.emit(event);

      expect(spy).toHaveBeenCalledWith(event);
    });
  });

  describe('template', () => {
    it('should render gateway id in heading', () => {
      const heading = fixture.nativeElement.querySelector('h3');
      expect(heading.textContent).toContain('gw-1');
    });

    it('should render sensor table', () => {
      const sensorTable = fixture.nativeElement.querySelector('app-dashboard-sensor-table');
      expect(sensorTable).toBeTruthy();
    });
  });
});
