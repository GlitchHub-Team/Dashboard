import { TestBed } from '@angular/core/testing';

import { GatewayCommandApiClientService } from './gateway-command-api-client.service';

describe('GatewayCommandApiClientService', () => {
  let service: GatewayCommandApiClientService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(GatewayCommandApiClientService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
