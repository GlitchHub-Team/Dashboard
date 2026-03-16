import { HttpInterceptorFn, HttpErrorResponse } from '@angular/common/http';
import { catchError, throwError } from 'rxjs';
import { ApiError } from '../../models/api-error.model';

export const httpErrorInterceptor: HttpInterceptorFn = (req, next) => {
  return next(req).pipe(
    catchError((response: HttpErrorResponse) => {
      const apiError: ApiError = {
        status: response.status,
        message: response.error?.message ?? fallbackMessage(response.status),
      };

      return throwError(() => apiError);
    }),
  );
};

function fallbackMessage(status: number): string {
  switch (status) {
    case 0:
      return 'Cannot reach server';
    case 400:
      return 'Invalid request';
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
