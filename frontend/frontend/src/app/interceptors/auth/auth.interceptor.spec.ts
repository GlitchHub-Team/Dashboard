import { TestBed } from '@angular/core/testing';
import { HttpRequest, HttpHandlerFn, HttpResponse, HttpErrorResponse } from '@angular/common/http';
import { Router } from '@angular/router';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { of, throwError, firstValueFrom } from 'rxjs';

import { TokenStorageService } from '../../services/token-storage/token-storage.service';
import { UserSessionService } from '../../services/user-session/user-session.service';
import { authInterceptor } from './auth.interceptor';

describe('authInterceptor', () => {
  let tokenStorageService: {
    getToken: ReturnType<typeof vi.fn>;
    clearToken: ReturnType<typeof vi.fn>;
  };
  let userSessionService: {
    clearSession: ReturnType<typeof vi.fn>;
  };
  let router: {
    navigate: ReturnType<typeof vi.fn>;
  };

  beforeEach(() => {
    tokenStorageService = {
      getToken: vi.fn(),
      clearToken: vi.fn(),
    };

    userSessionService = {
      clearSession: vi.fn(),
    };

    router = {
      navigate: vi.fn(),
    };

    TestBed.configureTestingModule({
      providers: [
        { provide: TokenStorageService, useValue: tokenStorageService },
        { provide: UserSessionService, useValue: userSessionService },
        { provide: Router, useValue: router },
      ],
    });
  });

  const executeInterceptor = (req: HttpRequest<unknown>, next: HttpHandlerFn) =>
    TestBed.runInInjectionContext(() => authInterceptor(req, next));

  const createRequest = (url: string) => new HttpRequest('GET', url);

  const createSuccessHandler = (): HttpHandlerFn =>
    (req) => of(new HttpResponse({ status: 200, body: {}, url: req.url }));

  const createErrorHandler = (status: number): HttpHandlerFn =>
    () => throwError(() => new HttpErrorResponse({ status, statusText: 'Error', url: 'test' }));

  it('should attach a Bearer Authorization header for non-login requests with a token', async () => {
    tokenStorageService.getToken.mockReturnValue('my-jwt-token');
    let capturedReq!: HttpRequest<unknown>;
    const next: HttpHandlerFn = (r) => { capturedReq = r; return of(new HttpResponse({ status: 200 })); };

    await firstValueFrom(executeInterceptor(createRequest('/api/data'), next));

    expect(capturedReq.headers.get('Authorization')).toBe('Bearer my-jwt-token');
  });

  it.each([
    ['login request with token',      'my-jwt-token', '/api/auth/login'],
    ['non-login request without token', null,          '/api/data'],
    ['login request without token',    null,           '/api/auth/login'],
  ])('should NOT attach Authorization header: %s', async (_, token, url) => {
    tokenStorageService.getToken.mockReturnValue(token);
    let capturedReq!: HttpRequest<unknown>;
    const next: HttpHandlerFn = (r) => { capturedReq = r; return of(new HttpResponse({ status: 200 })); };

    await firstValueFrom(executeInterceptor(createRequest(url), next));

    expect(capturedReq.headers.has('Authorization')).toBe(false);
  });

  it('should pass the original request body and method untouched', async () => {
    tokenStorageService.getToken.mockReturnValue('token');
    let capturedReq!: HttpRequest<unknown>;
    const next: HttpHandlerFn = (r) => { capturedReq = r; return of(new HttpResponse({ status: 200 })); };

    await firstValueFrom(executeInterceptor(new HttpRequest('POST', '/api/data', { key: 'value' }), next));

    expect(capturedReq.method).toBe('POST');
    expect(capturedReq.body).toEqual({ key: 'value' });
  });

  it('should clear token, session, navigate to /login, and rethrow on 401 (non-login, token present)', async () => {
    tokenStorageService.getToken
      .mockReturnValueOnce('my-jwt-token') 
      .mockReturnValueOnce('my-jwt-token'); 

    await expect(
      firstValueFrom(executeInterceptor(createRequest('/api/data'), createErrorHandler(401)))
    ).rejects.toBeInstanceOf(HttpErrorResponse);

    expect(tokenStorageService.clearToken).toHaveBeenCalledOnce();
    expect(userSessionService.clearSession).toHaveBeenCalledOnce();
    expect(router.navigate).toHaveBeenCalledOnce();
    expect(router.navigate).toHaveBeenCalledWith(['/login']);
  });

  it.each([
    ['token already cleared',  () => { tokenStorageService.getToken.mockReturnValueOnce('my-jwt-token').mockReturnValueOnce(null); }, '/api/data'],
    ['login request',          () => { tokenStorageService.getToken.mockReturnValue('my-jwt-token'); },                             '/api/auth/login'],
  ])('should NOT clear token/session or navigate on 401 when %s, but still rethrow', async (_label, setup, url) => {
    setup();

    await expect(
      firstValueFrom(executeInterceptor(createRequest(url), createErrorHandler(401)))
    ).rejects.toBeInstanceOf(HttpErrorResponse);

    expect(tokenStorageService.clearToken).not.toHaveBeenCalled();
    expect(userSessionService.clearSession).not.toHaveBeenCalled();
    expect(router.navigate).not.toHaveBeenCalled();
  });

  it.each([400, 403, 404, 500, 502, 503])(
    'should NOT clear token/session or navigate for non-401 status %d, but still rethrow',
    async (status) => {
      tokenStorageService.getToken.mockReturnValue('my-jwt-token');
      let thrownError: unknown;

      try {
        await firstValueFrom(executeInterceptor(createRequest('/api/data'), createErrorHandler(status)));
      } catch (e) {
        thrownError = e;
      }

      expect(thrownError).toBeInstanceOf(HttpErrorResponse);
      expect((thrownError as HttpErrorResponse).status).toBe(status);
      expect(tokenStorageService.clearToken).not.toHaveBeenCalled();
      expect(userSessionService.clearSession).not.toHaveBeenCalled();
      expect(router.navigate).not.toHaveBeenCalled();
    }
  );

  it('should pass through successful responses without clearing token, session, or navigating', async () => {
    tokenStorageService.getToken.mockReturnValue('token');

    const response = await firstValueFrom(
      executeInterceptor(createRequest('/api/data'), createSuccessHandler())
    );

    expect(response).toBeInstanceOf(HttpResponse);
    expect((response as HttpResponse<unknown>).status).toBe(200);
    expect(tokenStorageService.clearToken).not.toHaveBeenCalled();
    expect(userSessionService.clearSession).not.toHaveBeenCalled();
    expect(router.navigate).not.toHaveBeenCalled();
  });
});