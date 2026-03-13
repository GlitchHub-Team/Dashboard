import { Routes } from '@angular/router';
import { authGuard } from './guards/auth/auth.guard';
import { roleGuard } from './guards/role/role.guard';
import { Permission } from './models/permission.enum';

export const routes: Routes = [
  {
    path: 'login',
    loadComponent: () => import('./pages/login/login.page').then((m) => m.LoginPage),
  },
  {
    path: 'reset-password',
    loadComponent: () =>
      import('./pages/reset-password/reset-password.page').then((m) => m.ResetPasswordPage),
  },
  {
    path: 'dashboard',
    // Usare entrambe le guards, altrimenti si entra in un circolo vizioso (da canAny() manda a dashboard, ma
    // se non si è autenticati authGuard fallisce lo stesso). Se fallisce authGuard, non viene neanche valutato
    // roleGuard, quindi non c'è rischio di errori strani
    canActivate: [authGuard, roleGuard],
    loadComponent: () => import('./pages/app-shell/app-shell.page').then((m) => m.AppShellPage),
    data: {
      permissions: [Permission.DASHBOARD_ACCESS],
    },
  },
  {
    path: '',
    redirectTo: 'login',
    pathMatch: 'full',
  },
  {
    path: '**',
    redirectTo: 'login',
  },
];
