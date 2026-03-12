import { computed, Injectable, signal } from '@angular/core';

import { User } from '../../models/user.model';
import { UserRole } from '../../models/user-role.enum';

@Injectable({
  providedIn: 'root',
})
export class UserSessionService {
  private _currentUser = signal<User | null>(this.loadFromStorage());

  public readonly currentUser = this._currentUser.asReadonly();
  public readonly currentRole = computed<UserRole | null>(() => this.currentUser()?.role || null);
  // TODO: Meglio far in modo che appaia il nome del Tenant
  public readonly currentTenant = computed<string | null>(
    () => this.currentUser()?.tenantId || null,
  );

  public initSession(user: User): void {
    sessionStorage.setItem('currentUser', JSON.stringify(user));
    this._currentUser.set(user);
  }

  public clearSession(): void {
    sessionStorage.removeItem('currentUser');
    this._currentUser.set(null);
  }

  private loadFromStorage(): User | null {
    try {
      const userData = sessionStorage.getItem('currentUser');
      return userData ? (JSON.parse(userData) as User) : null;
    } catch {
      return null;
    }
  }
}
