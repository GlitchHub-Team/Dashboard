import { computed, inject, Injectable, signal } from '@angular/core';
import { Router } from '@angular/router';
import { Observable, tap, catchError, finalize, EMPTY } from 'rxjs';

import { UserSessionService } from '../user-session/user-session.service';
import { TokenStorageService } from '../token-storage/token-storage.service';
import { LoginRequest } from '../../models/auth/login-request.model';
import { AuthResponse } from '../../models/auth/auth-response.model';
import { ApiError } from '../../models/api-error.model';
import { AuthApiClientAdapter } from '../auth-api-client/auth-api-client-adapter.service';

@Injectable({
  providedIn: 'root',
})
export class AuthSessionService {
  private readonly authApiClient = inject(AuthApiClientAdapter);
  private readonly tokenStorage = inject(TokenStorageService);
  private readonly userSession = inject(UserSessionService);
  private readonly router = inject(Router);

  private readonly _loading = signal(false);
  private readonly _error = signal<string | null>(null);

  public readonly loading = this._loading.asReadonly();
  public readonly error = this._error.asReadonly();
  public readonly isAuthenticated = computed(
    () => this.tokenStorage.isValid() && this.userSession.currentUser() !== null,
  );

  public login(req: LoginRequest): Observable<AuthResponse> {
    this.setLoadingState();

    // Fa la richiesta di login e, in caso di successo, salva il token JWT
    // e inizializza la sessione utente
    return this.authApiClient.login(req).pipe(
      tap((response) => {
        this.tokenStorage.saveToken(response.jwt);
        this.userSession.initSession(response.jwt);
      }),
      catchError((err: ApiError) => {
        this._error.set(err.message ?? 'Login failed');
        return EMPTY;
      }),
      finalize(() => this._loading.set(false)),
    );
  }

  // In ogni caso di logout, sia che la chiamata al backend vada a buon fine o fallisca,
  // vogliamo pulire la sessione e reindirizzare al login
  public logout(): void {
    this.authApiClient.logout().subscribe({
      next: () => this.clearAndRedirect(),
      error: () => this.clearAndRedirect(),
    });
  }

  public clearError(): void {
    this._error.set(null);
  }

  private setLoadingState(): void {
    this._loading.set(true);
    this._error.set(null);
  }

  private clearAndRedirect(): void {
    this.tokenStorage.clearToken();
    this.userSession.clearSession();
    this.router.navigate(['/login']);
  }
}
