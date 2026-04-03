import { ActivatedRouteSnapshot, CanActivateFn, Router } from '@angular/router';
import { inject } from '@angular/core';

import { PermissionService } from '../../services/permission/permission.service';
import { Permission } from '../../models/permission.enum';

export const roleGuard: CanActivateFn = (route: ActivatedRouteSnapshot) => {
  const permissionService = inject(PermissionService);
  const router = inject(Router);

  const requiredPermissions = route.data['permissions'] as Permission[];

  if (!requiredPermissions?.length) {
    return true;
  }

  if (permissionService.canAny(requiredPermissions)) {
    return true;
  }

  router.navigate(['/dashboard']);
  return false;
};
