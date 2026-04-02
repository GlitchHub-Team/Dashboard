import { TestBed } from '@angular/core/testing';
import { ActivatedRouteSnapshot, Router } from '@angular/router';
import { describe, it, expect, vi, beforeEach } from 'vitest';

import { PermissionService } from '../../services/permission/permission.service';
import { Permission } from '../../models/permission.enum';
import { roleGuard } from './role.guard';

describe('roleGuard', () => {
  let permissionService: { canAny: ReturnType<typeof vi.fn> };
  let router: { navigate: ReturnType<typeof vi.fn> };

  beforeEach(() => {
    permissionService = {
      canAny: vi.fn(),
    };

    router = {
      navigate: vi.fn(),
    };

    TestBed.configureTestingModule({
      providers: [
        { provide: PermissionService, useValue: permissionService },
        { provide: Router, useValue: router },
      ],
    });
  });

  const executeGuard = (routeData: Record<string, unknown> = {}) => {
    const route = { data: routeData } as unknown as ActivatedRouteSnapshot;
    return TestBed.runInInjectionContext(() => roleGuard(route, {} as any));
  };

  const requiredPermissions = [Permission.GATEWAY_MANAGEMENT, Permission.SENSOR_MANAGEMENT];

  it.each([
    [{}],
    [{ permissions: undefined }],
    [{ permissions: [] }],
  ])('should return true and skip checks when no permissions are defined (%o)', (routeData) => {
    expect(executeGuard(routeData)).toBe(true);
    expect(permissionService.canAny).not.toHaveBeenCalled();
    expect(router.navigate).not.toHaveBeenCalled();
  });

  it('should return true and NOT navigate when the user HAS a required permission', () => {
    permissionService.canAny.mockReturnValue(true);

    expect(executeGuard({ permissions: requiredPermissions })).toBe(true);
    expect(permissionService.canAny).toHaveBeenCalledWith(requiredPermissions);
    expect(router.navigate).not.toHaveBeenCalled();
  });

  it('should return false and navigate to /dashboard when the user does NOT have any required permission', () => {
    permissionService.canAny.mockReturnValue(false);

    expect(executeGuard({ permissions: requiredPermissions })).toBe(false);
    expect(permissionService.canAny).toHaveBeenCalledWith(requiredPermissions);
    expect(router.navigate).toHaveBeenCalledOnce();
    expect(router.navigate).toHaveBeenCalledWith(['/dashboard']);
  });
});