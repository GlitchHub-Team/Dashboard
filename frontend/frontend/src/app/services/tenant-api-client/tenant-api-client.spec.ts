import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';

import { TenantApiClientService } from './tenant-api-client.service';
import { environment } from '../../../environments/environment';
import { PaginatedTenantResponse } from '../../models/tenant/paginated-tenant-response.model';
import { TenantBackend } from '../../models/tenant/tenant-backend.model';
import { TenantConfig } from '../../models/tenant/tenant-config.model';

describe('TenantApiClientService', () => {
  let service: TenantApiClientService;
  let httpMock: HttpTestingController;

  const apiUrl = `${environment.apiUrl}`;

  const paginatedTenantResponse: PaginatedTenantResponse<TenantBackend> = {
    count: 2,
    total: 2,
    tenants: [
      { tenant_id: 'tenant-01', tenant_name: 'Tenant 1', can_impersonate: false },
      { tenant_id: 'tenant-02', tenant_name: 'Tenant 2', can_impersonate: true },
    ],
  };

  const tenantConfig: TenantConfig = { name: 'Tenant 3', canImpersonate: false };
  const createdTenant: TenantBackend = {
    tenant_id: 'tenant-03',
    tenant_name: 'Tenant 3',
    can_impersonate: false,
  };

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
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
    it('should send GET request to fetch tenant by id', () => {
      service.getTenant('tenant-01').subscribe((tenant) => {
        expect(tenant).toEqual(paginatedTenantResponse.tenants[0]);
      });

      const req = httpMock.expectOne(`${apiUrl}/tenant/tenant-01`);
      expect(req.request.method).toBe('GET');
      req.flush(paginatedTenantResponse.tenants[0]);
    });
  });

  describe('getTenants', () => {
    it('should send GET request with page and limit query params', () => {
      service.getTenants(1, 20).subscribe((response) => {
        expect(response).toEqual(paginatedTenantResponse);
      });

      const req = httpMock.expectOne(
        (request) =>
          request.url === `${apiUrl}/tenants` &&
          request.params.get('page') === '1' &&
          request.params.get('limit') === '20',
      );
      expect(req.request.method).toBe('GET');
      req.flush(paginatedTenantResponse);
    });
  });

  describe('createTenant', () => {
    it('should send POST request with tenant config as body', () => {
      service.createTenant(tenantConfig).subscribe((tenant) => {
        expect(tenant).toEqual(createdTenant);
      });

      const req = httpMock.expectOne(`${apiUrl}/tenant`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual({
        tenant_name: tenantConfig.name,
        can_impersonate: tenantConfig.canImpersonate,
      });
      req.flush(createdTenant);
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
