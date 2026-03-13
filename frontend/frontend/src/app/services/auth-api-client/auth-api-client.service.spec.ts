import { TestBed } from '@angular/core/testing';

import { AuthApiClientService } from './auth-api-client.service';

describe('AuthApiClientService', () => {
  let service: AuthApiClientService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(AuthApiClientService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
