import { EnumMapper } from './enum.utils';
import { UserRole } from '../models/user/user-role.enum';

export const userRoleMapper = new EnumMapper<UserRole, string>(
  {
    [UserRole.SUPER_ADMIN]: 'super_admin',
    [UserRole.TENANT_ADMIN]: 'tenant_admin',
    [UserRole.TENANT_USER]: 'tenant_user',
  },
  UserRole.TENANT_USER,
);
