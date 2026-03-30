export interface LoginRequest {
  email: string;
  password: string;
  //TODO: togliere userRole
  userRole: string;
  tenantId?: string;
}
