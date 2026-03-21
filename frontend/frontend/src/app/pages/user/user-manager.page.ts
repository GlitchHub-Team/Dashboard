import { CommonModule } from '@angular/common';
import { Component, OnInit, inject } from '@angular/core';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { MatButtonModule } from '@angular/material/button';
import { PageEvent, MatPaginatorModule } from '@angular/material/paginator';
import { UserService } from '../../services/user/user.service';
import { UserFormDialogComponent } from './dialogs/user-form.dialog';
import { UserTableComponent } from './components/user-table.component';
import { ConfirmDeleteDialog } from '../tenant/dialogs/confirm-delete.dialog';
import { User } from '../../models/user.model';
import { ActivatedRoute } from '@angular/router';
import { UserRole } from '../../models/user-role.enum';

interface UserManagerContext {
  title: string;
  role: UserRole;
  tenantId?: string;
}

@Component({
  selector: 'app-user-manager-page',
  standalone: true,
  imports: [CommonModule, MatButtonModule, MatDialogModule, UserTableComponent, MatPaginatorModule],
  templateUrl: './user-manager.page.html',
  styleUrl: './user-manager.page.css',
})
export class UserManagerPage implements OnInit {
  private readonly userService = inject(UserService);
  private readonly dialog = inject(MatDialog);
  private readonly activatedRoute = inject(ActivatedRoute);

  // private readonly userSession = inject(UserSessionService); // Utile per gestire permessi o recuperare il Tenant ID corrente
  public context: UserManagerContext = { title: 'User Management', role: UserRole.TENANT_ADMIN };

  public users = this.userService.userList;
  public total = this.userService.total;
  public pageIndex = this.userService.pageIndex;
  public limit = this.userService.limit;
  public loading = this.userService.loading;


  ngOnInit(): void {
    this.activatedRoute.data.subscribe(data => {
      this.context = data['userManagerContext'] || this.context;
      this.userService.retrieveUser(this.context.role, this.context.tenantId);
    });
  }

  onCreateUser(): void {
    this.dialog.open(UserFormDialogComponent, {
      width: '400px',
      data: null,
    }).afterClosed().subscribe((result: User) => {
      if (result) {
        const userConfig = { email: result.email, role: this.context.role };
        this.userService.addNewUser(userConfig, this.context.tenantId).subscribe(() => {
          this.userService.retrieveUser(this.context.role, this.context.tenantId);
        });
      }
    });
  }

  onDeleteUser(user: User): void {
    this.dialog.open(ConfirmDeleteDialog, {
        width: '400px',
        data: {
          title: 'Elimina Utente',
          message: `Sei sicuro di voler eliminare "${user.email}"?`,
        },
      })
    .afterClosed()
    .subscribe((confirmed) => {
      if (confirmed) {
        this.userService.removeUser(user).subscribe(() => {
          this.userService.retrieveUser(this.context.role, this.context.tenantId);
        });
      }
    });
  }

  onPageChange(event: PageEvent): void {
    this.userService.changePage(event.pageIndex, event.pageSize, this.context.role, this.context.tenantId);
  }
}
