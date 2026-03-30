import { Component, inject, OnInit, signal } from '@angular/core';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { MatButtonModule } from '@angular/material/button';
import { PageEvent } from '@angular/material/paginator';
import { MatIconModule } from '@angular/material/icon';
import { combineLatest } from 'rxjs';

import { UserService } from '../../services/user/user.service';
import { UserFormDialogComponent } from './dialogs/user-form/user-form.dialog';
import { UserTableComponent } from './components/user-table/user-table.component';
import { ConfirmDeleteDialog } from '../gateway-sensor/dialogs/confirm-delete/confirm-delete.dialog';
import { User } from '../../models/user/user.model';
import { ActivatedRoute, Router } from '@angular/router';
import { UserRole } from '../../models/user/user-role.enum';
import { UserConfig } from '../../models/user/user-config.model';
import { UserSessionService } from '../../services/user-session/user-session.service';

interface UserManagerContext {
  title: string;
  role: UserRole;
  tenantId?: string;
}

@Component({
  selector: 'app-user-manager-page',
  standalone: true,
  imports: [MatButtonModule, MatDialogModule, UserTableComponent, MatIconModule],
  templateUrl: './user-manager.page.html',
  styleUrl: './user-manager.page.css',
})
export class UserManagerPage implements OnInit {
  private readonly userService = inject(UserService);
  private readonly dialog = inject(MatDialog);
  private readonly activatedRoute = inject(ActivatedRoute);
  private readonly router = inject(Router);

  protected readonly context = signal<UserManagerContext>({
    title: 'User Management',
    role: UserRole.TENANT_ADMIN,
  });

  protected readonly users = this.userService.userList;
  protected readonly total = this.userService.total;
  protected readonly pageIndex = this.userService.pageIndex;
  protected readonly limit = this.userService.limit;
  protected readonly loading = this.userService.loading;
  protected readonly error = this.userService.error;
  protected readonly UserRole = UserRole;
  private readonly userSession = inject(UserSessionService);
  protected readonly currentRole = this.userSession.currentRole;
  protected readonly activeTenantId = signal<string | null>(null);

  public ngOnInit(): void {
    combineLatest([
      this.activatedRoute.data,
      this.activatedRoute.queryParams
    ]).subscribe(([data, params]) => {
      const baseContext = data['userManagerContext'] || this.context();
      const urlTenantId = params['tenantId'];
      const currentUserRole = this.currentRole();
      const currentUserTenant = this.userSession.currentTenant();

      const resolvedTenantId = currentUserRole === UserRole.SUPER_ADMIN 
        ? (urlTenantId || null) 
        : (currentUserTenant || null);

      this.activeTenantId.set(resolvedTenantId);
      this.context.set({ ...baseContext, tenantId: resolvedTenantId || undefined });

      if (resolvedTenantId || baseContext.role !== UserRole.TENANT_USER) {
        this.refreshUsers();
      }
    });
  }

  protected onCreateUser(): void {
    const context = this.context();

    this.dialog
      .open(UserFormDialogComponent, {
        width: '400px',
        data: { user: null, role: context.role },
      })
      .afterClosed()
      .subscribe((result: (UserConfig & { tenantId?: string }) | undefined) => {
        if (result) {
          const userConfig: UserConfig = {
            email: result.email,
            username: result.username,
          };

          const tenantIdToPass = result.tenantId || context.tenantId;

          this.userService.addNewUser(userConfig, tenantIdToPass, context.role).subscribe(() => {
            this.refreshUsers();
          });
        }
      });
  }

  protected onDeleteUser(user: User): void {
    this.dialog
      .open(ConfirmDeleteDialog, {
        width: '400px',
        data: {
          title: 'Elimina Utente',
          message: `Sei sicuro di voler eliminare "${user.email}"?`,
        },
      })
      .afterClosed()
      .subscribe((confirmed) => {
        if (confirmed) {
          this.userService.removeUser(user).subscribe(() => {
            this.refreshUsers();
          });
        }
      });
  }

  protected onPageChange(event: PageEvent): void {
    const context = this.context();

    this.userService.changePage(event.pageIndex, event.pageSize, context.role, context.tenantId);
  }

  protected onBackToTenants(): void {
    this.router.navigate(['/tenant-management']);
  }

  private refreshUsers(): void {
    const context = this.context();
    this.userService.retrieveUser(context.role, context.tenantId);
  }
}
