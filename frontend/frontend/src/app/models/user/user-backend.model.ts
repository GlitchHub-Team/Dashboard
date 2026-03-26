export interface UserBackend {
  id: string;
  username: string;
  email: string;
  role: string;
  tenant_id?: string;
}
