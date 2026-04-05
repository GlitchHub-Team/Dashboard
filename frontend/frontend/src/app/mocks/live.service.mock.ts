import { Injectable } from '@angular/core';
import { Observable, interval, map, Subject, takeUntil, switchMap, throwError, timer } from 'rxjs';
import { HttpErrorResponse } from '@angular/common/http';
import { Sensor } from '../models/sensor/sensor.model';
import { RealTimeReading } from '../models/sensor-data/real-time-reading.model';
import { SensorProfiles } from '../models/sensor/sensor-profiles.enum';

@Injectable()
export class SensorRealTimeMockService {
  private readonly stop$ = new Subject<void>();
  private currentValue = 0;
  private readingCount = 0;

  // Toggle these to simulate different scenarios
  private readonly shouldFailConnection = false;
  private readonly shouldDisconnectAfter = 10; // 0 = never, N = after N readings

  connect(sensor: Sensor): Observable<RealTimeReading> {
    this.readingCount = 0;

    // Simulate connection failure
    if (this.shouldFailConnection) {
      return timer(1000).pipe(
        switchMap(() =>
          throwError(
            () =>
              new HttpErrorResponse({
                status: 0,
                statusText: 'Unknown Error',
                error: { error: 'WebSocket connection failed' },
              }),
          ),
        ),
      );
    }

    this.currentValue = this.getBaseValue(sensor.profile);

    return interval(1000).pipe(
      takeUntil(this.stop$),
      map(() => {
        this.readingCount++;

        // Simulate disconnection after N readings
        if (this.shouldDisconnectAfter > 0 && this.readingCount >= this.shouldDisconnectAfter) {
          throw new HttpErrorResponse({
            status: 0,
            statusText: 'Unknown Error',
            error: { error: 'WebSocket connection lost' },
          });
        }

        const delta = (Math.random() - 0.5) * this.getStep(sensor.profile);
        this.currentValue += delta;
        this.currentValue = Math.round(this.currentValue * 100) / 100;

        return {
          datum: this.currentValue,
          timestamp: Date.now(),
        };
      }),
    );
  }

  disconnect(): void {
    this.stop$.next();
  }

  private getBaseValue(profile: SensorProfiles): number {
    switch (profile) {
      case SensorProfiles.HEART_RATE_SERVICE:
        return 75;
      case SensorProfiles.PULSE_OXIMETER_SERVICE:
        return 97;
      case SensorProfiles.CUSTOM_ECG_SERVICE:
        return 0;
      case SensorProfiles.HEALTH_THERMOMETER_SERVICE:
        return 36.6;
      case SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE:
        return 22;
    }
  }

  private getStep(profile: SensorProfiles): number {
    switch (profile) {
      case SensorProfiles.HEART_RATE_SERVICE:
        return 3;
      case SensorProfiles.PULSE_OXIMETER_SERVICE:
        return 1;
      case SensorProfiles.CUSTOM_ECG_SERVICE:
        return 0.5;
      case SensorProfiles.HEALTH_THERMOMETER_SERVICE:
        return 0.2;
      case SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE:
        return 1;
    }
  }
}
