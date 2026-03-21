// mocks/sensor-service.mock.ts
import { Injectable } from '@angular/core';
import { Observable, of } from 'rxjs';
import { delay } from 'rxjs/operators';

import { SensorBackend } from '../models/sensor/sensor-backend.model';
import { PaginatedResponse } from '../models/paginated-response.model';
import { SensorProfiles } from '../models/sensor/sensor-profiles.enum';

@Injectable({
  providedIn: 'root',
})
export class SensorApiClientServiceMock {
  private readonly mockSensors: Record<string, SensorBackend[]> = {
    'gateway-01': [
      {
        sensor_id: 'sensor-01',
        gateway_id: 'gateway-01',
        sensor_name: 'Heart Rate Monitor A',
        sensor_interval: 60,
        profile: SensorProfiles.HEART_RATE_SERVICE,
        status: 'active',
      },
      {
        sensor_id: 'sensor-02',
        gateway_id: 'gateway-01',
        sensor_name: 'Thermometer A',
        sensor_interval: 60,
        profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
        status: 'active',
      },
      {
        sensor_id: 'sensor-03',
        gateway_id: 'gateway-01',
        sensor_name: 'ECG Monitor A',
        sensor_interval: 60,
        profile: SensorProfiles.CUSTOM_ECG_SERVICE,
        status: 'active',
      },
      {
        sensor_id: 'sensor-04',
        gateway_id: 'gateway-01',
        sensor_name: 'Pulse Oximeter A',
        sensor_interval: 60,
        profile: SensorProfiles.PULSE_OXIMETER_SERVICE,
        status: 'active',
      },
      {
        sensor_id: 'sensor-05',
        gateway_id: 'gateway-01',
        sensor_name: 'Environment Sensor A',
        sensor_interval: 60,
        profile: SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE,
        status: 'active',
      },
      {
        sensor_id: 'sensor-06',
        gateway_id: 'gateway-01',
        sensor_name: 'Heart Rate Monitor B',
        sensor_interval: 60,
        profile: SensorProfiles.HEART_RATE_SERVICE,
        status: 'active',
      },
      {
        sensor_id: 'sensor-07',
        gateway_id: 'gateway-01',
        sensor_name: 'Thermometer B',
        sensor_interval: 60,
        profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
        status: 'active',
      },
      {
        sensor_id: 'sensor-08',
        gateway_id: 'gateway-01',
        sensor_name: 'ECG Monitor B',
        sensor_interval: 60,
        profile: SensorProfiles.CUSTOM_ECG_SERVICE,
        status: 'active',
      },
      {
        sensor_id: 'sensor-09',
        gateway_id: 'gateway-01',
        sensor_name: 'Pulse Oximeter B',
        sensor_interval: 60,
        profile: SensorProfiles.PULSE_OXIMETER_SERVICE,
        status: 'active',
      },
      {
        sensor_id: 'sensor-10',
        gateway_id: 'gateway-01',
        sensor_name: 'Environment Sensor B',
        sensor_interval: 60,
        profile: SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE,
        status: 'active',
      },
      {
        sensor_id: 'sensor-11',
        gateway_id: 'gateway-01',
        sensor_name: 'Heart Rate Monitor C',
        sensor_interval: 60,
        profile: SensorProfiles.HEART_RATE_SERVICE,
        status: 'active',
      },
      {
        sensor_id: 'sensor-12',
        gateway_id: 'gateway-01',
        sensor_name: 'Thermometer C',
        sensor_interval: 60,
        profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
        status: 'active',
      },
      {
        sensor_id: 'sensor-13',
        gateway_id: 'gateway-01',
        sensor_name: 'ECG Monitor C',
        sensor_interval: 60,
        profile: SensorProfiles.CUSTOM_ECG_SERVICE,
        status: 'active',
      },
      {
        sensor_id: 'sensor-14',
        gateway_id: 'gateway-01',
        sensor_name: 'Pulse Oximeter C',
        sensor_interval: 60,
        profile: SensorProfiles.PULSE_OXIMETER_SERVICE,
        status: 'active',
      },
      {
        sensor_id: 'sensor-15',
        gateway_id: 'gateway-01',
        sensor_name: 'Environment Sensor C',
        sensor_interval: 60,
        profile: SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE,
        status: 'active',
      },
    ],
    'gateway-02': [
      {
        sensor_id: 'sensor-20',
        gateway_id: 'gateway-02',
        sensor_name: 'ICU Heart Rate',
        sensor_interval: 60,
        profile: SensorProfiles.HEART_RATE_SERVICE,
        status: 'active',
      },
      {
        sensor_id: 'sensor-21',
        gateway_id: 'gateway-02',
        sensor_name: 'ICU Pulse Oximeter',
        sensor_interval: 60,
        profile: SensorProfiles.PULSE_OXIMETER_SERVICE,
        status: 'active',
      },
      {
        sensor_id: 'sensor-22',
        gateway_id: 'gateway-02',
        sensor_name: 'ICU Thermometer',
        sensor_interval: 60,
        profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
        status: 'active',
      },
      {
        sensor_id: 'sensor-23',
        gateway_id: 'gateway-02',
        sensor_name: 'ICU ECG',
        sensor_interval: 60,
        profile: SensorProfiles.CUSTOM_ECG_SERVICE,
        status: 'active',
      },
      {
        sensor_id: 'sensor-24',
        gateway_id: 'gateway-02',
        sensor_name: 'ICU Environment',
        sensor_interval: 60,
        profile: SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE,
        status: 'active',
      },
    ],
    'gateway-03': [],
    'gateway-04': [
      {
        sensor_id: 'sensor-30',
        gateway_id: 'gateway-04',
        sensor_name: 'Ward C Heart Rate',
        sensor_interval: 60,
        profile: SensorProfiles.HEART_RATE_SERVICE,
        status: 'active',
      },
    ],
    'gateway-05': [],
  };

  private readonly tenantSensors: Record<string, SensorBackend[]> = {
    'tenant-01': [
      {
        sensor_id: 'sensor-01',
        gateway_id: 'gateway-01',
        sensor_name: 'Heart Rate Monitor A',
        sensor_interval: 60,
        profile: SensorProfiles.HEART_RATE_SERVICE,
        status: 'active',
      },
      {
        sensor_id: 'sensor-02',
        gateway_id: 'gateway-01',
        sensor_name: 'Thermometer A',
        sensor_interval: 60,
        profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
        status: 'active',
      },
      {
        sensor_id: 'sensor-03',
        gateway_id: 'gateway-01',
        sensor_name: 'ECG Monitor A',
        sensor_interval: 60,
        profile: SensorProfiles.CUSTOM_ECG_SERVICE,
        status: 'active',
      },
      {
        sensor_id: 'sensor-04',
        gateway_id: 'gateway-01',
        sensor_name: 'Pulse Oximeter A',
        sensor_interval: 60,
        profile: SensorProfiles.PULSE_OXIMETER_SERVICE,
        status: 'active',
      },
      {
        sensor_id: 'sensor-05',
        gateway_id: 'gateway-01',
        sensor_name: 'Environment Sensor A',
        sensor_interval: 60,
        profile: SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE,
        status: 'active',
      },
      {
        sensor_id: 'sensor-06',
        gateway_id: 'gateway-01',
        sensor_name: 'Heart Rate Monitor B',
        sensor_interval: 60,
        profile: SensorProfiles.HEART_RATE_SERVICE,
        status: 'active',
      },
      {
        sensor_id: 'sensor-07',
        gateway_id: 'gateway-01',
        sensor_name: 'Thermometer B',
        sensor_interval: 60,
        profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
        status: 'active',
      },
      {
        sensor_id: 'sensor-08',
        gateway_id: 'gateway-01',
        sensor_name: 'ECG Monitor B',
        sensor_interval: 60,
        profile: SensorProfiles.CUSTOM_ECG_SERVICE,
        status: 'active',
      },
      {
        sensor_id: 'sensor-09',
        gateway_id: 'gateway-01',
        sensor_name: 'Pulse Oximeter B',
        sensor_interval: 60,
        profile: SensorProfiles.PULSE_OXIMETER_SERVICE,
        status: 'active',
      },
      {
        sensor_id: 'sensor-10',
        gateway_id: 'gateway-01',
        sensor_name: 'Environment Sensor B',
        sensor_interval: 60,
        profile: SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE,
        status: 'active',
      },
      {
        sensor_id: 'sensor-20',
        gateway_id: 'gateway-02',
        sensor_name: 'ICU Heart Rate',
        sensor_interval: 60,
        profile: SensorProfiles.HEART_RATE_SERVICE,
        status: 'active',
      },
      {
        sensor_id: 'sensor-21',
        gateway_id: 'gateway-02',
        sensor_name: 'ICU Pulse Oximeter',
        sensor_interval: 60,
        profile: SensorProfiles.PULSE_OXIMETER_SERVICE,
        status: 'active',
      },
      {
        sensor_id: 'sensor-22',
        gateway_id: 'gateway-02',
        sensor_name: 'ICU Thermometer',
        sensor_interval: 60,
        profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
        status: 'active',
      },
    ],
    'tenant-02': [
      {
        sensor_id: 'sensor-30',
        gateway_id: 'gateway-04',
        sensor_name: 'Ward C Heart Rate',
        sensor_interval: 60,
        profile: SensorProfiles.HEART_RATE_SERVICE,
        status: 'active',
      },
      {
        sensor_id: 'sensor-31',
        gateway_id: 'gateway-04',
        sensor_name: 'Ward C Thermometer',
        sensor_interval: 60,
        profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
        status: 'active',
      },
    ],
    'tenant-03': [],
    'tenant-04': [
      {
        sensor_id: 'sensor-40',
        gateway_id: 'gateway-40',
        sensor_name: 'Clinic Heart Rate',
        sensor_interval: 60,
        profile: SensorProfiles.HEART_RATE_SERVICE,
        status: 'active',
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
      sensor_id: `sensor-${Date.now()}`,
      gateway_id: 'gateway-01',
      sensor_name: 'New Sensor',
      profile: SensorProfiles.HEART_RATE_SERVICE,
      sensor_interval: 60,
      status: 'active',
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
