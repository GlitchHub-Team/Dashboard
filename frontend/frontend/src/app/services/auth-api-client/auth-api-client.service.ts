import { HttpClient } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';

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

  // Riferimento su API DOG sezione User Auth https://app.apidog.com/project/1225781

  public login(req: LoginRequest): Observable<AuthResponse> {
    return this.http.post<AuthResponse>(`${this.apiUrl}/login`, {
      email: req.email,
      password: req.password,
      tenant_id: req.tenantId,
    });
  }

  public logout(): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/logout`, {});
  }

  public verifyForgotPasswordToken(token: string, tenantId?: string): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/forgot_password/verify_token`, {
      token: token,
      tenant_id: tenantId,
    });
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
      tenant_id: req.tenantId,
      new_password: req.newPassword,
    });
  }

  public confirmPasswordChange(data: PasswordChange): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/change_password`, {
      old_password: data.oldPassword,
      new_password: data.newPassword,
    });
  }

  public verifyAccountToken(token: string, tenantId?: string): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/confirm_account/verify_token/`, {
      token: token,
      tenant_id: tenantId,
    });
  }

  public confirmAccountCreation(req: ConfirmAccountResponse): Observable<AuthResponse> {
    return this.http.post<AuthResponse>(`${this.apiUrl}/confirm_account`, {
      token: req.token,
      tenant_id: req.tenantId,
      new_password: req.newPassword,
    });
  }
}
