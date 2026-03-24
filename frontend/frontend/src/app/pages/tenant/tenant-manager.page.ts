import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatDialog } from '@angular/material/dialog';
import { MatButtonModule } from '@angular/material/button';
import { PageEvent, MatPaginatorModule } from '@angular/material/paginator';
import { TenantService } from '../../services/tenant/tenant.service';
import { TenantFormDialog } from './dialogs/tenant-form.dialog';
import { TenantTableComponent } from './components/tenant-table.component';
import { ConfirmDeleteDialog } from './dialogs/confirm-delete.dialog';
import { Tenant } from '../../models/tenant.model';
import { Router } from '@angular/router';

@Component({
  selector: 'app-tenant-manager-page',
  standalone: true,
  imports: [CommonModule, MatButtonModule, TenantTableComponent, MatPaginatorModule],
  templateUrl: './tenant-manager.page.html',
  styleUrl: './tenant-manager.page.css',
})
export class TenantManagerPage implements OnInit {
  private readonly tenantService = inject(TenantService);
  private readonly dialog = inject(MatDialog);
  private readonly router = inject(Router);

  tenants = this.tenantService.tenantList;
  total = this.tenantService.total;
  pageIndex = this.tenantService.pageIndex;
  limit = this.tenantService.limit;
  loading = this.tenantService.loading;

  // non appena l'utente naviga su questa pagina, l'applicazione richiede automaticamente i dati dei tenant per poterli poi mostrare a schermo
  ngOnInit(): void {
    this.tenantService.retrieveTenant();
  }

  onCreateTenant(): void {
    this.dialog
      .open(TenantFormDialog, {
        width: '500px',
        data: null,
      })
      .afterClosed()
      .subscribe(() => {
        this.tenantService.retrieveTenant();
      });
  }

  onDeleteTenant(tenant: Tenant): void {
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
          this.tenantService.removeTenant(tenant.name).subscribe();
        }
      });
  }

  onPageChange(event: PageEvent): void {
    this.tenantService.changePage(event.pageIndex, event.pageSize);
  }

  onGoToDashboard(tenant: Tenant): void {
    // Adattamento per i mock: trasforma "Tenant 1" in "tenant-01"
    const mockTenantId = tenant.name.toLowerCase().replace(' ', '-0');
    this.router.navigate(['/dashboard'], { queryParams: { tenantId: mockTenantId } });
  }
}
