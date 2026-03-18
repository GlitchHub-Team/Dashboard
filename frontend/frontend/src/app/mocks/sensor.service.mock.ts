// mocks/sensor-service.mock.ts
import { Injectable } from '@angular/core';
import { Observable, of } from 'rxjs';
import { delay } from 'rxjs/operators';

import { SensorBackend } from '../models/sensor/sensor-backend.model';
import { PaginatedResponse } from '../models/paginated-response.model';

@Injectable({
  providedIn: 'root',
})
export class SensorApiClientServiceMock {
  private readonly mockSensors: Record<string, SensorBackend[]> = {
    'gateway-01': [
      {
        SensorId: 'sensor-01',
        GatewayId: 'gateway-01',
        Name: 'Heart Rate Monitor A',
        Profile: 'HEART_RATE_SERVICE',
      },
      {
        SensorId: 'sensor-02',
        GatewayId: 'gateway-01',
        Name: 'Thermometer A',
        Profile: 'HEALTH_THERMOMETER_SERVICE',
      },
      {
        SensorId: 'sensor-03',
        GatewayId: 'gateway-01',
        Name: 'ECG Monitor A',
        Profile: 'CUSTOM_ECG_SERVICE',
      },
      {
        SensorId: 'sensor-04',
        GatewayId: 'gateway-01',
        Name: 'Pulse Oximeter A',
        Profile: 'PULSE_OXIMETER_SERVICE',
      },
      {
        SensorId: 'sensor-05',
        GatewayId: 'gateway-01',
        Name: 'Environment Sensor A',
        Profile: 'ENVIRONMENTAL_SENSING_SERVICE',
      },
      {
        SensorId: 'sensor-06',
        GatewayId: 'gateway-01',
        Name: 'Heart Rate Monitor B',
        Profile: 'HEART_RATE_SERVICE',
      },
      {
        SensorId: 'sensor-07',
        GatewayId: 'gateway-01',
        Name: 'Thermometer B',
        Profile: 'HEALTH_THERMOMETER_SERVICE',
      },
      {
        SensorId: 'sensor-08',
        GatewayId: 'gateway-01',
        Name: 'ECG Monitor B',
        Profile: 'CUSTOM_ECG_SERVICE',
      },
      {
        SensorId: 'sensor-09',
        GatewayId: 'gateway-01',
        Name: 'Pulse Oximeter B',
        Profile: 'PULSE_OXIMETER_SERVICE',
      },
      {
        SensorId: 'sensor-10',
        GatewayId: 'gateway-01',
        Name: 'Environment Sensor B',
        Profile: 'ENVIRONMENTAL_SENSING_SERVICE',
      },
      {
        SensorId: 'sensor-11',
        GatewayId: 'gateway-01',
        Name: 'Heart Rate Monitor C',
        Profile: 'HEART_RATE_SERVICE',
      },
      {
        SensorId: 'sensor-12',
        GatewayId: 'gateway-01',
        Name: 'Thermometer C',
        Profile: 'HEALTH_THERMOMETER_SERVICE',
      },
      {
        SensorId: 'sensor-13',
        GatewayId: 'gateway-01',
        Name: 'ECG Monitor C',
        Profile: 'CUSTOM_ECG_SERVICE',
      },
      {
        SensorId: 'sensor-14',
        GatewayId: 'gateway-01',
        Name: 'Pulse Oximeter C',
        Profile: 'PULSE_OXIMETER_SERVICE',
      },
      {
        SensorId: 'sensor-15',
        GatewayId: 'gateway-01',
        Name: 'Environment Sensor C',
        Profile: 'ENVIRONMENTAL_SENSING_SERVICE',
      },
    ],
    'gateway-02': [
      {
        SensorId: 'sensor-20',
        GatewayId: 'gateway-02',
        Name: 'ICU Heart Rate',
        Profile: 'HEART_RATE_SERVICE',
      },
      {
        SensorId: 'sensor-21',
        GatewayId: 'gateway-02',
        Name: 'ICU Pulse Oximeter',
        Profile: 'PULSE_OXIMETER_SERVICE',
      },
      {
        SensorId: 'sensor-22',
        GatewayId: 'gateway-02',
        Name: 'ICU Thermometer',
        Profile: 'HEALTH_THERMOMETER_SERVICE',
      },
      {
        SensorId: 'sensor-23',
        GatewayId: 'gateway-02',
        Name: 'ICU ECG',
        Profile: 'CUSTOM_ECG_SERVICE',
      },
      {
        SensorId: 'sensor-24',
        GatewayId: 'gateway-02',
        Name: 'ICU Environment',
        Profile: 'ENVIRONMENTAL_SENSING_SERVICE',
      },
    ],
    'gateway-03': [],
    'gateway-04': [
      {
        SensorId: 'sensor-30',
        GatewayId: 'gateway-04',
        Name: 'Ward C Heart Rate',
        Profile: 'HEART_RATE_SERVICE',
      },
    ],
    'gateway-05': [],
  };

  private readonly tenantSensors: Record<string, SensorBackend[]> = {
    'tenant-01': [
      {
        SensorId: 'sensor-01',
        GatewayId: 'gateway-01',
        Name: 'Heart Rate Monitor A',
        Profile: 'HEART_RATE_SERVICE',
      },
      {
        SensorId: 'sensor-02',
        GatewayId: 'gateway-01',
        Name: 'Thermometer A',
        Profile: 'HEALTH_THERMOMETER_SERVICE',
      },
      {
        SensorId: 'sensor-03',
        GatewayId: 'gateway-01',
        Name: 'ECG Monitor A',
        Profile: 'CUSTOM_ECG_SERVICE',
      },
      {
        SensorId: 'sensor-04',
        GatewayId: 'gateway-01',
        Name: 'Pulse Oximeter A',
        Profile: 'PULSE_OXIMETER_SERVICE',
      },
      {
        SensorId: 'sensor-05',
        GatewayId: 'gateway-01',
        Name: 'Environment Sensor A',
        Profile: 'ENVIRONMENTAL_SENSING_SERVICE',
      },
      {
        SensorId: 'sensor-06',
        GatewayId: 'gateway-01',
        Name: 'Heart Rate Monitor B',
        Profile: 'HEART_RATE_SERVICE',
      },
      {
        SensorId: 'sensor-07',
        GatewayId: 'gateway-01',
        Name: 'Thermometer B',
        Profile: 'HEALTH_THERMOMETER_SERVICE',
      },
      {
        SensorId: 'sensor-08',
        GatewayId: 'gateway-01',
        Name: 'ECG Monitor B',
        Profile: 'CUSTOM_ECG_SERVICE',
      },
      {
        SensorId: 'sensor-09',
        GatewayId: 'gateway-01',
        Name: 'Pulse Oximeter B',
        Profile: 'PULSE_OXIMETER_SERVICE',
      },
      {
        SensorId: 'sensor-10',
        GatewayId: 'gateway-01',
        Name: 'Environment Sensor B',
        Profile: 'ENVIRONMENTAL_SENSING_SERVICE',
      },
      {
        SensorId: 'sensor-20',
        GatewayId: 'gateway-02',
        Name: 'ICU Heart Rate',
        Profile: 'HEART_RATE_SERVICE',
      },
      {
        SensorId: 'sensor-21',
        GatewayId: 'gateway-02',
        Name: 'ICU Pulse Oximeter',
        Profile: 'PULSE_OXIMETER_SERVICE',
      },
      {
        SensorId: 'sensor-22',
        GatewayId: 'gateway-02',
        Name: 'ICU Thermometer',
        Profile: 'HEALTH_THERMOMETER_SERVICE',
      },
    ],
    'tenant-02': [
      {
        SensorId: 'sensor-30',
        GatewayId: 'gateway-04',
        Name: 'Ward C Heart Rate',
        Profile: 'HEART_RATE_SERVICE',
      },
      {
        SensorId: 'sensor-31',
        GatewayId: 'gateway-04',
        Name: 'Ward C Thermometer',
        Profile: 'HEALTH_THERMOMETER_SERVICE',
      },
    ],
    'tenant-03': [],
    'tenant-04': [
      {
        SensorId: 'sensor-40',
        GatewayId: 'gateway-40',
        Name: 'Clinic Heart Rate',
        Profile: 'HEART_RATE_SERVICE',
      },
    ],
    'tenant-05': [],
  };

  public getSensorListByGateway(
    gatewayId: string,
    page: number,
    limit: number,
  ): Observable<PaginatedResponse<SensorBackend>> {
    const all = this.mockSensors[gatewayId] ?? [];
    return of(this.paginate(all, page, limit)).pipe(delay(600));
  }

  public getSensorListByTenant(
    tenantId: string,
    page: number,
    limit: number,
  ): Observable<PaginatedResponse<SensorBackend>> {
    const all = this.tenantSensors[tenantId] ?? [];
    return of(this.paginate(all, page, limit)).pipe(delay(800));
  }

  public addNewSensor(config: unknown): Observable<SensorBackend> {
    return of({
      SensorId: `sensor-${Date.now()}`,
      GatewayId: 'gateway-01',
      Name: 'New Sensor',
      Profile: 'HEART_RATE_SERVICE',
    }).pipe(delay(400));
  }

  public deleteSensor(id: string): Observable<void> {
    return of(undefined).pipe(delay(400));
  }

  private paginate<T>(items: T[], page: number, limit: number): PaginatedResponse<T> {
    const start = page * limit;
    const data = items.slice(start, start + limit);

    return {
      count: data.length,
      total: items.length,
      data,
    };
  }
}
