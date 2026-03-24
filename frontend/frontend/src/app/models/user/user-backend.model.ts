export interface UserBackend {
  id: string;
  username: string;
  email: string;
  role: string;
  tenantId?: string;
}
