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
    route: '/dashboard/gateway-management',
    permission: Permission.GATEWAY_MANAGEMENT,
  },
  {
    label: 'Tenant User Management',
    icon: 'people',
    route: '/dashboard/user-management/tenant-users',
    permission: [Permission.TENANT_USER_MANAGEMENT],
  },
  {
    label: 'Tenant Admin Management',
    icon: 'people',
    route: '/dashboard/user-management/tenant-admins',
    permission: [Permission.TENANT_ADMIN_MANAGEMENT],
  },
  {
    label: 'Tenant Management',
    icon: 'business',
    route: '/dashboard/tenant-management',
    permission: Permission.TENANT_MANAGEMENT,
  },
  {
    label: 'API Key Management',
    icon: 'vpn_key',
    route: '/dashboard/apikey-management',
    permission: Permission.APIKEY_MANAGEMENT,
  },
];
