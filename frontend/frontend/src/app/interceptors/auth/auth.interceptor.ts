import { HttpInterceptorFn, HttpErrorResponse } from '@angular/common/http';
import { inject } from '@angular/core';
import { Router } from '@angular/router';
import { catchError, throwError } from 'rxjs';
import { TokenStorageService } from '../../services/token-storage/token-storage.service';
import { UserSessionService } from '../../services/user-session/user-session.service';

export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const tokenStorage = inject(TokenStorageService);
  const userSession = inject(UserSessionService);
  const router = inject(Router);

  const token = tokenStorage.getToken();

  const authReq = token ? req.clone({ setHeaders: { Authorization: `Bearer ${token}` } }) : req;

  return next(authReq).pipe(
    catchError((error: HttpErrorResponse) => {
      if (error.status === 401) {
        tokenStorage.clearToken();
        userSession.clearSession();
        router.navigate(['/login']);
      }
      return throwError(() => error);
    }),
  );
};
