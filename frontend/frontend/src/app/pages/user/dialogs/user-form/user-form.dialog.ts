import { Component, inject } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatDialogModule, MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatButtonModule } from '@angular/material/button';
import { User } from '../../../../models/user/user.model';
import { UserRole } from '../../../../models/user/user-role.enum';
import { TenantService } from '../../../../services/tenant/tenant.service';

export interface UserFormDialogData {
  user: User | null;
  role: UserRole;
}

@Component({
  selector: 'app-user-form-dialog',
  standalone: true,
  imports: [
    ReactiveFormsModule,
    MatDialogModule,
    MatFormFieldModule,
    MatInputModule,
    MatSelectModule,
    MatButtonModule,
  ],
  templateUrl: './user-form.dialog.html',
  styleUrl: './user-form.dialog.css',
})
export class UserFormDialogComponent {
  private readonly formBuilder = inject(FormBuilder);
  private readonly dialogRef = inject(MatDialogRef<UserFormDialogComponent>);
  protected readonly data = inject<UserFormDialogData>(MAT_DIALOG_DATA);
  private readonly tenantService = inject(TenantService);

  protected readonly tenantList = this.tenantService.tenantList;
  protected readonly UserRole = UserRole;

  private get isTenantAdminRole(): boolean {
    return this.data.role === UserRole.TENANT_ADMIN;
  }

  protected readonly form = this.formBuilder.nonNullable.group({
    username: [this.data.user?.username || '', [Validators.required]],
    email: [this.data.user?.email || '', [Validators.required, Validators.email]],
    tenantId: [this.data.user?.tenantId || '', this.isTenantAdminRole ? [Validators.required] : []],
  });

  constructor() {
    if (this.isTenantAdminRole) {
      this.tenantService.retrieveTenants();
    }
  }

  protected onSave(): void {
    if (!this.form.valid) {
      this.form.markAllAsTouched();
      return;
    }

    this.dialogRef.close(this.form.value);
  }

  protected onCancel(): void {
    this.dialogRef.close();
  }
}
