import { TestBed } from '@angular/core/testing';
import { Router } from '@angular/router';
import { describe, it, expect, vi, beforeEach } from 'vitest';

import { AuthSessionService } from '../../services/auth/auth-session.service';
import { authGuard } from './auth.guard';

describe('authGuard', () => {
  let authSessionService: { isAuthenticated: ReturnType<typeof vi.fn> };
  let router: { navigate: ReturnType<typeof vi.fn> };

  beforeEach(() => {
    authSessionService = {
      isAuthenticated: vi.fn(),
    };

    router = {
      navigate: vi.fn(),
    };

    TestBed.configureTestingModule({
      providers: [
        { provide: AuthSessionService, useValue: authSessionService },
        { provide: Router, useValue: router },
      ],
    });
  });

  const executeGuard = () => {
    return TestBed.runInInjectionContext(() => authGuard({} as any, {} as any));
  };

  it('should return true and NOT navigate when the user is authenticated', () => {
    authSessionService.isAuthenticated.mockReturnValue(true);

    expect(executeGuard()).toBe(true);
    expect(router.navigate).not.toHaveBeenCalled();
  });

  it('should return false and navigate to /login when the user is NOT authenticated', () => {
    authSessionService.isAuthenticated.mockReturnValue(false);

    expect(executeGuard()).toBe(false);
    expect(router.navigate).toHaveBeenCalledOnce();
    expect(router.navigate).toHaveBeenCalledWith(['/login']);
  });
});