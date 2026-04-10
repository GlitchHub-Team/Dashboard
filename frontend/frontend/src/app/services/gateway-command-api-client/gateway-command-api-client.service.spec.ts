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
    it('should POST tenant_id and token in body and return GatewayBackend', () => {
      service.commissionGateway('gw-1', 'tenant-1', 'commission-token').subscribe((gateway) => {
        expect(gateway).toEqual(mockGateway);
        expect(gateway.gateway_id).toBe('gw-1');
        expect(gateway.name).toBe('Gateway 1');
        expect(gateway.status).toBe('active');
      });

      const req = httpMock.expectOne(`${apiUrl}/gateway/gw-1/commission`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual({
        tenant_id: 'tenant-1',
        commission_token: 'commission-token',
      });
      req.flush(mockGateway);
    });
  });

  describe('void gateway commands', () => {
    it.each([
      ['decommissionGateway', 'decommission'] as const,
      ['resetGateway', 'reset'] as const,
      ['rebootGateway', 'reboot'] as const,
      ['interruptGateway', 'interrupt'] as const,
      ['resumeGateway', 'resume'] as const,
    ])('%s should POST empty body to /gateway/gw-1/%s and return void', (method, path) => {
      service[method]('gw-1').subscribe((result) => {
        expect(result).toBeNull();
      });

      const req = httpMock.expectOne(`${apiUrl}/gateway/gw-1/${path}`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual({});
      req.flush(null);
    });
  });
});
