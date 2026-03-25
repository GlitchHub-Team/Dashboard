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
  // TODO: Double check dell'api url e delle routes rispetto al backend
  private http = inject(HttpClient);
  private apiUrl = `${environment.apiUrl}/auth`;

  public login(req: LoginRequest): Observable<AuthResponse> {
    return this.http.post<AuthResponse>(`${this.apiUrl}/login`, req);
  }

  public logout(userId: string): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/logout`, { userId });
  }

  // API DOG - Password Dimenticata
  public verifyForgotPasswordToken(token: string): Observable<boolean> {
    return this.http.get<boolean>(`${this.apiUrl}/forgot_password/verify_token/${token}`, {});
  }

  // Serve anche il tenantId
  public forgotPasswordRequest(forgotPasswordRequest: ForgotPasswordRequest): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/forgot_password/request`, forgotPasswordRequest);
  }

  public confirmPasswordReset(req: ForgotPasswordResponse): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/forgot_password`, req);
  }

  // API DOG - Cambia password (serve old e new password)
  public confirmPasswordChange(data: PasswordChange): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/change_password`, data);
  }

  // API DOG - Conferma account
  public verifyAccountToken(token: string): Observable<boolean> {
    return this.http.get<boolean>(`${this.apiUrl}/confirm_account/verify_token/${token}`, {});
  }

  public confirmAccountCreation(req: ConfirmAccountResponse): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/confirm_account`, req);
  }
}
