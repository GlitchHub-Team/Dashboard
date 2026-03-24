import { Component, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
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
    CommonModule,
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
  private readonly fb = inject(FormBuilder);
  private readonly dialogRef = inject(MatDialogRef<UserFormDialogComponent>);
  public data = inject<UserFormDialogData>(MAT_DIALOG_DATA);
  private readonly tenantService = inject(TenantService);

  protected form: FormGroup;
  protected generalError = signal<string | null>(null);
  protected serverErrors = signal<Record<string, string>>({});
  protected tenantList = this.tenantService.tenantList;
  protected UserRole = UserRole;

  constructor() {
    this.form = this.fb.group({
      id: [this.data.user?.id || ''],
      username: [this.data.user?.username || '', [Validators.required]],
      email: [this.data.user?.email || '', [Validators.required, Validators.email]],
      tenantId: [
        this.data.user?.tenantId || '',
        this.data.role === UserRole.TENANT_ADMIN ? [Validators.required] : [],
      ],
    });

    if (this.data.role === UserRole.TENANT_ADMIN) {
      this.tenantService.retrieveTenant();
    }

    // Resetta gli errori quando l'utente digita qualcosa
    this.form.valueChanges.subscribe(() => {
      this.serverErrors.set({});
      this.generalError.set(null);
    });
  }

  protected onSave(): void {
    if (this.form.invalid) return;
    this.dialogRef.close(this.form.value);
  }

  protected onCancel(): void {
    this.dialogRef.close();
  }
}
