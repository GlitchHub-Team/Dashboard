import { Injectable, signal, inject } from '@angular/core';
import { Observable, tap } from 'rxjs';
import { UserApiClientService, UserConfig } from './user-api-client.service';
import { UserRole } from '../../models/user-role.enum';
import { User } from '../../models/user.model';

@Injectable({
  providedIn: 'root'
})
export class UserService {
  private readonly userApi = inject(UserApiClientService);

  private readonly _loading = signal<boolean>(false);
  private readonly _error = signal<string | null>(null);
  private readonly _userList = signal<User[]>([]);

  public loading = this._loading.asReadonly();
  public error = this._error.asReadonly();
  public userList = this._userList.asReadonly();

  public retrieveUser(role?: UserRole): void {
    this._loading.set(true);
    this._error.set(null);
    
    this.userApi.getUsers(role).subscribe({
      next: (users) => {
        this._userList.set(users);
        this._loading.set(false);
      },
      error: (err: Error) => {
        this._error.set(err.message || 'Errore nel recupero degli utenti');
        this._loading.set(false);
      }
    });
  }

  public addNewUser(config: UserConfig): Observable<User> {
    this._loading.set(true);
    return this.userApi.createUser(config).pipe(
      tap({
        next: () => this._loading.set(false),
        error: () => this._loading.set(false)
      })
    );
  }

  public removeUser(email: string): Observable<void> {
    this._loading.set(true);
    return this.userApi.deleteUser(email).pipe(
      tap({
        next: () => this._loading.set(false),
        error: () => this._loading.set(false)
      })
    );
  }
}
