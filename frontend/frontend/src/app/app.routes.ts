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
    // Usare entrambe le guards, altrimenti si entra in un circolo vizioso (da canAny() manda a dashboard, ma
    // se non si è autenticati authGuard fallisce lo stesso). Se fallisce authGuard, non viene neanche valutato
    // roleGuard, quindi non c'è rischio di errori strani
    path: '',
    canActivate: [authGuard],
    loadComponent: () => import('./pages/app-shell/app-shell.page').then((m) => m.AppShellPage),
    children: [
      {
        path: 'dashboard',
        canActivate: [roleGuard],
        data: { permissions: [Permission.DASHBOARD_ACCESS] },
        loadComponent: () =>
          import('./pages/dashboard/dashboard.page').then((m) => m.DashboardPage),
      },
      {
        path: '',
        redirectTo: 'dashboard',
        pathMatch: 'full',
      },
    ],
  },
  {
    path: '**',
    redirectTo: 'login',
  },
];
