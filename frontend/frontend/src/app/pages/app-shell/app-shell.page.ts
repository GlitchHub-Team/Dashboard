import { Component, computed, inject } from '@angular/core';
import { MatDialog } from '@angular/material/dialog';
import { Router, RouterOutlet } from '@angular/router';

import { UserSessionService } from '../../services/user-session/user-session.service';
import { AuthSessionService } from '../../services/auth/auth-session.service';
import { PermissionService } from '../../services/permission/permission.service';
import { ChangePasswordDialog } from './dialogs/change-password/change-password.dialog';
import { HeaderComponent } from './components/header/header.component';
import { SideBarComponent } from './components/side-bar/side-bar.component';

import { NAV_ITEMS } from '../../models/nav-config.model';

@Component({
  selector: 'app-shell',
  imports: [RouterOutlet, SideBarComponent, HeaderComponent],
  templateUrl: './app-shell.page.html',
  styleUrl: './app-shell.page.css',
})
export class AppShellPage {
  private readonly userSession = inject(UserSessionService);
  private readonly authSessionService = inject(AuthSessionService);
  private readonly permissionService = inject(PermissionService);
  private readonly dialog = inject(MatDialog);
  private readonly router = inject(Router);

  protected readonly currentUser = this.userSession.currentUser;
  protected readonly currentUserRole = this.userSession.currentRole;
  // TODO: per ora mostra solo il TenantID
  protected readonly currentTenant = this.userSession.currentTenant;
  protected readonly navItems = computed(() =>
    NAV_ITEMS.filter((item) => {
      if (!item.permission) {
        return true;
      }
      const permissions = Array.isArray(item.permission) ? item.permission : [item.permission];
      return this.permissionService.canAny(permissions);
    }),
  );

  protected onLogout(): void {
    this.authSessionService.logout();
    this.router.navigate(['/login']);
  }

  protected onChangePassword(): void {
    this.dialog.open(ChangePasswordDialog, {
      width: '800px',
      disableClose: true,
    });
  }
}
