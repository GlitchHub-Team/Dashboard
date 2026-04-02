import { TestBed } from '@angular/core/testing';
import { HttpRequest, HttpHandlerFn, HttpResponse, HttpErrorResponse } from '@angular/common/http';
import { describe, it, expect, beforeEach } from 'vitest';
import { of, throwError, firstValueFrom } from 'rxjs';

import { httpErrorInterceptor } from './http-error.interceptor';
import { ApiError } from '../../models/api-error.model';

describe('httpErrorInterceptor', () => {
  beforeEach(() => TestBed.configureTestingModule({}));

  const run = (next: HttpHandlerFn) =>
    TestBed.runInInjectionContext(() =>
      httpErrorInterceptor(new HttpRequest('GET', '/api/test'), next),
    );

  const catchApiError = async (status: number, body: unknown = null): Promise<ApiError> => {
    try {
      await firstValueFrom(
        run(() => throwError(() => new HttpErrorResponse({ status, error: body }))),
      );
      expect.unreachable('Should have thrown');
    } catch (e) {
      return e as ApiError;
    }
    throw new Error();
  };

  it('should pass through successful responses untouched', async () => {
    const res = await firstValueFrom(run(() => of(new HttpResponse({ status: 200, body: { data: 'ok' } }))));
    expect(res).toBeInstanceOf(HttpResponse);
    expect((res as HttpResponse<{ data: string }>).body).toEqual({ data: 'ok' });
  });

  it('should rethrow errors as ApiError (not HttpErrorResponse)', async () => {
    const err = await catchApiError(500);
    expect(err).not.toBeInstanceOf(HttpErrorResponse);
    expect(err).toHaveProperty('status');
    expect(err).toHaveProperty('message');
  });

  // object body with string `error` property -> use it as message
  it.each([
    [400, { error: 'Email is required' },                                     'Email is required'],
    [422, { error: 'Validation failed on field X' },                          'Validation failed on field X'],
    [400, { error: 'Bad input', details: ['f1'], code: 'V' },                 'Bad input'],
    [400, { error: '' },                                                       ''],
  ])('object body with error string (status %d, body %o) -> message "%s"', async (status, body, msg) => {
    expect(await catchApiError(status, body)).toEqual({ status, message: msg });
  });

  // plain non-empty string body -> use as message
  it.each([
    [400, 'Something went wrong'],
    [503, 'Service Unavailable'],
  ])('string body (status %d) -> message "%s"', async (status, body) => {
    expect(await catchApiError(status, body)).toEqual({ status, message: body });
  });

  // fallback cases — body carries no usable message
  it.each([
    // invalid object bodies
    [500, { error: 123 },           'Server error'],
    [404, { error: undefined },     'Not found'],
    [403, { code: 'FORBIDDEN' },    'Access denied'],
    [401, {},                       'Unauthorized'],
    // empty / null / undefined / primitive bodies
    [500, '',                       'Server error'],
    [404, null,                     'Not found'],
    [401, undefined,                'Unauthorized'],
    [400, ['e1', 'e2'],             'Invalid request'],
    [500, 42,                       'Server error'],
    [500, false,                    'Server error'],
  ])('fallback body (status %d, body %o) -> message "%s"', async (status, body, msg) => {
    expect(await catchApiError(status, body)).toEqual({ status, message: msg });
  });

  // fallbackMessage — mapped status codes
  it.each([
    [0,   'Cannot reach server'],
    [400, 'Invalid request'],
    [401, 'Unauthorized'],
    [403, 'Access denied'],
    [404, 'Not found'],
    [409, 'Already exists'],
    [422, 'Validation failed'],
    [500, 'Server error'],
  ])('fallback message for status %d -> "%s"', async (status, msg) => {
    expect(await catchApiError(status, null)).toEqual({ status, message: msg });
  });

  // fallbackMessage — unmapped status codes
  it.each([301, 405, 408, 429, 502, 503, 504])(
    'fallback message for unmapped status %d -> "Unexpected error"',
    async (status) => {
      expect(await catchApiError(status, null)).toEqual({ status, message: 'Unexpected error' });
    },
  );
});