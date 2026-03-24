import { TestBed } from '@angular/core/testing';

import { UserSessionService } from './user-session.service';
import { User } from '../../models/user.model';
import { UserRole } from '../../models/user-role.enum';

describe('UserSessionService', () => {
  let service: UserSessionService;

  const mockUser: User = {
    id: '1',
    username: 'admin',
    email: 'admin@test.com',
    role: UserRole.SUPER_ADMIN,
    tenantId: 'tenant-1',
  };

  beforeEach(() => {
    vi.resetAllMocks();
    sessionStorage.clear();

    TestBed.configureTestingModule({});
    service = TestBed.inject(UserSessionService);
  });

  describe('initial state', () => {
    it('should be created', () => {
      expect(service).toBeTruthy();
    });

    it('should start with null user when storage is empty', () => {
      expect(service.currentUser()).toBeNull();
    });

    it('should start with null role when no user', () => {
      expect(service.currentRole()).toBeNull();
    });

    it('should start with null tenant when no user', () => {
      expect(service.currentTenant()).toBeNull();
    });
  });

  describe('loadFromStorage', () => {
    it('should restore user from sessionStorage on creation', () => {
      sessionStorage.setItem('currentUser', JSON.stringify(mockUser));

      // Note: since it's providedIn: 'root', we need a fresh TestBed
      TestBed.resetTestingModule();
      TestBed.configureTestingModule({});
      const newService = TestBed.inject(UserSessionService);

      expect(newService.currentUser()).toEqual(mockUser);
    });

    it('should return null when sessionStorage has invalid JSON', () => {
      sessionStorage.setItem('currentUser', 'not-json');

      TestBed.resetTestingModule();
      TestBed.configureTestingModule({});
      const newService = TestBed.inject(UserSessionService);

      expect(newService.currentUser()).toBeNull();
    });

    it('should return null when sessionStorage is empty', () => {
      TestBed.resetTestingModule();
      TestBed.configureTestingModule({});
      const newService = TestBed.inject(UserSessionService);

      expect(newService.currentUser()).toBeNull();
    });
  });

  describe('initSession', () => {
    it('should set current user', () => {
      service.initSession(mockUser);

      expect(service.currentUser()).toEqual(mockUser);
    });

    it('should persist user to sessionStorage', () => {
      service.initSession(mockUser);

      const stored = JSON.parse(sessionStorage.getItem('currentUser')!);
      expect(stored).toEqual(mockUser);
    });

    it('should update currentRole', () => {
      service.initSession(mockUser);

      expect(service.currentRole()).toBe(UserRole.SUPER_ADMIN);
    });

    it('should update currentTenant', () => {
      service.initSession(mockUser);

      expect(service.currentTenant()).toBe('tenant-1');
    });

    it('should overwrite previous session', () => {
      service.initSession(mockUser);

      const newUser: User = {
        id: '2',
        username: 'testuser',
        email: 'user@test.com',
        role: UserRole.TENANT_USER,
        tenantId: 'tenant-2',
      };
      service.initSession(newUser);

      expect(service.currentUser()).toEqual(newUser);
      expect(service.currentRole()).toBe(UserRole.TENANT_USER);
      expect(service.currentTenant()).toBe('tenant-2');
    });
  });

  describe('clearSession', () => {
    it('should set current user to null', () => {
      service.initSession(mockUser);
      service.clearSession();

      expect(service.currentUser()).toBeNull();
    });

    it('should remove user from sessionStorage', () => {
      service.initSession(mockUser);
      service.clearSession();

      expect(sessionStorage.getItem('currentUser')).toBeNull();
    });

    it('should set currentRole to null', () => {
      service.initSession(mockUser);
      service.clearSession();

      expect(service.currentRole()).toBeNull();
    });

    it('should set currentTenant to null', () => {
      service.initSession(mockUser);
      service.clearSession();

      expect(service.currentTenant()).toBeNull();
    });
  });

  describe('computed values', () => {
    it('should derive role from each user role type', () => {
      const roles = [UserRole.TENANT_USER, UserRole.TENANT_ADMIN, UserRole.SUPER_ADMIN];

      for (const role of roles) {
        service.initSession({ ...mockUser, role });
        expect(service.currentRole()).toBe(role);
      }
    });

    it('should derive tenant from user', () => {
      service.initSession({ ...mockUser, tenantId: 'custom-tenant' });

      expect(service.currentTenant()).toBe('custom-tenant');
    });
  });
});
