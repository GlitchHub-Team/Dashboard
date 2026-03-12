import { Permission } from './permission.enum';
import { NavItem } from './nav-item.model';

export const NAV_ITEMS: NavItem[] = [
  {
    label: 'Dashboard',
    icon: 'dashboard',
    route: '/dashboard',
    permission: Permission.DASHBOARD_ACCESS,
  },
  {
    label: 'Gateway Management',
    icon: 'settings',
    route: '/gateway-management',
    permission: Permission.GATEWAY_MANAGEMENT,
  },
  {
    label: 'User Management',
    icon: 'people',
    route: '/user-management',
    permission: [Permission.TENANT_USER_MANAGEMENT, Permission.TENANT_ADMIN_MANAGEMENT],
  },
  {
    label: 'Tenant Management',
    icon: 'business',
    route: '/tenant-management',
    permission: Permission.TENANT_MANAGEMENT,
  },
  {
    label: 'API Key Management',
    icon: 'vpn_key',
    route: '/apikey-management',
    permission: Permission.APIKEY_MANAGEMENT,
  },
];
