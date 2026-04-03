export interface ForgotPasswordResponse {
  tenantId?: string;
  token: string;
  newPassword: string;
}
