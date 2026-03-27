import { Injectable } from '@angular/core';

import { SensorLiveReadingAdapter } from './sensor-live-reading.adapter';
import { SensorReading } from '../models/sensor-data/sensor-reading.model';
import { RealTimeReading } from '../models/sensor-data/real-time-reading.model';

@Injectable()
export class SensorLiveReadingApiAdapter extends SensorLiveReadingAdapter {
  fromDTO(dto: RealTimeReading): SensorReading {
    return {
      value: dto.datum,
      timestamp: new Date(dto.timestamp).toISOString(),
    };
  }
}
