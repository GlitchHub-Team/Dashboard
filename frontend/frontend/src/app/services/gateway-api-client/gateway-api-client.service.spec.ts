import { TestBed } from '@angular/core/testing';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { provideHttpClient } from '@angular/common/http';

import { GatewayApiClientService } from './gateway-api-client.service';
import { environment } from '../../../environments/environment';
import { GatewayBackend } from '../../models/gateway/gateway-backend.model';
import { GatewayConfig } from '../../models/gateway/gateway-config.model';
import { PaginatedResponse } from '../../models/paginated-response.model';

describe('GatewayApiClientService', () => {
  let service: GatewayApiClientService;
  let httpMock: HttpTestingController;

  const apiUrl = `${environment.apiUrl}/gateway`;

  const mockGateways: GatewayBackend[] = [
    {
      gateway_id: 'gw-1',
      name: 'Gateway 1',
      tenant_id: 'tenant-1',
      status: 'active',
      interval: 60,
    },
    {
      gateway_id: 'gw-2',
      name: 'Gateway 2',
      tenant_id: 'tenant-1',
      status: 'inactive',
      interval: 120,
    },
  ];

  const mockPaginatedResponse: PaginatedResponse<GatewayBackend> = {
    count: 2,
    total: 10,
    data: mockGateways,
  };

  beforeEach(() => {
    vi.resetAllMocks();

    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });

    service = TestBed.inject(GatewayApiClientService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  describe('getGatewayListByTenant', () => {
    it('should send GET request with correct URL and query params', () => {
      service.getGatewayListByTenant('tenant-1', 1, 20).subscribe((response) => {
        expect(response).toEqual(mockPaginatedResponse);
      });

      const req = httpMock.expectOne(`${apiUrl}/tenant-1/list?page=1&limit=20`);
      expect(req.request.method).toBe('GET');
      expect(req.request.params.get('page')).toBe('1');
      expect(req.request.params.get('limit')).toBe('20');
      req.flush(mockPaginatedResponse);
    });

    it('should return a PaginatedResponse of GatewayBackend', () => {
      service.getGatewayListByTenant('tenant-1', 0, 10).subscribe((response) => {
        expect(response.count).toBe(2);
        expect(response.total).toBe(10);
        expect(response.data.length).toBe(2);
        expect(response.data[0].gateway_id).toBe('gw-1');
        expect(response.data[1].gateway_id).toBe('gw-2');
      });

      const req = httpMock.expectOne(`${apiUrl}/tenant-1/list?page=0&limit=10`);
      req.flush(mockPaginatedResponse);
    });
  });

  describe('getGatewayList', () => {
    it('should send GET request with correct URL and query params', () => {
      service.getGatewayList(0, 10).subscribe((response) => {
        expect(response).toEqual(mockPaginatedResponse);
      });

      const req = httpMock.expectOne(`${apiUrl}/list?page=0&limit=10`);
      expect(req.request.method).toBe('GET');
      expect(req.request.params.get('page')).toBe('0');
      expect(req.request.params.get('limit')).toBe('10');
      req.flush(mockPaginatedResponse);
    });

    it('should return a PaginatedResponse of GatewayBackend', () => {
      service.getGatewayList(2, 25).subscribe((response) => {
        expect(response.count).toBe(2);
        expect(response.total).toBe(10);
        expect(response.data.length).toBe(2);
      });

      const req = httpMock.expectOne(`${apiUrl}/list?page=2&limit=25`);
      req.flush(mockPaginatedResponse);
    });
  });

  describe('addNewGateway', () => {
    const mockConfig: GatewayConfig = {
      name: 'New Gateway',
      interval: 60,
    };

    const mockResponse: GatewayBackend = {
      gateway_id: 'gw-3',
      name: 'New Gateway',
      tenant_id: 'tenant-1',
      status: 'active',
      interval: 60,
    };

    it('should send POST request with gateway config as body', () => {
      service.addNewGateway(mockConfig).subscribe((gateway) => {
        expect(gateway).toEqual(mockResponse);
      });

      const req = httpMock.expectOne(`${apiUrl}/add`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual(mockConfig);
      req.flush(mockResponse);
    });

    it('should return a GatewayBackend', () => {
      service.addNewGateway(mockConfig).subscribe((gateway) => {
        expect(gateway.gateway_id).toBe('gw-3');
        expect(gateway.name).toBe('New Gateway');
      });

      const req = httpMock.expectOne(`${apiUrl}/add`);
      req.flush(mockResponse);
    });
  });

  describe('deleteGateway', () => {
    it('should send DELETE request with gateway id in the URL', () => {
      service.deleteGateway('gw-1').subscribe();

      const req = httpMock.expectOne(`${apiUrl}/delete/gw-1`);
      expect(req.request.method).toBe('DELETE');
      req.flush(null);
    });

    it('should return an observable of void', () => {
      service.deleteGateway('gw-1').subscribe((result) => {
        expect(result).toBeNull();
      });

      const req = httpMock.expectOne(`${apiUrl}/delete/gw-1`);
      req.flush(null);
    });
  });
});
