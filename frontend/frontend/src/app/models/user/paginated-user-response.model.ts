import { PaginatedResponse } from '../paginated-response.model';

export interface PaginatedUserResponse<T> extends PaginatedResponse {
  users: T[];
}
