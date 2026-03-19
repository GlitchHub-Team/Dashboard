import { TestBed } from '@angular/core/testing';

import { SensorHistoricApiService } from './sensor-historic-api.service';

describe('SensorHistoricApiService', () => {
  let service: SensorHistoricApiService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(SensorHistoricApiService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
