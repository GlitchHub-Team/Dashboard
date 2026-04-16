import { Observable } from 'rxjs';

import { Sensor } from '../../models/sensor/sensor.model';
import { SensorConfig } from '../../models/sensor/sensor-config.model';
import { PaginatedSensorResponse } from '../../models/sensor/paginated-sensor-response.model';

export abstract class SensorApiClientAdapter {
  abstract getSensorListByGateway(
    gatewayId: string,
    page: number,
    limit: number,
  ): Observable<PaginatedSensorResponse<Sensor>>;

  abstract getSensorListByTenant(
    tenantId: string,
    page: number,
    limit: number,
  ): Observable<PaginatedSensorResponse<Sensor>>;

  abstract addNewSensor(config: SensorConfig): Observable<Sensor>;

  abstract deleteSensor(sensorId: string): Observable<void>;
}