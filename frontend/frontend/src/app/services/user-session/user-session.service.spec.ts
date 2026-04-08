import { TestBed } from '@angular/core/testing';

import { UserSessionService } from './user-session.service';
import { TokenStorageService } from '../token-storage/token-storage.service';
import { UserSession } from '../../models/auth/user-session.model';
import { UserRole } from '../../models/user/user-role.enum';

function buildJwt(payload: Record<string, unknown>): string {
  const header = btoa(JSON.stringify({ alg: 'HS256', typ: 'JWT' }));
  const body = btoa(JSON.stringify(payload));
  return `${header}.${body}.signature`;
}

describe('UserSessionService', () => {
  let service: UserSessionService;

  const tokenStorageMock = {
    getToken: vi.fn<() => string | null>().mockReturnValue(null),
    isTokenValid: vi.fn<() => boolean>().mockReturnValue(false),
  };

  const mockSession: UserSession = {
    userId: '1',
    role: UserRole.SUPER_ADMIN,
    tenantId: 'tenant-1',
  };

  const mockToken = buildJwt({
    uid: '1',
    rol: 'sa',
    tid: 'tenant-1',
  });

  beforeEach(() => {
    vi.resetAllMocks();
    tokenStorageMock.getToken.mockReturnValue(null);
    tokenStorageMock.isTokenValid.mockReturnValue(false);
    sessionStorage.clear();

    TestBed.configureTestingModule({
      providers: [{ provide: TokenStorageService, useValue: tokenStorageMock }],
    });
    service = TestBed.inject(UserSessionService);
  });

  describe('initial state', () => {
    it('should be created with null currentUser', () => {
      expect(service).toBeTruthy();
      expect(service.currentUser()).toBeNull();
    });
  });

  describe('restoreSession', () => {
    it('should restore user from sessionStorage on creation', () => {
      sessionStorage.setItem('currentUser', JSON.stringify(mockSession));

      TestBed.resetTestingModule();
      TestBed.configureTestingModule({
        providers: [{ provide: TokenStorageService, useValue: tokenStorageMock }],
      });
      const newService = TestBed.inject(UserSessionService);

      expect(newService.currentUser()).toEqual(mockSession);
    });

    it('should decode token from TokenStorageService when sessionStorage is empty', () => {
      tokenStorageMock.getToken.mockReturnValue(mockToken);
      tokenStorageMock.isTokenValid.mockReturnValue(true);

      TestBed.resetTestingModule();
      TestBed.configureTestingModule({
        providers: [{ provide: TokenStorageService, useValue: tokenStorageMock }],
      });
      const newService = TestBed.inject(UserSessionService);

      expect(newService.currentUser()).toEqual(mockSession);
    });

    it('should return null when sessionStorage has invalid JSON and no valid token', () => {
      sessionStorage.setItem('currentUser', 'not-json');

      TestBed.resetTestingModule();
      TestBed.configureTestingModule({
        providers: [{ provide: TokenStorageService, useValue: tokenStorageMock }],
      });
      const newService = TestBed.inject(UserSessionService);

      expect(newService.currentUser()).toBeNull();
    });
  });

  describe('initSession', () => {
    it('should decode token, set currentUser, and persist to sessionStorage', () => {
      service.initSession(mockToken);

      expect(service.currentUser()).toEqual(mockSession);
      expect(JSON.parse(sessionStorage.getItem('currentUser')!)).toEqual(mockSession);
    });

    it('should not set user when token is invalid', () => {
      service.initSession('not.a.valid.token');

      expect(service.currentUser()).toBeNull();
    });

    it('should overwrite previous session', () => {
      service.initSession(mockToken);

      const newToken = buildJwt({
        uid: '2',
        rol: 'tu',
        tid: 'tenant-2',
      });
      service.initSession(newToken);

      expect(service.currentUser()).toEqual({
        userId: '2',
        role: UserRole.TENANT_USER,
        tenantId: 'tenant-2',
      });
    });
  });

  describe('clearSession', () => {
    it('should set currentUser to null and remove from sessionStorage', () => {
      service.initSession(mockToken);
      service.clearSession();

      expect(service.currentUser()).toBeNull();
      expect(sessionStorage.getItem('currentUser')).toBeNull();
    });
  });
});
