import { TestBed } from '@angular/core/testing';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { provideHttpClient } from '@angular/common/http';

import { GatewayApiClientService } from './gateway-api-client.service';
import { environment } from '../../../environments/environment';
import { Gateway } from '../../models/gateway.model';
import { GatewayConfig } from '../../models/gateway-config.model';
import { GatewayStatus } from '../../models/gateway-status.enum';

describe('GatewayApiClientService', () => {
  let service: GatewayApiClientService;
  let httpMock: HttpTestingController;

  const apiUrl = `${environment.apiUrl}/gateway`;

  const mockGateways: Gateway[] = [
    {
      id: 'gw-1',
      tenantId: 'tenant-1',
      name: 'Gateway 1',
      status: GatewayStatus.ONLINE,
    },
    {
      id: 'gw-2',
      tenantId: 'tenant-1',
      name: 'Gateway 2',
      status: GatewayStatus.OFFLINE,
    },
  ];

  beforeEach(() => {
    vi.resetAllMocks();

    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });

    service = TestBed.inject(GatewayApiClientService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  describe('getGatewayListByTenant', () => {
    it('should send GET request with tenantId param', () => {
      service.getGatewayListByTenant('tenant-1').subscribe((gateways) => {
        expect(gateways).toEqual(mockGateways);
      });

      const req = httpMock.expectOne(`${apiUrl}/list/?tenantId=tenant-1`);
      expect(req.request.method).toBe('GET');
      req.flush(mockGateways);
    });

    it('should return an observable of Gateway[]', () => {
      service.getGatewayListByTenant('tenant-1').subscribe((gateways) => {
        expect(gateways).toEqual(mockGateways);
      });

      const req = httpMock.expectOne(`${apiUrl}/list/?tenantId=tenant-1`);
      expect(req.request.method).toBe('GET');
      expect(req.request.params.get('tenantId')).toBe('tenant-1');
      req.flush(mockGateways);
    });
  });

  describe('getGatewayList', () => {
    it('should send GET request to list endpoint', () => {
      service.getGatewayList().subscribe((gateways) => {
        expect(gateways).toEqual(mockGateways);
      });

      const req = httpMock.expectOne(`${apiUrl}/list`);
      expect(req.request.method).toBe('GET');
      req.flush(mockGateways);
    });

    it('should return an observable of Gateway[]', () => {
      service.getGatewayList().subscribe((gateways) => {
        expect(gateways).toEqual(mockGateways);
      });

      const req = httpMock.expectOne(`${apiUrl}/list`);
      expect(req.request.method).toBe('GET');
      req.flush(mockGateways);
    });
  });

  describe('addNewGateway', () => {
    const mockConfig: GatewayConfig = {
      name: 'New Gateway',
    };

    const mockResponse: Gateway = {
      id: 'gw-new',
      tenantId: 'tenant-1',
      name: 'New Gateway',
      status: GatewayStatus.ONLINE,
    };

    it('should send POST request with gateway config', () => {
      service.addNewGateway(mockConfig).subscribe((gateway) => {
        expect(gateway).toEqual(mockResponse);
      });

      const req = httpMock.expectOne(`${apiUrl}/add`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual(mockConfig);
      req.flush(mockResponse);
    });

    it('should return an observable of Gateway', () => {
      service.addNewGateway(mockConfig).subscribe((gateway) => {
        expect(gateway.id).toBe('gw-new');
        expect(gateway.name).toBe('New Gateway');
      });

      const req = httpMock.expectOne(`${apiUrl}/add`);
      req.flush(mockResponse);
    });
  });

  describe('deleteGateway', () => {
    it('should send DELETE request with gateway id', () => {
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
      expect(req.request.method).toBe('DELETE');
      req.flush(null);
    });
  });
});
