// mocks/sensor-service.mock.ts
import { Injectable } from '@angular/core';
import { Observable, of, throwError } from 'rxjs';
import { delay } from 'rxjs/operators';

import { SensorBackend } from '../models/sensor/sensor-backend.model';
import { SensorConfig } from '../models/sensor/sensor-config.model';
import { PaginatedSensorResponse } from '../models/sensor/paginated-sensor-response.model';
import { HttpErrorResponse } from '@angular/common/http';

@Injectable({
  providedIn: 'root',
})
export class SensorApiClientServiceMock {
  private readonly mockSensors = new Map<string, SensorBackend[]>([
    [
      'gateway-01',
      [
        {
          sensor_id: 'sensor-01',
          gateway_id: 'gateway-01',
          sensor_name: 'Heart Rate Monitor A',
          data_interval: 60,
          profile: 'heart_rate',
          status: 'inactive',
        },
        {
          sensor_id: 'sensor-02',
          gateway_id: 'gateway-01',
          sensor_name: 'Thermometer A',
          data_interval: 60,
          profile: 'health_thermometer',
          status: 'active',
        },
        {
          sensor_id: 'sensor-03',
          gateway_id: 'gateway-01',
          sensor_name: 'ECG Monitor A',
          data_interval: 60,
          profile: 'custom_ecg',
          status: 'active',
        },
        {
          sensor_id: 'sensor-04',
          gateway_id: 'gateway-01',
          sensor_name: 'Pulse Oximeter A',
          data_interval: 60,
          profile: 'pulse_oximeter',
          status: 'inactive',
        },
        {
          sensor_id: 'sensor-05',
          gateway_id: 'gateway-01',
          sensor_name: 'Environment Sensor A',
          data_interval: 60,
          profile: 'environmental_sensing',
          status: 'active',
        },
        {
          sensor_id: 'sensor-06',
          gateway_id: 'gateway-01',
          sensor_name: 'Heart Rate Monitor B',
          data_interval: 60,
          profile: 'heart_rate',
          status: 'active',
        },
        {
          sensor_id: 'sensor-07',
          gateway_id: 'gateway-01',
          sensor_name: 'Thermometer B',
          data_interval: 60,
          profile: 'health_thermometer',
          status: 'inactive',
        },
        {
          sensor_id: 'sensor-08',
          gateway_id: 'gateway-01',
          sensor_name: 'ECG Monitor B',
          data_interval: 60,
          profile: 'custom_ecg',
          status: 'active',
        },
        {
          sensor_id: 'sensor-09',
          gateway_id: 'gateway-01',
          sensor_name: 'Pulse Oximeter B',
          data_interval: 60,
          profile: 'pulse_oximeter',
          status: 'active',
        },
        {
          sensor_id: 'sensor-10',
          gateway_id: 'gateway-01',
          sensor_name: 'Environment Sensor B',
          data_interval: 60,
          profile: 'environmental_sensing',
          status: 'active',
        },
        {
          sensor_id: 'sensor-11',
          gateway_id: 'gateway-01',
          sensor_name: 'Heart Rate Monitor C',
          data_interval: 60,
          profile: 'heart_rate',
          status: 'active',
        },
        {
          sensor_id: 'sensor-12',
          gateway_id: 'gateway-01',
          sensor_name: 'Thermometer C',
          data_interval: 60,
          profile: 'health_thermometer',
          status: 'active',
        },
        {
          sensor_id: 'sensor-13',
          gateway_id: 'gateway-01',
          sensor_name: 'ECG Monitor C',
          data_interval: 60,
          profile: 'custom_ecg',
          status: 'active',
        },
        {
          sensor_id: 'sensor-14',
          gateway_id: 'gateway-01',
          sensor_name: 'Pulse Oximeter C',
          data_interval: 60,
          profile: 'pulse_oximeter',
          status: 'active',
        },
        {
          sensor_id: 'sensor-15',
          gateway_id: 'gateway-01',
          sensor_name: 'Environment Sensor C',
          data_interval: 60,
          profile: 'environmental_sensing',
          status: 'active',
        },
      ],
    ],
    [
      'gateway-02',
      [
        {
          sensor_id: 'sensor-20',
          gateway_id: 'gateway-02',
          sensor_name: 'ICU Heart Rate',
          data_interval: 60,
          profile: 'heart_rate',
          status: 'active',
        },
        {
          sensor_id: 'sensor-21',
          gateway_id: 'gateway-02',
          sensor_name: 'ICU Pulse Oximeter',
          data_interval: 60,
          profile: 'pulse_oximeter',
          status: 'active',
        },
        {
          sensor_id: 'sensor-22',
          gateway_id: 'gateway-02',
          sensor_name: 'ICU Thermometer',
          data_interval: 60,
          profile: 'health_thermometer',
          status: 'active',
        },
        {
          sensor_id: 'sensor-23',
          gateway_id: 'gateway-02',
          sensor_name: 'ICU ECG',
          data_interval: 60,
          profile: 'ecg_custom',
          status: 'active',
        },
        {
          sensor_id: 'sensor-24',
          gateway_id: 'gateway-02',
          sensor_name: 'ICU Environment',
          data_interval: 60,
          profile: 'environmental_sensing',
          status: 'active',
        },
      ],
    ],
    ['gateway-03', []],
    [
      'gateway-04',
      [
        {
          sensor_id: 'sensor-30',
          gateway_id: 'gateway-04',
          sensor_name: 'Ward C Heart Rate',
          data_interval: 60,
          profile: 'heart_rate',
          status: 'active',
        },
      ],
    ],
    ['gateway-05', []],
  ]);

  private readonly tenantGatewayMap: Record<string, string[]> = {
    'tenant-1': ['gateway-01', 'gateway-02'],
    'tenant-2': ['gateway-04'],
    'tenant-3': [],
    'tenant-4': [],
    'tenant-5': [],
  };

  public getSensorListByGateway(
    gatewayId: string,
    page: number,
    limit: number,
  ): Observable<PaginatedSensorResponse<SensorBackend>> {
    const shouldFail = false;

    if (shouldFail) {
      return throwError(
        () =>
          new HttpErrorResponse({
            status: 400,
            statusText: 'Bad Request',
            error: { error: 'tenant already exists' },
          }),
      ).pipe(delay(500));
    }
    const all = this.mockSensors.get(gatewayId) ?? [];
    return of(this.paginate(all, page, limit)).pipe(delay(600));
  }

  public getSensorListByTenant(
    tenantId: string,
    page: number,
    limit: number,
  ): Observable<PaginatedSensorResponse<SensorBackend>> {
    const gatewayIds = this.tenantGatewayMap[tenantId] ?? [];
    const all = gatewayIds.flatMap((id) => this.mockSensors.get(id) ?? []);
    return of(this.paginate(all, page, limit)).pipe(delay(800));
  }

  public addNewSensor(config: SensorConfig): Observable<SensorBackend> {
    const gatewaySensors = this.mockSensors.get(config.gatewayId);

    if (!gatewaySensors) {
      return throwError(() => ({
        status: 404,
        message: `Gateway ${config.gatewayId} not found`,
      })).pipe(delay(400));
    }

    const newSensor: SensorBackend = {
      sensor_id: `sensor-${Date.now()}`,
      gateway_id: config.gatewayId,
      sensor_name: config.name,
      data_interval: config.dataInterval,
      profile: config.profile,
      status: 'attivo',
    };

    gatewaySensors.push(newSensor);

    return of(newSensor).pipe(delay(400));
  }

  public deleteSensor(sensorId: string): Observable<void> {
    let found = false;

    for (const [gatewayId, sensors] of this.mockSensors) {
      const index = sensors.findIndex((s) => s.sensor_id === sensorId);
      if (index !== -1) {
        sensors.splice(index, 1);
        found = true;
        break;
      }
    }

    if (!found) {
      return throwError(() => ({
        status: 404,
        message: `Sensor ${sensorId} not found`,
      })).pipe(delay(400));
    }

    return of(undefined).pipe(delay(400));
  }

  private paginate<T>(items: T[], page: number, limit: number): PaginatedSensorResponse<T> {
    const start = (page - 1) * limit;
    const sensors = items.slice(start, start + limit);

    return {
      count: sensors.length,
      total: items.length,
      sensors,
    };
  }
}
