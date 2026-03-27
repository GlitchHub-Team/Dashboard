export interface UserBackend {
  user_id: string;
  username: string;
  email: string;
  user_role: string;
  tenant_id?: string;
}
