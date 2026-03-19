import { TestBed } from '@angular/core/testing';

import { SensorLiveReadingsApiService } from './sensor-live-readings-api.service';

describe('SensorLiveReadingsApiService', () => {
  let service: SensorLiveReadingsApiService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(SensorLiveReadingsApiService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
