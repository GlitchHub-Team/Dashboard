import { Component, input, output } from '@angular/core';
import { MatProgressSpinner } from '@angular/material/progress-spinner';
import { MatTableModule } from '@angular/material/table';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatPaginatorModule, PageEvent } from '@angular/material/paginator';
import { User } from '../../../../models/user/user.model';

@Component({
  selector: 'app-user-table',
  standalone: true,
  imports: [
    MatProgressSpinner,
    MatTableModule,
    MatButtonModule,
    MatIconModule,
    MatTooltipModule,
    MatPaginatorModule,
  ],
  templateUrl: './user-table.component.html',
  styleUrl: './user-table.component.css',
})
export class UserTableComponent {
  public readonly users = input.required<User[]>();
  public readonly loading = input<boolean>();
  public readonly total = input<number>(0);
  public readonly pageIndex = input<number>(0);
  public readonly limit = input<number>(10);
  public readonly deleteRequested = output<User>();
  public readonly pageChange = output<PageEvent>();
  public readonly displayedColumns = input<string[]>([]);

  protected onDelete(user: User): void {
    this.deleteRequested.emit(user);
  }

  protected onPageChange(event: PageEvent): void {
    this.pageChange.emit(event);
  }
}
