import { Injectable, signal, inject } from '@angular/core';
import { Observable, tap } from 'rxjs';
import { UserApiClientService, UserConfig } from './user-api-client.service';
import { UserRole } from '../../models/user-role.enum';
import { User } from '../../models/user.model';

@Injectable({ providedIn: 'root' })
export class UserService {
  private readonly userApi = inject(UserApiClientService);

  public readonly loading = signal<boolean>(false);
  public readonly error = signal<string | null>(null);
  public readonly userList = signal<User[]>([]);

  public retrieveUser(role?: UserRole): void {
    this.loading.set(true);
    this.error.set(null);
    
    this.userApi.getUsers(role).subscribe({
      next: (users) => {
        this.userList.set(users);
        this.loading.set(false);
      },
      error: (err: Error) => {
        this.error.set(err.message || 'Errore nel recupero degli utenti');
        this.loading.set(false);
      }
    });
  }

  public addNewUser(config: UserConfig): Observable<User> {
    this.loading.set(true);
    return this.userApi.createUser(config).pipe(
      tap({
        next: () => this.loading.set(false),
        error: () => this.loading.set(false)
      })
    );
  }

  public removeUser(email: string): Observable<void> {
    this.loading.set(true);
    return this.userApi.deleteUser(email).pipe(
      tap({
        next: () => this.loading.set(false),
        error: () => this.loading.set(false)
      })
    );
  }
}
