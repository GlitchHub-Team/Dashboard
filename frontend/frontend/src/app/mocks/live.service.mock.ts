import { Injectable } from '@angular/core';
import { Observable, interval, map, Subject, takeUntil } from 'rxjs';
import { Sensor } from '../models/sensor/sensor.model';
import { RealTimeReading } from '../models/sensor-data/real-time-reading.model';
import { SensorProfiles } from '../models/sensor/sensor-profiles.enum';

@Injectable()
export class SensorRealTimeMockService {
  private readonly stop$ = new Subject<void>();
  private currentValue = 0;

  connect(sensor: Sensor): Observable<RealTimeReading> {
    this.currentValue = this.getBaseValue(sensor.profile);

    return interval(1000).pipe(
      // one reading every 1 second
      takeUntil(this.stop$),
      map(() => {
        // random walk: small increments/decrements
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
