import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';

import { GatewayService } from './gateway.service';
import { GatewayApiClientService } from '../gateway-api-client/gateway-api-client.service';
import { Gateway } from '../../models/gateway.model';
import { GatewayConfig } from '../../models/gateway-config.model';
import { GatewayStatus } from '../../models/gateway-status.enum';
import { ApiError } from '../../models/api-error.model';

describe('GatewayService', () => {
  let service: GatewayService;

  const mockGateways: Gateway[] = [
    { id: 'gw-1', name: 'Gateway Alpha', tenantId: 'tenant-1', status: GatewayStatus.ONLINE },
    { id: 'gw-2', name: 'Gateway Beta', tenantId: 'tenant-1', status: GatewayStatus.OFFLINE },
  ];

  const mockNewGateway: Gateway = {
    id: 'gw-3',
    name: 'Gateway Gamma',
    tenantId: 'tenant-1',
    status: GatewayStatus.ONLINE,
  };

  const mockConfig: GatewayConfig = {
    name: 'Gateway Gamma',
  };

  const gatewayApiMock = {
    getGatewayListByTenant: vi.fn(),
    getGatewayList: vi.fn(),
    addNewGateway: vi.fn(),
    deleteGateway: vi.fn(),
  };

  beforeEach(() => {
    vi.resetAllMocks();

    TestBed.configureTestingModule({
      providers: [GatewayService, { provide: GatewayApiClientService, useValue: gatewayApiMock }],
    });

    service = TestBed.inject(GatewayService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  describe('initial state', () => {
    it('should have empty gateway list', () => {
      expect(service.gatewayList()).toEqual([]);
    });

    it('should not be loading', () => {
      expect(service.loading()).toBe(false);
    });

    it('should have no error', () => {
      expect(service.error()).toBeNull();
    });
  });

  describe('getGatewaysByTenant', () => {
    it('should call api with tenantId', () => {
      gatewayApiMock.getGatewayListByTenant.mockReturnValue(of(mockGateways));

      service.getGatewaysByTenant('tenant-1');

      expect(gatewayApiMock.getGatewayListByTenant).toHaveBeenCalledWith('tenant-1');
    });

    it('should populate gateway list on success', () => {
      gatewayApiMock.getGatewayListByTenant.mockReturnValue(of(mockGateways));

      service.getGatewaysByTenant('tenant-1');

      expect(service.gatewayList()).toEqual(mockGateways);
      expect(service.loading()).toBe(false);
      expect(service.error()).toBeNull();
    });

    it('should clear previous gateway list before fetching', () => {
      gatewayApiMock.getGatewayListByTenant.mockReturnValue(of(mockGateways));
      service.getGatewaysByTenant('tenant-1');

      gatewayApiMock.getGatewayListByTenant.mockReturnValue(of([]));
      service.getGatewaysByTenant('tenant-2');

      expect(service.gatewayList()).toEqual([]);
    });

    it('should set error on failure with message', () => {
      const apiError: ApiError = { status: 404, message: 'Tenant not found' };
      gatewayApiMock.getGatewayListByTenant.mockReturnValue(throwError(() => apiError));

      service.getGatewaysByTenant('tenant-1');

      expect(service.error()).toBe('Tenant not found');
      expect(service.loading()).toBe(false);
      expect(service.gatewayList()).toEqual([]);
    });

    it('should set default error message when error has no message', () => {
      const apiError: ApiError = { status: 500 } as ApiError;
      gatewayApiMock.getGatewayListByTenant.mockReturnValue(throwError(() => apiError));

      service.getGatewaysByTenant('tenant-1');

      expect(service.error()).toBe('Failed to load gateways');
    });

    it('should clear previous error before fetching', () => {
      const apiError: ApiError = { status: 500, message: 'Error' };
      gatewayApiMock.getGatewayListByTenant.mockReturnValue(throwError(() => apiError));
      service.getGatewaysByTenant('tenant-1');
      expect(service.error()).toBe('Error');

      gatewayApiMock.getGatewayListByTenant.mockReturnValue(of(mockGateways));
      service.getGatewaysByTenant('tenant-1');
      expect(service.error()).toBeNull();
    });
  });

  describe('getGateways', () => {
    it('should call api without params', () => {
      gatewayApiMock.getGatewayList.mockReturnValue(of(mockGateways));

      service.getGateways();

      expect(gatewayApiMock.getGatewayList).toHaveBeenCalled();
    });

    it('should populate gateway list on success', () => {
      gatewayApiMock.getGatewayList.mockReturnValue(of(mockGateways));

      service.getGateways();

      expect(service.gatewayList()).toEqual(mockGateways);
      expect(service.loading()).toBe(false);
      expect(service.error()).toBeNull();
    });

    it('should clear previous gateway list before fetching', () => {
      gatewayApiMock.getGatewayList.mockReturnValue(of(mockGateways));
      service.getGateways();

      gatewayApiMock.getGatewayList.mockReturnValue(of([]));
      service.getGateways();

      expect(service.gatewayList()).toEqual([]);
    });

    it('should set error on failure with message', () => {
      const apiError: ApiError = { status: 500, message: 'Server error' };
      gatewayApiMock.getGatewayList.mockReturnValue(throwError(() => apiError));

      service.getGateways();

      expect(service.error()).toBe('Server error');
      expect(service.loading()).toBe(false);
      expect(service.gatewayList()).toEqual([]);
    });

    it('should set default error message when error has no message', () => {
      const apiError: ApiError = { status: 500 } as ApiError;
      gatewayApiMock.getGatewayList.mockReturnValue(throwError(() => apiError));

      service.getGateways();

      expect(service.error()).toBe('Failed to load gateways');
    });

    it('should clear previous error before fetching', () => {
      const apiError: ApiError = { status: 500, message: 'Error' };
      gatewayApiMock.getGatewayList.mockReturnValue(throwError(() => apiError));
      service.getGateways();
      expect(service.error()).toBe('Error');

      gatewayApiMock.getGatewayList.mockReturnValue(of(mockGateways));
      service.getGateways();
      expect(service.error()).toBeNull();
    });
  });

  describe('addNewGateway', () => {
    it('should call api with gateway config', () => {
      gatewayApiMock.addNewGateway.mockReturnValue(of(mockNewGateway));

      service.addNewGateway(mockConfig).subscribe();

      expect(gatewayApiMock.addNewGateway).toHaveBeenCalledWith(mockConfig);
    });

    it('should append new gateway to list on success', () => {
      gatewayApiMock.getGatewayListByTenant.mockReturnValue(of(mockGateways));
      service.getGatewaysByTenant('tenant-1');

      gatewayApiMock.addNewGateway.mockReturnValue(of(mockNewGateway));
      service.addNewGateway(mockConfig).subscribe();

      expect(service.gatewayList()).toEqual([...mockGateways, mockNewGateway]);
      expect(service.loading()).toBe(false);
      expect(service.error()).toBeNull();
    });

    it('should set loading to false after success', () => {
      gatewayApiMock.addNewGateway.mockReturnValue(of(mockNewGateway));

      service.addNewGateway(mockConfig).subscribe();

      expect(service.loading()).toBe(false);
    });

    it('should set error on failure with message', () => {
      const apiError: ApiError = { status: 500, message: 'Duplicate gateway' };
      gatewayApiMock.addNewGateway.mockReturnValue(throwError(() => apiError));

      service.addNewGateway(mockConfig).subscribe();

      expect(service.error()).toBe('Duplicate gateway');
      expect(service.loading()).toBe(false);
    });

    it('should set default error message when error has no message', () => {
      const apiError: ApiError = { status: 500 } as ApiError;
      gatewayApiMock.addNewGateway.mockReturnValue(throwError(() => apiError));

      service.addNewGateway(mockConfig).subscribe();

      expect(service.error()).toBe('Failed to add gateway');
    });

    it('should clear previous error before adding', () => {
      const apiError: ApiError = { status: 500, message: 'Error' };
      gatewayApiMock.addNewGateway.mockReturnValue(throwError(() => apiError));
      service.addNewGateway(mockConfig).subscribe();
      expect(service.error()).toBe('Error');

      gatewayApiMock.addNewGateway.mockReturnValue(of(mockNewGateway));
      service.addNewGateway(mockConfig).subscribe();
      expect(service.error()).toBeNull();
    });

    it('should return the new gateway on success', () => {
      gatewayApiMock.addNewGateway.mockReturnValue(of(mockNewGateway));

      let result: Gateway | undefined;
      service.addNewGateway(mockConfig).subscribe((gateway) => {
        result = gateway;
      });

      expect(result).toEqual(mockNewGateway);
    });

    it('should complete without emitting on error', () => {
      const apiError: ApiError = { status: 500, message: 'Error' };
      gatewayApiMock.addNewGateway.mockReturnValue(throwError(() => apiError));

      const nextSpy = vi.fn();
      const errorSpy = vi.fn();
      const completeSpy = vi.fn();

      service.addNewGateway(mockConfig).subscribe({
        next: nextSpy,
        error: errorSpy,
        complete: completeSpy,
      });

      expect(nextSpy).not.toHaveBeenCalled();
      expect(errorSpy).not.toHaveBeenCalled();
      expect(completeSpy).toHaveBeenCalled();
    });
  });

  describe('deleteGateway', () => {
    beforeEach(() => {
      gatewayApiMock.getGatewayListByTenant.mockReturnValue(of(mockGateways));
      service.getGatewaysByTenant('tenant-1');
    });

    it('should call api with gateway id', () => {
      gatewayApiMock.deleteGateway.mockReturnValue(of(undefined));

      service.deleteGateway('gw-1').subscribe();

      expect(gatewayApiMock.deleteGateway).toHaveBeenCalledWith('gw-1');
    });

    it('should remove gateway from list on success', () => {
      gatewayApiMock.deleteGateway.mockReturnValue(of(undefined));

      service.deleteGateway('gw-1').subscribe();

      expect(service.gatewayList()).toEqual([mockGateways[1]]);
      expect(service.loading()).toBe(false);
      expect(service.error()).toBeNull();
    });

    it('should set loading to false after success', () => {
      gatewayApiMock.deleteGateway.mockReturnValue(of(undefined));

      service.deleteGateway('gw-1').subscribe();

      expect(service.loading()).toBe(false);
    });

    it('should set error on failure with message', () => {
      const apiError: ApiError = { status: 500, message: 'Gateway in use' };
      gatewayApiMock.deleteGateway.mockReturnValue(throwError(() => apiError));

      service.deleteGateway('gw-1').subscribe();

      expect(service.error()).toBe('Gateway in use');
      expect(service.loading()).toBe(false);
    });

    it('should set default error message when error has no message', () => {
      const apiError: ApiError = { status: 500 } as ApiError;
      gatewayApiMock.deleteGateway.mockReturnValue(throwError(() => apiError));

      service.deleteGateway('gw-1').subscribe();

      expect(service.error()).toBe('Failed to delete gateway');
    });

    it('should not remove gateway from list on failure', () => {
      const apiError: ApiError = { status: 500, message: 'Error' };
      gatewayApiMock.deleteGateway.mockReturnValue(throwError(() => apiError));

      service.deleteGateway('gw-1').subscribe();

      expect(service.gatewayList()).toEqual(mockGateways);
    });

    it('should clear previous error before deleting', () => {
      const apiError: ApiError = { status: 500, message: 'Error' };
      gatewayApiMock.deleteGateway.mockReturnValue(throwError(() => apiError));
      service.deleteGateway('gw-1').subscribe();
      expect(service.error()).toBe('Error');

      gatewayApiMock.deleteGateway.mockReturnValue(of(undefined));
      service.deleteGateway('gw-1').subscribe();
      expect(service.error()).toBeNull();
    });

    it('should complete without emitting on error', () => {
      const apiError: ApiError = { status: 500, message: 'Error' };
      gatewayApiMock.deleteGateway.mockReturnValue(throwError(() => apiError));

      const nextSpy = vi.fn();
      const errorSpy = vi.fn();
      const completeSpy = vi.fn();

      service.deleteGateway('gw-1').subscribe({
        next: nextSpy,
        error: errorSpy,
        complete: completeSpy,
      });

      expect(nextSpy).not.toHaveBeenCalled();
      expect(errorSpy).not.toHaveBeenCalled();
      expect(completeSpy).toHaveBeenCalled();
    });
  });
});
