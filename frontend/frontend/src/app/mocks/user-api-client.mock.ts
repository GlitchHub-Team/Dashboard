import { Injectable } from '@angular/core';
import { Observable, of, delay } from 'rxjs';
import { UserConfig } from '../models/user/user-config.model';
import { UserRole } from '../models/user/user-role.enum';
import { UserBackend } from '../models/user/user-backend.model';
import { PaginatedResponse } from '../models/paginated-response.model';
import { userRoleMapper } from '../utils/user-role.utils';

@Injectable({ providedIn: 'root' })
export class UserApiClientMockService {
  private mockUsers: UserBackend[] = [
    {
      id: 'user-5741',
      username: 'admin',
      email: 'admin@example.com',
      role: 'tenant_admin',
      tenantId: 'tenant-01',
    },
    {
      id: 'user-3592',
      username: 'editor',
      email: 'editor@example.com',
      role: 'tenant_user',
      tenantId: 'tenant-01',
    },
    {
      id: 'user-3876',
      username: 'viewer',
      email: 'viewer@example.com',
      role: 'super_admin',
      tenantId: 'tenant-02',
    },
    {
      id: 'user-6779',
      username: 'supersuper',
      email: 'supersuper@example.com',
      role: 'super_admin',
      tenantId: 'tenant-01',
    },
    {
      id: 'user-4160',
      username: 'alice.smith',
      email: 'alice.smith@tenant-1.com',
      role: 'tenant_user',
      tenantId: 'tenant-01',
    },
    {
      id: 'user-5386',
      username: 'bob.jones',
      email: 'bob.jones@tenant-1.com',
      role: 'tenant_user',
      tenantId: 'tenant-01',
    },
    {
      id: 'user-6397',
      username: 'charlie.brown',
      email: 'charlie.brown@tenant-2.com',
      role: 'tenant_admin',
      tenantId: 'tenant-02',
    },
    {
      id: 'user-7351',
      username: 'diana.prince',
      email: 'diana.prince@tenant-2.com',
      role: 'tenant_user',
      tenantId: 'tenant-02',
    },
    {
      id: 'user-8888',
      username: 'eve.davis',
      email: 'eve.davis@tenant-3.com',
      role: 'tenant_admin',
      tenantId: 'tenant-03',
    },
    {
      id: 'user-9765',
      username: 'frank.miller',
      email: 'frank.miller@tenant-3.com',
      role: 'tenant_user',
      tenantId: 'tenant-03',
    },
    {
      id: 'user-1027',
      username: 'grace.hopper',
      email: 'grace.hopper@tenant-3.com',
      role: 'tenant_user',
      tenantId: 'tenant-03',
    },
    {
      id: 'user-1136',
      username: 'heidi.klum',
      email: 'heidi.klum@tenant-1.com',
      role: 'tenant_admin',
      tenantId: 'tenant-01',
    },
    {
      id: 'user-1283',
      username: 'ivan.drago',
      email: 'ivan.drago@tenant-2.com',
      role: 'tenant_user',
      tenantId: 'tenant-02',
    },
    {
      id: 'user-1390',
      username: 'judy.garland',
      email: 'judy.garland@tenant-3.com',
      role: 'tenant_user',
      tenantId: 'tenant-03',
    },
    {
      id: 'user-1482',
      username: 'kevin.bacon',
      email: 'kevin.bacon@tenant-1.com',
      role: 'tenant_user',
      tenantId: 'tenant-01',
    },
    {
      id: 'user-1516',
      username: 'laura.croft',
      email: 'laura.croft@tenant-2.com',
      role: 'tenant_admin',
      tenantId: 'tenant-02',
    },
  ];

  public getUsers(
    role: UserRole,
    page = 0,
    size = 10,
    tenantId?: string,
  ): Observable<PaginatedResponse<UserBackend>> {
    const roleString = userRoleMapper.toBackend(role);
    let filteredUsers = [...this.mockUsers];
    if (role) {
      filteredUsers = filteredUsers.filter((user) => user.role === roleString);
    }
    if (tenantId) {
      filteredUsers = filteredUsers.filter((user) => user.tenantId === tenantId);
    }
    const total = filteredUsers.length;
    const data = filteredUsers.slice(page * size, (page + 1) * size);
    return of({ count: data.length, total, data }).pipe(delay(500));
  }

  public getUser(id: string, role: UserRole, tenantId?: string): Observable<UserBackend> {
    const roleString = userRoleMapper.toBackend(role);
    const user = this.mockUsers.find(
      (u) =>
        u.id === id &&
        u.role === roleString &&
        (role === UserRole.SUPER_ADMIN || u.tenantId === tenantId),
    );
    if (!user) {
      throw new Error('User not found');
    }
    return of(user).pipe(delay(500));
  }

  public createUser(config: UserConfig, tenantId?: string, role?: UserRole): Observable<UserBackend> {
    const newId = `user-${Math.floor(Math.random() * 10000)}`;
    const newTenantId = tenantId || 'mock-tenant-id';

    const newUser: UserBackend = {
      id: newId,
      username: config.username || config.email.split('@')[0],
      email: config.email,
      role: userRoleMapper.toBackend(role ?? UserRole.TENANT_USER),
      tenantId: newTenantId,
    };
    this.mockUsers.push(newUser);
    return of(newUser).pipe(delay(500));
  }

  public deleteUser(id: string, role: UserRole, tenantId?: string): Observable<void> {
    const roleString = userRoleMapper.toBackend(role);
    this.mockUsers = this.mockUsers.filter(
      (u) =>
        !(
          u.id === id &&
          u.role === roleString &&
          (role === UserRole.SUPER_ADMIN || u.tenantId === tenantId)
        ),
    );
    return of(void 0).pipe(delay(500));
  }
}
