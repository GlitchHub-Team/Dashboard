import { CommonModule } from '@angular/common';
import { Component, OnInit, inject } from '@angular/core';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { MatButtonModule } from '@angular/material/button';
import { PageEvent, MatPaginatorModule } from '@angular/material/paginator';
import { UserService } from '../../services/user/user.service';
import { UserFormDialogComponent } from './dialogs/user-form/user-form.dialog';
import { UserTableComponent } from './components/user-table/user-table.component';
import { ConfirmDeleteDialog } from '../gateway-sensor/dialogs/confirm-delete/confirm-delete.dialog';
import { User } from '../../models/user/user.model';
import { ActivatedRoute } from '@angular/router';
import { UserRole } from '../../models/user/user-role.enum';

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

  // Configurazione dinamica delle colonne: mostriamo il Tenant ID solo per i TENANT_ADMIN
  public get columnConfig() {
    const cols: { key: keyof User; label: string }[] = [
      { key: 'username' as keyof User, label: 'Username' },
      { key: 'email' as keyof User, label: 'Email' },
    ];
    if (this.context.role === UserRole.TENANT_ADMIN) {
      cols.push({ key: 'tenantId' as keyof User, label: 'Tenant ID' });
    }
    return cols;
  }

  ngOnInit(): void {
    this.activatedRoute.data.subscribe((data) => {
      this.context = data['userManagerContext'] || this.context;
      this.userService.retrieveUser(this.context.role, this.context.tenantId);
    });
  }

  onCreateUser(): void {
    this.dialog
      .open(UserFormDialogComponent, {
        width: '400px',
        data: { user: null, role: this.context.role },
      })
      .afterClosed()
      .subscribe((result: { username: string; email: string; tenantId?: string } | undefined) => {
        if (result) {
          const userConfig = {
            username: result.username,
            email: result.email,
            role: this.context.role,
          };

          let tenantIdToPass = this.context.tenantId;
          if (this.context.role === UserRole.TENANT_ADMIN && result.tenantId) {
            tenantIdToPass = result.tenantId.toLowerCase().replace(' ', '-0'); // Adattamento per i mock
          }

          this.userService.addNewUser(userConfig, tenantIdToPass).subscribe(() => {
            this.userService.retrieveUser(this.context.role, this.context.tenantId);
          });
        }
      });
  }

  onDeleteUser(user: User): void {
    this.dialog
      .open(ConfirmDeleteDialog, {
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
    this.userService.changePage(
      event.pageIndex,
      event.pageSize,
      this.context.role,
      this.context.tenantId,
    );
  }
}
