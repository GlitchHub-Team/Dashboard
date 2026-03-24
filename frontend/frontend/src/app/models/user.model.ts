import { UserRole } from './user-role.enum';

// TODO: User deve modellare anche super admin?
export interface User {
  id: string;
  username: string;
  email: string;
  role: UserRole;
  tenantId: string;
}
