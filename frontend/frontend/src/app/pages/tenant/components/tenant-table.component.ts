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
  template: `
    <table mat-table [dataSource]="tenants()" class="tenant-table">
      
      <!-- Ciclo dinamico per le colonne dati -->
      @for (col of columnConfig(); track col.key) {
        <ng-container [matColumnDef]="col.key">
          <th mat-header-cell *matHeaderCellDef>{{ col.label }}</th>
          <!-- Leggiamo il valore della proprietà dinamicamente usando la chiave -->
          <td mat-cell *matCellDef="let element">{{ element[col.key] }}</td>
        </ng-container>
      }

      <ng-container matColumnDef="actions">
        <th mat-header-cell *matHeaderCellDef class="actions-header">Azioni</th>
        <td mat-cell *matCellDef="let element" class="actions-cell">
          <button
            mat-icon-button
            (click)="onDelete(element)"
            matTooltip="Elimina"
            color="warn"
            [disabled]="loading()"
          >
            <mat-icon>delete</mat-icon>
          </button>
        </td>
      </ng-container>

      <!-- Eseguiamo il binding alla funzione Signal -->
      <tr mat-header-row *matHeaderRowDef="displayedColumns()"></tr>
      <tr mat-row *matRowDef="let row; columns: displayedColumns();"></tr>
    </table>
  `,
  styles: [
    `
      .tenant-table {
        width: 100%;
      }
      
      ::ng-deep tr.mat-header-row,
      ::ng-deep tr.mat-mdc-header-row {
        background-color: #dedcdcff !important;
      }

      ::ng-deep th.mat-header-cell,
      ::ng-deep th.mat-mdc-header-cell {
        border-bottom: 2px solid #000000 !important;
        font-weight: bold !important;
      }

      .actions-header {
        text-align: right;
        padding-right: 1rem;
      }

      .actions-cell {
        text-align: right;
        padding-right: 1rem;
      }
    `,
  ],
})
export class TenantTableComponent {
  tenants = input.required<Tenant[]>();
  loading = input<boolean>(false);
  
  // Accetta la configurazione dall'esterno (con un valore di default sensato)
  columnConfig = input<ColumnConfig<Tenant>[]>([
    { key: 'name', label: 'Nome' }
  ]);

  deleteRequested = output<Tenant>();

  // Calcola automaticamente le colonne da mostrare unendo le chiavi dinamiche a "actions"
  displayedColumns = computed(() => [...this.columnConfig().map(c => c.key as string), 'actions']);

  onDelete(tenant: Tenant): void {
    this.deleteRequested.emit(tenant);
  }
}
