import { computed, inject, Injectable, signal } from '@angular/core';
import { Router } from '@angular/router';
import { UserSessionService } from '../user-session/user-session.service';
import { TokenStorageService } from '../token-storage/token-storage.service';
import { AuthApiClientService } from '../auth-api-client/auth-api-client.service';
import { LoginRequest } from '../../models/login-request.model';
import { PasswordReset } from '../../models/password-reset.model';
import { PasswordChange } from '../../models/password-change.model';
import { tap, Observable } from 'rxjs';
import { AuthResponse } from '../../models/auth-response.model';

@Injectable({
  providedIn: 'root',
})
export class AuthService {
  private authApiClient = inject(AuthApiClientService);
  private tokenStorage = inject(TokenStorageService);
  private userSession = inject(UserSessionService);
  private router = inject(Router);

  // TODO: Da rivedere per come utilizziamo il token che viene (forse) inviato dal backend
  public readonly isAuthenticated = computed(
    () => this.tokenStorage.isValid() && this.userSession.currentUser() !== null,
  );

  private _passwordChangeResult = signal<boolean | null>(null);
  readonly passwordChangeResult = this._passwordChangeResult.asReadonly();

  // Anche il service deve ritornare un Observable, così da poter gestire errori e successi in modo più flessibile nei componenti che lo utilizzano

  public login(req: LoginRequest): Observable<AuthResponse> {
    return this.authApiClient.login(req).pipe(
      tap((response) => {
        this.tokenStorage.saveToken(response.token);
        this.userSession.initSession(response.user);
      }),
    );
  }

  public logout(): void {
    this.authApiClient.logout().subscribe({
      next: () => this.clearAndRedirect(),
      error: () => this.clearAndRedirect(),
    });
  }

  public requestPasswordReset(email: string): Observable<void> {
    return this.authApiClient.requestPasswordReset(email);
  }

  // TODO: Validare i model creati confrontandosi con il backend

  public resetPassword(resetPasswordData: PasswordReset): Observable<void> {
    return this.authApiClient.resetPassword(resetPasswordData);
  }

  public changePassword(changePasswordData: PasswordChange): Observable<void> {
    return this.authApiClient
      .changePassword(changePasswordData)
      .pipe(tap(() => this._passwordChangeResult.set(true)));
  }

  private clearAndRedirect(): void {
    this.tokenStorage.clearToken();
    this.userSession.clearSession();
    this._passwordChangeResult.set(null);
    this.router.navigate(['/login']);
  }
}
