import { Injectable, signal, inject } from '@angular/core';
import { Observable, tap, catchError, EMPTY, finalize, map } from 'rxjs';
import { UserApiClientService } from '../user-api-client/user-api-client.service';
import { UserConfig } from '../../models/user/user-config.model';
import { UserRole } from '../../models/user/user-role.enum';
import { User } from '../../models/user/user.model';
import { ApiError } from '../../models/api-error.model';
import { UserAdapter } from '../../adapters/user.adapter';

@Injectable({ providedIn: 'root' })
export class UserService {
  private readonly userApi = inject(UserApiClientService);
  private readonly adapter = inject(UserAdapter);

  private readonly _loading = signal(false);
  private readonly _error = signal<string | null>(null);
  private readonly _userList = signal<User[]>([]);
  private readonly _total = signal(0);
  private readonly _pageIndex = signal(0);
  private readonly _limit = signal(10);

  public readonly loading = this._loading.asReadonly();
  public readonly error = this._error.asReadonly();
  public readonly userList = this._userList.asReadonly();
  public readonly total = this._total.asReadonly();
  public readonly pageIndex = this._pageIndex.asReadonly();
  public readonly limit = this._limit.asReadonly();

  public getUser(userId: string, role: UserRole, tenantId?: string): Observable<User> {
    this._loading.set(true);

    return this.userApi.getUser(userId, role, tenantId).pipe(
      map((dto) => this.adapter.fromDTO(dto)),
      tap({
        error: (err: ApiError) => {
          this._error.set(err.message ?? 'Failed to load user');
        },
      }),
      finalize(() => this._loading.set(false)),
    );
  }

  public retrieveUser(role: UserRole, tenantId?: string): void {
    this.setGettingUsersState();

    this.userApi
      .getUsers(role, this._pageIndex(), this._limit(), tenantId)
      .pipe(
        map((response) => this.adapter.fromPaginatedDTO(response)),
        tap((result) => {
          this._userList.set(result.users);
          this._total.set(result.total);
        }),
        catchError((err: ApiError) => {
          this._error.set(err.message ?? 'Failed to load users');
          return EMPTY;
        }),
        finalize(() => this._loading.set(false)),
      )
      .subscribe();
  }

  public changePage(pageIndex: number, limit: number, role: UserRole, tenantId?: string): void {
    this._pageIndex.set(pageIndex);
    this._limit.set(limit);
    this.retrieveUser(role, tenantId);
  }

  public addNewUser(config: UserConfig, role: UserRole, tenantId?: string): Observable<User> {
    this.setLoadingState();

    return this.userApi.createUser(config, role, tenantId).pipe(
      map((dto) => this.adapter.fromDTO(dto)),
      tap({
        error: (err: ApiError) => {
          this._error.set(err.message ?? 'Failed to create user');
        },
      }),
      finalize(() => this._loading.set(false)),
    );
  }

  public removeUser(user: User): Observable<void> {
    this.setLoadingState();

    return this.userApi.deleteUser(user.id, user.role, user.tenantId).pipe(
      tap({
        error: (err: ApiError) => {
          this._error.set(err.message ?? 'Failed to delete user');
        },
      }),
      finalize(() => this._loading.set(false)),
    );
  }

  private setGettingUsersState(): void {
    this._userList.set([]);
    this._loading.set(true);
    this._error.set(null);
  }

  private setLoadingState(): void {
    this._loading.set(true);
    this._error.set(null);
  }
}
