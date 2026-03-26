import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { UserRole } from '../../models/user/user-role.enum';
import { environment } from '../../../environments/environment';
import { PaginatedUserResponse } from '../../models/user/paginated-user-response.model';
import { UserBackend } from '../../models/user/user-backend.model';
import { UserConfig } from '../../models/user/user-config.model';

@Injectable({ providedIn: 'root' })
export class UserApiClientService {
  private readonly http = inject(HttpClient);

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
    page: number,
    size: number,
    tenantId?: string,
  ): Observable<PaginatedUserResponse<UserBackend>> {
    const params = new HttpParams().set('page', page.toString()).set('size', size.toString());

    const url = this.getBaseUrl(role, tenantId, true);

    return this.http.get<PaginatedUserResponse<UserBackend>>(url, { params });
  }

  public getUser(id: string, role: UserRole, tenantId?: string): Observable<UserBackend> {
    const url = `${this.getBaseUrl(role, tenantId, false)}/${id}`;
    return this.http.get<UserBackend>(url);
  }

  public createUser(
    config: UserConfig,
    role: UserRole,
    tenantId?: string,
  ): Observable<UserBackend> {
    const url = this.getBaseUrl(role, tenantId, false);
    return this.http.post<UserBackend>(url, config);
  }

  public deleteUser(id: string, role: UserRole, tenantId?: string): Observable<void> {
    const url = `${this.getBaseUrl(role, tenantId, false)}/${id}`;
    return this.http.delete<void>(url);
  }
}
