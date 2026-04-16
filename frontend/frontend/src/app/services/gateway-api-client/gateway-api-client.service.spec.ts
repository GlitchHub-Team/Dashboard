import { TestBed } from '@angular/core/testing';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { provideHttpClient } from '@angular/common/http';

import { GatewayApiClientService } from './gateway-api-client.service';
import { GatewayApiAdapter } from '../../adapters/gateway/gateway-api.adapter';
import { environment } from '../../../environments/environment';
import { GatewayBackend } from '../../models/gateway/gateway-backend.model';
import { GatewayConfig } from '../../models/gateway/gateway-config.model';
import { PaginatedGatewayResponse } from '../../models/gateway/paginated-gateway-response.model';
import { Gateway } from '../../models/gateway/gateway.model';
import { GatewayStatus } from '../../models/gateway-status.enum';

describe('GatewayApiClientService', () => {
  let service: GatewayApiClientService;
  let httpMock: HttpTestingController;

  const apiUrl = `${environment.apiUrl}`;

  const mockBackendGateways: GatewayBackend[] = [
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

  const mockBackendPaginatedResponse: PaginatedGatewayResponse<GatewayBackend> = {
    count: 2,
    total: 10,
    gateways: mockBackendGateways,
  };

  const mockMappedGateways: Gateway[] = [
    {
      id: 'gw-1',
      name: 'Gateway 1',
      tenantId: 'tenant-1',
      status: GatewayStatus.ACTIVE,
      interval: 60,
    },
    {
      id: 'gw-2',
      name: 'Gateway 2',
      tenantId: 'tenant-1',
      status: GatewayStatus.INACTIVE,
      interval: 120,
    },
  ];

  const mockMappedPaginatedResponse: PaginatedGatewayResponse<Gateway> = {
    count: 2,
    total: 10,
    gateways: mockMappedGateways,
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
        GatewayApiClientService,
        { provide: GatewayApiAdapter, useValue: mapperMock },
      ],
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
    it('should send GET with correct URL and params, map through adapter, and return mapped response', () => {
      mapperMock.fromPaginatedDTO.mockReturnValue(mockMappedPaginatedResponse);

      service.getGatewayListByTenant('tenant-1', 1, 20).subscribe((response) => {
        expect(response).toEqual(mockMappedPaginatedResponse);
        expect(response.count).toBe(2);
        expect(response.total).toBe(10);
        expect(response.gateways[0].id).toBe('gw-1');
        expect(response.gateways[1].id).toBe('gw-2');
      });

      const req = httpMock.expectOne(`${apiUrl}/tenant/tenant-1/gateways?page=1&limit=20`);
      expect(req.request.method).toBe('GET');
      expect(req.request.params.get('page')).toBe('1');
      expect(req.request.params.get('limit')).toBe('20');
      req.flush(mockBackendPaginatedResponse);

      expect(mapperMock.fromPaginatedDTO).toHaveBeenCalledWith(mockBackendPaginatedResponse);
    });
  });

  describe('getGatewayList', () => {
    it('should send GET with correct URL and params, map through adapter, and return mapped response', () => {
      mapperMock.fromPaginatedDTO.mockReturnValue(mockMappedPaginatedResponse);

      service.getGatewayList(0, 10).subscribe((response) => {
        expect(response).toEqual(mockMappedPaginatedResponse);
        expect(response.count).toBe(2);
        expect(response.total).toBe(10);
        expect(response.gateways.length).toBe(2);
      });

      const req = httpMock.expectOne(`${apiUrl}/gateways?page=0&limit=10`);
      expect(req.request.method).toBe('GET');
      expect(req.request.params.get('page')).toBe('0');
      expect(req.request.params.get('limit')).toBe('10');
      req.flush(mockBackendPaginatedResponse);

      expect(mapperMock.fromPaginatedDTO).toHaveBeenCalledWith(mockBackendPaginatedResponse);
    });
  });

  describe('addNewGateway', () => {
    const mockConfig: GatewayConfig = {
      name: 'New Gateway',
      interval: 60,
    };

    const mockBackendResponse: GatewayBackend = {
      gateway_id: 'gw-3',
      name: 'New Gateway',
      tenant_id: 'tenant-1',
      status: 'active',
      interval: 60,
    };

    const mockMappedGateway: Gateway = {
      id: 'gw-3',
      name: 'New Gateway',
      tenantId: 'tenant-1',
      status: GatewayStatus.ACTIVE,
      interval: 60,
    };

    it('should send POST with mapped body, map response through adapter, and return domain model', () => {
      mapperMock.fromDTO.mockReturnValue(mockMappedGateway);

      service.addNewGateway(mockConfig).subscribe((gateway) => {
        expect(gateway).toEqual(mockMappedGateway);
        expect(gateway.id).toBe('gw-3');
        expect(gateway.name).toBe('New Gateway');
      });

      const req = httpMock.expectOne(`${apiUrl}/gateway`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual({
        name: mockConfig.name,
        interval: mockConfig.interval,
      });
      req.flush(mockBackendResponse);

      expect(mapperMock.fromDTO).toHaveBeenCalledWith(mockBackendResponse);
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