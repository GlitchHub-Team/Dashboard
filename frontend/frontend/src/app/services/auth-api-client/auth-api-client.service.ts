import { HttpClient } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { LoginRequest } from '../../models/login-request.model';
import { AuthResponse } from '../../models/auth-response.model';
import { Observable } from 'rxjs';
import { PasswordReset } from '../../models/password-reset.model';
import { PasswordChange } from '../../models/password-change.model';
import { environment } from '../../../environments/environment';

@Injectable({
  providedIn: 'root',
})
export class AuthApiClientService {
  private http = inject(HttpClient);
  private apiUrl = `${environment.apiUrl}/auth`;

  public login(req: LoginRequest): Observable<AuthResponse> {
    return this.http.post<AuthResponse>(`${this.apiUrl}/login`, req);
  }

  public logout(): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/logout`, {});
  }

  public requestPasswordReset(email: string): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/forgot-password`, { email });
  }

  public resetPassword(resetPasswordData: PasswordReset): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/reset-password`, resetPasswordData);
  }

  public changePassword(changePasswordData: PasswordChange): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/change-password`, changePasswordData);
  }
}
