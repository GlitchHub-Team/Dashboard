import { Component, DestroyRef, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatDialogModule, MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinner } from '@angular/material/progress-spinner';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatButtonModule } from '@angular/material/button';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';

import { UserRole } from '../../../../models/user/user-role.enum';
import { TenantService } from '../../../../services/tenant/tenant.service';
import { UserConfig } from '../../../../models/user/user-config.model';
import { UserService } from '../../../../services/user/user.service';
import { ApiError } from '../../../../models/api-error.model';
import { Tenant } from '../../../../models/tenant/tenant.model';

export interface UserFormDialogData {
  role: UserRole;
  tenantId?: string;
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
    MatIconModule,
    MatProgressSpinner,
  ],
  templateUrl: './user-form.dialog.html',
  styleUrl: './user-form.dialog.css',
})
export class UserFormDialogComponent {
  private readonly tenantService = inject(TenantService);
  private readonly userService = inject(UserService);
  private readonly formBuilder = inject(FormBuilder);
  private readonly dialogRef = inject(MatDialogRef<UserFormDialogComponent>);
  private readonly destroyRef = inject(DestroyRef);

  protected readonly data = inject<UserFormDialogData>(MAT_DIALOG_DATA);
  protected readonly tenantList = signal<Tenant[]>([]);
  protected readonly UserRole = UserRole;

  // Indica se sto creando Tenant Admin?
  private get isTenantAdminRole(): boolean {
    return this.data.role === UserRole.TENANT_ADMIN;
  }

  // tenantId già noto (TENANT_ADMIN che crea, o SUPER_ADMIN dopo aver selezionato il tenant)
  protected get isTenantIdLocked(): boolean {
    return this.isTenantAdminRole && !!this.data.tenantId;
  }

  protected readonly form = this.formBuilder.nonNullable.group({
    username: ['', [Validators.required]],
    email: ['', [Validators.required, Validators.email]],
    tenantId: ['', this.isTenantAdminRole ? [Validators.required] : []],
  });

  protected readonly isSubmitting = signal(false);
  protected readonly generalError = signal('');
  protected readonly lockedTenantName = signal<string | null>(null);

  constructor() {
    if (this.isTenantAdminRole) {
      if (this.data.tenantId) {
        this.form.controls.tenantId.setValue(this.data.tenantId);
        this.tenantService
          .getTenant(this.data.tenantId)
          .pipe(takeUntilDestroyed(this.destroyRef))
          .subscribe((tenant) => this.lockedTenantName.set(tenant.name));
      } else {
        this.tenantService.getAllTenants().subscribe({
          next: (tenants) => this.tenantList.set(tenants),
          error: (err: ApiError) => this.generalError.set(err.message ?? 'Failed to fetch tenants'),
        });
      }
    }
  }

  protected onSave(): void {
    if (!this.form.valid) {
      this.form.markAllAsTouched();
      return;
    }

    if (this.isSubmitting()) return;

    this.isSubmitting.set(true);
    this.generalError.set('');

    const config: UserConfig = {
      email: this.form.value.email!,
      username: this.form.value.username!,
    };

    this.userService
      .addNewUser(
        config,
        this.data.role,
        this.isTenantAdminRole ? this.form.value.tenantId : this.data.tenantId,
      )
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: () => this.dialogRef.close(true),
        error: (err: ApiError) => {
          this.generalError.set(err.message ?? 'Failed to create user');
          this.isSubmitting.set(false);
        },
      });
  }

  protected onCancel(): void {
    this.dialogRef.close();
  }

  protected dismissError(): void {
    this.generalError.set('');
  }
}
