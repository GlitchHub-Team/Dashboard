import { Component, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatDialogModule, MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { TenantService } from '../../../../services/tenant/tenant.service';
import { RawTenantConfig } from '../../../../models/tenant/raw-tenant-config.model';

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
  ],
  templateUrl: './tenant-form.dialog.html',
  styleUrl: './tenant-form.dialog.css',
  providers: [TenantService],
})
export class TenantFormDialog {
  private readonly fb = inject(FormBuilder);
  private readonly tenantService = inject(TenantService);
  private readonly dialogRef = inject(MatDialogRef<TenantFormDialog>);
  public data = inject<RawTenantConfig | null>(MAT_DIALOG_DATA);

  protected formBuilder: FormGroup;
  public loading = signal(false);
  protected generalError = signal<string | null>(null);
  protected serverErrors = signal<Record<string, string>>({});

  constructor() {
    this.formBuilder = this.fb.group({
      name: ['', [Validators.required]],
    });

    if (this.data) {
      this.formBuilder.patchValue(this.data);
    }

    // Resetta gli errori quando l'utente digita qualcosa
    this.formBuilder.valueChanges.subscribe(() => {
      this.serverErrors.set({});
      this.generalError.set(null);
    });
  }

  onSave(): void {
    if (this.formBuilder.invalid) return;

    this.loading.set(true);
    this.serverErrors.set({});
    this.generalError.set(null);

    const config: RawTenantConfig = this.formBuilder.value;

    this.tenantService.addNewTenant(config).subscribe({
      next: (tenant: unknown) => {
        this.loading.set(false);
        this.dialogRef.close(tenant);
      },
      error: (err: HttpErrorResponse) => {
        this.loading.set(false);

        if (err.error && err.error.fieldErrors) {
          this.serverErrors.set(err.error.fieldErrors);
          Object.keys(err.error.fieldErrors).forEach((field) => {
            this.formBuilder.get(field)?.setErrors({ serverError: true });
          });
        } else {
          this.generalError.set(err.message || 'Failed to save tenant');
        }
      },
    });
  }

  onCancel(): void {
    this.dialogRef.close();
  }
}
