import { Injectable } from '@angular/core';
import { SensorAdapter } from './sensor.adapter';
import { SensorBackend } from '../models/sensor/sensor-backend.model';
import { Sensor } from '../models/sensor/sensor.model';
import { PaginatedResponse } from '../models/paginated-response.model';
import { SensorProfiles } from '../models/sensor/sensor-profiles.enum';
import { Status } from '../models/gateway-sensor-status.enum';

@Injectable()
export class SensorApiAdapter extends SensorAdapter {
  fromDTO(dto: SensorBackend): Sensor {
    return {
      id: dto.sensor_id,
      gatewayId: dto.gateway_id,
      name: dto.sensor_name,
      status: dto.status as Status,
      profile: dto.profile as SensorProfiles,
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
