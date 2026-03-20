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
  public readonly totalUsers = signal<number>(0);

  public retrieveUser(role: UserRole, tenantId?: string, page = 0, size = 10): void {
    this.loading.set(true);
    this.error.set(null);
    
    this.userApi.getUsers(role, tenantId, page, size).subscribe({
      next: (res) => {
        this.userList.set(res.items);
        this.totalUsers.set(res.totalCount);
        this.loading.set(false);
      },
      error: (err: Error) => {
        this.error.set(err.message || 'Errore nel recupero degli utenti');
        this.loading.set(false);
      }
    });
  }

  public addNewUser(config: UserConfig, tenantId?: string): Observable<User> {
    this.loading.set(true);
    return this.userApi.createUser(config, tenantId).pipe(
      tap({
        next: () => this.loading.set(false),
        error: () => this.loading.set(false)
      })
    );
  }

  public removeUser(user: User): Observable<void> {
    this.loading.set(true);
    return this.userApi.deleteUser(user.id, user.role, user.tenantId).pipe(
      tap({
        next: () => this.loading.set(false),
        error: () => this.loading.set(false)
      })
    );
  }
}
