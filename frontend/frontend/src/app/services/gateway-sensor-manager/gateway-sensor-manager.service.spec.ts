import { describe, it, expect, vi, beforeEach } from 'vitest';
import { TestBed } from '@angular/core/testing';
import { of, EMPTY } from 'rxjs';

import { GatewaySensorManagerService } from './gateway-sensor-manager.service';
import { GatewayService } from '../gateway/gateway.service';
import { SensorService } from '../sensor/sensor.service';
import { Gateway } from '../../models/gateway/gateway.model';
import { Status } from '../../models/gateway-sensor-status.enum';
import { Sensor } from '../../models/sensor/sensor.model';
import { SensorProfiles } from '../../models/sensor/sensor-profiles.enum';
import { signal } from '@angular/core';

describe('GatewaySensorManagerService', () => {
  let service: GatewaySensorManagerService;

  let gatewayPageIndexSig = signal(0);
  let gatewayLimitSig = signal(10);
  let sensorPageIndexSig = signal(0);
  let sensorLimitSig = signal(10);

  let gatewayServiceMock: any;
  let sensorServiceMock: any;

  const mockGateway: Gateway = {
    id: 'gw-1',
    tenantId: 'tenant-01',
    name: 'Gateway 1',
    status: Status.ACTIVE,
    interval: 60,
  };

  const mockGateway2: Gateway = {
    id: 'gw-2',
    tenantId: 'tenant-01',
    name: 'Gateway 2',
    status: Status.INACTIVE,
    interval: 120,
  };

  const mockSensor: Sensor = {
    id: 'sensor-1',
    gatewayId: 'gw-1',
    name: 'Heart Rate Sensor',
    status: Status.ACTIVE,
    profile: SensorProfiles.HEART_RATE_SERVICE,
    dataInterval: 1000,
  };

  beforeEach(() => {
    gatewayPageIndexSig = signal(0);
    gatewayLimitSig = signal(10);
    sensorPageIndexSig = signal(0);
    sensorLimitSig = signal(10);

    gatewayServiceMock = {
      gatewayList: signal<Gateway[]>([]),
      total: signal(0),
      pageIndex: gatewayPageIndexSig,
      limit: gatewayLimitSig,
      loading: signal(false),
      error: signal<string | null>(null),
      getGateways: vi.fn(),
      getGatewaysByTenant: vi.fn(),
      changePage: vi.fn(),
      deleteGateway: vi.fn().mockReturnValue(of(undefined)),
    };

    sensorServiceMock = {
      sensorList: signal<Sensor[]>([]),
      total: signal(0),
      pageIndex: sensorPageIndexSig,
      limit: sensorLimitSig,
      loading: signal(false),
      error: signal<string | null>(null),
      getSensorsByGateway: vi.fn(),
      getSensorsByTenant: vi.fn(),
      changePage: vi.fn(),
      clearSensors: vi.fn(),
      deleteSensor: vi.fn().mockReturnValue(of(undefined)),
    };

    TestBed.configureTestingModule({
      providers: [
        GatewaySensorManagerService,
        { provide: GatewayService, useValue: gatewayServiceMock },
        { provide: SensorService, useValue: sensorServiceMock },
      ],
    });

    service = TestBed.inject(GatewaySensorManagerService);
  });

  describe('signal proxies', () => {
    it.each([
      ['gatewayList', () => gatewayServiceMock.gatewayList],
      ['gatewayTotal', () => gatewayServiceMock.total],
      ['gatewayPageIndex', () => gatewayServiceMock.pageIndex],
      ['gatewayLimit', () => gatewayServiceMock.limit],
      ['gatewayLoading', () => gatewayServiceMock.loading],
      ['gatewayError', () => gatewayServiceMock.error],
      ['sensorList', () => sensorServiceMock.sensorList],
      ['sensorTotal', () => sensorServiceMock.total],
      ['sensorPageIndex', () => sensorServiceMock.pageIndex],
      ['sensorLimit', () => sensorServiceMock.limit],
      ['sensorLoading', () => sensorServiceMock.loading],
      ['sensorError', () => sensorServiceMock.error],
    ] as [string, () => any][])('should expose %s from underlying service', (prop, getter) => {
      expect((service as any)[prop]).toBe(getter());
    });
  });

  it('should initialize expandedGateway as null', () => {
    expect(service.expandedGateway()).toBeNull();
  });

  describe('loadGateways', () => {
    it.each([
      [0, 10],
      [3, 25],
    ])('should call getGateways with page %i and limit %i', (page, limit) => {
      gatewayPageIndexSig.set(page);
      gatewayLimitSig.set(limit);
      service.loadGateways();
      expect(gatewayServiceMock.getGateways).toHaveBeenCalledWith(page, limit);
    });
  });

  describe('toggleExpandedGateway', () => {
    it.each([
      [0, 10],
      [2, 5],
    ])('should expand gateway and load sensors with page %i and limit %i', (page, limit) => {
      sensorPageIndexSig.set(page);
      sensorLimitSig.set(limit);

      service.toggleExpandedGateway(mockGateway);

      expect(service.expandedGateway()).toEqual(mockGateway);
      expect(sensorServiceMock.getSensorsByGateway).toHaveBeenCalledWith('gw-1', page, limit);
    });

    it('should collapse gateway and clear sensors when toggling the same gateway', () => {
      service.toggleExpandedGateway(mockGateway);
      sensorServiceMock.clearSensors.mockClear();

      service.toggleExpandedGateway(mockGateway);

      expect(service.expandedGateway()).toBeNull();
      expect(sensorServiceMock.clearSensors).toHaveBeenCalledOnce();
    });

    it('should switch to a different gateway', () => {
      service.toggleExpandedGateway(mockGateway);
      service.toggleExpandedGateway(mockGateway2);

      expect(service.expandedGateway()).toEqual(mockGateway2);
      expect(sensorServiceMock.getSensorsByGateway).toHaveBeenLastCalledWith('gw-2', 0, 10);
    });
  });

  describe('changeGatewayPage', () => {
    it('should collapse expanded gateway and clear sensors', () => {
      service.toggleExpandedGateway(mockGateway);
      sensorServiceMock.clearSensors.mockClear();

      service.changeGatewayPage(1, 10);

      expect(service.expandedGateway()).toBeNull();
      expect(sensorServiceMock.clearSensors).toHaveBeenCalledOnce();
    });

    it('should call gatewayService.changePage', () => {
      service.changeGatewayPage(2, 25);
      expect(gatewayServiceMock.changePage).toHaveBeenCalledWith(2, 25);
    });
  });

  it('should call sensorService.changePage', () => {
    service.changeSensorPage(3, 15);
    expect(sensorServiceMock.changePage).toHaveBeenCalledWith(3, 15);
  });

  describe('refreshGateways', () => {
    it.each([
      [0, 10],
      [2, 25],
    ])('should call gatewayService.changePage with page %i and limit %i', (page, limit) => {
      gatewayPageIndexSig.set(page);
      gatewayLimitSig.set(limit);
      service.refreshGateways();
      expect(gatewayServiceMock.changePage).toHaveBeenCalledWith(page, limit);
    });
  });

  describe('refreshSensors', () => {
    it.each([
      [0, 10],
      [1, 5],
    ])('should call sensorService.getSensorsByGateway with page %i and limit %i', (page, limit) => {
      sensorPageIndexSig.set(page);
      sensorLimitSig.set(limit);
      service.refreshSensors('gw-1');
      expect(sensorServiceMock.getSensorsByGateway).toHaveBeenCalledWith('gw-1', page, limit);
    });
  });

  describe('deleteGateway', () => {
    it('should call gatewayService.deleteGateway with gateway id', () => {
      service.deleteGateway(mockGateway);
      expect(gatewayServiceMock.deleteGateway).toHaveBeenCalledWith('gw-1');
    });

    it.each([
      [0, 10],
      [3, 20],
    ])('should refresh gateways at page %i with limit %i after deletion', (page, limit) => {
      gatewayPageIndexSig.set(page);
      gatewayLimitSig.set(limit);
      service.deleteGateway(mockGateway);
      expect(gatewayServiceMock.changePage).toHaveBeenCalledWith(page, limit);
    });

    it('should collapse and clear sensors if deleted gateway was expanded', () => {
      service.toggleExpandedGateway(mockGateway);
      sensorServiceMock.clearSensors.mockClear();

      service.deleteGateway(mockGateway);

      expect(service.expandedGateway()).toBeNull();
      expect(sensorServiceMock.clearSensors).toHaveBeenCalled();
    });

    it.each([
      ['different gateway expanded', () => service.toggleExpandedGateway(mockGateway2)],
      ['no gateway expanded', () => {}],
    ] as [string, () => void][])('should not collapse when %s', (_, setup) => {
      setup();
      const expandedBefore = service.expandedGateway();
      sensorServiceMock.clearSensors.mockClear();

      service.deleteGateway(mockGateway);

      expect(service.expandedGateway()).toEqual(expandedBefore);
      expect(sensorServiceMock.clearSensors).not.toHaveBeenCalled();
    });

    it('should not refresh if deleteGateway returns EMPTY', () => {
      gatewayServiceMock.deleteGateway.mockReturnValue(EMPTY);
      service.deleteGateway(mockGateway);
      expect(gatewayServiceMock.changePage).not.toHaveBeenCalled();
    });
  });

  describe('deleteSensor', () => {
    it('should call sensorService.deleteSensor with sensor id', () => {
      service.deleteSensor(mockSensor);
      expect(sensorServiceMock.deleteSensor).toHaveBeenCalledWith('sensor-1');
    });

    it.each([
      [0, 10],
      [2, 15],
    ])('should refresh sensors at page %i with limit %i after deletion', (page, limit) => {
      sensorPageIndexSig.set(page);
      sensorLimitSig.set(limit);
      service.deleteSensor(mockSensor);
      expect(sensorServiceMock.getSensorsByGateway).toHaveBeenCalledWith('gw-1', page, limit);
    });

    it('should not refresh if deleteSensor returns EMPTY', () => {
      sensorServiceMock.deleteSensor.mockReturnValue(EMPTY);
      service.deleteSensor(mockSensor);
      expect(sensorServiceMock.getSensorsByGateway).not.toHaveBeenCalled();
    });
  });
});
