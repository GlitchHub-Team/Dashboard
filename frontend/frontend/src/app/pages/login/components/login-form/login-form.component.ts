import { Component, inject, input, output, signal } from '@angular/core';
import { FormBuilder, Validators, ReactiveFormsModule } from '@angular/forms';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatSelectModule } from '@angular/material/select';

import { LoginRequest } from '../../../../models/auth/login-request.model';
import { TenantService } from '../../../../services/tenant/tenant.service';
import { Tenant } from '../../../../models/tenant/tenant.model';
import { ApiError } from '../../../../models/api-error.model';

@Component({
  selector: 'app-login-form',
  standalone: true,
  imports: [
    ReactiveFormsModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatIconModule,
    MatProgressBarModule,
    MatSelectModule,
  ],
  templateUrl: './login-form.component.html',
  styleUrl: './login-form.component.css',
})
export class LoginFormComponent {
  private readonly formBuilder = inject(FormBuilder);
  private readonly tenantService = inject(TenantService);

  public loading = input(false);
  public generalError = input<string | null>(null);

  public submitLogin = output<LoginRequest>();
  public forgotPassword = output<void>();
  public dismissError = output<void>();

  protected readonly displayedTenants = signal<Tenant[]>([]);
  protected readonly tenantLoadingError = signal<string | null>(null);

  protected loginForm = this.formBuilder.nonNullable.group({
    email: ['', [Validators.required, Validators.email]],
    password: ['', [Validators.required]],
    tenantId: [''],
  });

  constructor() {
    this.tenantService.getAllTenants().subscribe({
      next: (tenants) => this.displayedTenants.set(tenants),
      error: (err: ApiError) => this.tenantLoadingError.set(err.message ?? 'Failed to fetch tenants'),
    });
  }

  protected onSubmit(): void {
    if (!this.loginForm.valid) {
      this.loginForm.markAllAsTouched();
      return;
    }

    const loginRequest: LoginRequest = {
      email: this.loginForm.get('email')!.value,
      password: this.loginForm.get('password')!.value,
      tenantId: this.loginForm.get('tenantId')?.value || undefined,
    };

    this.submitLogin.emit(loginRequest);
  }
}
