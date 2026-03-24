import { TestBed } from '@angular/core/testing';

import { PermissionService } from './permission.service';
import { UserSessionService } from '../user-session/user-session.service';
import { UserRole } from '../../models/user/user-role.enum';
import { Permission } from '../../models/permission.enum';

describe('PermissionService', () => {
  let service: PermissionService;

  const userSessionMock = { currentRole: vi.fn() };

  beforeEach(() => {
    vi.resetAllMocks();
    TestBed.configureTestingModule({
      providers: [PermissionService, { provide: UserSessionService, useValue: userSessionMock }],
    });
    service = TestBed.inject(PermissionService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  describe('can', () => {
    it('should return false when role is null', () => {
      userSessionMock.currentRole.mockReturnValue(null);
      expect(service.can(Permission.DASHBOARD_ACCESS)).toBe(false);
    });

    it.each([
      // TENANT_USER
      [UserRole.TENANT_USER, Permission.DASHBOARD_ACCESS, true],
      [UserRole.TENANT_USER, Permission.GATEWAY_COMMANDS, false],
      [UserRole.TENANT_USER, Permission.TENANT_USER_MANAGEMENT, false],
      // TENANT_ADMIN
      [UserRole.TENANT_ADMIN, Permission.DASHBOARD_ACCESS, true],
      [UserRole.TENANT_ADMIN, Permission.GATEWAY_COMMANDS, true],
      [UserRole.TENANT_ADMIN, Permission.TENANT_USER_MANAGEMENT, true],
      [UserRole.TENANT_ADMIN, Permission.GATEWAY_MANAGEMENT, false],
      // SUPER_ADMIN
      [UserRole.SUPER_ADMIN, Permission.DASHBOARD_ACCESS, true],
      [UserRole.SUPER_ADMIN, Permission.GATEWAY_MANAGEMENT, true],
      [UserRole.SUPER_ADMIN, Permission.SENSOR_MANAGEMENT, true],
      [UserRole.SUPER_ADMIN, Permission.GATEWAY_COMMANDS, true],
      [UserRole.SUPER_ADMIN, Permission.TENANT_ADMIN_MANAGEMENT, true],
      [UserRole.SUPER_ADMIN, Permission.TENANT_MANAGEMENT, true],
      [UserRole.SUPER_ADMIN, Permission.APIKEY_MANAGEMENT, true],
      [UserRole.SUPER_ADMIN, Permission.TENANT_USER_MANAGEMENT, false],
    ])('%s / %s => %s', (role: UserRole, permission: Permission, expected: boolean) => {
      userSessionMock.currentRole.mockReturnValue(role);
      expect(service.can(permission)).toBe(expected);
    });
  });

  describe('canAny', () => {
    it('should return false when role is null', () => {
      userSessionMock.currentRole.mockReturnValue(null);
      expect(service.canAny([Permission.DASHBOARD_ACCESS])).toBe(false);
    });

    it.each([
      [
        UserRole.TENANT_USER,
        [Permission.DASHBOARD_ACCESS, Permission.GATEWAY_MANAGEMENT],
        true,
        'at least one matches',
      ],
      [
        UserRole.TENANT_USER,
        [Permission.GATEWAY_MANAGEMENT, Permission.SENSOR_MANAGEMENT],
        false,
        'none match',
      ],
      [
        UserRole.TENANT_ADMIN,
        [Permission.DASHBOARD_ACCESS, Permission.GATEWAY_COMMANDS],
        true,
        'all match',
      ],
      [UserRole.SUPER_ADMIN, [], false, 'empty array'],
    ])('%s: %s => %s (%s)', (role: UserRole, permissions: Permission[], expected: boolean) => {
      userSessionMock.currentRole.mockReturnValue(role);
      expect(service.canAny(permissions)).toBe(expected);
    });
  });

  describe('canAll', () => {
    it('should return false when role is null', () => {
      userSessionMock.currentRole.mockReturnValue(null);
      expect(service.canAll([Permission.DASHBOARD_ACCESS])).toBe(false);
    });

    it.each([
      [UserRole.TENANT_USER, [Permission.DASHBOARD_ACCESS], true, 'single granted permission'],
      [UserRole.TENANT_USER, [], true, 'empty array'],
      [
        UserRole.TENANT_ADMIN,
        [Permission.DASHBOARD_ACCESS, Permission.GATEWAY_COMMANDS],
        true,
        'all granted',
      ],
      [
        UserRole.TENANT_ADMIN,
        [Permission.DASHBOARD_ACCESS, Permission.GATEWAY_COMMANDS, Permission.GATEWAY_MANAGEMENT],
        false,
        'one missing',
      ],
    ])('%s: %s => %s (%s)', (role: UserRole, permissions: Permission[], expected: boolean) => {
      userSessionMock.currentRole.mockReturnValue(role);
      expect(service.canAll(permissions)).toBe(expected);
    });
  });
});
