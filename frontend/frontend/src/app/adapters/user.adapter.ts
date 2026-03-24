import { User } from '../models/user/user.model';
import { PaginatedResponse } from '../models/paginated-response.model';

export abstract class UserAdapter {
  abstract fromDTO(dto: unknown): User;
  abstract fromPaginatedDTO(response: PaginatedResponse<unknown>): PaginatedResponse<User>;
}
