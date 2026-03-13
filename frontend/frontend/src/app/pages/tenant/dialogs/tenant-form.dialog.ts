import { Component, Inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatDialogModule, MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { TenantService } from '../../../services/tenant/tenant.service';
import { RawTenantConfig } from '../../../models/raw-tenant-config.model';
import { Signal, signal } from '@angular/core';

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
  template: `
    <h2 mat-dialog-title>Add Tenant</h2>
    <mat-dialog-content>
      <form [formGroup]="formBuilder">
        <mat-form-field appearance="outline" class="w-100">
          <mat-label>Name</mat-label>
          <input matInput formControlName="name" required />
        </mat-form-field>
        <div *ngIf="generalError()" class="error-text">
          {{ generalError() }}
        </div>
      </form>
    </mat-dialog-content>
    <mat-dialog-actions align="end">
      <button mat-button (click)="onCancel()">Cancel</button>
      <button
        mat-raised-button
        color="primary"
        (click)="onSave()"
        [disabled]="formBuilder.invalid || loading()"
      >
        Save
      </button>
    </mat-dialog-actions>
  `,
  styles: [
    `
      .w-100 {
        width: 100%;
      }
      .error-text {
        color: red;
        margin-top: 0.5rem;
        font-size: 0.875rem;
      }
    `,
  ],  
  providers: [TenantService],
})
export class TenantFormDialog {
  formBuilder: FormGroup;
  loading: Signal<boolean>;
  generalError: Signal<string | null>;

  constructor(
    private fb: FormBuilder,
    private tenantService: TenantService,
    private dialogRef: MatDialogRef<TenantFormDialog>,
    @Inject(MAT_DIALOG_DATA) public data: RawTenantConfig | null
  ) {
    this.formBuilder = this.fb.group({
      name: ['', [Validators.required]],
    });

    if (data) {
      this.formBuilder.patchValue(data);
    }

    this.loading = signal(false);
    this.generalError = signal(null);
  }

  onSave(): void {
    if (this.formBuilder.invalid) return;

    this.loading = signal(true);
    const config: RawTenantConfig = this.formBuilder.value;

    this.tenantService.addNewTenant(config).subscribe({
      next: (tenant: any) => {
        this.loading = signal(false);
        this.dialogRef.close(tenant);
      },
      error: (err: any) => {
        this.loading = signal(false);
        this.generalError = signal(err.message || 'Failed to save tenant');
      },
    });
  }

  onCancel(): void {
    this.dialogRef.close();
  }
}