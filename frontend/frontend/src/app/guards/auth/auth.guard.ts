import { inject } from '@angular/core/primitives/di';
import { CanActivateFn, Router } from '@angular/router';

import { AuthSessionService } from '../../services/auth/auth-session.service';

export const authGuard: CanActivateFn = () => {
  const authSessionService = inject(AuthSessionService);
  const router = inject(Router);

  if (authSessionService.isAuthenticated()) {
    return true;
  }

  router.navigate(['/login']);
  return false;
};
