import { Component, computed, DestroyRef, inject, OnInit, signal } from '@angular/core';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { MatButtonModule } from '@angular/material/button';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatSelectModule } from '@angular/material/select';
import { PageEvent } from '@angular/material/paginator';
import { MatIconModule } from '@angular/material/icon';
import { combineLatest } from 'rxjs';
import { filter, switchMap } from 'rxjs';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { MatSnackBar } from '@angular/material/snack-bar';

import { UserService } from '../../services/user/user.service';
import { UserSessionService } from '../../services/user-session/user-session.service';
import { UserFormDialogComponent } from './dialogs/user-form/user-form.dialog';
import { UserTableComponent } from './components/user-table/user-table.component';
import { ConfirmDeleteDialog } from '../shared/dialogs/confirm-delete/confirm-delete.dialog';
import { User } from '../../models/user/user.model';
import { ActivatedRoute, Router } from '@angular/router';
import { UserRole } from '../../models/user/user-role.enum';
import { TenantService } from '../../services/tenant/tenant.service';

interface UserManagerContext {
  title: string;
  role: UserRole;
  tenantId?: string;
}

@Component({
  selector: 'app-user-manager-page',
  standalone: true,
  imports: [MatButtonModule, MatDialogModule, UserTableComponent, MatIconModule, MatFormFieldModule, MatSelectModule],
  templateUrl: './user-manager.page.html',
  styleUrl: './user-manager.page.css',
})
export class UserManagerPage implements OnInit {
  private readonly userService = inject(UserService);
  private readonly tenantService = inject(TenantService);
  private readonly userSession = inject(UserSessionService);
  private readonly dialog = inject(MatDialog);
  private readonly activatedRoute = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly destroyRef = inject(DestroyRef);
  private readonly snackBar = inject(MatSnackBar);

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
  protected readonly tenants = this.tenantService.tenantList;

  private readonly _dismissedError = signal<string | null>(null);

  protected readonly visibleError = computed(() => {
    const err = this.error();
    return err === this._dismissedError() ? null : err;
  });

  protected dismissError(): void {
    this._dismissedError.set(this.error());
  }

  protected readonly UserRole = UserRole;
  protected readonly currentRole = this.userSession.currentUser()!.role;
  protected readonly activeTenantId = signal<string | null>(null);

  public ngOnInit(): void {
    combineLatest([this.activatedRoute.data, this.activatedRoute.queryParams]).subscribe(
      ([data, params]) => {
        const baseContext = data['userManagerContext'] || this.context();
        const urlTenantId = params['tenantId'];
        const currentUserRole = this.currentRole;
        const currentUserTenant = this.userSession.currentUser()?.tenantId;

        if (urlTenantId) {
          this.tenantService.getTenant(urlTenantId).subscribe((tenant) => {
            if (!tenant.canImpersonate) {
              this.router.navigate(['/user-management/tenant-users']);
              return;
            }
          });
        }

        const resolvedTenantId =
          currentUserRole === UserRole.SUPER_ADMIN
            ? urlTenantId || null
            : currentUserTenant || null;

        this.activeTenantId.set(resolvedTenantId);
        this.context.set({ ...baseContext, tenantId: resolvedTenantId || undefined });

        if (currentUserRole === UserRole.SUPER_ADMIN && baseContext.role === UserRole.TENANT_ADMIN) {
          this.tenantService.retrieveTenants(true);
        }

        const needsTenantId =
          baseContext.role === UserRole.TENANT_USER || baseContext.role === UserRole.TENANT_ADMIN;
        if (!needsTenantId || resolvedTenantId) {
          this.refreshUsers();
        }
      },
    );
  }

  protected onCreateUser(): void {
    const context = this.context();

    this.dialog
      .open(UserFormDialogComponent, {
        width: '400px',
        data: { role: context.role, tenantId: context.tenantId },
      })
      .afterClosed()
      .pipe(
        filter((result) => !!result),
        takeUntilDestroyed(this.destroyRef),
      )
      .subscribe(() => {
        this.refreshUsers();
        this.snackBar.open('Utente creato con successo', 'Chiudi', { duration: 3000 });
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
      .pipe(
        filter((confirmed) => !!confirmed),
        switchMap(() => this.userService.removeUser(user)),
        takeUntilDestroyed(this.destroyRef),
      )
      .subscribe(() => {
        this.refreshUsers();
        this.snackBar.open('Utente eliminato con successo', 'Chiudi', { duration: 3000 });
      });
  }

  protected onPageChange(event: PageEvent): void {
    const context = this.context();

    this.userService.changePage(event.pageIndex, event.pageSize, context.role, context.tenantId);
  }

  protected onBackToTenants(): void {
    this.router.navigate(['/tenant-management']);
  }

  protected onTenantSelected(tenantId: string): void {
    this.activeTenantId.set(tenantId);
    this.context.update((ctx) => ({ ...ctx, tenantId }));
    this.refreshUsers();
  }

  protected onBackToDashboard(): void {
    if (this.activeTenantId()) {
      this.router.navigate(['/dashboard'], {
        queryParams: { tenantId: this.activeTenantId() },
      });
    }
  }

  private refreshUsers(): void {
    const context = this.context();
    this.userService.retrieveUsers(context.role, context.tenantId);
  }
}
