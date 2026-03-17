import { Component, inject, signal, WritableSignal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatDialogModule, MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { User } from '../../../models/user.model';

@Component({
  selector: 'app-user-form-dialog',
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
    <h2 mat-dialog-title>{{ data ? 'Modifica' : 'Aggiungi' }} Utente</h2>
    <mat-dialog-content>
      <form [formGroup]="form">
        <mat-form-field appearance="outline" class="w-100">
          <mat-label>Email</mat-label>
          <input matInput formControlName="email" type="email" required />
        </mat-form-field>

        @if (generalError()) {
          <div class="error-text">
            {{ generalError() }}
          </div>
        }
      </form>
    </mat-dialog-content>
    <mat-dialog-actions align="end">
      <button mat-button (click)="onCancel()">Annulla</button>
      <button
        mat-raised-button
        color="primary"
        (click)="onSave()"
        [disabled]="form.invalid"
      >
        Salva
      </button>
    </mat-dialog-actions>
  `,
  styles: [
    `
      .w-100 {
        width: 100%;
        margin-bottom: 0.5rem;
      }
      .error-text {
        color: red;
        margin-top: 0.5rem;
        font-size: 0.875rem;
      }
    `,
  ],
})
export class UserFormDialogComponent {
  protected form: FormGroup;
  protected generalError: WritableSignal<string> = signal<string>('');
  
  private serverError: Map<string, string> = new Map<string, string>();

  private formBuilder = inject(FormBuilder);
  private dialogRef = inject(MatDialogRef<UserFormDialogComponent>);
  public data = inject<User | null>(MAT_DIALOG_DATA);

  constructor() {
    this.form = this.formBuilder.group({
      id: [this.data?.id || ''],
      email: [this.data?.email || '', [Validators.required, Validators.email]],
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
