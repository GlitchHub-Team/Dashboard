import { TestBed } from '@angular/core/testing';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { of, throwError } from 'rxjs';

import { TenantService } from './tenant.service';
import { TenantApiAdapter } from '../../adapters/tenant/tenant-api.adapter';
import { TenantApiClientService } from '../tenant-api-client/tenant-api-client.service';
import { PaginatedTenantResponse } from '../../models/tenant/paginated-tenant-response.model';
import { ApiError } from '../../models/api-error.model';
import { TenantBackend } from '../../models/tenant/tenant-backend.model';
import { Tenant } from '../../models/tenant/tenant.model';
import { TenantConfig } from '../../models/tenant/tenant-config.model';

describe('TenantService', () => {
  let service: TenantService;

  const tenantBackendList: TenantBackend[] = [
    { tenant_id: 'tenant-01', tenant_name: 'Tenant 1', can_impersonate: false },
    { tenant_id: 'tenant-02', tenant_name: 'Tenant 2', can_impersonate: true },
  ];

  const paginatedBackendResponse: PaginatedTenantResponse<TenantBackend> = {
    count: 2,
    total: 2,
    tenants: tenantBackendList,
  };

  const mappedTenants: Tenant[] = [
    { id: 'tenant-01', name: 'Tenant 1', canImpersonate: false },
    { id: 'tenant-02', name: 'Tenant 2', canImpersonate: true },
  ];

  const mappedPaginatedResponse = {
    count: 2,
    total: 2,
    tenants: mappedTenants,
  };

  const mockConfig: TenantConfig = { name: 'Tenant 3', canImpersonate: false };
  const createdTenantBackend: TenantBackend = {
    tenant_id: 'tenant-03',
    tenant_name: 'Tenant 3',
    can_impersonate: false,
  };
  const createdTenant: Tenant = {
    id: 'tenant-03',
    name: 'Tenant 3',
    canImpersonate: false,
  };

  const tenantApiMock = {
    getTenant: vi.fn(),
    getTenants: vi.fn(),
    createTenant: vi.fn(),
    deleteTenant: vi.fn(),
  };

  const tenantAdapterMock = {
    fromPaginatedDTO: vi.fn(),
    fromDTO: vi.fn(),
  };

  beforeEach(() => {
    vi.resetAllMocks();
    TestBed.configureTestingModule({
      providers: [
        TenantService,
        { provide: TenantApiClientService, useValue: tenantApiMock },
        { provide: TenantApiAdapter, useValue: tenantAdapterMock },
      ],
    });
    service = TestBed.inject(TenantService);
  });

  it('should be created with default state', () => {
    expect(service).toBeTruthy();
    expect(service.loading()).toBe(false);
    expect(service.error()).toBeNull();
    expect(service.tenantList()).toEqual([]);
    expect(service.total()).toBe(0);
    expect(service.pageIndex()).toBe(0);
    expect(service.limit()).toBe(10);
  });

  describe('getTenant', () => {
    const rawDto: TenantBackend = tenantBackendList[0];
    const adaptedTenant: Tenant = mappedTenants[0];

    it('should call API with correct id, adapt the DTO and return the tenant', () => {
      tenantApiMock.getTenant.mockReturnValue(of(rawDto));
      tenantAdapterMock.fromDTO.mockReturnValue(adaptedTenant);

      let result: Tenant | undefined;
      service.getTenant('tenant-01').subscribe((tenant) => {
        result = tenant;
      });

      expect(tenantApiMock.getTenant).toHaveBeenCalledWith('tenant-01');
      expect(tenantAdapterMock.fromDTO).toHaveBeenCalledWith(rawDto);
      expect(result).toEqual(adaptedTenant);
      expect(service.loading()).toBe(false);
      expect(service.error()).toBeNull();
    });

    it.each([
      { error: { status: 500, message: 'Server error' } as ApiError, expected: 'Server error' },
      { error: { status: 500 } as ApiError, expected: 'Failed to fetch tenant' },
    ])('should set error "$expected", reset loading and propagate error', ({ error, expected }) => {
      tenantApiMock.getTenant.mockReturnValue(throwError(() => error));

      let propagatedError: ApiError | undefined;
      service.getTenant('tenant-01').subscribe({ error: (err) => (propagatedError = err) });

      expect(service.error()).toBe(expected);
      expect(service.loading()).toBe(false);
      expect(propagatedError).toEqual(error);
    });
  });

  describe('retrieveTenants', () => {
    it('should retrieve tenants and update state', () => {
      tenantApiMock.getTenants.mockReturnValue(of(paginatedBackendResponse));
      tenantAdapterMock.fromPaginatedDTO.mockReturnValue(mappedPaginatedResponse);

      service.retrieveTenants();

      expect(tenantApiMock.getTenants).toHaveBeenCalledWith(1, 10);
      expect(tenantAdapterMock.fromPaginatedDTO).toHaveBeenCalledWith(paginatedBackendResponse);
      expect(service.loading()).toBe(false);
      expect(service.tenantList()).toEqual(mappedTenants);
      expect(service.total()).toBe(2);
      expect(service.error()).toBeNull();
    });

    it.each([
      {
        error: { status: 500, message: 'Failed to fetch' } as ApiError,
        expected: 'Failed to fetch',
      },
      { error: { status: 500 } as ApiError, expected: 'Failed to fetch tenants' },
    ])('should handle retrieval errors', ({ error, expected }) => {
      tenantApiMock.getTenants.mockReturnValue(throwError(() => error));

      service.retrieveTenants();

      expect(tenantApiMock.getTenants).toHaveBeenCalledWith(1, 10);
      expect(service.loading()).toBe(false);
      expect(service.tenantList()).toEqual([]);
      expect(service.error()).toBe(expected);
    });
  });

  describe('changePage', () => {
    it('should update pagination state and retrieve tenants', () => {
      tenantApiMock.getTenants.mockReturnValue(of(paginatedBackendResponse));
      tenantAdapterMock.fromPaginatedDTO.mockReturnValue(mappedPaginatedResponse);

      service.changePage(2, 25);

      expect(service.pageIndex()).toBe(2);
      expect(service.limit()).toBe(25);
      expect(tenantApiMock.getTenants).toHaveBeenCalledWith(3, 25);
    });
  });

  describe('addNewTenant', () => {
    it('should create a tenant and emit adapted tenant', () => {
      tenantApiMock.createTenant.mockReturnValue(of(createdTenantBackend));
      tenantAdapterMock.fromDTO.mockReturnValue(createdTenant);

      let result: Tenant | undefined;
      service.addNewTenant(mockConfig).subscribe((tenant) => {
        result = tenant;
      });

      expect(tenantApiMock.createTenant).toHaveBeenCalledWith(mockConfig);
      expect(tenantAdapterMock.fromDTO).toHaveBeenCalledWith(createdTenantBackend);
      expect(result).toEqual(createdTenant);
      expect(service.loading()).toBe(false);
      expect(service.error()).toBeNull();
    });
  });

  describe('removeTenant', () => {
    it('should delete a tenant and refetch current page', () => {
      tenantApiMock.deleteTenant.mockReturnValue(of(void 0));
      tenantApiMock.getTenants.mockReturnValue(of(paginatedBackendResponse));
      tenantAdapterMock.fromPaginatedDTO.mockReturnValue(mappedPaginatedResponse);

      let completed = false;
      service.removeTenant('tenant-03').subscribe({
        complete: () => {
          completed = true;
        },
      });

      expect(tenantApiMock.deleteTenant).toHaveBeenCalledWith('tenant-03');
      expect(tenantApiMock.getTenants).toHaveBeenCalledWith(1, 10);
      expect(service.tenantList()).toEqual(mappedTenants);
      expect(service.loading()).toBe(false);
      expect(service.error()).toBeNull();
      expect(completed).toBe(true);
    });

    it.each([
      {
        error: { status: 500, message: 'Failed to delete' } as ApiError,
        expected: 'Failed to delete',
      },
      { error: { status: 500 } as ApiError, expected: 'Failed to delete tenant' },
    ])('should handle delete errors', ({ error, expected }) => {
      tenantApiMock.deleteTenant.mockReturnValue(throwError(() => error));

      service.removeTenant('tenant-01').subscribe();

      expect(tenantApiMock.deleteTenant).toHaveBeenCalledWith('tenant-01');
      expect(service.tenantList()).toEqual([]);
      expect(service.loading()).toBe(false);
      expect(service.error()).toBe(expected);
    });
  });
});
