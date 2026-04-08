import { PaginatedResponse } from '../paginated-response.model';

export interface PaginatedSensorResponse<T> extends PaginatedResponse {
  sensors: T[];
}
