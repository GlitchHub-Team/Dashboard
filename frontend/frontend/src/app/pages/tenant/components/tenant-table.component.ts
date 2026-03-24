import { Component, input, output, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatTableModule } from '@angular/material/table';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatTooltipModule } from '@angular/material/tooltip';
import { Tenant } from './../../../models/tenant.model';

export interface ColumnConfig<T> {
  key: keyof T;
  label: string;
}

@Component({
  selector: 'app-tenant-table',
  standalone: true,
  imports: [CommonModule, MatTableModule, MatButtonModule, MatIconModule, MatTooltipModule],
  templateUrl: './tenant-table.component.html',
  styleUrl: './tenant-table.component.css',
})
export class TenantTableComponent {
  tenants = input.required<Tenant[]>();
  loading = input<boolean>(false);
  
  // Accetta la configurazione dall'esterno (con un valore di default sensato)
  columnConfig = input<ColumnConfig<Tenant>[]>([
    { key: 'name', label: 'Nome' }
  ]);

  deleteRequested = output<Tenant>();
  dashboardRequested = output<Tenant>();

  // Calcola automaticamente le colonne da mostrare unendo le chiavi dinamiche a "actions"
  displayedColumns = computed(() => [...this.columnConfig().map(c => c.key as string), 'actions']);

  onDelete(tenant: Tenant): void {
    this.deleteRequested.emit(tenant);
  }

  onGoToDashboard(tenant: Tenant): void {
    this.dashboardRequested.emit(tenant);
  }
}
