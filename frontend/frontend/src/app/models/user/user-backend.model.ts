export interface UserBackend {
  user_id: number;
  username: string;
  email: string;
  user_role: string;
  tenant_id?: string;
}
