import { UserRole } from './user-role.enum';

export interface User {
  id: string;
  username: string;
  email: string;
  role: UserRole;
  tenantId?: string;
}
