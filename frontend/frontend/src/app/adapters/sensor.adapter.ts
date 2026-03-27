import { Sensor } from '../models/sensor/sensor.model';
import { PaginatedSensorResponse } from '../models/sensor/paginated-sensor-response.model';
import { SensorBackend } from '../models/sensor/sensor-backend.model';

export abstract class SensorAdapter {
  abstract fromDTO(dto: SensorBackend): Sensor;
  abstract fromPaginatedDTO(
    response: PaginatedSensorResponse<SensorBackend>,
  ): PaginatedSensorResponse<Sensor>;
}
