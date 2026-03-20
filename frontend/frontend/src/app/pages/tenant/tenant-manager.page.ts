import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { MatButtonModule } from '@angular/material/button';
import { TenantService } from '../../services/tenant/tenant.service';
import { TenantFormDialog } from './dialogs/tenant-form.dialog';
import { TenantTableComponent } from './components/tenant-table.component';
import { ConfirmDeleteDialog } from './dialogs/confirm-delete.dialog';
import { Tenant } from '../../models/tenant.model';

@Component({
  selector: 'app-tenant-manager-page',
  standalone: true,
  imports: [CommonModule, MatButtonModule, MatDialogModule, TenantTableComponent],
  templateUrl: './tenant-manager.page.html',
  styleUrl: './tenant-manager.page.css',
})
export class TenantManagerPage implements OnInit {
  private readonly tenantService = inject(TenantService);
  private readonly dialog = inject(MatDialog);

  tenants = this.tenantService.tenantList;
  loading = this.tenantService.loading;

  // non appena l'utente naviga su questa pagina, l'applicazione richiede automaticamente i dati dei tenant per poterli poi mostrare a schermo
  ngOnInit(): void {
    this.tenantService.retrieveTenant();
  }

  onCreateTenant(): void {
    this.dialog.open(TenantFormDialog, {
      width: '500px',
      data: null,
    }).afterClosed().subscribe(() => {
      this.tenantService.retrieveTenant();
    });
  }

  onDeleteTenant(tenant: Tenant): void {
    this.dialog.open(ConfirmDeleteDialog, {
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
}