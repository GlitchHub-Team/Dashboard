import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';

import { TenantApiClientService } from './tenant-api-client.service';
import { TenantApiAdapter } from '../../adapters/tenant/tenant-api.adapter';
import { environment } from '../../../environments/environment';
import { PaginatedTenantResponse } from '../../models/tenant/paginated-tenant-response.model';
import { TenantBackend } from '../../models/tenant/tenant-backend.model';
import { TenantConfig } from '../../models/tenant/tenant-config.model';
import { Tenant } from '../../models/tenant/tenant.model';

describe('TenantApiClientService', () => {
  let service: TenantApiClientService;
  let httpMock: HttpTestingController;

  const apiUrl = `${environment.apiUrl}`;

  const mockBackendTenants: TenantBackend[] = [
    { tenant_id: 'tenant-01', tenant_name: 'Tenant 1', can_impersonate: false },
    { tenant_id: 'tenant-02', tenant_name: 'Tenant 2', can_impersonate: true },
  ];

  const mockBackendPaginatedResponse: PaginatedTenantResponse<TenantBackend> = {
    count: 2,
    total: 2,
    tenants: mockBackendTenants,
  };

  const mockMappedTenants: Tenant[] = [
    { id: 'tenant-01', name: 'Tenant 1', canImpersonate: false },
    { id: 'tenant-02', name: 'Tenant 2', canImpersonate: true },
  ];

  const mockMappedPaginatedResponse: PaginatedTenantResponse<Tenant> = {
    count: 2,
    total: 2,
    tenants: mockMappedTenants,
  };

  const tenantConfig: TenantConfig = { name: 'Tenant 3', canImpersonate: false };

  const mockBackendCreated: TenantBackend = {
    tenant_id: 'tenant-03',
    tenant_name: 'Tenant 3',
    can_impersonate: false,
  };

  const mockMappedCreated: Tenant = {
    id: 'tenant-03',
    name: 'Tenant 3',
    canImpersonate: false,
  };

  const mapperMock = {
    fromPaginatedDTO: vi.fn(),
    fromDTO: vi.fn(),
  };

  beforeEach(() => {
    vi.resetAllMocks();

    TestBed.configureTestingModule({
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        TenantApiClientService,
        { provide: TenantApiAdapter, useValue: mapperMock },
      ],
    });

    service = TestBed.inject(TenantApiClientService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  describe('getTenant', () => {
    it('should send GET request, map through adapter, and return domain model', () => {
      mapperMock.fromDTO.mockReturnValue(mockMappedTenants[0]);

      service.getTenant('tenant-01').subscribe((tenant) => {
        expect(tenant).toEqual(mockMappedTenants[0]);
        expect(tenant.id).toBe('tenant-01');
        expect(tenant.name).toBe('Tenant 1');
      });

      const req = httpMock.expectOne(`${apiUrl}/tenant/tenant-01`);
      expect(req.request.method).toBe('GET');
      req.flush(mockBackendTenants[0]);

      expect(mapperMock.fromDTO).toHaveBeenCalledWith(mockBackendTenants[0]);
    });
  });

  describe('getTenants', () => {
    it('should send GET with page and limit params, map through adapter, and return mapped response', () => {
      mapperMock.fromPaginatedDTO.mockReturnValue(mockMappedPaginatedResponse);

      service.getTenants(1, 20).subscribe((response) => {
        expect(response).toEqual(mockMappedPaginatedResponse);
        expect(response.tenants[0].id).toBe('tenant-01');
        expect(response.tenants[1].id).toBe('tenant-02');
      });

      const req = httpMock.expectOne(
        (request) =>
          request.url === `${apiUrl}/tenants` &&
          request.params.get('page') === '1' &&
          request.params.get('limit') === '20',
      );
      expect(req.request.method).toBe('GET');
      req.flush(mockBackendPaginatedResponse);

      expect(mapperMock.fromPaginatedDTO).toHaveBeenCalledWith(mockBackendPaginatedResponse);
    });
  });

  describe('getAllTenants', () => {
    it('should send GET request, map each DTO through adapter, and return domain models', () => {
      mapperMock.fromDTO
        .mockReturnValueOnce(mockMappedTenants[0])
        .mockReturnValueOnce(mockMappedTenants[1]);

      service.getAllTenants().subscribe((tenants) => {
        expect(tenants).toEqual(mockMappedTenants);
      });

      const req = httpMock.expectOne(`${apiUrl}/all_tenants`);
      expect(req.request.method).toBe('GET');
      req.flush({ tenants: mockBackendTenants });

      expect(mapperMock.fromDTO).toHaveBeenCalledTimes(2);
      expect(mapperMock.fromDTO).toHaveBeenCalledWith(mockBackendTenants[0]);
      expect(mapperMock.fromDTO).toHaveBeenCalledWith(mockBackendTenants[1]);
    });
  });

  describe('createTenant', () => {
    it('should send POST with mapped body, map response through adapter, and return domain model', () => {
      mapperMock.fromDTO.mockReturnValue(mockMappedCreated);

      service.createTenant(tenantConfig).subscribe((tenant) => {
        expect(tenant).toEqual(mockMappedCreated);
        expect(tenant.id).toBe('tenant-03');
        expect(tenant.name).toBe('Tenant 3');
      });

      const req = httpMock.expectOne(`${apiUrl}/tenant`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual({
        tenant_name: tenantConfig.name,
        can_impersonate: tenantConfig.canImpersonate,
      });
      req.flush(mockBackendCreated);

      expect(mapperMock.fromDTO).toHaveBeenCalledWith(mockBackendCreated);
    });
  });

  describe('deleteTenant', () => {
    it('should send DELETE with tenant id in URL and return void', () => {
      service.deleteTenant('tenant-01').subscribe((result) => {
        expect(result).toBeNull();
      });

      const req = httpMock.expectOne(`${apiUrl}/tenant/tenant-01`);
      expect(req.request.method).toBe('DELETE');
      req.flush(null);
    });
  });
});