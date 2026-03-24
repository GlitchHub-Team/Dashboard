import { TestBed } from '@angular/core/testing';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { of, throwError } from 'rxjs';

import { TenantService } from './tenant.service';
import { TenantApiClientService } from '../tenant-api-client/tenant-api-client.service';
import { Tenant } from '../../models/tenant.model';
import { RawTenantConfig } from '../../models/raw-tenant-config.model';

describe('TenantService', () => {
  let service: TenantService;

  const mockTenants: Tenant[] = [{ name: 'Tenant 1' }, { name: 'Tenant 2' }];
  const mockConfig: RawTenantConfig = { name: 'Tenant 3' };
  const newTenant: Tenant = { name: 'Tenant 3' };

  const tenantApiMock = {
    getTenant: vi.fn(),
    createTenant: vi.fn(),
    deleteTenant: vi.fn(),
  };

  beforeEach(() => {
    vi.resetAllMocks();
    TestBed.configureTestingModule({
      providers: [TenantService, { provide: TenantApiClientService, useValue: tenantApiMock }],
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

  describe('retrieveTenant', () => {
    it('should retrieve tenants and update state', () => {
      tenantApiMock.getTenant.mockReturnValue(of({ items: mockTenants, totalCount: 2 }));

      service.retrieveTenant();

      expect(tenantApiMock.getTenant).toHaveBeenCalledWith(0, 10);
      expect(service.loading()).toBe(false);
      expect(service.tenantList()).toEqual(mockTenants);
      expect(service.total()).toBe(2);
      expect(service.error()).toBeNull();
    });

    it.each([
      { error: new Error('Failed to fetch'), expected: 'Failed to fetch' },
      { error: { message: '' } as Error, expected: 'Failed to fetch tenants' },
    ])('should handle retrieval errors', ({ error, expected }) => {
      tenantApiMock.getTenant.mockReturnValue(throwError(() => error));

      service.retrieveTenant();

      expect(tenantApiMock.getTenant).toHaveBeenCalledWith(0, 10);
      expect(service.loading()).toBe(false);
      expect(service.tenantList()).toEqual([]);
      expect(service.error()).toBe(expected);
    });
  });

  describe('changePage', () => {
    it('should update pagination state and retrieve tenants', () => {
      tenantApiMock.getTenant.mockReturnValue(of({ items: mockTenants, totalCount: 2 }));

      service.changePage(2, 25);

      expect(service.pageIndex()).toBe(2);
      expect(service.limit()).toBe(25);
      expect(tenantApiMock.getTenant).toHaveBeenCalledWith(2, 25);
    });
  });

  describe('addNewTenant', () => {
    it('should create a tenant, append it to the list, and emit the new tenant', () => {
      service.tenantList.set(mockTenants);
      tenantApiMock.createTenant.mockReturnValue(of(newTenant));

      let result: Tenant | undefined;
      service.addNewTenant(mockConfig).subscribe((tenant) => {
        result = tenant;
      });

      expect(tenantApiMock.createTenant).toHaveBeenCalledWith(mockConfig);
      expect(result).toEqual(newTenant);
      expect(service.tenantList()).toEqual([...mockTenants, newTenant]);
      expect(service.loading()).toBe(false);
      expect(service.error()).toBeNull();
    });

    it.each([
      { error: new Error('Failed to create'), expected: 'Failed to create' },
      { error: { message: '' } as Error, expected: 'Failed to create tenant' },
    ])('should handle create errors', ({ error, expected }) => {
      tenantApiMock.createTenant.mockReturnValue(throwError(() => error));

      let thrownError: unknown;
      service.addNewTenant(mockConfig).subscribe({
        error: (err) => {
          thrownError = err;
        },
      });

      expect(tenantApiMock.createTenant).toHaveBeenCalledWith(mockConfig);
      expect(thrownError).toBe(error);
      expect(service.loading()).toBe(false);
      expect(service.error()).toBe(expected);
    });
  });

  describe('removeTenant', () => {
    it('should delete a tenant, remove it from the list, and complete successfully', () => {
      service.tenantList.set([...mockTenants, newTenant]);
      tenantApiMock.deleteTenant.mockReturnValue(of(void 0));

      let completed = false;
      service.removeTenant('Tenant 3').subscribe({
        complete: () => {
          completed = true;
        },
      });

      expect(tenantApiMock.deleteTenant).toHaveBeenCalledWith('Tenant 3');
      expect(service.tenantList()).toEqual(mockTenants);
      expect(service.loading()).toBe(false);
      expect(service.error()).toBeNull();
      expect(completed).toBe(true);
    });

    it.each([
      { error: new Error('Failed to delete'), expected: 'Failed to delete' },
      { error: { message: '' } as Error, expected: 'Failed to delete tenant' },
    ])('should handle delete errors', ({ error, expected }) => {
      service.tenantList.set(mockTenants);
      tenantApiMock.deleteTenant.mockReturnValue(throwError(() => error));

      let thrownError: unknown;
      service.removeTenant('Tenant 1').subscribe({
        error: (err) => {
          thrownError = err;
        },
      });

      expect(tenantApiMock.deleteTenant).toHaveBeenCalledWith('Tenant 1');
      expect(thrownError).toBe(error);
      expect(service.tenantList()).toEqual(mockTenants);
      expect(service.loading()).toBe(false);
      expect(service.error()).toBe(expected);
    });
  });
});
