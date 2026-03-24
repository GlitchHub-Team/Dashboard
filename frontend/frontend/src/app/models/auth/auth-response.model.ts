import { User } from '../user/user.model';

export interface AuthResponse {
  user: User;
  token: string;
}
