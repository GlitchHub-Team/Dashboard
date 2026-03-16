import { Injectable } from '@angular/core';
import { Observable, of, delay } from 'rxjs';
import { User } from '../../models/user.model';
import { UserConfig } from './user-api-client.service';
import { UserRole } from '../../models/user-role.enum';

@Injectable({ providedIn: 'root' })
export class UserApiClientMockService {
  private mockUsers: User[] = [
    { id: 'user-1', email: 'admin@example.com', role: UserRole.TENANT_ADMIN, tenantId: 'tenant-a' },
    { id: 'user-2', email: 'editor@example.com', role: UserRole.TENANT_USER, tenantId: 'tenant-a' },
    { id: 'user-3', email: 'viewer@example.com', role: UserRole.SUPER_ADMIN, tenantId: 'tenant-b' },
    { id: 'user-4', email: 'alice.smith@tenant-a.com', role: UserRole.TENANT_USER, tenantId: 'tenant-a' },
    { id: 'user-5', email: 'bob.jones@tenant-a.com', role: UserRole.TENANT_USER, tenantId: 'tenant-a' },
    { id: 'user-6', email: 'charlie.brown@tenant-b.com', role: UserRole.TENANT_ADMIN, tenantId: 'tenant-b' },
    { id: 'user-7', email: 'diana.prince@tenant-b.com', role: UserRole.TENANT_USER, tenantId: 'tenant-b' },
    { id: 'user-8', email: 'eve.davis@tenant-c.com', role: UserRole.TENANT_ADMIN, tenantId: 'tenant-c' },
    { id: 'user-9', email: 'frank.miller@tenant-c.com', role: UserRole.TENANT_USER, tenantId: 'tenant-c' },
    { id: 'user-10', email: 'grace.hopper@tenant-c.com', role: UserRole.TENANT_USER, tenantId: 'tenant-c' },
    { id: 'user-11', email: 'heidi.klum@tenant-a.com', role: UserRole.TENANT_ADMIN, tenantId: 'tenant-a' },
    { id: 'user-12', email: 'ivan.drago@tenant-b.com', role: UserRole.TENANT_USER, tenantId: 'tenant-b' },
    { id: 'user-13', email: 'judy.garland@tenant-c.com', role: UserRole.TENANT_USER, tenantId: 'tenant-c' },
    { id: 'user-14', email: 'kevin.bacon@tenant-a.com', role: UserRole.TENANT_USER, tenantId: 'tenant-a' },
    { id: 'user-15', email: 'laura.croft@tenant-b.com', role: UserRole.TENANT_ADMIN, tenantId: 'tenant-b' },
  ];

  public getUsers(role?: UserRole): Observable<User[]> {
    let filteredUsers = [...this.mockUsers];
    if (role) {
      filteredUsers = filteredUsers.filter(user => user.role === role);
    }
    return of(filteredUsers).pipe(delay(500));
  }

  public createUser(config: UserConfig): Observable<User> {
    const newId = `user-${Date.now()}`;
    const newTenantId = 'mock-tenant-id';

    const newUser: User = {
      id: newId,
      email: config.email,
      role: config.role,
      tenantId: newTenantId,
    };
    this.mockUsers.push(newUser);
    return of(newUser).pipe(delay(500));
  }

  public deleteUser(email: string): Observable<void> {
    this.mockUsers = this.mockUsers.filter((u) => u.email !== email);
    return of(void 0).pipe(delay(500));
  }
}
