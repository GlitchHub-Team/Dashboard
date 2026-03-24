import { Injectable } from '@angular/core';
import { UserAdapter } from './user.adapter';
import { UserBackend } from '../models/user/user-backend.model';
import { PaginatedResponse } from '../models/paginated-response.model';
import { User } from '../models/user/user.model';
import { UserRole } from '../models/user/user-role.enum';

@Injectable({ providedIn: 'root' })
export class UserApiAdapter extends UserAdapter {
  fromDTO(dto: UserBackend): User {
    return {
      id: dto.id,
      username: dto.username,
      email: dto.email,
      role: dto.role as UserRole,
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
