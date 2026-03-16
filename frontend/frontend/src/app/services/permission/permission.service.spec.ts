import { TestBed } from '@angular/core/testing';

import { PermissionService } from './permission.service';
import { UserSessionService } from '../user-session/user-session.service';
import { UserRole } from '../../models/user-role.enum';
import { Permission } from '../../models/permission.enum';

describe('PermissionService', () => {
  let service: PermissionService;

  const userSessionMock = {
    currentRole: vi.fn(),
  };

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

    it('should allow DASHBOARD_ACCESS for TENANT_USER', () => {
      userSessionMock.currentRole.mockReturnValue(UserRole.TENANT_USER);

      expect(service.can(Permission.DASHBOARD_ACCESS)).toBe(true);
    });

    it('should deny GATEWAY_COMMANDS for TENANT_USER', () => {
      userSessionMock.currentRole.mockReturnValue(UserRole.TENANT_USER);

      expect(service.can(Permission.GATEWAY_COMMANDS)).toBe(false);
    });

    it('should deny TENANT_USER_MANAGEMENT for TENANT_USER', () => {
      userSessionMock.currentRole.mockReturnValue(UserRole.TENANT_USER);

      expect(service.can(Permission.TENANT_USER_MANAGEMENT)).toBe(false);
    });

    it('should allow DASHBOARD_ACCESS for TENANT_ADMIN', () => {
      userSessionMock.currentRole.mockReturnValue(UserRole.TENANT_ADMIN);

      expect(service.can(Permission.DASHBOARD_ACCESS)).toBe(true);
    });

    it('should allow GATEWAY_COMMANDS for TENANT_ADMIN', () => {
      userSessionMock.currentRole.mockReturnValue(UserRole.TENANT_ADMIN);

      expect(service.can(Permission.GATEWAY_COMMANDS)).toBe(true);
    });

    it('should allow TENANT_USER_MANAGEMENT for TENANT_ADMIN', () => {
      userSessionMock.currentRole.mockReturnValue(UserRole.TENANT_ADMIN);

      expect(service.can(Permission.TENANT_USER_MANAGEMENT)).toBe(true);
    });

    it('should deny GATEWAY_MANAGEMENT for TENANT_ADMIN', () => {
      userSessionMock.currentRole.mockReturnValue(UserRole.TENANT_ADMIN);

      expect(service.can(Permission.GATEWAY_MANAGEMENT)).toBe(false);
    });

    it('should allow all permissions for SUPER_ADMIN', () => {
      userSessionMock.currentRole.mockReturnValue(UserRole.SUPER_ADMIN);

      expect(service.can(Permission.DASHBOARD_ACCESS)).toBe(true);
      expect(service.can(Permission.GATEWAY_MANAGEMENT)).toBe(true);
      expect(service.can(Permission.SENSOR_MANAGEMENT)).toBe(true);
      expect(service.can(Permission.GATEWAY_COMMANDS)).toBe(true);
      expect(service.can(Permission.TENANT_ADMIN_MANAGEMENT)).toBe(true);
      expect(service.can(Permission.TENANT_MANAGEMENT)).toBe(true);
      expect(service.can(Permission.APIKEY_MANAGEMENT)).toBe(true);
    });

    it('should deny TENANT_USER_MANAGEMENT for SUPER_ADMIN', () => {
      userSessionMock.currentRole.mockReturnValue(UserRole.SUPER_ADMIN);

      expect(service.can(Permission.TENANT_USER_MANAGEMENT)).toBe(false);
    });
  });

  describe('canAny', () => {
    it('should return false when role is null', () => {
      userSessionMock.currentRole.mockReturnValue(null);

      expect(service.canAny([Permission.DASHBOARD_ACCESS])).toBe(false);
    });

    it('should return true if at least one permission matches', () => {
      userSessionMock.currentRole.mockReturnValue(UserRole.TENANT_USER);

      expect(service.canAny([Permission.DASHBOARD_ACCESS, Permission.GATEWAY_MANAGEMENT])).toBe(
        true,
      );
    });

    it('should return false if no permissions match', () => {
      userSessionMock.currentRole.mockReturnValue(UserRole.TENANT_USER);

      expect(service.canAny([Permission.GATEWAY_MANAGEMENT, Permission.SENSOR_MANAGEMENT])).toBe(
        false,
      );
    });

    it('should return true if all permissions match', () => {
      userSessionMock.currentRole.mockReturnValue(UserRole.TENANT_ADMIN);

      expect(service.canAny([Permission.DASHBOARD_ACCESS, Permission.GATEWAY_COMMANDS])).toBe(true);
    });

    it('should return false for empty array', () => {
      userSessionMock.currentRole.mockReturnValue(UserRole.SUPER_ADMIN);

      expect(service.canAny([])).toBe(false);
    });
  });

  describe('canAll', () => {
    it('should return false when role is null', () => {
      userSessionMock.currentRole.mockReturnValue(null);

      expect(service.canAll([Permission.DASHBOARD_ACCESS])).toBe(false);
    });

    it('should return true when user has all permissions', () => {
      userSessionMock.currentRole.mockReturnValue(UserRole.TENANT_ADMIN);

      expect(service.canAll([Permission.DASHBOARD_ACCESS, Permission.GATEWAY_COMMANDS])).toBe(true);
    });

    it('should return false when user is missing one permission', () => {
      userSessionMock.currentRole.mockReturnValue(UserRole.TENANT_ADMIN);

      expect(
        service.canAll([
          Permission.DASHBOARD_ACCESS,
          Permission.GATEWAY_COMMANDS,
          Permission.GATEWAY_MANAGEMENT,
        ]),
      ).toBe(false);
    });

    it('should return true for single permission the user has', () => {
      userSessionMock.currentRole.mockReturnValue(UserRole.TENANT_USER);

      expect(service.canAll([Permission.DASHBOARD_ACCESS])).toBe(true);
    });

    it('should return true for empty array', () => {
      userSessionMock.currentRole.mockReturnValue(UserRole.TENANT_USER);

      expect(service.canAll([])).toBe(true);
    });
  });
});
