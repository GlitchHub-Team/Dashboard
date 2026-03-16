import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable, map } from 'rxjs';
import { UserDataAdapter } from './user-data.adapter';
import { User } from '../../models/user.model';
import { UserRole } from '../../models/user-role.enum';
import { environment } from '../../../environments/environment';
import { RawUserConfig } from '../../models/raw-user-config.model';

export interface UserConfig {
  email: string;
  role: UserRole;
}

@Injectable({ providedIn: 'root'})
export class UserApiClientService {
  private readonly http = inject(HttpClient);
  private readonly userAdapter = inject(UserDataAdapter);
  private readonly apiUrl = `${environment.apiUrl}/users`;

  public getUsers(role?: UserRole): Observable<User[]> {
    let params = new HttpParams();
    if (role) {
      params = params.set('role', role);
    }
    return this.http.get<RawUserConfig[]>(this.apiUrl, { params }).pipe(
      map(data => this.userAdapter.adaptArray(data))
    );
  }

  public createUser(config: UserConfig): Observable<User> {
    return this.http.post<RawUserConfig>(this.apiUrl, config).pipe(
      map(data => this.userAdapter.adapt(data))
    );
  }

  public deleteUser(email: string): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/${email}`);
  }
}
