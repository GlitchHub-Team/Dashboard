import { inject, Injectable, signal } from '@angular/core';
import { Observable, tap, catchError, finalize, EMPTY, switchMap } from 'rxjs';

import { ApiError } from '../../models/api-error.model';
import { PasswordChange } from '../../models/auth/password-change.model';
import { AuthApiClientService } from '../auth-api-client/auth-api-client.service';
import { ConfirmAccountResponse } from '../../models/auth/confirm-account.model';
import { ForgotPasswordRequest } from '../../models/auth/forgot-password-request.model';
import { ForgotPasswordResponse } from '../../models/auth/forgot-password.model';
import { AuthResponse } from '../../models/auth/auth-response.model';
import { UserSessionService } from '../user-session/user-session.service';
import { TokenStorageService } from '../token-storage/token-storage.service';

@Injectable({
  providedIn: 'root',
})
export class AuthActionsService {
  private readonly authApiClient = inject(AuthApiClientService);
  private readonly tokenStorage = inject(TokenStorageService);
  private readonly userSession = inject(UserSessionService);

  private readonly _loading = signal(false);
  private readonly _error = signal<string | null>(null);
  private readonly _passwordChangeResult = signal<boolean | null>(null);

  public readonly loading = this._loading.asReadonly();
  public readonly error = this._error.asReadonly();
  public readonly passwordChangeResult = this._passwordChangeResult.asReadonly();

  // Manda la mail per il reset della password (quindi utente non loggato)
  public forgotPassword(req: ForgotPasswordRequest): Observable<void> {
    this.setLoadingState();

    return this.authApiClient.forgotPasswordRequest(req).pipe(
      catchError((err: ApiError) => {
        this._error.set(err.message ?? 'Failed to send reset email');
        return EMPTY;
      }),
      finalize(() => this._loading.set(false)),
    );
  }

  // Cambia la password (utente loggato)
  public confirmPasswordChange(req: PasswordChange): Observable<void> {
    this.setLoadingState();
    this._passwordChangeResult.set(null);

    return this.authApiClient.confirmPasswordChange(req).pipe(
      tap(() => this._passwordChangeResult.set(true)),
      catchError((err: ApiError) => {
        this._passwordChangeResult.set(false);
        this._error.set(err.message ?? 'Failed to change password');
        return EMPTY;
      }),
      finalize(() => this._loading.set(false)),
    );
  }

  // Cambia la password (utente non loggato, reset password)
  public confirmPasswordReset(req: ForgotPasswordResponse): Observable<void> {
    this.setLoadingState();

    return this.authApiClient.verifyForgotPasswordToken(req.token).pipe(
      switchMap(() => {
        // Non ritorna niente ma semplicemente aggiorna i propri signal per indicare il successo
        return this.authApiClient.confirmPasswordReset(req).pipe(
          tap(() => {
            this._passwordChangeResult.set(true);
          }),
        );
      }),
      catchError((err: ApiError) => {
        this._error.set(err.message ?? 'Failed to reset password');
        return EMPTY;
      }),
      finalize(() => this._loading.set(false)),
    );
  }

  // Conferma la creazione dell'account (dopo che l'utente ha cliccato sul link di conferma ricevuto via mail)
  public confirmAccount(req: ConfirmAccountResponse): Observable<AuthResponse> {
    this.setLoadingState();

    // confirmAccountCreation ritorna il JWT legato all'account confermato
    return this.authApiClient.verifyAccountToken(req.token).pipe(
      switchMap(() => {
        return this.authApiClient.confirmAccountCreation(req).pipe(
          tap((response) => {
            // Salva il token JWT e inizializza la sessione utente
            // per loggare automaticamente l'utente dopo la conferma dell'account
            this.tokenStorage.saveToken(response.jwt);
            this.userSession.initSession(response.jwt);
          }),
        );
      }),
      catchError((err: ApiError) => {
        this._error.set(err.message ?? 'Failed to confirm account');
        return EMPTY;
      }),
      finalize(() => this._loading.set(false)),
    );
  }

  public clearMessages(): void {
    this._error.set(null);
    this._passwordChangeResult.set(null);
  }

  private setLoadingState(): void {
    this._loading.set(true);
    this._error.set(null);
  }
}
