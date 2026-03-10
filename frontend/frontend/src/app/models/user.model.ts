import { UserRole } from './user-role.enum';

export interface User {
  id: number;
  name: string;
  email: string;
  role: UserRole;
  tenantId: number;
}
