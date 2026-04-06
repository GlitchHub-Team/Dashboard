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

  it.each([
    [true,  false],
    [false, true],
  ])('isAuthenticated=%s => guard returns %s and navigate called=%s', (isAuthenticated, expectNavigate) => {
    authSessionService.isAuthenticated.mockReturnValue(isAuthenticated);

    expect(executeGuard()).toBe(isAuthenticated);
    if (expectNavigate) {
      expect(router.navigate).toHaveBeenCalledWith(['/login']);
    } else {
      expect(router.navigate).not.toHaveBeenCalled();
    }
  });
});