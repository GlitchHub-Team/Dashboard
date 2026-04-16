import { Observable } from 'rxjs';

import { AuthResponse } from '../../models/auth/auth-response.model';
import { ConfirmAccountResponse } from '../../models/auth/confirm-account.model';
import { ForgotPasswordRequest } from '../../models/auth/forgot-password-request.model';
import { ForgotPasswordResponse } from '../../models/auth/forgot-password.model';
import { LoginRequest } from '../../models/auth/login-request.model';
import { PasswordChange } from '../../models/auth/password-change.model';

export abstract class AuthApiClientAdapter {
  abstract login(req: LoginRequest): Observable<AuthResponse>;
  abstract logout(): Observable<void>;
  abstract verifyForgotPasswordToken(token: string, tenantId?: string): Observable<void>;
  abstract forgotPasswordRequest(req: ForgotPasswordRequest): Observable<void>;
  abstract confirmPasswordReset(req: ForgotPasswordResponse): Observable<void>;
  abstract confirmPasswordChange(req: PasswordChange): Observable<void>;
  abstract verifyAccountToken(token: string, tenantId?: string): Observable<void>;
  abstract confirmAccountCreation(req: ConfirmAccountResponse): Observable<AuthResponse>;
}