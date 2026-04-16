import { Injectable } from '@angular/core';
import { SensorAdapter } from './sensor.adapter';
import { SensorBackend } from '../../models/sensor/sensor-backend.model';
import { Sensor } from '../../models/sensor/sensor.model';
import { sensorStatusMapper } from '../../utils/sensor-status.utils';
import { sensorProfilesMapper } from '../../utils/sensor-profile.utils';
import { PaginatedSensorResponse } from '../../models/sensor/paginated-sensor-response.model';

@Injectable({ providedIn: 'root' })
export class SensorApiAdapter extends SensorAdapter {
  fromDTO(dto: SensorBackend): Sensor {
    return {
      id: dto.sensor_id,
      gatewayId: dto.gateway_id,
      name: dto.sensor_name,
      status: sensorStatusMapper.fromBackend(dto.status),
      profile: sensorProfilesMapper.fromBackend(dto.profile),
      dataInterval: dto.data_interval,
    };
  }

  fromPaginatedDTO(
    response: PaginatedSensorResponse<SensorBackend>,
  ): PaginatedSensorResponse<Sensor> {
    return {
      count: response.count,
      total: response.total,
      sensors: response.sensors.map((dto) => this.fromDTO(dto)),
    };
  }
}
