// mocks/sensor-service.mock.ts
import { Injectable } from '@angular/core';
import { Observable, of, throwError } from 'rxjs';
import { delay } from 'rxjs/operators';

import { SensorBackend } from '../models/sensor/sensor-backend.model';
import { SensorConfig } from '../models/sensor/sensor-config.model';
import { PaginatedSensorResponse } from '../models/sensor/paginated-sensor-response.model';
import { SensorProfiles } from '../models/sensor/sensor-profiles.enum';

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
          sensor_interval: 60,
          profile: 'heart_rate_service',
          status: 'active',
        },
        {
          sensor_id: 'sensor-02',
          gateway_id: 'gateway-01',
          sensor_name: 'Thermometer A',
          sensor_interval: 60,
          profile: 'health_thermometer_service',
          status: 'active',
        },
        {
          sensor_id: 'sensor-03',
          gateway_id: 'gateway-01',
          sensor_name: 'ECG Monitor A',
          sensor_interval: 60,
          profile: 'custom_ecg_service',
          status: 'active',
        },
        {
          sensor_id: 'sensor-04',
          gateway_id: 'gateway-01',
          sensor_name: 'Pulse Oximeter A',
          sensor_interval: 60,
          profile: 'pulse_oximeter_service',
          status: 'active',
        },
        {
          sensor_id: 'sensor-05',
          gateway_id: 'gateway-01',
          sensor_name: 'Environment Sensor A',
          sensor_interval: 60,
          profile: 'environmental_sensing_service',
          status: 'active',
        },
        {
          sensor_id: 'sensor-06',
          gateway_id: 'gateway-01',
          sensor_name: 'Heart Rate Monitor B',
          sensor_interval: 60,
          profile: 'heart_rate_service',
          status: 'active',
        },
        {
          sensor_id: 'sensor-07',
          gateway_id: 'gateway-01',
          sensor_name: 'Thermometer B',
          sensor_interval: 60,
          profile: 'health_thermometer_service',
          status: 'active',
        },
        {
          sensor_id: 'sensor-08',
          gateway_id: 'gateway-01',
          sensor_name: 'ECG Monitor B',
          sensor_interval: 60,
          profile: 'custom_ecg_service',
          status: 'active',
        },
        {
          sensor_id: 'sensor-09',
          gateway_id: 'gateway-01',
          sensor_name: 'Pulse Oximeter B',
          sensor_interval: 60,
          profile: 'pulse_oximeter_service',
          status: 'active',
        },
        {
          sensor_id: 'sensor-10',
          gateway_id: 'gateway-01',
          sensor_name: 'Environment Sensor B',
          sensor_interval: 60,
          profile: 'environmental_sensing_service',
          status: 'active',
        },
        {
          sensor_id: 'sensor-11',
          gateway_id: 'gateway-01',
          sensor_name: 'Heart Rate Monitor C',
          sensor_interval: 60,
          profile: 'heart_rate_service',
          status: 'active',
        },
        {
          sensor_id: 'sensor-12',
          gateway_id: 'gateway-01',
          sensor_name: 'Thermometer C',
          sensor_interval: 60,
          profile: 'health_thermometer_service',
          status: 'active',
        },
        {
          sensor_id: 'sensor-13',
          gateway_id: 'gateway-01',
          sensor_name: 'ECG Monitor C',
          sensor_interval: 60,
          profile: 'custom_ecg_service',
          status: 'active',
        },
        {
          sensor_id: 'sensor-14',
          gateway_id: 'gateway-01',
          sensor_name: 'Pulse Oximeter C',
          sensor_interval: 60,
          profile: 'pulse_oximeter_service',
          status: 'active',
        },
        {
          sensor_id: 'sensor-15',
          gateway_id: 'gateway-01',
          sensor_name: 'Environment Sensor C',
          sensor_interval: 60,
          profile: 'environmental_sensing_service',
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
          sensor_interval: 60,
          profile: 'heart_rate_service',
          status: 'active',
        },
        {
          sensor_id: 'sensor-21',
          gateway_id: 'gateway-02',
          sensor_name: 'ICU Pulse Oximeter',
          sensor_interval: 60,
          profile: 'pulse_oximeter_service',
          status: 'active',
        },
        {
          sensor_id: 'sensor-22',
          gateway_id: 'gateway-02',
          sensor_name: 'ICU Thermometer',
          sensor_interval: 60,
          profile: 'health_thermometer_service',
          status: 'active',
        },
        {
          sensor_id: 'sensor-23',
          gateway_id: 'gateway-02',
          sensor_name: 'ICU ECG',
          sensor_interval: 60,
          profile: 'custom_ecg_service',
          status: 'active',
        },
        {
          sensor_id: 'sensor-24',
          gateway_id: 'gateway-02',
          sensor_name: 'ICU Environment',
          sensor_interval: 60,
          profile: 'environmental_sensing_service',
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
          sensor_interval: 60,
          profile: 'heart_rate_service',
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
      sensor_interval: config.dataInterval,
      profile: SensorProfiles[config.profile as keyof typeof SensorProfiles],
      status: 'active',
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
    const start = page * limit;
    const sensors = items.slice(start, start + limit);

    return {
      count: sensors.length,
      total: items.length,
      sensors,
    };
  }
}
