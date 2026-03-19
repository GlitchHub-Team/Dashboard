import { TestBed } from '@angular/core/testing';

import { SensorChartService } from './sensor-chart.service';

describe('SensorChartService', () => {
  let service: SensorChartService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(SensorChartService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
