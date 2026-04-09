import { Injectable } from '@angular/core';
import { delay, Observable, of, switchMap, throwError, timer } from 'rxjs';

import { ApiError } from '../models/api-error.model';

@Injectable({ providedIn: 'root' })
export class SensorCommandApiClientMockService {
  private readonly shouldFailInterrupt = true;
  private readonly shouldFailResume = true;

  public interruptSensor(_sensorId: string): Observable<void> {
    if (this.shouldFailInterrupt) {
      return this.delayedError(400, 'Failed to interrupt sensor');
    }
    return of(void 0).pipe(delay(500));
  }

  public resumeSensor(_sensorId: string): Observable<void> {
    if (this.shouldFailResume) {
      return this.delayedError(400, 'Failed to resume sensor');
    }
    return of(void 0).pipe(delay(500));
  }

  private delayedError(status: number, message: string): Observable<never> {
    return timer(500).pipe(switchMap(() => throwError(() => ({ status, message }) as ApiError)));
  }
}
