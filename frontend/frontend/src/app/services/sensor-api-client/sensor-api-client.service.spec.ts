import { TestBed } from '@angular/core/testing';

import { SensorApiClientService } from './sensor-api-client.service';

describe('SensorApiClientService', () => {
  let service: SensorApiClientService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(SensorApiClientService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
