import { Component, inject, OnInit, signal } from '@angular/core';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { MatButtonModule } from '@angular/material/button';
import { PageEvent } from '@angular/material/paginator';
import { MatIconModule } from '@angular/material/icon';

import { UserService } from '../../services/user/user.service';
import { UserSessionService } from '../../services/user-session/user-session.service';
import { UserFormDialogComponent } from './dialogs/user-form/user-form.dialog';
import { UserTableComponent } from './components/user-table/user-table.component';
import { ConfirmDeleteDialog } from '../gateway-sensor/dialogs/confirm-delete/confirm-delete.dialog';
import { User } from '../../models/user/user.model';
import { ActivatedRoute } from '@angular/router';
import { UserRole } from '../../models/user/user-role.enum';
import { UserConfig } from '../../models/user/user-config.model';

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
  private readonly userSession = inject(UserSessionService);
  private readonly dialog = inject(MatDialog);
  private readonly activatedRoute = inject(ActivatedRoute);

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

  public ngOnInit(): void {
    this.activatedRoute.data.subscribe((data) => {
      const routeContext = data['userManagerContext'];
      this.context.set({
        ...routeContext,
        tenantId: this.userSession.currentUser()?.tenantId,
      });
      this.refreshUsers();
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

          this.userService.addNewUser(userConfig, context.role, tenantIdToPass).subscribe(() => {
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

  private refreshUsers(): void {
    const context = this.context();
    this.userService.retrieveUser(context.role, context.tenantId);
  }
}
