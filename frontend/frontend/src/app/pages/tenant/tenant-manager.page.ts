import { Component, computed, DestroyRef, inject, OnInit, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { CommonModule } from '@angular/common';
import { MatDialog } from '@angular/material/dialog';
import { MatButtonModule } from '@angular/material/button';
import { PageEvent } from '@angular/material/paginator';
import { Router } from '@angular/router';
import { MatIconModule } from '@angular/material/icon';
import { filter, switchMap } from 'rxjs';

import { TenantService } from '../../services/tenant/tenant.service';
import { TenantFormDialog } from './dialogs/tenant-form/tenant-form.dialog';
import { TenantTableComponent } from './components/tenant-table/tenant-table.component';
import { ConfirmDeleteDialog } from '../shared/dialogs/confirm-delete/confirm-delete.dialog';
import { Tenant } from '../../models/tenant/tenant.model';
import { MatSnackBar } from '@angular/material/snack-bar';

@Component({
  selector: 'app-tenant-manager-page',
  standalone: true,
  imports: [CommonModule, MatButtonModule, TenantTableComponent, MatIconModule],
  templateUrl: './tenant-manager.page.html',
  styleUrl: './tenant-manager.page.css',
})
export class TenantManagerPage implements OnInit {
  private readonly tenantService = inject(TenantService);
  private readonly dialog = inject(MatDialog);
  private readonly router = inject(Router);
  private readonly destroyRef = inject(DestroyRef);
  private readonly snackBar = inject(MatSnackBar);

  protected readonly tenants = this.tenantService.tenantList;
  protected readonly total = this.tenantService.total;
  protected readonly pageIndex = this.tenantService.pageIndex;
  protected readonly limit = this.tenantService.limit;
  protected readonly loading = this.tenantService.loading;
  protected readonly error = this.tenantService.error;

  private readonly _dismissedError = signal<string | null>(null);

  protected readonly visibleError = computed(() => {
    const err = this.error();
    return err === this._dismissedError() ? null : err;
  });

  protected dismissError(): void {
    this._dismissedError.set(this.error());
  }

  public ngOnInit(): void {
    this.tenantService.retrieveTenants();
  }

  protected onCreateTenant(): void {
    this.dialog
      .open(TenantFormDialog, {
        width: '500px',
        data: null,
      })
      .afterClosed()
      .pipe(
        filter((created) => !!created),
        takeUntilDestroyed(this.destroyRef),
      )
      .subscribe(() => {
        this.tenantService.retrieveTenants();
        this.snackBar.open('Tenant creato con successo', 'Chiudi', { duration: 3000 });
      });
  }

  protected onDeleteTenant(tenant: Tenant): void {
    this.dialog
      .open(ConfirmDeleteDialog, {
        width: '400px',
        data: {
          title: 'Delete Tenant',
          message: `Sei sicuro di voler eliminare il tenant "${tenant.name}"?`,
        },
      })
      .afterClosed()
      .pipe(
        filter((confirmed) => !!confirmed),
        switchMap(() => this.tenantService.removeTenant(tenant.id)),
        takeUntilDestroyed(this.destroyRef),
      )
      .subscribe(() => {
        this.snackBar.open('Tenant eliminato con successo', 'Chiudi', { duration: 3000 });
      });
  }

  protected onPageChange(event: PageEvent): void {
    this.tenantService.changePage(event.pageIndex, event.pageSize);
  }

  protected onGoToDashboard(tenant: Tenant): void {
    this.router.navigate(['/dashboard'], { queryParams: { tenantId: tenant.id } });
  }

  protected onGoToTenantUserManagement(tenant: Tenant): void {
    this.router.navigate(['/user-management/tenant-users'], {
      queryParams: { tenantId: tenant.id },
    });
  }
}
