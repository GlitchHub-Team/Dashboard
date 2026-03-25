import { Injectable } from '@angular/core';
import { SensorAdapter } from './sensor.adapter';
import { SensorBackend } from '../models/sensor/sensor-backend.model';
import { Sensor } from '../models/sensor/sensor.model';
import { PaginatedResponse } from '../models/paginated-response.model';
import { statusMapper } from '../utils/status.utils';
import { sensorProfilesMapper } from '../utils/sensor-profile.utils';

@Injectable()
export class SensorApiAdapter extends SensorAdapter {
  fromDTO(dto: SensorBackend): Sensor {
    return {
      id: dto.sensor_id,
      gatewayId: dto.gateway_id,
      name: dto.sensor_name,
      status: statusMapper.fromBackend(dto.status),
      profile: sensorProfilesMapper.fromBackend(dto.profile),
      dataInterval: dto.sensor_interval,
    };
  }

  fromPaginatedDTO(response: PaginatedResponse<SensorBackend>): PaginatedResponse<Sensor> {
    return {
      count: response.count,
      total: response.total,
      data: response.data.map((dto) => this.fromDTO(dto)),
    };
  }
}
