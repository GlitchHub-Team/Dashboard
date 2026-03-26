import { inject, Injectable, signal } from '@angular/core';

import { UserSession } from '../../models/auth/user-session.model';
import { TokenStorageService } from '../token-storage/token-storage.service';
import { userRoleMapperJWT } from '../../utils/user-role-jwt.utils';

@Injectable({
  providedIn: 'root',
})
export class UserSessionService {
  private tokenStorage = inject(TokenStorageService);
  private _currentUser = signal<UserSession | null>(null);

  public readonly currentUser = this._currentUser.asReadonly();

  constructor() {
    this.restoreSession();
  }

  // Dai campi del JWT costruisce la sessione utente e la salva
  // sia nello stato che nel sessionStorage.
  // Il JWT dovrebbe contenere info su userId, userRole e tenantId (quando non SUPER_ADMIN)
  public initSession(token: string): void {
    const session = this.decodeToken(token);
    if (!session) {
      console.warn('Failed to decode JWT');
      return;
    }

    this._currentUser.set(session);
    sessionStorage.setItem('currentUser', JSON.stringify(session));
  }

  public clearSession(): void {
    sessionStorage.removeItem('currentUser');
    this._currentUser.set(null);
  }

  private restoreSession(): void {
    const stored = this.loadFromStorage();
    if (stored) {
      this._currentUser.set(stored);
      return;
    }

    const token = this.tokenStorage.getToken();
    if (token && this.tokenStorage.isTokenValid()) {
      this.initSession(token);
    } else {
      this.clearSession();
    }
  }

  // Mappa i campi del JWT alla struttura UserSession.
  // Se il token non è valido o mancano campi, ritorna null
  private decodeToken(token: string): UserSession | null {
    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      return {
        userId: payload.uid.toString(),
        role: userRoleMapperJWT.fromBackend(payload.rol),
        tenantId: payload.tid || undefined,
      };
    } catch {
      return null;
    }
  }

  private loadFromStorage(): UserSession | null {
    try {
      const data = sessionStorage.getItem('currentUser');
      return data ? (JSON.parse(data) as UserSession) : null;
    } catch {
      return null;
    }
  }
}
