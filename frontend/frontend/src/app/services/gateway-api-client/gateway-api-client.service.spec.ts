import { TestBed } from '@angular/core/testing';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { provideHttpClient } from '@angular/common/http';

import { GatewayApiClientService } from './gateway-api-client.service';
import { environment } from '../../../environments/environment';
import { GatewayBackend } from '../../models/gateway/gateway-backend.model';
import { GatewayConfig } from '../../models/gateway/gateway-config.model';
import { PaginatedGatewayResponse } from '../../models/gateway/paginated-gateway-response.model';

describe('GatewayApiClientService', () => {
  let service: GatewayApiClientService;
  let httpMock: HttpTestingController;

  const apiUrl = `${environment.apiUrl}`;

  const mockGateways: GatewayBackend[] = [
    {
      gateway_id: 'gw-1',
      name: 'Gateway 1',
      tenant_id: 'tenant-1',
      status: 'attivo',
      interval: 60,
    },
    {
      gateway_id: 'gw-2',
      name: 'Gateway 2',
      tenant_id: 'tenant-1',
      status: 'inattivo',
      interval: 120,
    },
  ];

  const mockPaginatedResponse: PaginatedGatewayResponse<GatewayBackend> = {
    count: 2,
    total: 10,
    gateways: mockGateways,
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
    it('should send GET with correct URL, params, and return a PaginatedResponse', () => {
      service.getGatewayListByTenant('tenant-1', 1, 20).subscribe((response) => {
        expect(response).toEqual(mockPaginatedResponse);
        expect(response.count).toBe(2);
        expect(response.total).toBe(10);
        expect(response.gateways[0].gateway_id).toBe('gw-1');
        expect(response.gateways[1].gateway_id).toBe('gw-2');
      });

      const req = httpMock.expectOne(`${apiUrl}/tenant/tenant-1/gateways?page=1&limit=20`);
      expect(req.request.method).toBe('GET');
      expect(req.request.params.get('page')).toBe('1');
      expect(req.request.params.get('limit')).toBe('20');
      req.flush(mockPaginatedResponse);
    });
  });

  describe('getGatewayList', () => {
    it('should send GET with correct URL, params, and return a PaginatedResponse', () => {
      service.getGatewayList(0, 10).subscribe((response) => {
        expect(response).toEqual(mockPaginatedResponse);
        expect(response.count).toBe(2);
        expect(response.total).toBe(10);
        expect(response.gateways.length).toBe(2);
      });

      const req = httpMock.expectOne(`${apiUrl}/gateways?page=0&limit=10`);
      expect(req.request.method).toBe('GET');
      expect(req.request.params.get('page')).toBe('0');
      expect(req.request.params.get('limit')).toBe('10');
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

    it('should send POST with gateway config body and return a GatewayBackend', () => {
      service.addNewGateway(mockConfig).subscribe((gateway) => {
        expect(gateway).toEqual(mockResponse);
        expect(gateway.gateway_id).toBe('gw-3');
        expect(gateway.name).toBe('New Gateway');
      });

      const req = httpMock.expectOne(`${apiUrl}/gateway`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual(mockConfig);
      req.flush(mockResponse);
    });
  });

  describe('deleteGateway', () => {
    it('should send DELETE with gateway id in URL and complete with void', () => {
      service.deleteGateway('gw-1').subscribe((result) => {
        expect(result).toBeNull();
      });

      const req = httpMock.expectOne(`${apiUrl}/gateway/gw-1`);
      expect(req.request.method).toBe('DELETE');
      req.flush(null);
    });
  });
});
