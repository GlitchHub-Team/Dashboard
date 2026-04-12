import { Permission } from '../permission.enum';

export interface NavItem {
  label: string;
  route: string;
  icon: string;
  permission?: Permission | Permission[];
  separator?: boolean;
  sectionTitle?: string;
  tenantSectionTitle?: string;
}
