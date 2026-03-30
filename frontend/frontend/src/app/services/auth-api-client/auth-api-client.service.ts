import { HttpClient } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { map, Observable } from 'rxjs';

import { LoginRequest } from '../../models/auth/login-request.model';
import { AuthResponse } from '../../models/auth/auth-response.model';
import { PasswordChange } from '../../models/auth/password-change.model';
import { environment } from '../../../environments/environment';
import { ConfirmAccountResponse } from '../../models/auth/confirm-account.model';
import { ForgotPasswordRequest } from '../../models/auth/forgot-password-request.model';
import { ForgotPasswordResponse } from '../../models/auth/forgot-password.model';

@Injectable({
  providedIn: 'root',
})
export class AuthApiClientService {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = `${environment.apiUrl}/auth`;

  // Si aspetta di ricevere solo il token JWT
  public login(req: LoginRequest): Observable<AuthResponse> {
    return this.http.post<AuthResponse>(`${this.apiUrl}/login`, {
      email: req.email,
      password: req.password,
      user_role: req.userRole,
      tenant_id: req.tenantId,
    });
  }

  public logout(): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/logout`, {});
  }

  // API DOG - Password Dimenticata
  // Da dove lo tiro fuori il tenant id ???
  public verifyForgotPasswordToken(token: string): Observable<boolean> {
    return this.http
      .get<{ result: boolean }>(`${this.apiUrl}/forgot_password/verify_token/${token}`, {})
      .pipe(map((response) => response.result));
  }

  public forgotPasswordRequest(forgotPasswordRequest: ForgotPasswordRequest): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/forgot_password/request`, {
      email: forgotPasswordRequest.email,
      tenant_id: forgotPasswordRequest.tenantId,
    });
  }

  public confirmPasswordReset(req: ForgotPasswordResponse): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/forgot_password`, {
      token: req.token,
      new_password: req.newPassword,
    });
  }

  // API DOG - Cambia password (serve old e new password)
  public confirmPasswordChange(data: PasswordChange): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/change_password`, {
      old_password: data.oldPassword,
      new_password: data.newPassword,
    });
  }

  // API DOG - Conferma account
  // Da dove lo tiro fuori il tenant id ???
  public verifyAccountToken(token: string): Observable<boolean> {
    return this.http
      .get<{ result: boolean }>(`${this.apiUrl}/confirm_account/verify_token/${token}`, {})
      .pipe(map((response) => response.result));
  }

  public confirmAccountCreation(req: ConfirmAccountResponse): Observable<AuthResponse> {
    return this.http.post<AuthResponse>(`${this.apiUrl}/confirm_account`, {
      token: req.token,
      new_password: req.newPassword,
    });
  }
}
