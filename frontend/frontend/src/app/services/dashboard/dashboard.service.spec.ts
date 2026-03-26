import { TestBed } from '@angular/core/testing';
import { signal } from '@angular/core';

import { DashboardService } from './dashboard.service';
import { GatewayService } from '../gateway/gateway.service';
import { SensorService } from '../sensor/sensor.service';
import { PermissionService } from '../permission/permission.service';
import { Gateway } from '../../models/gateway/gateway.model';
import { Sensor } from '../../models/sensor/sensor.model';
import { ChartRequest } from '../../models/chart/chart-request.model';
import { Permission } from '../../models/permission.enum';
import { ChartType } from '../../models/chart/chart-type.enum';
import { SensorProfiles } from '../../models/sensor/sensor-profiles.enum';
import { Status } from '../../models/gateway-sensor-status.enum';

describe('DashboardService', () => {
  let service: DashboardService;

  const mockGateway: Gateway = {
    id: 'gw-1',
    name: 'Gateway 1',
    status: Status.ACTIVE,
    interval: 60,
  };
  const mockGateway2: Gateway = {
    id: 'gw-2',
    name: 'Gateway 2',
    status: Status.INACTIVE,
    interval: 120,
  };
  const mockSensor: Sensor = {
    id: 's-1',
    gatewayId: 'gw-1',
    name: 'Temp',
    profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
    status: Status.ACTIVE,
    dataInterval: 60,
  };
  const mockChartRequest: ChartRequest = { sensor: mockSensor, chartType: ChartType.HISTORIC };

  const _gwList = signal<Gateway[]>([mockGateway]);
  const _gwTotal = signal(5);
  const _gwPageIndex = signal(1);
  const _gwLimit = signal(20);
  const _gwLoading = signal(true);
  const _gwError = signal<string | null>('gw-error');

  const _sList = signal<Sensor[]>([mockSensor]);
  const _sTotal = signal(3);
  const _sPageIndex = signal(2);
  const _sLimit = signal(15);
  const _sLoading = signal(true);
  const _sError = signal<string | null>('s-error');

  const gatewayServiceMock = {
    gatewayList: _gwList.asReadonly(),
    total: _gwTotal.asReadonly(),
    pageIndex: _gwPageIndex.asReadonly(),
    limit: _gwLimit.asReadonly(),
    loading: _gwLoading.asReadonly(),
    error: _gwError.asReadonly(),
    getGatewaysByTenant: vi.fn(),
    changePage: vi.fn(),
  };

  const sensorServiceMock = {
    sensorList: _sList.asReadonly(),
    total: _sTotal.asReadonly(),
    pageIndex: _sPageIndex.asReadonly(),
    limit: _sLimit.asReadonly(),
    loading: _sLoading.asReadonly(),
    error: _sError.asReadonly(),
    getSensorsByTenant: vi.fn(),
    getSensorsByGateway: vi.fn(),
    changePage: vi.fn(),
    clearSensors: vi.fn(),
  };

  const permissionServiceMock = { can: vi.fn() };

  beforeEach(() => {
    vi.resetAllMocks();
    permissionServiceMock.can.mockReturnValue(true);

    TestBed.configureTestingModule({
      providers: [
        DashboardService,
        { provide: GatewayService, useValue: gatewayServiceMock },
        { provide: SensorService, useValue: sensorServiceMock },
        { provide: PermissionService, useValue: permissionServiceMock },
      ],
    });

    service = TestBed.inject(DashboardService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('should have null expandedGateway and selectedChart initially', () => {
    expect(service.expandedGateway()).toBeNull();
    expect(service.selectedChart()).toBeNull();
  });

  it('should forward all gateway service signals', () => {
    expect(service.gatewayList()).toEqual([mockGateway]);
    expect(service.gatewayTotal()).toBe(5);
    expect(service.gatewayPageIndex()).toBe(1);
    expect(service.gatewayLimit()).toBe(20);
    expect(service.gatewayLoading()).toBe(true);
    expect(service.gatewayError()).toBe('gw-error');
  });

  it('should forward all sensor service signals', () => {
    expect(service.sensorList()).toEqual([mockSensor]);
    expect(service.sensorTotal()).toBe(3);
    expect(service.sensorPageIndex()).toBe(2);
    expect(service.sensorLimit()).toBe(15);
    expect(service.sensorLoading()).toBe(true);
    expect(service.sensorError()).toBe('s-error');
  });

  describe('canSendCommands', () => {
    it('should return true when permission is granted', () => {
      permissionServiceMock.can.mockReturnValue(true);
      expect(service.canSendCommands()).toBe(true);
      expect(permissionServiceMock.can).toHaveBeenCalledWith(Permission.GATEWAY_COMMANDS);
    });

    it('should return false when permission is denied', () => {
      permissionServiceMock.can.mockReturnValue(false);
      expect(service.canSendCommands()).toBe(false);
    });
  });

  describe('loadDashboard', () => {
    it('should do nothing when no tenantId is provided', () => {
      service.loadDashboard();

      expect(gatewayServiceMock.getGatewaysByTenant).not.toHaveBeenCalled();
      expect(sensorServiceMock.getSensorsByTenant).not.toHaveBeenCalled();
    });

    it('should call getGatewaysByTenant when canSendCommands is true', () => {
      permissionServiceMock.can.mockReturnValue(true);
      service.loadDashboard('tenant-01');

      expect(gatewayServiceMock.getGatewaysByTenant).toHaveBeenCalledWith('tenant-01', 0, 10);
      expect(sensorServiceMock.getSensorsByTenant).not.toHaveBeenCalled();
    });

    it('should call getSensorsByTenant when canSendCommands is false', () => {
      permissionServiceMock.can.mockReturnValue(false);
      service.loadDashboard('tenant-01');

      expect(sensorServiceMock.getSensorsByTenant).toHaveBeenCalledWith('tenant-01', 0, 10);
      expect(gatewayServiceMock.getGatewaysByTenant).not.toHaveBeenCalled();
    });
  });

  describe('changeGatewayPage', () => {
    it('should collapse gateway, clear sensors, and delegate to gateway service', () => {
      service.toggleExpandedGateway(mockGateway);
      expect(service.expandedGateway()).toEqual(mockGateway);

      service.changeGatewayPage(2, 25);

      expect(service.expandedGateway()).toBeNull();
      expect(sensorServiceMock.clearSensors).toHaveBeenCalled();
      expect(gatewayServiceMock.changePage).toHaveBeenCalledWith(2, 25);
    });
  });

  describe('changeSensorPage', () => {
    it('should delegate to sensor service', () => {
      service.changeSensorPage(3, 15);

      expect(sensorServiceMock.changePage).toHaveBeenCalledWith(3, 15);
    });
  });

  describe('toggleExpandedGateway', () => {
    it('should set expandedGateway and fetch sensors when none is expanded', () => {
      service.toggleExpandedGateway(mockGateway);

      expect(service.expandedGateway()).toEqual(mockGateway);
      expect(sensorServiceMock.getSensorsByGateway).toHaveBeenCalledWith('gw-1', 2, 15);
    });

    it('should collapse gateway and clear sensors when the same gateway is toggled again', () => {
      service.toggleExpandedGateway(mockGateway);
      sensorServiceMock.clearSensors.mockClear();

      service.toggleExpandedGateway(mockGateway);

      expect(service.expandedGateway()).toBeNull();
      expect(sensorServiceMock.clearSensors).toHaveBeenCalled();
    });

    it('should expand the new gateway when a different one is toggled', () => {
      service.toggleExpandedGateway(mockGateway);
      sensorServiceMock.getSensorsByGateway.mockClear();

      service.toggleExpandedGateway(mockGateway2);

      expect(service.expandedGateway()).toEqual(mockGateway2);
      expect(sensorServiceMock.getSensorsByGateway).toHaveBeenCalledWith('gw-2', 2, 15);
    });
  });

  describe('openChart / closeChart', () => {
    it('should set selectedChart on openChart and clear it on closeChart', () => {
      service.openChart(mockChartRequest);
      expect(service.selectedChart()).toEqual(mockChartRequest);

      service.closeChart();
      expect(service.selectedChart()).toBeNull();
    });
  });
});
