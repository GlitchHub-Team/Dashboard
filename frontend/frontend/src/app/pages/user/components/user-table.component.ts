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
  templateUrl: './user-table.component.html',
  styleUrl: './user-table.component.css',
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
