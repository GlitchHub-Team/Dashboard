import { Injectable } from '@angular/core';
import { SensorAdapter } from './sensor.adapter';
import { SensorBackend } from '../models/sensor/sensor-backend.model';
import { Sensor } from '../models/sensor/sensor.model';
import { PaginatedResponse } from '../models/paginated-response.model';
import { SensorProfiles } from '../models/sensor/sensor-profiles.enum';

@Injectable()
export class SensorApiAdapter extends SensorAdapter {
  fromDTO(dto: SensorBackend): Sensor {
    return {
      id: dto.SensorId,
      gatewayId: dto.GatewayId,
      name: dto.Name,
      profile: dto.Profile as SensorProfiles,
      ...(dto.DataInterval != null && { dataInterval: dto.DataInterval }),
    };
  }

  toDTO(sensor: Partial<Sensor>): Partial<SensorBackend> {
    return {
      Name: sensor.name,
      GatewayId: sensor.gatewayId,
      Profile: sensor.profile,
      // TODO: Aggiungere DataInterval
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
