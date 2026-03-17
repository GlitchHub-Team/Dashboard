import { TestBed } from '@angular/core/testing';

import { GatewayApiClientService } from './gateway-api-client.service';

describe('GatewayApiClientService', () => {
  let service: GatewayApiClientService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(GatewayApiClientService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
