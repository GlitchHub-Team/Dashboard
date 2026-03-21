import { Sensor } from '../models/sensor/sensor.model';
import { PaginatedResponse } from '../models/paginated-response.model';

export abstract class SensorAdapter {
  abstract fromDTO(dto: unknown): Sensor;
  abstract fromPaginatedDTO(response: PaginatedResponse<unknown>): PaginatedResponse<Sensor>;
}
