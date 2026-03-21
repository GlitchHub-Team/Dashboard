import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';

import { GatewayService } from './gateway.service';
import { GatewayApiClientService } from '../gateway-api-client/gateway-api-client.service';
import { GatewayAdapter } from '../../adapters/gateway.adapter';
import { Gateway } from '../../models/gateway/gateway.model';
import { GatewayBackend } from '../../models/gateway/gateway-backend.model';
import { GatewayConfig } from '../../models/gateway/gateway-config.model';
import { PaginatedResponse } from '../../models/paginated-response.model';
import { Status } from '../../models/gateway-sensor-status.enum';

describe('GatewayService', () => {
  let service: GatewayService;

  const mockGateways: Gateway[] = [
    { id: 'gw-1', tenantId: 'tenant-1', name: 'Gateway 1', status: Status.ACTIVE, interval: 60 },
    { id: 'gw-2', tenantId: 'tenant-1', name: 'Gateway 2', status: Status.INACTIVE, interval: 60 },
  ];

  const mockBackendResponse: PaginatedResponse<GatewayBackend> = {
    count: 2,
    total: 10,
    data: [
      {
        gateway_id: 'gw-1',
        name: 'Gateway 1',
        tenant_id: 'tenant-1',
        status: 'active',
        intervals: 60,
      },
      {
        gateway_id: 'gw-2',
        name: 'Gateway 2',
        tenant_id: 'tenant-1',
        status: 'inactive',
        intervals: 60,
      },
    ],
  };

  const mockAdaptedResponse: PaginatedResponse<Gateway> = {
    count: 2,
    total: 10,
    data: mockGateways,
  };
  const emptyBackend: PaginatedResponse<GatewayBackend> = { count: 0, total: 0, data: [] };
  const emptyAdapted: PaginatedResponse<Gateway> = { count: 0, total: 0, data: [] };

  const mockNewBackend: GatewayBackend = {
    gateway_id: 'gw-3',
    name: 'New Gateway',
    tenant_id: 'tenant-1',
    status: 'active',
    intervals: 60,
  };
  const mockNewGateway: Gateway = {
    id: 'gw-3',
    name: 'New Gateway',
    tenantId: 'tenant-1',
    status: Status.ACTIVE,
    interval: 60,
  };
  const mockConfig: GatewayConfig = { name: 'New Gateway', interval: 60 };

  const gatewayApiMock = {
    getGatewayListByTenant: vi.fn(),
    getGatewayList: vi.fn(),
    addNewGateway: vi.fn(),
    deleteGateway: vi.fn(),
    sendCommandToGateway: vi.fn(),
  };

  const adapterMock = {
    fromPaginatedDTO: vi.fn(),
    fromDTO: vi.fn(),
  };

  function mockTenantSuccess(
    backendRes = mockBackendResponse,
    adaptedRes = mockAdaptedResponse,
  ): void {
    gatewayApiMock.getGatewayListByTenant.mockReturnValue(of(backendRes));
    adapterMock.fromPaginatedDTO.mockReturnValue(adaptedRes);
  }

  function mockListSuccess(
    backendRes = mockBackendResponse,
    adaptedRes = mockAdaptedResponse,
  ): void {
    gatewayApiMock.getGatewayList.mockReturnValue(of(backendRes));
    adapterMock.fromPaginatedDTO.mockReturnValue(adaptedRes);
  }

  beforeEach(() => {
    vi.resetAllMocks();
    TestBed.configureTestingModule({
      providers: [
        GatewayService,
        { provide: GatewayApiClientService, useValue: gatewayApiMock },
        { provide: GatewayAdapter, useValue: adapterMock },
      ],
    });
    service = TestBed.inject(GatewayService);
  });

  it('should be created with default state', () => {
    expect(service).toBeTruthy();
    expect(service.gatewayList()).toEqual([]);
    expect(service.loading()).toBe(false);
    expect(service.error()).toBeNull();
    expect(service.pageIndex()).toBe(0);
    expect(service.limit()).toBe(10);
    expect(service.total()).toBe(0);
  });

  describe('getGatewaysByTenant', () => {
    it('should call api and map through adapter with tenantId', () => {
      mockTenantSuccess();
      service.getGatewaysByTenant('tenant-1', 0, 10);
      expect(gatewayApiMock.getGatewayListByTenant).toHaveBeenCalledWith('tenant-1', 0, 10);
      expect(adapterMock.fromPaginatedDTO).toHaveBeenCalledWith(mockBackendResponse, 'tenant-1');
    });

    it('should set success state, update pagination, clear previous error, and reset list on refetch', () => {
      gatewayApiMock.getGatewayListByTenant.mockReturnValue(
        throwError(() => ({ status: 500, message: 'previous error' })),
      );
      service.getGatewaysByTenant('tenant-1', 0, 10);

      mockTenantSuccess();
      service.getGatewaysByTenant('tenant-1', 2, 25);
      expect(service.gatewayList()).toEqual(mockGateways);
      expect(service.total()).toBe(10);
      expect(service.loading()).toBe(false);
      expect(service.error()).toBeNull();
      expect(service.pageIndex()).toBe(2);
      expect(service.limit()).toBe(25);

      mockTenantSuccess(emptyBackend, emptyAdapted);
      service.getGatewaysByTenant('tenant-2', 0, 10);
      expect(service.gatewayList()).toEqual([]);
    });

    it.each([
      { error: { status: 500, message: 'Tenant not found' }, expected: 'Tenant not found' },
      { error: { status: 500 }, expected: 'Failed to load gateways' },
    ])('should set error "$expected" on failure', ({ error, expected }) => {
      gatewayApiMock.getGatewayListByTenant.mockReturnValue(throwError(() => error));
      service.getGatewaysByTenant('tenant-1', 0, 10);
      expect(service.error()).toBe(expected);
      expect(service.loading()).toBe(false);
      expect(service.gatewayList()).toEqual([]);
    });
  });

  describe('getGateways', () => {
    it('should call api and map through adapter without tenantId', () => {
      mockListSuccess();
      service.getGateways(0, 10);
      expect(gatewayApiMock.getGatewayList).toHaveBeenCalledWith(0, 10);
      expect(adapterMock.fromPaginatedDTO).toHaveBeenCalledWith(mockBackendResponse);
    });

    it('should set success state, update pagination, clear previous error, and reset list on refetch', () => {
      gatewayApiMock.getGatewayList.mockReturnValue(
        throwError(() => ({ status: 500, message: 'previous error' })),
      );
      service.getGateways(0, 10);

      mockListSuccess();
      service.getGateways(3, 50);
      expect(service.gatewayList()).toEqual(mockGateways);
      expect(service.total()).toBe(10);
      expect(service.loading()).toBe(false);
      expect(service.error()).toBeNull();
      expect(service.pageIndex()).toBe(3);
      expect(service.limit()).toBe(50);

      mockListSuccess(emptyBackend, emptyAdapted);
      service.getGateways(0, 10);
      expect(service.gatewayList()).toEqual([]);
    });

    it('should clear tenant context so changePage uses getGatewayList', () => {
      mockTenantSuccess();
      service.getGatewaysByTenant('tenant-1', 0, 10);

      mockListSuccess();
      service.getGateways(0, 10);

      gatewayApiMock.getGatewayList.mockClear();
      gatewayApiMock.getGatewayListByTenant.mockClear();
      mockListSuccess();

      service.changePage(1, 10);

      expect(gatewayApiMock.getGatewayList).toHaveBeenCalled();
      expect(gatewayApiMock.getGatewayListByTenant).not.toHaveBeenCalled();
    });

    it.each([
      { error: { status: 500, message: 'Server error' }, expected: 'Server error' },
      { error: { status: 500 }, expected: 'Failed to load gateways' },
    ])('should set error "$expected" on failure', ({ error, expected }) => {
      gatewayApiMock.getGatewayList.mockReturnValue(throwError(() => error));
      service.getGateways(0, 10);
      expect(service.error()).toBe(expected);
      expect(service.loading()).toBe(false);
      expect(service.gatewayList()).toEqual([]);
    });
  });

  describe('addNewGateway', () => {
    it('should call api, map through adapter, and return adapted gateway', () => {
      gatewayApiMock.addNewGateway.mockReturnValue(of(mockNewBackend));
      adapterMock.fromDTO.mockReturnValue(mockNewGateway);
      mockListSuccess();

      let result: Gateway | undefined;
      service.addNewGateway(mockConfig).subscribe((gw) => (result = gw));

      expect(gatewayApiMock.addNewGateway).toHaveBeenCalledWith(mockConfig);
      expect(adapterMock.fromDTO).toHaveBeenCalledWith(mockNewBackend, undefined);
      expect(result).toEqual(mockNewGateway);
    });

    it('should pass tenantId to adapter when in tenant context', () => {
      mockTenantSuccess();
      service.getGatewaysByTenant('tenant-1', 0, 10);

      gatewayApiMock.addNewGateway.mockReturnValue(of(mockNewBackend));
      adapterMock.fromDTO.mockReturnValue(mockNewGateway);
      mockTenantSuccess();

      service.addNewGateway(mockConfig).subscribe();

      expect(adapterMock.fromDTO).toHaveBeenCalledWith(mockNewBackend, 'tenant-1');
    });

    it('should refetch by tenant and set loading false after success', () => {
      mockTenantSuccess();
      service.getGatewaysByTenant('tenant-1', 0, 10);
      gatewayApiMock.getGatewayListByTenant.mockClear();
      mockTenantSuccess();

      gatewayApiMock.addNewGateway.mockReturnValue(of(mockNewBackend));
      adapterMock.fromDTO.mockReturnValue(mockNewGateway);

      service.addNewGateway(mockConfig).subscribe();

      expect(gatewayApiMock.getGatewayListByTenant).toHaveBeenCalledWith('tenant-1', 0, 10);
      expect(service.loading()).toBe(false);
    });

    it('should refetch via getGatewayList when no tenant context', () => {
      mockListSuccess();
      service.getGateways(0, 10);
      gatewayApiMock.getGatewayList.mockClear();
      mockListSuccess();

      gatewayApiMock.addNewGateway.mockReturnValue(of(mockNewBackend));
      adapterMock.fromDTO.mockReturnValue(mockNewGateway);

      service.addNewGateway(mockConfig).subscribe();

      expect(gatewayApiMock.getGatewayList).toHaveBeenCalledWith(0, 10);
    });

    it.each([
      { error: { status: 409, message: 'Duplicate gateway' }, expected: 'Duplicate gateway' },
      { error: { status: 500 }, expected: 'Failed to add gateway' },
    ])('should set error "$expected" on failure', ({ error, expected }) => {
      gatewayApiMock.addNewGateway.mockReturnValue(throwError(() => error));
      service.addNewGateway(mockConfig).subscribe();
      expect(service.error()).toBe(expected);
      expect(service.loading()).toBe(false);
    });

    it('should not refetch on error and clear previous error on retry', () => {
      mockTenantSuccess();
      service.getGatewaysByTenant('tenant-1', 0, 10);
      gatewayApiMock.getGatewayListByTenant.mockClear();

      gatewayApiMock.addNewGateway.mockReturnValue(
        throwError(() => ({ status: 500, message: 'Error' })),
      );
      service.addNewGateway(mockConfig).subscribe();
      expect(gatewayApiMock.getGatewayListByTenant).not.toHaveBeenCalled();
      expect(service.error()).toBe('Error');

      gatewayApiMock.addNewGateway.mockReturnValue(of(mockNewBackend));
      adapterMock.fromDTO.mockReturnValue(mockNewGateway);
      mockTenantSuccess();
      service.addNewGateway(mockConfig).subscribe();
      expect(service.error()).toBeNull();
    });

    it('should complete without emitting on error', () => {
      gatewayApiMock.addNewGateway.mockReturnValue(
        throwError(() => ({ status: 500, message: 'Error' })),
      );
      const nextSpy = vi.fn();
      const errorSpy = vi.fn();
      const completeSpy = vi.fn();

      service
        .addNewGateway(mockConfig)
        .subscribe({ next: nextSpy, error: errorSpy, complete: completeSpy });

      expect(nextSpy).not.toHaveBeenCalled();
      expect(errorSpy).not.toHaveBeenCalled();
      expect(completeSpy).toHaveBeenCalled();
    });
  });

  describe('deleteGateway', () => {
    beforeEach(() => {
      mockTenantSuccess();
      service.getGatewaysByTenant('tenant-1', 0, 10);
    });

    it('should call api, refetch current page, and set loading false after success', () => {
      gatewayApiMock.getGatewayListByTenant.mockClear();
      gatewayApiMock.deleteGateway.mockReturnValue(of(undefined));
      mockTenantSuccess();

      service.deleteGateway('gw-1').subscribe();

      expect(gatewayApiMock.deleteGateway).toHaveBeenCalledWith('gw-1');
      expect(gatewayApiMock.getGatewayListByTenant).toHaveBeenCalledWith('tenant-1', 0, 10);
      expect(service.loading()).toBe(false);
    });

    it.each([
      { error: { status: 500, message: 'Gateway in use' }, expected: 'Gateway in use' },
      { error: { status: 500 }, expected: 'Failed to delete gateway' },
    ])('should set error "$expected" on failure', ({ error, expected }) => {
      gatewayApiMock.deleteGateway.mockReturnValue(throwError(() => error));
      service.deleteGateway('gw-1').subscribe();
      expect(service.error()).toBe(expected);
      expect(service.loading()).toBe(false);
    });

    it('should not refetch on failure and clear previous error on retry', () => {
      gatewayApiMock.getGatewayListByTenant.mockClear();
      gatewayApiMock.deleteGateway.mockReturnValue(
        throwError(() => ({ status: 500, message: 'Error' })),
      );
      service.deleteGateway('gw-1').subscribe();
      expect(gatewayApiMock.getGatewayListByTenant).not.toHaveBeenCalled();
      expect(service.error()).toBe('Error');

      gatewayApiMock.deleteGateway.mockReturnValue(of(undefined));
      mockTenantSuccess();
      service.deleteGateway('gw-1').subscribe();
      expect(service.error()).toBeNull();
    });

    it('should complete without emitting on error', () => {
      gatewayApiMock.deleteGateway.mockReturnValue(
        throwError(() => ({ status: 500, message: 'Error' })),
      );
      const nextSpy = vi.fn();
      const errorSpy = vi.fn();
      const completeSpy = vi.fn();

      service
        .deleteGateway('gw-1')
        .subscribe({ next: nextSpy, error: errorSpy, complete: completeSpy });

      expect(nextSpy).not.toHaveBeenCalled();
      expect(errorSpy).not.toHaveBeenCalled();
      expect(completeSpy).toHaveBeenCalled();
    });
  });

  describe('changePage', () => {
    it('should refetch by tenant when tenant context is active', () => {
      mockTenantSuccess();
      service.getGatewaysByTenant('tenant-1', 0, 10);
      gatewayApiMock.getGatewayListByTenant.mockClear();
      mockTenantSuccess();

      service.changePage(2, 20);

      expect(gatewayApiMock.getGatewayListByTenant).toHaveBeenCalledWith('tenant-1', 2, 20);
    });

    it('should refetch all gateways when no tenant context is set', () => {
      mockListSuccess();
      service.getGateways(0, 10);
      gatewayApiMock.getGatewayList.mockClear();
      mockListSuccess();

      service.changePage(3, 15);

      expect(gatewayApiMock.getGatewayList).toHaveBeenCalledWith(3, 15);
    });

    it('should call getGateways by default when never fetched before', () => {
      mockListSuccess();

      service.changePage(1, 10);

      expect(gatewayApiMock.getGatewayList).toHaveBeenCalledWith(1, 10);
      expect(gatewayApiMock.getGatewayListByTenant).not.toHaveBeenCalled();
    });
  });
});
