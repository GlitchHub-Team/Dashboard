import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatDialogModule, MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
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
  ],
  templateUrl: './tenant-form.dialog.html',
  styleUrl: './tenant-form.dialog.css',
})
export class TenantFormDialog {
  private readonly tenantService = inject(TenantService);
  private readonly dialogRef = inject(MatDialogRef<TenantFormDialog>);
  protected readonly data = inject<TenantConfig | null>(MAT_DIALOG_DATA);

  protected readonly tenantForm = inject(FormBuilder).nonNullable.group({
    name: ['', [Validators.required]],
    canImpersonate: [false],
  });

  protected isSubmitting = false;
  protected generalError = '';

  constructor() {
    if (this.data) {
      this.tenantForm.patchValue(this.data);
    }
  }

  protected onSave(): void {
    if (!this.tenantForm.valid) {
      this.tenantForm.markAllAsTouched();
      return;
    }

    this.isSubmitting = true;
    this.generalError = '';

    const config: TenantConfig = {
      name: this.tenantForm.value.name!,
      canImpersonate: this.tenantForm.value.canImpersonate ?? false,
    };

    this.tenantService.addNewTenant(config).subscribe({
      next: () => {
        this.dialogRef.close(true);
      },
      error: (err: ApiError) => {
        this.generalError = err.message ?? 'Failed to save tenant';
        this.isSubmitting = false;
      },
    });
  }

  protected onCancel(): void {
    this.dialogRef.close(false);
  }
}
