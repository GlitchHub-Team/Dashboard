import { User } from '../models/user/user.model';
import { PaginatedUserResponse } from '../models/user/paginated-user-response.model';
import { UserBackend } from '../models/user/user-backend.model';

export abstract class UserAdapter {
  abstract fromDTO(dto: UserBackend): User;
  abstract fromPaginatedDTO(
    response: PaginatedUserResponse<UserBackend>,
  ): PaginatedUserResponse<User>;
}
