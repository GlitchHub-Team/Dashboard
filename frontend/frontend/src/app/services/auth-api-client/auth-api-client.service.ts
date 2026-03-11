import { HttpClient } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { LoginRequest } from '../../models/login-request.model';
import { AuthResponse } from '../../models/auth-response.model';
import { Observable } from 'rxjs';
import { PasswordChange } from '../../models/password-change.model';
import { environment } from '../../../environments/environment';

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

  public logout(userId: number): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/logout`, { userId });
  }

  public forgotPassword(email: string): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/forgot-password`, { email });
  }

  public requestPasswordChange(userId: number): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/request-password-change`, { userId });
  }

  public confirmPasswordChange(data: PasswordChange): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/confirm-password-change`, data);
  }
}
