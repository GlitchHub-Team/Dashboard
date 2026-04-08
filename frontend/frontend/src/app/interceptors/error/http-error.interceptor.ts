import { HttpInterceptorFn, HttpErrorResponse } from '@angular/common/http';
import { catchError, throwError } from 'rxjs';
import { ApiError } from '../../models/api-error.model';

interface BackendError {
  error?: string;
}

export const httpErrorInterceptor: HttpInterceptorFn = (req, next) => {
  return next(req).pipe(
    catchError((response: HttpErrorResponse) => {
      const apiError = buildApiError(response);
      return throwError(() => apiError);
    }),
  );
};

function buildApiError(response: HttpErrorResponse): ApiError {
  const body = response.error as BackendError | string | null;

  if (body && typeof body === 'object' && typeof body.error === 'string') {
    return {
      status: response.status,
      message: body.error,
    };
  }

  if (typeof body === 'string' && body.length > 0) {
    return {
      status: response.status,
      message: body,
    };
  }

  return {
    status: response.status,
    message: fallbackMessage(response.status),
  };
}

function fallbackMessage(status: number): string {
  switch (status) {
    case 0:
      return 'Cannot reach server';
    case 400:
      return 'Invalid request';
    case 401:
      return 'Unauthorized';
    case 403:
      return 'Access denied';
    case 404:
      return 'Not found';
    case 409:
      return 'Already exists';
    case 422:
      return 'Validation failed';
    case 500:
      return 'Server error';
    default:
      return 'Unexpected error';
  }
}
