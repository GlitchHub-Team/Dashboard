import { UserRole } from './user-role.enum';

export interface RawUserConfig {
  id: string;
  email: string;
  role: UserRole;
  tenantId: string;
}
