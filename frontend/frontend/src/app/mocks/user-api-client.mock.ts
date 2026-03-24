import { Injectable } from '@angular/core';
import { Observable, of, delay } from 'rxjs';
import { User } from '../models/user.model';
import { UserConfig } from '../services/user/user-api-client.service';
import { UserRole } from '../models/user-role.enum';

@Injectable({ providedIn: 'root' })
export class UserApiClientMockService {
  private mockUsers: User[] = [
    { id: 'user-5741', email: 'admin@example.com', role: UserRole.TENANT_ADMIN, tenantId: 'tenant-a' },
    { id: 'user-3592', email: 'editor@example.com', role: UserRole.TENANT_USER, tenantId: 'tenant-a' },
    { id: 'user-3876', email: 'viewer@example.com', role: UserRole.SUPER_ADMIN, tenantId: 'tenant-b' },
    { id: 'user-4160', email: 'alice.smith@tenant-a.com', role: UserRole.TENANT_USER, tenantId: 'tenant-a' },
    { id: 'user-5386', email: 'bob.jones@tenant-a.com', role: UserRole.TENANT_USER, tenantId: 'tenant-a' },
    { id: 'user-6397', email: 'charlie.brown@tenant-b.com', role: UserRole.TENANT_ADMIN, tenantId: 'tenant-b' },
    { id: 'user-7351', email: 'diana.prince@tenant-b.com', role: UserRole.TENANT_USER, tenantId: 'tenant-b' },
    { id: 'user-8888', email: 'eve.davis@tenant-c.com', role: UserRole.TENANT_ADMIN, tenantId: 'tenant-c' },
    { id: 'user-9765', email: 'frank.miller@tenant-c.com', role: UserRole.TENANT_USER, tenantId: 'tenant-c' },
    { id: 'user-1027', email: 'grace.hopper@tenant-c.com', role: UserRole.TENANT_USER, tenantId: 'tenant-c' },
    { id: 'user-1136', email: 'heidi.klum@tenant-a.com', role: UserRole.TENANT_ADMIN, tenantId: 'tenant-a' },
    { id: 'user-1283', email: 'ivan.drago@tenant-b.com', role: UserRole.TENANT_USER, tenantId: 'tenant-b' },
    { id: 'user-1390', email: 'judy.garland@tenant-c.com', role: UserRole.TENANT_USER, tenantId: 'tenant-c' },
    { id: 'user-1482', email: 'kevin.bacon@tenant-a.com', role: UserRole.TENANT_USER, tenantId: 'tenant-a' },
    { id: 'user-1516', email: 'laura.croft@tenant-b.com', role: UserRole.TENANT_ADMIN, tenantId: 'tenant-b' },
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
