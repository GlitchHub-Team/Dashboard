import { Injectable } from '@angular/core';
import { Observable, of, delay } from 'rxjs';
import { User } from '../models/user.model';
import { UserConfig } from '../services/user/user-api-client.service';
import { UserRole } from '../models/user-role.enum';

@Injectable({ providedIn: 'root' })
export class UserApiClientMockService {
  private mockUsers: User[] = [
    { id: 'user-5741', username: 'admin', email: 'admin@example.com', role: UserRole.TENANT_ADMIN, tenantId: 'tenant 1' },
    { id: 'user-3592', username: 'editor', email: 'editor@example.com', role: UserRole.TENANT_USER, tenantId: 'tenant 1' },
    { id: 'user-3876', username: 'viewer', email: 'viewer@example.com', role: UserRole.SUPER_ADMIN, tenantId: 'tenant 2' },
    { id: 'user-4160', username: 'alice.smith', email: 'alice.smith@tenant-1.com', role: UserRole.TENANT_USER, tenantId: 'tenant 1' },
    { id: 'user-5386', username: 'bob.jones', email: 'bob.jones@tenant-1.com', role: UserRole.TENANT_USER, tenantId: 'tenant 1' },
    { id: 'user-6397', username: 'charlie.brown', email: 'charlie.brown@tenant-2.com', role: UserRole.TENANT_ADMIN, tenantId: 'tenant 2' },
    { id: 'user-7351', username: 'diana.prince', email: 'diana.prince@tenant-2.com', role: UserRole.TENANT_USER, tenantId: 'tenant 2' },
    { id: 'user-8888', username: 'eve.davis', email: 'eve.davis@tenant-3.com', role: UserRole.TENANT_ADMIN, tenantId: 'tenant 3' },
    { id: 'user-9765', username: 'frank.miller', email: 'frank.miller@tenant-3.com', role: UserRole.TENANT_USER, tenantId: 'tenant 3' },
    { id: 'user-1027', username: 'grace.hopper', email: 'grace.hopper@tenant-3.com', role: UserRole.TENANT_USER, tenantId: 'tenant 3' },
    { id: 'user-1136', username: 'heidi.klum', email: 'heidi.klum@tenant-1.com', role: UserRole.TENANT_ADMIN, tenantId: 'tenant 1' },
    { id: 'user-1283', username: 'ivan.drago', email: 'ivan.drago@tenant-2.com', role: UserRole.TENANT_USER, tenantId: 'tenant 2' },
    { id: 'user-1390', username: 'judy.garland', email: 'judy.garland@tenant-3.com', role: UserRole.TENANT_USER, tenantId: 'tenant 3' },
    { id: 'user-1482', username: 'kevin.bacon', email: 'kevin.bacon@tenant-1.com', role: UserRole.TENANT_USER, tenantId: 'tenant 1' },
    { id: 'user-1516', username: 'laura.croft', email: 'laura.croft@tenant-2.com', role: UserRole.TENANT_ADMIN, tenantId: 'tenant 2' },
  ];

  public getUsers(role: UserRole, tenantId?: string, page = 0, size = 10): Observable<{ items: User[]; totalCount: number }> {
    let filteredUsers = [...this.mockUsers];
    if (role) {
      filteredUsers = filteredUsers.filter(user => user.role === role);
    }
    if (tenantId) {
      filteredUsers = filteredUsers.filter(user => user.tenantId === tenantId);
    }
    const totalCount = filteredUsers.length;
    const items = filteredUsers.slice(page * size, (page + 1) * size);
    return of({ items, totalCount }).pipe(delay(500));
  }

  public getUser(id: string, role: UserRole, tenantId?: string): Observable<User> {
    const user = this.mockUsers.find(u => u.id === id && u.role === role && (role === UserRole.SUPER_ADMIN || u.tenantId === tenantId));
    if (!user) {
      throw new Error('User not found');
    }
    return of(user).pipe(delay(500));
  }

  public createUser(config: UserConfig, tenantId?: string): Observable<User> {
    const newId = `user-${Math.floor(Math.random() * 10000)}`;
    const newTenantId = tenantId || 'mock-tenant-id';

    const newUser: User = {
      id: newId,
      username: (config as UserConfig & { username?: string }).username || config.email.split('@')[0],
      email: config.email,
      role: config.role,
      tenantId: newTenantId,
    };
    this.mockUsers.push(newUser);
    return of(newUser).pipe(delay(500));
  }

  public deleteUser(id: string): Observable<void> {
    this.mockUsers = this.mockUsers.filter((u) => u.id !== id);
    return of(void 0).pipe(delay(500));
  }
}
