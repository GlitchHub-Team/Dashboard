import { Observable } from 'rxjs';

import { User } from '../../models/user/user.model';
import { UserConfig } from '../../models/user/user-config.model';
import { UserRole } from '../../models/user/user-role.enum';
import { PaginatedUserResponse } from '../../models/user/paginated-user-response.model';

export abstract class UserApiClientAdapter {
  abstract getUser(userId: string, role: UserRole, tenantId?: string): Observable<User>;
  abstract getUsers(
    role: UserRole,
    page: number,
    limit: number,
    tenantId?: string,
  ): Observable<PaginatedUserResponse<User>>;
  abstract createUser(config: UserConfig, role: UserRole, tenantId?: string): Observable<User>;
  abstract deleteUser(id: string, role: UserRole, tenantId?: string): Observable<void>;
}