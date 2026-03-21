import { Injectable } from '@angular/core';

import { SensorHistoricAdapter } from './sensor-historic.adapter';
import { HistoricDataPoint } from '../models/sensor-data/historic-data-point.model';
import { HistoricResponse } from '../models/sensor-data/historic-response.model';
import { SensorReading } from '../models/sensor-data/sensor-reading.model';
import { HistoricReadings } from '../models/sensor-data/historic-readings.model';

@Injectable()
export class SensorHistoricApiAdapter extends SensorHistoricAdapter {
  fromDTO(dto: HistoricDataPoint): SensorReading {
    return {
      value: dto.value,
      timestamp: new Date(dto.timestamp).toISOString(),
    };
  }

  fromResponse(response: HistoricResponse): HistoricReadings {
    return {
      resolution: response.resolution,
      readings: response.data.map((dto) => this.fromDTO(dto)),
    };
  }
}
