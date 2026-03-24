import { Injectable, inject } from '@angular/core';

import { UserSessionService } from '../user-session/user-session.service';
import { UserRole } from '../../models/user/user-role.enum';
import { Permission } from '../../models/permission.enum';

@Injectable({
  providedIn: 'root',
})
export class PermissionService {
  private userSession = inject(UserSessionService);

  private ROLE_PERMISSIONS: Record<UserRole, Permission[]> = {
    [UserRole.TENANT_USER]: [Permission.DASHBOARD_ACCESS],
    [UserRole.TENANT_ADMIN]: [
      Permission.DASHBOARD_ACCESS,
      Permission.GATEWAY_COMMANDS,
      Permission.TENANT_USER_MANAGEMENT,
    ],
    [UserRole.SUPER_ADMIN]: [
      Permission.DASHBOARD_ACCESS,
      Permission.GATEWAY_MANAGEMENT,
      Permission.SENSOR_MANAGEMENT,
      Permission.GATEWAY_COMMANDS,
      Permission.TENANT_ADMIN_MANAGEMENT,
      Permission.TENANT_MANAGEMENT,
      Permission.APIKEY_MANAGEMENT,
    ],
  };

  public can(permission: Permission): boolean {
    const role = this.userSession.currentRole();
    if (!role) {
      return false;
    }
    return this.ROLE_PERMISSIONS[role].includes(permission);
  }

  public canAny(permissions: Permission[]): boolean {
    const role = this.userSession.currentRole();
    if (!role) {
      return false;
    }
    return permissions.some((perm) => this.ROLE_PERMISSIONS[role].includes(perm));
  }

  public canAll(permissions: Permission[]): boolean {
    const role = this.userSession.currentRole();
    if (!role) {
      return false;
    }
    return permissions.every((perm) => this.ROLE_PERMISSIONS[role].includes(perm));
  }
}
