import { computed, inject, Injectable } from '@angular/core';
import { Router } from '@angular/router';
import { UserSessionService } from '../user-session/user-session.service';
import { TokenStorageService } from '../token-storage/token-storage.service';
import { AuthApiClientService } from '../auth-api-client/auth-api-client.service';
import { LoginRequest } from '../../models/login-request.model';
import { tap, Observable } from 'rxjs';
import { AuthResponse } from '../../models/auth-response.model';

@Injectable({
  providedIn: 'root',
})
export class AuthSessionService {
  private authApiClient = inject(AuthApiClientService);
  private tokenStorage = inject(TokenStorageService);
  private userSession = inject(UserSessionService);
  private router = inject(Router);

  // TODO: Da rivedere per come utilizziamo il token che viene (forse) inviato dal backend
  public readonly isAuthenticated = computed(
    () => this.tokenStorage.isValid() && this.userSession.currentUser() !== null,
  );

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
    const user = this.userSession.currentUser();

    if (user?.id) {
      this.authApiClient.logout(user.id).subscribe({
        next: () => this.clearAndRedirect(),
        error: () => this.clearAndRedirect(),
      });
    } else {
      this.clearAndRedirect();
    }
  }

  private clearAndRedirect(): void {
    this.tokenStorage.clearToken();
    this.userSession.clearSession();
    this.router.navigate(['/login']);
  }
}
