import { UserRole } from '../user/user-role.enum';

export interface UserSession {
  userId: string;
  tenantId?: string;
  role: UserRole;
}
