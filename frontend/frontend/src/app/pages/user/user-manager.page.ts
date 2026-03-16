import { CommonModule } from '@angular/common';
import { Component, OnInit, inject } from '@angular/core';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { MatButtonModule } from '@angular/material/button';
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
}

@Component({
  selector: 'app-user-manager-page',
  standalone: true,
  imports: [CommonModule, MatButtonModule, MatDialogModule, UserTableComponent],
  templateUrl: './user-manager.page.html',
  styleUrl: './user-manager.page.css',
})
export class UserManagerPage implements OnInit {
  readonly userService = inject(UserService);
  readonly dialog = inject(MatDialog);
  private readonly activatedRoute = inject(ActivatedRoute);
  //userSession: UserSessionService;
  context: UserManagerContext = { title: 'User Management', role: UserRole.TENANT_ADMIN };

  users = this.userService.userList;
  loading = this.userService.loading;


  ngOnInit(): void {
    this.activatedRoute.data.subscribe(data => {
      this.context = data['userManagerContext'] || this.context;
      this.userService.retrieveUser(this.context.role);
    });
  }

  onCreateUser(): void {
    this.dialog.open(UserFormDialogComponent, {
      width: '400px',
      data: null,
    }).afterClosed().subscribe((result: User) => {
      if (result) {
        const userConfig = { email: result.email, role: result.role };
        this.userService.addNewUser(userConfig).subscribe(() => {
          this.userService.retrieveUser(this.context.role);
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
        this.userService.removeUser(user.email).subscribe(() => {
          this.userService.retrieveUser(this.context.role);
        });
      }
    });
  }
}
