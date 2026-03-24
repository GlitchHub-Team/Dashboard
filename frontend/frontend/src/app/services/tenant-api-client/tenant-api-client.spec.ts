import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';

import { TenantApiClientService } from './tenant-api-client.service';
import { environment } from '../../../environments/environment';
import { TenantDataAdapter, RawPaginatedTenantResponse } from '../../adapters/tenant-data.adapter';
import { Tenant } from '../../models/tenant/tenant.model';
import { RawTenantConfig } from '../../models/tenant/raw-tenant-config.model';

describe('TenantApiClientService', () => {
  let service: TenantApiClientService;
  let httpMock: HttpTestingController;

  const apiUrl = `${environment.apiUrl}/tenants`;

  const adapterMock = {
    adaptPaginated: vi.fn(),
    adapt: vi.fn(),
  };

  const rawPaginatedResponse: RawPaginatedTenantResponse = {
    items: [{ name: 'Tenant 1' }, { name: 'Tenant 2' }],
    totalCount: 2,
  };

  const adaptedPaginatedResponse: { items: Tenant[]; totalCount: number } = {
    items: [{ name: 'Tenant 1' }, { name: 'Tenant 2' }],
    totalCount: 2,
  };

  const rawTenant: RawTenantConfig = { name: 'Tenant 3' };
  const adaptedTenant: Tenant = { name: 'Tenant 3' };

  beforeEach(() => {
    vi.resetAllMocks();

    TestBed.configureTestingModule({
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        { provide: TenantDataAdapter, useValue: adapterMock },
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
    it('should send GET request with page and size query params', () => {
      adapterMock.adaptPaginated.mockReturnValue(adaptedPaginatedResponse);

      service.getTenant(1, 20).subscribe((response) => {
        expect(response).toEqual(adaptedPaginatedResponse);
      });

      const req = httpMock.expectOne(
        (request) =>
          request.url === apiUrl &&
          request.params.get('page') === '1' &&
          request.params.get('size') === '20',
      );
      expect(req.request.method).toBe('GET');
      req.flush(rawPaginatedResponse);
    });

    it('should use default page and size values when omitted', () => {
      adapterMock.adaptPaginated.mockReturnValue(adaptedPaginatedResponse);

      service.getTenant().subscribe();

      const req = httpMock.expectOne(
        (request) =>
          request.url === apiUrl &&
          request.params.get('page') === '0' &&
          request.params.get('size') === '10',
      );
      expect(req.request.method).toBe('GET');
      req.flush(rawPaginatedResponse);
    });

    it('should map the raw paginated response through TenantDataAdapter', () => {
      adapterMock.adaptPaginated.mockReturnValue(adaptedPaginatedResponse);

      let result: { items: Tenant[]; totalCount: number } | undefined;
      service.getTenant(0, 10).subscribe((response) => {
        result = response;
      });

      const req = httpMock.expectOne((request) => request.url === apiUrl);
      req.flush(rawPaginatedResponse);

      expect(adapterMock.adaptPaginated).toHaveBeenCalledWith(rawPaginatedResponse);
      expect(result).toEqual(adaptedPaginatedResponse);
    });
  });

  describe('createTenant', () => {
    it('should send POST request with tenant config as body', () => {
      adapterMock.adapt.mockReturnValue(adaptedTenant);

      service.createTenant(rawTenant).subscribe((tenant) => {
        expect(tenant).toEqual(adaptedTenant);
      });

      const req = httpMock.expectOne(apiUrl);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual(rawTenant);
      req.flush(rawTenant);
    });

    it('should map the raw tenant response through TenantDataAdapter', () => {
      adapterMock.adapt.mockReturnValue(adaptedTenant);

      let result: Tenant | undefined;
      service.createTenant(rawTenant).subscribe((tenant) => {
        result = tenant;
      });

      const req = httpMock.expectOne(apiUrl);
      req.flush(rawTenant);

      expect(adapterMock.adapt).toHaveBeenCalledWith(rawTenant);
      expect(result).toEqual(adaptedTenant);
    });
  });

  describe('deleteTenant', () => {
    it('should send DELETE request with tenant id in the URL', () => {
      service.deleteTenant('Tenant 1').subscribe();

      const req = httpMock.expectOne(`${apiUrl}/Tenant 1`);
      expect(req.request.method).toBe('DELETE');
      req.flush(null);
    });

    it('should return an observable of void', () => {
      let result: void | null | undefined;

      service.deleteTenant('Tenant 1').subscribe((response) => {
        result = response;
      });

      const req = httpMock.expectOne(`${apiUrl}/Tenant 1`);
      req.flush(null);

      expect(result).toBeNull();
    });
  });
});
