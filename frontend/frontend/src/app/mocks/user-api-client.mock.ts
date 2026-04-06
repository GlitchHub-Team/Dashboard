import { Injectable } from '@angular/core';
import { Observable, of, delay } from 'rxjs';
import { UserConfig } from '../models/user/user-config.model';
import { UserRole } from '../models/user/user-role.enum';
import { UserBackend } from '../models/user/user-backend.model';
import { PaginatedUserResponse } from '../models/user/paginated-user-response.model';
import { userRoleMapper } from '../utils/user-role.utils';

@Injectable({ providedIn: 'root' })
export class UserApiClientMockService {
  private mockUsers: UserBackend[] = [
    {
      user_id: 1,
      username: 'Tullio',
      email: 'super@test.com',
      user_role: 'super_admin',
    },
    {
      user_id: 2,
      username: 'Stefano',
      email: 'admin@test.com',
      user_role: 'tenant_admin',
      tenant_id: 'tenant-1',
    },
    {
      user_id: 3,
      username: 'Tullio x Stefano',
      email: 'user@test.com',
      user_role: 'tenant_user',
      tenant_id: 'tenant-1',
    },
    {
      user_id: 4,
      username: 'admin',
      email: 'admin@example.com',
      user_role: 'tenant_admin',
      tenant_id: 'tenant-1',
    },
    {
      user_id: 5,
      username: 'editor',
      email: 'editor@example.com',
      user_role: 'tenant_user',
      tenant_id: 'tenant-1',
    },
    {
      user_id: 6,
      username: 'viewer',
      email: 'viewer@example.com',
      user_role: 'super_admin',
    },
    {
      user_id: 7,
      username: 'supersuper',
      email: 'supersuper@example.com',
      user_role: 'super_admin',
    },
    {
      user_id: 8,
      username: 'meraviglia',
      email: 'meraviglia@example.com',
      user_role: 'super_admin',
    },
    {
      user_id: 9,
      username: 'alice.smith',
      email: 'alice.smith@tenant-1.com',
      user_role: 'tenant_user',
      tenant_id: 'tenant-1',
    },
    {
      user_id: 10,
      username: 'bob.jones',
      email: 'bob.jones@tenant-1.com',
      user_role: 'tenant_user',
      tenant_id: 'tenant-1',
    },
    {
      user_id: 11,
      username: 'charlie.brown',
      email: 'charlie.brown@tenant-2.com',
      user_role: 'tenant_admin',
      tenant_id: 'tenant-2',
    },
    {
      user_id: 12,
      username: 'diana.prince',
      email: 'diana.prince@tenant-2.com',
      user_role: 'tenant_user',
      tenant_id: 'tenant-2',
    },
    {
      user_id: 13,
      username: 'eve.davis',
      email: 'eve.davis@tenant-3.com',
      user_role: 'tenant_admin',
      tenant_id: 'tenant-3',
    },
    {
      user_id: 14,
      username: 'frank.miller',
      email: 'frank.miller@tenant-3.com',
      user_role: 'tenant_user',
      tenant_id: 'tenant-3',
    },
    {
      user_id: 15,
      username: 'grace.hopper',
      email: 'grace.hopper@tenant-3.com',
      user_role: 'tenant_user',
      tenant_id: 'tenant-3',
    },
    {
      user_id: 16,
      username: 'heidi.klum',
      email: 'heidi.klum@tenant-1.com',
      user_role: 'tenant_admin',
      tenant_id: 'tenant-1',
    },
    {
      user_id: 17,
      username: 'ivan.drago',
      email: 'ivan.drago@tenant-2.com',
      user_role: 'tenant_user',
      tenant_id: 'tenant-2',
    },
    {
      user_id: 18,
      username: 'judy.garland',
      email: 'judy.garland@tenant-3.com',
      user_role: 'tenant_user',
      tenant_id: 'tenant-3',
    },
    {
      user_id: 19,
      username: 'kevin.bacon',
      email: 'kevin.bacon@tenant-1.com',
      user_role: 'tenant_user',
      tenant_id: 'tenant-1',
    },
    {
      user_id: 20,
      username: 'laura.croft',
      email: 'laura.croft@tenant-2.com',
      user_role: 'tenant_admin',
      tenant_id: 'tenant-2',
    },
  ];

  public getUsers(
    role: UserRole,
    page = 1,
    limit = 10,
    tenantId?: string,
  ): Observable<PaginatedUserResponse<UserBackend>> {
    const roleString = userRoleMapper.toBackend(role);
    let filteredUsers = [...this.mockUsers];
    if (role) {
      filteredUsers = filteredUsers.filter((user) => user.user_role === roleString);
    }
    if (tenantId) {
      filteredUsers = filteredUsers.filter((user) => user.tenant_id === tenantId);
    }
    const total = filteredUsers.length;
    const start = (page - 1) * limit;
    const users = filteredUsers.slice(start, start + limit);
    return of({ count: users.length, total, users }).pipe(delay(500));
  }

  public getUser(id: string, role: UserRole, tenantId?: string): Observable<UserBackend> {
    const roleString = userRoleMapper.toBackend(role);
    const user = this.mockUsers.find(
      (u) =>
        u.user_id === Number(id) &&
        u.user_role === roleString &&
        (role === UserRole.SUPER_ADMIN || u.tenant_id === tenantId),
    );
    if (!user) {
      throw new Error('User not found');
    }
    return of(user).pipe(delay(500));
  }

  public createUser(
    config: UserConfig,
    role: UserRole,
    tenantId?: string,
  ): Observable<UserBackend> {
    const newId = Math.floor(Math.random() * 10000);

    const newUser: UserBackend = {
      user_id: newId,
      username: config.username || config.email.split('@')[0],
      email: config.email,
      user_role: userRoleMapper.toBackend(role),
    };

    if (role !== UserRole.SUPER_ADMIN) {
      newUser.tenant_id = tenantId || 'mock-tenant-id';
    }

    this.mockUsers.push(newUser);
    return of(newUser).pipe(delay(500));
  }

  public deleteUser(id: string, role: UserRole, tenantId?: string): Observable<void> {
    const roleString = userRoleMapper.toBackend(role);
    this.mockUsers = this.mockUsers.filter(
      (u) =>
        !(
          u.user_id === Number(id) &&
          u.user_role === roleString &&
          (role === UserRole.SUPER_ADMIN || u.tenant_id === tenantId)
        ),
    );
    return of(void 0).pipe(delay(500));
  }
}
