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
    it('should be created with isValid false when no token in storage', () => {
      expect(service).toBeTruthy();
      expect(service.isValid()).toBe(false);
    });
  });

  describe('initialization from window.sessionStorage', () => {
    it.each([
      [validToken, true],
      [expiredToken, false],
    ])('should init isValid=%s from stored token', (token, expected) => {
      window.sessionStorage.setItem('jwt', token);

      TestBed.resetTestingModule();
      TestBed.configureTestingModule({});
      const freshService = TestBed.inject(TokenStorageService);

      expect(freshService.isValid()).toBe(expected);
    });
  });

  describe('saveToken', () => {
    it('should persist valid token to sessionStorage and set isValid to true', () => {
      service.saveToken(validToken);

      expect(window.sessionStorage.getItem('jwt')).toBe(validToken);
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
    it('should remove token from sessionStorage and set isValid to false', () => {
      service.saveToken(validToken);
      expect(service.isValid()).toBe(true);

      service.clearToken();

      expect(window.sessionStorage.getItem('jwt')).toBeNull();
      expect(service.isValid()).toBe(false);
    });
  });

  describe('isTokenValid', () => {
    it.each([
      ['no token', () => {}, false],
      ['valid token', () => service.saveToken(validToken), true],
      ['expired token', () => service.saveToken(expiredToken), false],
    ] as [string, () => void, boolean][])('should return %s => %s', (_label, setup, expected) => {
      setup();
      expect(service.isTokenValid()).toBe(expected);
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
