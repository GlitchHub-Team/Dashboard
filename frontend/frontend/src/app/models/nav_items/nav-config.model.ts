import { Permission } from '../permission.enum';
import { NavItem } from './nav-item.model';

export const NAV_ITEMS: NavItem[] = [
  {
    label: 'Dashboard',
    icon: 'dashboard',
    route: '/dashboard',
    permission: Permission.DASHBOARD_ACCESS,
  },
  {
    label: 'Gestione Gateway',
    icon: 'settings',
    route: '/gateway-management',
    permission: Permission.GATEWAY_MANAGEMENT,
  },
  {
    label: 'Gestione Tenant User',
    icon: 'people',
    route: '/user-management/tenant-users',
    permission: Permission.TENANT_USER_MANAGEMENT,
  },
  {
    label: 'Gestione Tenant Admin',
    icon: 'people',
    route: '/user-management/tenant-admins',
    permission: Permission.TENANT_ADMIN_MANAGEMENT,
  },
  {
    label: 'Gestione Super Admin',
    icon: 'people',
    route: '/user-management/super-admins',
    permission: Permission.SUPER_ADMIN_MANAGEMENT,
  },
  {
    label: 'Gestione Tenant',
    icon: 'business',
    route: '/tenant-management',
    permission: Permission.TENANT_MANAGEMENT,
  },
  {
    label: 'Gestione Api Key',
    icon: 'vpn_key',
    route: '/apikey-management',
    permission: Permission.APIKEY_MANAGEMENT,
  },
];
