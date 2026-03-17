import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { TenantService } from './tenant.service';
import { TenantApiClientService } from './tenant-api-client.service';

class MockTenantApiClientService {
  getTenantResult = of([{ name: 'Tenant 1' }, { name: 'Tenant 2' }]);
  getTenantCalled = false;

  getTenant() {
    this.getTenantCalled = true;
    return this.getTenantResult;
  }
  createTenant() {
    // no-op
  }
  deleteTenant() {
    // no-op
  }
}

describe('TenantService', () => {
  let service: TenantService;
  let apiClient: MockTenantApiClientService;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [
        TenantService,
        { provide: TenantApiClientService, useClass: MockTenantApiClientService },
      ],
    });
    service = TestBed.inject(TenantService);
    apiClient = TestBed.inject(TenantApiClientService) as unknown as MockTenantApiClientService;
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  describe('retrieveTenant', () => {
    it('should retrieve tenants and update the list', () => {
      apiClient.getTenantResult = of([{ name: 'Tenant 1' }, { name: 'Tenant 2' }]);

      service.retrieveTenant();

      expect(apiClient.getTenantCalled).toBe(true);
      expect(service.loading()).toBe(false);
      expect(service.tenantList().length).toBe(2);
      expect(service.error()).toBeNull();
    });

    it('should handle errors when retrieving tenants', () => {
      const error = new Error('Failed to fetch');
      apiClient.getTenantResult = throwError(() => error);

      service.retrieveTenant();

      expect(apiClient.getTenantCalled).toBe(true);
      expect(service.loading()).toBe(false);
      expect(service.tenantList()).toEqual([]);
      expect(service.error()).toBe('Failed to fetch');
    });
  });
});
