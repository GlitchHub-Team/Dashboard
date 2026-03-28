import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';

import { GatewayCommandApiClientService } from './gateway-command-api-client.service';
import { environment } from '../../../environments/environment';
import { GatewayBackend } from '../../models/gateway/gateway-backend.model';

describe('GatewayCommandApiClientService', () => {
  let service: GatewayCommandApiClientService;
  let httpMock: HttpTestingController;

  const apiUrl = `${environment.apiUrl}`;

  const mockGateway: GatewayBackend = {
    gateway_id: 'gw-1',
    name: 'Gateway 1',
    tenant_id: 'tenant-1',
    status: 'active',
    interval: 60,
  };

  beforeEach(() => {
    vi.resetAllMocks();

    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });

    service = TestBed.inject(GatewayCommandApiClientService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  describe('commissionGateway', () => {
    it('should send POST request with gateway id in the URL and tenant id and commission token in the body', () => {
      service.commissionGateway('gw-1', 'tenant-1', 'commission-token').subscribe((gateway) => {
        expect(gateway).toEqual(mockGateway);
      });

      const req = httpMock.expectOne(`${apiUrl}/gateway/gw-1/commission`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual({
        tenant_id: 'tenant-1',
        commission_token: 'commission-token',
      });
      req.flush(mockGateway);
    });

    it('should return a GatewayBackend', () => {
      service.commissionGateway('gw-1', 'tenant-1', 'commission-token').subscribe((gateway) => {
        expect(gateway.gateway_id).toBe('gw-1');
        expect(gateway.name).toBe('Gateway 1');
        expect(gateway.status).toBe('active');
      });

      const req = httpMock.expectOne(`${apiUrl}/gateway/gw-1/commission`);
      req.flush(mockGateway);
    });
  });

  describe('decommissionGateway', () => {
    it('should send POST request with gateway id in the URL', () => {
      service.decommissionGateway('gw-1').subscribe();

      const req = httpMock.expectOne(`${apiUrl}/gateway/gw-1/decommission`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual({});
      req.flush(null);
    });

    it('should return an observable of void', () => {
      service.decommissionGateway('gw-1').subscribe((result) => {
        expect(result).toBeNull();
      });

      const req = httpMock.expectOne(`${apiUrl}/gateway/gw-1/decommission`);
      req.flush(null);
    });
  });

  describe('resetGateway', () => {
    it('should send POST request with gateway id in the URL', () => {
      service.resetGateway('gw-1').subscribe();

      const req = httpMock.expectOne(`${apiUrl}/gateway/gw-1/reset`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual({});
      req.flush(null);
    });

    it('should return an observable of void', () => {
      service.resetGateway('gw-1').subscribe((result) => {
        expect(result).toBeNull();
      });

      const req = httpMock.expectOne(`${apiUrl}/gateway/gw-1/reset`);
      req.flush(null);
    });
  });

  describe('rebootGateway', () => {
    it('should send POST request with gateway id in the URL', () => {
      service.rebootGateway('gw-1').subscribe();

      const req = httpMock.expectOne(`${apiUrl}/gateway/gw-1/reboot`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual({});
      req.flush(null);
    });

    it('should return an observable of void', () => {
      service.rebootGateway('gw-1').subscribe((result) => {
        expect(result).toBeNull();
      });

      const req = httpMock.expectOne(`${apiUrl}/gateway/gw-1/reboot`);
      req.flush(null);
    });
  });
});
