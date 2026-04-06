import { Component, DestroyRef, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatIconModule } from '@angular/material/icon';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatProgressSpinner } from '@angular/material/progress-spinner';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatCheckboxModule } from '@angular/material/checkbox';

import { ApiError } from '../../../../models/api-error.model';
import { TenantConfig } from '../../../../models/tenant/tenant-config.model';
import { TenantService } from '../../../../services/tenant/tenant.service';

@Component({
  selector: 'app-tenant-form-dialog',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatDialogModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatCheckboxModule,
    MatIconModule,
    MatProgressSpinner,
  ],
  templateUrl: './tenant-form.dialog.html',
  styleUrl: './tenant-form.dialog.css',
})
export class TenantFormDialog {
  private readonly tenantService = inject(TenantService);
  private readonly dialogRef = inject(MatDialogRef<TenantFormDialog>);
  private readonly destroyRef = inject(DestroyRef);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly tenantForm = this.formBuilder.nonNullable.group({
    name: ['', [Validators.required]],
    canImpersonate: [false, Validators.required],
  });

  protected readonly isSubmitting = signal(false);
  protected readonly generalError = signal('');

  protected onSave(): void {
    if (!this.tenantForm.valid) {
      this.tenantForm.markAllAsTouched();
      return;
    }

    if (this.isSubmitting()) return;

    this.isSubmitting.set(true);
    this.generalError.set('');

    const config: TenantConfig = {
      name: this.tenantForm.value.name!,
      canImpersonate: this.tenantForm.value.canImpersonate ?? false,
    };

    this.tenantService
      .addNewTenant(config)
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: () => this.dialogRef.close(true),
        error: (err: ApiError) => {
          this.generalError.set(err.message ?? 'Failed to create tenant');
          this.isSubmitting.set(false);
        },
      });
  }

  protected onCancel(): void {
    this.dialogRef.close(false);
  }
}
