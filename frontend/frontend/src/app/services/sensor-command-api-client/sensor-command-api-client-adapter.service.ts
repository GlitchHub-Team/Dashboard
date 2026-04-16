import { Observable } from 'rxjs';

export abstract class SensorCommandApiClientAdapter {
  abstract interruptSensor(sensorId: string): Observable<void>;
  abstract resumeSensor(sensorId: string): Observable<void>;
}