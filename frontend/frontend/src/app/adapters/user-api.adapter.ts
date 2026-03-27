import { Injectable } from '@angular/core';
import { UserAdapter } from './user.adapter';
import { UserBackend } from '../models/user/user-backend.model';
import { PaginatedUserResponse } from '../models/user/paginated-user-response.model';
import { User } from '../models/user/user.model';
import { userRoleMapper } from '../utils/user-role.utils';

@Injectable({ providedIn: 'root' })
export class UserApiAdapter extends UserAdapter {
  fromDTO(dto: UserBackend): User {
    return {
      id: dto.user_id,
      username: dto.username,
      email: dto.email,
      role: userRoleMapper.fromBackend(dto.user_role),
      tenantId: dto.tenant_id || '',
    };
  }

  fromPaginatedDTO(response: PaginatedUserResponse<UserBackend>): PaginatedUserResponse<User> {
    return {
      count: response.count,
      total: response.total,
      users: response.users.map((dto) => this.fromDTO(dto)),
    };
  }
}
