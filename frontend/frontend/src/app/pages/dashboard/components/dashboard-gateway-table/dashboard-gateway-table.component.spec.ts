import { ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';

import { DashboardGatewayTableComponent } from './dashboard-gateway-table.component';
import { DashboardGatewayExpandedComponent } from '../dashboard-gateway-expanded/dashboard-gateway-expanded.component';
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
    { id: 'gw-1', name: 'Gateway Alpha', status: GatewayStatus.ONLINE },
    { id: 'gw-2', name: 'Gateway Beta', status: GatewayStatus.OFFLINE },
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
      imports: [DashboardGatewayTableComponent],
    })
      .overrideComponent(DashboardGatewayTableComponent, {
        remove: { imports: [DashboardGatewayExpandedComponent] },
        add: { schemas: [CUSTOM_ELEMENTS_SCHEMA] },
      })
      .compileComponents();

    fixture = TestBed.createComponent(DashboardGatewayTableComponent);
    component = fixture.componentInstance;

    // Required inputs
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

  describe('displayedColumns', () => {
    it('should not include commands column when canSendCommands is false', () => {
      fixture.componentRef.setInput('canSendCommands', false);
      fixture.detectChanges();

      expect(component['displayedColumns']()).toEqual(['id', 'name', 'status']);
    });

    it('should include commands column when canSendCommands is true', () => {
      fixture.componentRef.setInput('canSendCommands', true);
      fixture.detectChanges();

      expect(component['displayedColumns']()).toEqual(['id', 'name', 'status', 'commands']);
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

  describe('outputs', () => {
    it('should emit commandRequested', () => {
      const spy = vi.fn();
      component.commandRequested.subscribe(spy);

      component.commandRequested.emit(mockGateways[0]);

      expect(spy).toHaveBeenCalledWith(mockGateways[0]);
    });

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

    it('should emit expandedGatewayChange', () => {
      const spy = vi.fn();
      component.expandedGatewayChange.subscribe(spy);

      component.expandedGatewayChange.emit(mockGateways[1]);

      expect(spy).toHaveBeenCalledWith(mockGateways[1]);
    });
  });
});
