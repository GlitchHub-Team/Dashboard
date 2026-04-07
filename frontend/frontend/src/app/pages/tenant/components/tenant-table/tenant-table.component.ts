import { Component, input, output, computed } from '@angular/core';
import { MatTableModule } from '@angular/material/table';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatProgressSpinner } from '@angular/material/progress-spinner';
import { Tenant } from '../../../../models/tenant/tenant.model';

@Component({
  selector: 'app-tenant-table',
  standalone: true,
  imports: [MatProgressSpinner, MatTableModule, MatButtonModule, MatIconModule, MatTooltipModule],
  templateUrl: './tenant-table.component.html',
  styleUrl: './tenant-table.component.css',
})
export class TenantTableComponent {
  public readonly tenants = input.required<Tenant[]>();
  public readonly loading = input<boolean>(false);
  public readonly total = input<number>(0);
  public readonly pageIndex = input<number>(0);
  public readonly limit = input<number>(10);

  public readonly deleteRequested = output<Tenant>();
  public readonly dashboardRequested = output<Tenant>();
  public readonly tenantUserManagementRequested = output<Tenant>();


  protected readonly displayedColumns = computed(() => {
    const cols: string[] = ['id', 'name'];
    cols.push('actions');
    return cols;
  });

  protected onDelete(tenant: Tenant): void {
    this.deleteRequested.emit(tenant);
  }

  protected onGoToDashboard(tenant: Tenant): void {
    this.dashboardRequested.emit(tenant);
  }

  protected onGoToTenantUserManagement(tenant: Tenant): void {
    this.tenantUserManagementRequested.emit(tenant);
  }
}
