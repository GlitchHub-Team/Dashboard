import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';

import { TenantApiClientService } from './tenant-api-client.service';
import { environment } from '../../../environments/environment';
import { PaginatedResponse } from '../../models/paginated-response.model';
import { TenantBackend } from '../../models/tenant/tenant-backend.model';
import { TenantConfig } from '../../models/tenant/tenant-config.model';

describe('TenantApiClientService', () => {
  let service: TenantApiClientService;
  let httpMock: HttpTestingController;

  const apiUrl = `${environment.apiUrl}`;

  const paginatedTenantResponse: PaginatedResponse<TenantBackend> = {
    count: 2,
    total: 2,
    data: [
      { tenant_id: 'tenant-01', name: 'Tenant 1', can_impersonate: false },
      { tenant_id: 'tenant-02', name: 'Tenant 2', can_impersonate: true },
    ],
  };

  const tenantConfig: TenantConfig = { name: 'Tenant 3', canImpersonate: false };
  const createdTenant: TenantBackend = {
    tenant_id: 'tenant-03',
    name: 'Tenant 3',
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
    it('should send GET request with page and limit query params', () => {
      service.getTenant(1, 20).subscribe((response) => {
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
      expect(req.request.body).toEqual(tenantConfig);
      req.flush(createdTenant);
    });
  });

  describe('deleteTenant', () => {
    it('should send DELETE request with tenant id in the URL', () => {
      service.deleteTenant('tenant-01').subscribe();

      const req = httpMock.expectOne(`${apiUrl}/tenant/tenant-01`);
      expect(req.request.method).toBe('DELETE');
      req.flush(null);
    });

    it('should return an observable of void', () => {
      let result: void | null | undefined;

      service.deleteTenant('tenant-01').subscribe((response) => {
        result = response;
      });

      const req = httpMock.expectOne(`${apiUrl}/tenant/tenant-01`);
      req.flush(null);

      expect(result).toBeNull();
    });
  });
});
