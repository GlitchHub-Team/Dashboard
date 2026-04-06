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

  it.each([
    [true,  false],
    [false, true],
  ])('canAny=%s => guard returns %s and navigate called=%s', (hasPermission, expectNavigate) => {
    permissionService.canAny.mockReturnValue(hasPermission);

    expect(executeGuard({ permissions: requiredPermissions })).toBe(hasPermission);
    expect(permissionService.canAny).toHaveBeenCalledWith(requiredPermissions);
    if (expectNavigate) {
      expect(router.navigate).toHaveBeenCalledWith(['/dashboard']);
    } else {
      expect(router.navigate).not.toHaveBeenCalled();
    }
  });
});