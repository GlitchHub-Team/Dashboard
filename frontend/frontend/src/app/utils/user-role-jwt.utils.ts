import { EnumMapper } from './enum.utils';
import { UserRole } from '../models/user/user-role.enum';

export const userRoleMapperJWT = new EnumMapper<UserRole, string>(
  {
    [UserRole.SUPER_ADMIN]: 'sa',
    [UserRole.TENANT_ADMIN]: 'ta',
    [UserRole.TENANT_USER]: 'tu',
  },
  UserRole.TENANT_USER,
);
