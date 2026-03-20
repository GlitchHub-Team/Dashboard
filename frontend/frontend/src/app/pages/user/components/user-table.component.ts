import { Component, input, output, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatTableModule } from '@angular/material/table';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatTooltipModule } from '@angular/material/tooltip';
import { User } from '../../../models/user.model';

export interface ColumnConfig<T> {
  key: keyof T;
  label: string;
}

@Component({
  selector: 'app-user-table',
  standalone: true,
  imports: [CommonModule, MatTableModule, MatButtonModule, MatIconModule, MatTooltipModule],
  template: `
    <table mat-table [dataSource]="users()" class="user-table">
      
      @for (col of columnConfig(); track col.key) {
        <ng-container [matColumnDef]="col.key">
          <th mat-header-cell *matHeaderCellDef>{{ col.label }}</th>
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
      .user-table {
        width: 100%;
      }
    
      ::ng-deep tr.mat-header-row,
      ::ng-deep tr.mat-mdc-header-row {
        background-color: #d0d0d0 !important;
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
export class UserTableComponent {
  users = input.required<User[]>();
  loading = input<boolean>(false);
  deleteRequested = output<User>();

  // Accetta la configurazione dall'esterno (con i valori di default sensati per User)
  columnConfig = input<ColumnConfig<User>[]>([
    { key: 'id', label: 'Id' },
    { key: 'email', label: 'Email' },
    { key: 'role', label: 'Ruolo' }
  ]);

  displayedColumns = computed(() => [...this.columnConfig().map(c => c.key as string), 'actions']);

  onDelete(user: User): void {
    this.deleteRequested.emit(user);
  }
}
