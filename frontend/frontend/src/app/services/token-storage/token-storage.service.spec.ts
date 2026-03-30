import { TestBed } from '@angular/core/testing';

import { TokenStorageService } from './token-storage.service';

describe('TokenStorageService', () => {
  let service: TokenStorageService;

  // Helper to create a fake JWT with a given exp
  const createFakeToken = (exp: number): string => {
    const header = btoa(JSON.stringify({ alg: 'none' }));
    const payload = btoa(JSON.stringify({ exp }));
    return `${header}.${payload}.fake-signature`;
  };

  const futureExp = Math.floor(Date.now() / 1000) + 3600; // 1 hour from now
  const pastExp = Math.floor(Date.now() / 1000) - 3600; // 1 hour ago
  const validToken = createFakeToken(futureExp);
  const expiredToken = createFakeToken(pastExp);

  beforeEach(() => {
    window.sessionStorage.clear();

    TestBed.resetTestingModule();
    TestBed.configureTestingModule({});
    service = TestBed.inject(TokenStorageService);
  });

  describe('initial state', () => {
    it('should be created', () => {
      expect(service).toBeTruthy();
    });

    it('should start with isValid false when no token in storage', () => {
      expect(service.isValid()).toBe(false);
    });
  });

  describe('initialization from window.sessionStorage', () => {
    it('should be valid if localStorage has a non-expired token', () => {
      window.sessionStorage.setItem('jwt', validToken);

      TestBed.resetTestingModule();
      TestBed.configureTestingModule({});
      const freshService = TestBed.inject(TokenStorageService);

      expect(freshService.isValid()).toBe(true);
    });

    it('should be invalid if window.sessionStorage has an expired token', () => {
      window.sessionStorage.setItem('jwt', expiredToken);

      TestBed.resetTestingModule();
      TestBed.configureTestingModule({});
      const freshService = TestBed.inject(TokenStorageService);

      expect(freshService.isValid()).toBe(false);
    });
  });

  describe('saveToken', () => {
    it('should persist token to window.sessionStorage', () => {
      service.saveToken(validToken);

      expect(window.sessionStorage.getItem('jwt')).toBe(validToken);
    });

    it('should set isValid to true for a non-expired token', () => {
      service.saveToken(validToken);

      expect(service.isValid()).toBe(true);
    });

    it('should set isValid to false for an expired token', () => {
      service.saveToken(expiredToken);

      expect(service.isValid()).toBe(false);
    });

    it('should overwrite previous token', () => {
      service.saveToken(expiredToken);
      expect(service.isValid()).toBe(false);

      service.saveToken(validToken);
      expect(service.isValid()).toBe(true);
      expect(window.sessionStorage.getItem('jwt')).toBe(validToken);
    });
  });

  describe('getToken', () => {
    it('should return null when no token exists', () => {
      expect(service.getToken()).toBeNull();
    });

    it('should return the stored token', () => {
      service.saveToken(validToken);

      expect(service.getToken()).toBe(validToken);
    });
  });

  describe('clearToken', () => {
    it('should remove token from window.sessionStorage', () => {
      service.saveToken(validToken);
      service.clearToken();

      expect(window.sessionStorage.getItem('jwt')).toBeNull();
    });

    it('should set isValid to false', () => {
      service.saveToken(validToken);
      expect(service.isValid()).toBe(true);

      service.clearToken();

      expect(service.isValid()).toBe(false);
    });
  });

  describe('isTokenValid', () => {
    it('should return false when no token exists', () => {
      expect(service.isTokenValid()).toBe(false);
    });

    it('should return true for a non-expired token', () => {
      service.saveToken(validToken);

      expect(service.isTokenValid()).toBe(true);
    });

    it('should return false for an expired token', () => {
      service.saveToken(expiredToken);

      expect(service.isTokenValid()).toBe(false);
    });

    it('should return false for a malformed token', () => {
      window.sessionStorage.setItem('', 'not-a-jwt');

      expect(service.isTokenValid()).toBe(false);
    });

    it('should return false for a token with invalid base64 payload', () => {
      window.sessionStorage.setItem('jwt', 'header.!!!invalid!!!.signature');

      expect(service.isTokenValid()).toBe(false);
    });

    it('should return false for a token with no exp claim', () => {
      const header = btoa(JSON.stringify({ alg: 'none' }));
      const payload = btoa(JSON.stringify({ email: 'test@test.com' }));
      localStorage.setItem('', `${header}.${payload}.sig`);

      expect(service.isTokenValid()).toBe(false);
    });
  });
});
