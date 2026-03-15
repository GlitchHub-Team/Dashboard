import { Injectable } from '@angular/core';
import { Observable, of } from 'rxjs';
import { delay } from 'rxjs/operators';

//import { SensorConfig } from '../models/sensor-config.model';
import { Sensor } from '../models/sensor.model';

@Injectable({
  providedIn: 'root',
})
export class SensorServiceMock {
  private readonly mockSensors: Record<string, Sensor[]> = {
    'gateway-01': [
      {
        id: 'sensor-01',
      },
      {
        id: 'sensor-02',
      },
      {
        id: 'sensor-03',
      },
    ],
    'gateway-02': [
      {
        id: 'sensor-04',
      },
      {
        id: 'sensor-05',
      },
    ],
    'gateway-03': [],
    'gateway-04': [
      {
        id: 'sensor-06',
      },
    ],
    'gateway-05': [],
  };

  private readonly tenantSensors: Record<string, Sensor[]> = {
    'tenant-01': [
      {
        id: 'sensor-01',
      },
      {
        id: 'sensor-02',
      },
      {
        id: 'sensor-03',
      },
    ],
    'tenant-02': [
      {
        id: 'sensor-04',
      },
      {
        id: 'sensor-05',
      },
    ],
    'tenant-03': [],
    'tenant-04': [
      {
        id: 'sensor-06',
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

  /*   public addNewSensor(config: SensorConfig): Observable<Sensor> {
    const newSensor: Sensor = {
      id: config.id,
    };
    const gatewaySensors = this.mockSensors[config.gatewayId] ?? [];
    gatewaySensors.push(newSensor);
    this.mockSensors[config.gatewayId] = gatewaySensors;
    return of(newSensor).pipe(delay(500));
  } */

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
