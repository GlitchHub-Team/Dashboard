import { Injectable } from '@angular/core';

import { SensorHistoricAdapter } from './sensor-historic.adapter';
import { HistoricResponse } from '../../models/sensor-data/historic-response.model';
import { SensorReading } from '../../models/sensor-data/sensor-reading.model';
import { HistoricReadings } from '../../models/sensor-data/historic-readings.model';

@Injectable()
export class SensorHistoricApiAdapter extends SensorHistoricAdapter {
  fromDTO(value: number, timestamp: number): SensorReading {
    return {
      value: value,
      timestamp: new Date(timestamp).toISOString(),
    };
  }

  fromResponse(response: HistoricResponse): HistoricReadings {
    return {
      dataCount: response.count.current,
      readings: response.dataset.timestamps.map((timestamp, index) =>
        this.fromDTO(response.dataset.values[index], timestamp),
      ),
    };
  }
}
