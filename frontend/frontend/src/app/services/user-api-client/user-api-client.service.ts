import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable, map } from 'rxjs';
import { UserDataAdapter, RawPaginatedResponse } from '../../adapters/user-data.adapter';
import { User } from '../../models/user/user.model';
import { UserRole } from '../../models/user/user-role.enum';
import { environment } from '../../../environments/environment';
import { RawUserConfig } from '../../models/user/raw-user-config.model';

export interface UserConfig {
  email: string;
  role: UserRole;
}

@Injectable({ providedIn: 'root' })
export class UserApiClientService {
  private readonly http = inject(HttpClient);
  private readonly userAdapter = inject(UserDataAdapter);

  private getBaseUrl(role: UserRole, tenantId?: string, isPlural = false): string {
    const baseUrl = environment.apiUrl;
    switch (role) {
      case UserRole.SUPER_ADMIN:
        return `${baseUrl}/super_admin${isPlural ? 's' : ''}`;
      case UserRole.TENANT_ADMIN:
        if (!tenantId) throw new Error('tenantId is required for TENANT_ADMIN');
        return `${baseUrl}/tenant/${tenantId}/tenant_admin${isPlural ? 's' : ''}`;
      case UserRole.TENANT_USER:
        if (!tenantId) throw new Error('tenantId is required for TENANT_USER');
        return `${baseUrl}/tenant/${tenantId}/tenant_user${isPlural ? 's' : ''}`;
      default:
        throw new Error(`Unsupported role: ${role}`);
    }
  }

  public getUsers(
    role: UserRole,
    tenantId?: string,
    page = 0,
    size = 10,
  ): Observable<{ items: User[]; totalCount: number }> {
    const params = new HttpParams().set('page', page.toString()).set('size', size.toString());

    const url = this.getBaseUrl(role, tenantId, true);

    return this.http
      .get<RawPaginatedResponse>(url, { params })
      .pipe(map((response) => this.userAdapter.adaptPaginated(response)));
  }

  public getUser(id: string, role: UserRole, tenantId?: string): Observable<User> {
    const url = `${this.getBaseUrl(role, tenantId, false)}/${id}`;
    return this.http.get<RawUserConfig>(url).pipe(map((data) => this.userAdapter.adapt(data)));
  }

  public createUser(config: UserConfig, tenantId?: string): Observable<User> {
    const url = this.getBaseUrl(config.role, tenantId, false);
    return this.http
      .post<RawUserConfig>(url, config)
      .pipe(map((data) => this.userAdapter.adapt(data)));
  }

  public deleteUser(id: string, role: UserRole, tenantId?: string): Observable<void> {
    const url = `${this.getBaseUrl(role, tenantId, false)}/${id}`;
    return this.http.delete<void>(url);
  }
}
