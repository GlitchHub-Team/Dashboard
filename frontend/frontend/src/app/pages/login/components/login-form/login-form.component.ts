import { Component, DestroyRef, inject, input, output } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { FormBuilder, Validators, ReactiveFormsModule } from '@angular/forms';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatSelectModule } from '@angular/material/select';

import { LoginRequest } from '../../../../models/auth/login-request.model';
import { TenantService } from '../../../../services/tenant/tenant.service';
import { UserRole } from '../../../../models/user/user-role.enum';
import { userRoleMapper } from '../../../../utils/user-role.utils';

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
  private readonly destroyRef = inject(DestroyRef);

  public loading = input(false);
  public generalError = input<string | null>(null);

  public submitLogin = output<LoginRequest>();
  public forgotPassword = output<void>();
  public dismissError = output<void>();

  protected readonly displayedTenants = this.tenantService.tenantList;

  protected readonly roles = [
    { value: UserRole.SUPER_ADMIN, label: 'Super Admin' },
    { value: UserRole.TENANT_ADMIN, label: 'Tenant Admin' },
    { value: UserRole.TENANT_USER, label: 'Tenant User' },
  ];

  protected loginForm = this.formBuilder.nonNullable.group({
    email: ['', [Validators.required, Validators.email]],
    password: ['', [Validators.required]],
    role: ['' as UserRole, Validators.required],
    tenantId: ['', Validators.required],
  });

  protected get showTenantDropdown(): boolean {
    const role = this.loginForm.controls.role.value;
    return !!role && role !== UserRole.SUPER_ADMIN;
  }

  // Gestisce dinamicamente la validazione del campo tenantId in base al ruolo selezionato.
  // Se il ruolo è SUPER_ADMIN, il campo tenantId non è richiesto e viene rimosso il validatore.
  // Per gli altri ruoli, il campo tenantId è obbligatorio e viene aggiunto il validatore.
  constructor() {
    this.tenantService.retrieveTenants();

    this.loginForm.controls.role.valueChanges
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe((role) => {
        const tenantIdCtrl = this.loginForm.controls.tenantId;
        if (role === UserRole.SUPER_ADMIN) {
          tenantIdCtrl.removeValidators(Validators.required);
          tenantIdCtrl.setValue('');
        } else {
          tenantIdCtrl.addValidators(Validators.required);
        }
        tenantIdCtrl.updateValueAndValidity();
      });
  }

  protected onSubmit(): void {
    if (!this.loginForm.valid) {
      this.loginForm.markAllAsTouched();
      return;
    }

    // Formatta la loginRequest, mappa il ruolo selezionato da ENUM a string e include
    // tenantId solo se necessario (non per SUPER_ADMIN)
    const loginRequest: LoginRequest = {
      email: this.loginForm.get('email')!.value,
      password: this.loginForm.get('password')!.value,
      userRole: userRoleMapper.toBackend(this.loginForm.get('role')!.value),
      tenantId: this.showTenantDropdown
        ? this.loginForm.get('tenantId')?.value || undefined
        : undefined,
    };

    this.submitLogin.emit(loginRequest);
  }
}
