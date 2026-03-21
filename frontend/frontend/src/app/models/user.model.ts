import { UserRole } from './user-role.enum';

// TODO: User deve modellare anche super admin?
export interface User {
  id: string;
  email: string;
  role: UserRole;
  tenantId?: string;
}
