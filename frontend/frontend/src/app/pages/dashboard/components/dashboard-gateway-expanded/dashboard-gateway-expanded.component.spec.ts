// dashboard-gateway-expanded.component.spec.ts

import { ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';

import { DashboardGatewayExpandedComponent } from './dashboard-gateway-expanded.component';
import { DashboardSensorTableComponent } from '../dashboard-sensor-table/dashboard-sensor-table.component';
import { Gateway } from '../../../../models/gateway/gateway.model';
import { GatewayStatus } from '../../../../models/gateway/gateway-status.enum';
import { Sensor } from '../../../../models/sensor/sensor.model';
import { SensorProfiles } from '../../../../models/sensor/sensor-profiles.enum';
import { ChartRequest } from '../../../../models/chart-request.model';
import { ChartType } from '../../../../models/chart-type.enum';

describe('DashboardGatewayExpandedComponent', () => {
  let component: DashboardGatewayExpandedComponent;
  let fixture: ComponentFixture<DashboardGatewayExpandedComponent>;

  // Adjust to match your actual models
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
    })
      .overrideComponent(DashboardGatewayExpandedComponent, {
        remove: { imports: [DashboardSensorTableComponent] },
        add: { schemas: [CUSTOM_ELEMENTS_SCHEMA] },
      })
      .compileComponents();

    fixture = TestBed.createComponent(DashboardGatewayExpandedComponent);
    component = fixture.componentInstance;

    // Required inputs
    fixture.componentRef.setInput('sensors', mockSensors);
    fixture.componentRef.setInput('gateway', mockGateway);
    fixture.detectChanges();
  });

  describe('initial state', () => {
    it('should create', () => {
      expect(component).toBeTruthy();
    });

    it('should default loading to undefined', () => {
      expect(component.loading()).toBeUndefined();
    });
  });

  describe('inputs', () => {
    it('should accept sensors', () => {
      expect(component.sensors()).toEqual(mockSensors);
    });

    it('should accept gateway', () => {
      expect(component.gateway()).toEqual(mockGateway);
    });

    it('should accept loading', () => {
      fixture.componentRef.setInput('loading', true);
      fixture.detectChanges();

      expect(component.loading()).toBe(true);
    });
  });

  describe('outputs', () => {
    it('should emit chartRequested', () => {
      const spy = vi.fn();
      component.chartRequested.subscribe(spy);

      const mockRequest: ChartRequest = {
        sensor: mockSensors[0],
        chartType: ChartType.HISTORIC,
        timeInterval: null!,
      };
      component.chartRequested.emit(mockRequest);

      expect(spy).toHaveBeenCalledWith(mockRequest);
    });
  });
});
