import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';

import { SensorCommandApiClientService } from './sensor-command-api-client.service';
import { environment } from '../../../environments/environment';

describe('SensorCommandApiClientService', () => {
  let service: SensorCommandApiClientService;
  let httpMock: HttpTestingController;

  const apiUrl = `${environment.apiUrl}`;

  beforeEach(() => {
    vi.resetAllMocks();

    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });

    service = TestBed.inject(SensorCommandApiClientService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  describe('sensor commands', () => {
    it.each([
      ['interruptSensor', 'interrupt'] as const,
      ['resumeSensor', 'resume'] as const,
    ])('%s should POST empty body to /sensor/sensor-1/%s and return void', (method, path) => {
      service[method]('sensor-1').subscribe((result) => {
        expect(result).toBeNull();
      });

      const req = httpMock.expectOne(`${apiUrl}/sensor/sensor-1/${path}`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual({});
      req.flush(null);
    });
  });
});
