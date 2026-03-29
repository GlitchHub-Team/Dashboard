import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatDialog } from '@angular/material/dialog';
import { MatButtonModule } from '@angular/material/button';
import { MatPaginatorModule, PageEvent } from '@angular/material/paginator';
import { Router } from '@angular/router';
import { MatIconModule } from '@angular/material/icon';

import { TenantService } from '../../services/tenant/tenant.service';
import { TenantFormDialog } from './dialogs/tenant-form/tenant-form.dialog';
import { TenantTableComponent } from './components/tenant-table/tenant-table.component';
import { ConfirmDeleteDialog } from '../gateway-sensor/dialogs/confirm-delete/confirm-delete.dialog';
import { Tenant } from '../../models/tenant/tenant.model';

@Component({
  selector: 'app-tenant-manager-page',
  standalone: true,
  imports: [CommonModule, MatButtonModule, TenantTableComponent, MatPaginatorModule, MatIconModule],
  templateUrl: './tenant-manager.page.html',
  styleUrl: './tenant-manager.page.css',
})
export class TenantManagerPage implements OnInit {
  private readonly tenantService = inject(TenantService);
  private readonly dialog = inject(MatDialog);
  private readonly router = inject(Router);

  protected readonly tenants = this.tenantService.tenantList;
  protected readonly total = this.tenantService.total;
  protected readonly pageIndex = this.tenantService.pageIndex;
  protected readonly limit = this.tenantService.limit;
  protected readonly loading = this.tenantService.loading;
  protected readonly error = this.tenantService.error;

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
      .subscribe((created) => {
        if (created) {
          this.tenantService.retrieveTenants();
        }
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
      .subscribe((confirmed) => {
        if (confirmed) {
          this.tenantService.removeTenant(tenant.id).subscribe();
        }
      });
  }

  protected onPageChange(event: PageEvent): void {
    this.tenantService.changePage(event.pageIndex, event.pageSize);
  }

  protected onGoToDashboard(tenant: Tenant): void {
    this.router.navigate(['/dashboard'], { queryParams: { tenantId: tenant.id } });
  }
}
