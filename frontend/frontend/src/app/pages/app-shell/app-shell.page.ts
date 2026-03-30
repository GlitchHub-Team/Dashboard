import { Component, computed, inject, signal } from '@angular/core';
import { MatDialog } from '@angular/material/dialog';
import { RouterOutlet } from '@angular/router';
import { toSignal } from '@angular/core/rxjs-interop';

import { UserSessionService } from '../../services/user-session/user-session.service';
import { AuthSessionService } from '../../services/auth/auth-session.service';
import { PermissionService } from '../../services/permission/permission.service';
import { ChangePasswordDialog } from './dialogs/change-password/change-password.dialog';
import { HeaderComponent } from './components/header/header.component';
import { SideBarComponent } from './components/side-bar/side-bar.component';
import { NAV_ITEMS } from '../../models/nav_items/nav-config.model';
import { UserService } from '../../services/user/user.service';
import { TenantService } from '../../services/tenant/tenant.service';
import { UserRole } from '../../models/user/user-role.enum';

@Component({
  selector: 'app-shell',
  imports: [RouterOutlet, SideBarComponent, HeaderComponent],
  templateUrl: './app-shell.page.html',
  styleUrl: './app-shell.page.css',
})
export class AppShellPage {
  private readonly userSession = inject(UserSessionService);
  private readonly authSessionService = inject(AuthSessionService);
  private readonly userService = inject(UserService);
  private readonly tenantService = inject(TenantService);
  private readonly permissionService = inject(PermissionService);
  private readonly dialog = inject(MatDialog);

  // Recupera singolarmente i campi dell'utente loggato (ID e ruolo sono sicuro di trovarli),
  // mentre il tenantId potrebbe essere null per i super admin
  private readonly currentUserId = this.userSession.currentUser()!.userId;
  private readonly currentTenantId = this.userSession.currentUser()?.tenantId;
  private readonly currentUserRole = this.userSession.currentUser()!.role;

  // Dai dati recuperati uso i services User e Tenant per recuperare i dettagli completi che passo
  // ai componenti figli quali header e side-bar per mostrare nomi e non ID
  protected readonly currentUser = toSignal(
    this.userService.getUser(
      this.currentUserId,
      this.currentUserRole,
      this.currentTenantId ?? undefined,
    ),
  );

  protected readonly currentTenant = this.currentTenantId
    ? toSignal(this.tenantService.getTenant(this.currentTenantId))
    : signal(null);

  protected readonly navItems = computed(() => {
    const isSuperAdmin = this.currentUserRole === UserRole.SUPER_ADMIN;
    return NAV_ITEMS.filter((item) => {
      if (!item.permission) {
        return true;
      }
      const permissions = Array.isArray(item.permission) ? item.permission : [item.permission];
      return this.permissionService.canAny(permissions);
    }).map((item) => (item.separator && !isSuperAdmin ? { ...item, separator: false } : item));
  });

  // Logout non richiede campi da passare
  protected onLogout(): void {
    this.authSessionService.logout();
  }

  // Cambia la password per utente loggato, quindi non richiede campi speciali.
  // Si può segnalare esito positivo direttamente nel dialog
  protected onChangePassword(): void {
    this.dialog.open(ChangePasswordDialog, {
      width: '800px',
      disableClose: true,
    });
  }
}
