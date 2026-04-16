import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';

import { GatewayCommandApiClientService } from './gateway-command-api-client.service';
import { GatewayApiAdapter } from '../../adapters/gateway/gateway-api.adapter';
import { environment } from '../../../environments/environment';
import { GatewayBackend } from '../../models/gateway/gateway-backend.model';
import { Gateway } from '../../models/gateway/gateway.model';
import { GatewayStatus } from '../../models/gateway-status.enum';

describe('GatewayCommandApiClientService', () => {
  let service: GatewayCommandApiClientService;
  let httpMock: HttpTestingController;

  const apiUrl = `${environment.apiUrl}`;

  const mockBackendGateway: GatewayBackend = {
    gateway_id: 'gw-1',
    name: 'Gateway 1',
    tenant_id: 'tenant-1',
    status: 'active',
    interval: 60,
  };

  const mockMappedGateway: Gateway = {
    id: 'gw-1',
    name: 'Gateway 1',
    tenantId: 'tenant-1',
    status: GatewayStatus.ACTIVE,
    interval: 60,
  };

  const mapperMock = {
    fromDTO: vi.fn(),
  };

  beforeEach(() => {
    vi.resetAllMocks();

    TestBed.configureTestingModule({
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        GatewayCommandApiClientService,
        { provide: GatewayApiAdapter, useValue: mapperMock },
      ],
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
    it('should POST tenant_id and token in body, map through adapter, and return domain model', () => {
      mapperMock.fromDTO.mockReturnValue(mockMappedGateway);

      service.commissionGateway('gw-1', 'tenant-1', 'commission-token').subscribe((gateway) => {
        expect(gateway).toEqual(mockMappedGateway);
        expect(gateway.id).toBe('gw-1');
        expect(gateway.name).toBe('Gateway 1');
        expect(gateway.status).toBe(GatewayStatus.ACTIVE);
      });

      const req = httpMock.expectOne(`${apiUrl}/gateway/gw-1/commission`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual({
        tenant_id: 'tenant-1',
        commission_token: 'commission-token',
      });
      req.flush(mockBackendGateway);

      expect(mapperMock.fromDTO).toHaveBeenCalledWith(mockBackendGateway);
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