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

  // Non attaccare token alle richieste di login
  const isLoginRequest = req.url.includes('/auth/login');
  const token = tokenStorage.getToken();

  const authReq =
    token && !isLoginRequest
      ? req.clone({ setHeaders: { Authorization: `Bearer ${token}` } })
      : req;

  return next(authReq).pipe(
    catchError((error: HttpErrorResponse) => {
      if (error.status === 401 && !isLoginRequest) {
        // Redirect solo se la sessione non è stata ancora pulita
        if (tokenStorage.getToken()) {
          tokenStorage.clearToken();
          userSession.clearSession();
          router.navigate(['/login']);
        }
      }
      return throwError(() => error);
    }),
  );
};
