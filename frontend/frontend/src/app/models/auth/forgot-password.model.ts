export interface ForgotPasswordResponse {
  // TODO: mettere tenant_id?: string;
  token: string;
  newPassword: string;
}
