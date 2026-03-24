import { Injectable, signal, inject } from '@angular/core';
import { Observable, tap } from 'rxjs';
import { UserApiClientService, UserConfig } from '../user-api-client/user-api-client.service';
import { UserRole } from '../../models/user/user-role.enum';
import { User } from '../../models/user/user.model';

@Injectable({ providedIn: 'root' })
export class UserService {
  private readonly userApi = inject(UserApiClientService);

  public readonly loading = signal<boolean>(false);
  public readonly error = signal<string | null>(null);
  public readonly userList = signal<User[]>([]);
  public readonly total = signal<number>(0);
  public readonly pageIndex = signal<number>(0);
  public readonly limit = signal<number>(10);

  public retrieveUser(role: UserRole, tenantId?: string): void {
    this.loading.set(true);
    this.error.set(null);

    this.userApi.getUsers(role, tenantId, this.pageIndex(), this.limit()).subscribe({
      next: (res) => {
        this.userList.set(res.items);
        this.total.set(res.totalCount);
        this.loading.set(false);
      },
      error: (err: Error) => {
        this.error.set(err.message || 'Errore nel recupero degli utenti');
        this.loading.set(false);
      },
    });
  }

  public changePage(pageIndex: number, limit: number, role: UserRole, tenantId?: string): void {
    this.pageIndex.set(pageIndex);
    this.limit.set(limit);
    this.retrieveUser(role, tenantId);
  }

  public addNewUser(config: UserConfig, tenantId?: string): Observable<User> {
    this.loading.set(true);
    return this.userApi.createUser(config, tenantId).pipe(
      tap({
        next: () => this.loading.set(false),
        error: () => this.loading.set(false),
      }),
    );
  }

  public removeUser(user: User): Observable<void> {
    this.loading.set(true);
    return this.userApi.deleteUser(user.id, user.role, user.tenantId).pipe(
      tap({
        next: () => this.loading.set(false),
        error: () => this.loading.set(false),
      }),
    );
  }
}
