export interface PaginatedResponse<T> {
  count: number;
  total: number;
  data: T[];
}
