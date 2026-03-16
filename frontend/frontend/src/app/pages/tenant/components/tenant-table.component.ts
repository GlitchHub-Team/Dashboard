import { Component, Input, Output, EventEmitter } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatTableModule } from '@angular/material/table';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatTooltipModule } from '@angular/material/tooltip';
import { Tenant } from './../../../models/tenant.model';

@Component({
  selector: 'app-tenant-table',
  standalone: true,
  imports: [CommonModule, MatTableModule, MatButtonModule, MatIconModule, MatTooltipModule],
  template: `
    <table mat-table [dataSource]="tenants" class="tenant-table">
      <ng-container matColumnDef="name">
        <th mat-header-cell *matHeaderCellDef>Name</th>
        <td mat-cell *matCellDef="let element">{{ element.name }}</td>
      </ng-container>

      <ng-container matColumnDef="actions">
        <th mat-header-cell *matHeaderCellDef class="actions-header">Actions</th>
        <td mat-cell *matCellDef="let element" class="actions-cell">
          <button
            mat-icon-button
            (click)="onDelete(element)"
            matTooltip="Delete"
            color="warn"
          >
            <mat-icon>delete</mat-icon>
          </button>
        </td>
      </ng-container>

      <tr mat-header-row *matHeaderRowDef="displayedColumns"></tr>
      <tr mat-row *matRowDef="let row; columns: displayedColumns;"></tr>
    </table>
  `,
  styles: [
    `
      .tenant-table {
        width: 100%;
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
  @Input() tenants: Tenant[] = [];
  @Input() loading = false;
  @Output() deleteRequested = new EventEmitter<Tenant>();

  displayedColumns: string[] = ['name', 'actions'];

  onDelete(tenant: Tenant): void {
    this.deleteRequested.emit(tenant);
  }
}