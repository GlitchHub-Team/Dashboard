import { Injectable } from '@angular/core';
import { UserAdapter } from './user.adapter';
import { UserBackend } from '../models/user/user-backend.model';
import { PaginatedResponse } from '../models/paginated-response.model';
import { User } from '../models/user/user.model';
import { userRoleMapper } from '../utils/user-role.utils';

@Injectable({ providedIn: 'root' })
export class UserApiAdapter extends UserAdapter {
  fromDTO(dto: UserBackend): User {
    return {
      id: dto.id,
      username: dto.username,
      email: dto.email,
      role: userRoleMapper.fromBackend(dto.role),
      tenantId: dto.tenantId || '',
    };
  }

  fromPaginatedDTO(response: PaginatedResponse<UserBackend>): PaginatedResponse<User> {
    return {
      count: response.count,
      total: response.total,
      data: response.data.map((dto) => this.fromDTO(dto)),
    };
  }
}
