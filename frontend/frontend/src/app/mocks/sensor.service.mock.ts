import { Injectable } from '@angular/core';
import { Observable, of } from 'rxjs';
import { delay } from 'rxjs/operators';

import { Sensor } from '../models/sensor.model';
import { SensorProfiles } from '../models/sensor-profiles.enum';

@Injectable({
  providedIn: 'root',
})
export class SensorServiceMock {
  private readonly mockSensors: Record<string, Sensor[]> = {
    'gateway-01': [
      {
        id: 'sensor-01',
        gatewayId: 'gateway-01',
        name: 'Heart Rate Monitor A',
        profile: SensorProfiles.HEART_RATE_SERVICE,
      },
      {
        id: 'sensor-02',
        gatewayId: 'gateway-01',
        name: 'Thermometer A',
        profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
      },
      {
        id: 'sensor-03',
        gatewayId: 'gateway-01',
        name: 'ECG Monitor A',
        profile: SensorProfiles.CUSTOM_ECG_SERVICE,
      },
    ],
    'gateway-02': [
      {
        id: 'sensor-04',
        gatewayId: 'gateway-02',
        name: 'Pulse Oximeter A',
        profile: SensorProfiles.PULSE_OXIMETER_SERVICE,
      },
      {
        id: 'sensor-05',
        gatewayId: 'gateway-02',
        name: 'Environment Sensor A',
        profile: SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE,
      },
    ],
    'gateway-03': [],
    'gateway-04': [
      {
        id: 'sensor-06',
        gatewayId: 'gateway-04',
        name: 'Heart Rate Monitor B',
        profile: SensorProfiles.HEART_RATE_SERVICE,
      },
    ],
    'gateway-05': [],
  };

  private readonly tenantSensors: Record<string, Sensor[]> = {
    'tenant-01': [
      {
        id: 'sensor-01',
        gatewayId: 'gateway-01',
        name: 'Heart Rate Monitor A',
        profile: SensorProfiles.HEART_RATE_SERVICE,
      },
      {
        id: 'sensor-02',
        gatewayId: 'gateway-01',
        name: 'Thermometer A',
        profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
      },
      {
        id: 'sensor-03',
        gatewayId: 'gateway-01',
        name: 'ECG Monitor A',
        profile: SensorProfiles.CUSTOM_ECG_SERVICE,
      },
    ],
    'tenant-02': [
      {
        id: 'sensor-04',
        gatewayId: 'gateway-02',
        name: 'Pulse Oximeter A',
        profile: SensorProfiles.PULSE_OXIMETER_SERVICE,
      },
      {
        id: 'sensor-05',
        gatewayId: 'gateway-02',
        name: 'Environment Sensor A',
        profile: SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE,
      },
    ],
    'tenant-03': [],
    'tenant-04': [
      {
        id: 'sensor-06',
        gatewayId: 'gateway-04',
        name: 'Heart Rate Monitor B',
        profile: SensorProfiles.HEART_RATE_SERVICE,
      },
    ],
    'tenant-05': [],
  };

  public getSensorListByGateway(gatewayId: string): Observable<Sensor[]> {
    const sensors = this.mockSensors[gatewayId] ?? [];
    return of(sensors).pipe(delay(600));
  }

  public getSensorListByTenant(tenantId: string): Observable<Sensor[]> {
    const sensors = this.tenantSensors[tenantId] ?? [];
    return of(sensors).pipe(delay(800));
  }

  public removeSensor(id: string): Observable<void> {
    for (const gatewayId of Object.keys(this.mockSensors)) {
      const index = this.mockSensors[gatewayId].findIndex((s) => s.id === id);
      if (index !== -1) {
        this.mockSensors[gatewayId].splice(index, 1);
        break;
      }
    }
    return of(void 0).pipe(delay(500));
  }
}
