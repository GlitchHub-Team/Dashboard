import { CommonModule } from '@angular/common';
import { Component, inject, OnInit, signal } from '@angular/core';
import { FormBuilder, Validators, ReactiveFormsModule } from '@angular/forms';
import { MatDialogRef, MatDialogModule } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { Subject } from 'rxjs';

import { ApiError } from '../../../../models/api-error.model';

@Component({
  selector: 'app-forgot-password.dialog',
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatDialogModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatIconModule,
    MatProgressBarModule,
  ],
  templateUrl: './forgot-password.dialog.html',
  styleUrl: './forgot-password.dialog.css',
})
export class ForgotPasswordDialog implements OnInit {
  private dialogRef = inject(MatDialogRef<ForgotPasswordDialog>);
  private formBuilder = inject(FormBuilder);

  protected forgotPasswordForm = this.formBuilder.nonNullable.group({
    email: ['', [Validators.required, Validators.email]],
  });

  protected loading = signal<boolean>(false);
  protected generalError = signal<string>('');
  protected serverError = signal<Map<string, string>>(new Map());

  public save$ = new Subject<string>();

  ngOnInit(): void {
    this.setupAutoClear();
  }

  // Imposta gli errori ricevuti dal backend (se sono arrivati) sui campi del form
  public setServerError(error: ApiError): void {
    const fieldErrors = new Map<string, string>();

    if (error.errors?.length) {
      for (const fe of error.errors) {
        const control = this.forgotPasswordForm.get(fe.field);
        if (control) {
          control.setErrors({ serverError: true });
          control.markAsTouched();
          fieldErrors.set(fe.field, fe.message);
        } else {
          this.generalError.set(fe.message);
        }
      }
    }

    this.serverError.set(fieldErrors);
  }

  public setLoading(value: boolean): void {
    this.loading.set(value);
  }

  protected onSubmit(): void {
    if (!this.forgotPasswordForm.valid) {
      this.forgotPasswordForm.markAllAsTouched();
      return;
    }

    this.generalError.set('');
    this.serverError.set(new Map());
    this.loading.set(true);

    this.save$.next(this.forgotPasswordForm.getRawValue().email);
  }

  protected onCancel(): void {
    this.dialogRef.close();
  }

  protected dismissError(): void {
    this.generalError.set('');
  }

  // Qui possiamo incontrare errori dal backend e errori di validazione del form, quindi li gestiamo con la funzione apposita
  protected getFieldError(field: string, label: string): string {
    const control = this.forgotPasswordForm.get(field);
    if (!control?.errors) return '';

    if (control.hasError('serverError')) {
      return this.serverError().get(field) ?? '';
    }

    if (control.hasError('required')) {
      return `${label} is required.`;
    }

    if (control.hasError('email')) {
      return `Please enter a valid email address.`;
    }

    return 'Invalid value';
  }

  // Questa funzione serve a cancellare gli errori di validazione e quelli provenienti dal backend quando l'utente modifica i campi,
  // così da non confondere l'utente con errori che magari ha già corretto
  private setupAutoClear(): void {
    for (const key of Object.keys(this.forgotPasswordForm.controls)) {
      this.forgotPasswordForm.get(key)!.valueChanges.subscribe(() => {
        const current = new Map(this.serverError());
        if (current.has(key)) {
          current.delete(key);
          this.serverError.set(current);
        }
        if (this.generalError()) {
          this.generalError.set('');
        }
      });
    }
  }
}
