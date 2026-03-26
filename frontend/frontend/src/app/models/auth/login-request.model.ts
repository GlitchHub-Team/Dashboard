export interface LoginRequest {
  email: string;
  password: string;
  userRole: string;
  tenantId?: string;
}
