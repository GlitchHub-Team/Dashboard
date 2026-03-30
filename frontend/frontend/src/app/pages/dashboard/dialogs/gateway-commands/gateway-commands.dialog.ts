import { Component, inject, OnInit } from '@angular/core';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatButtonModule } from '@angular/material/button';
import { MatIcon } from '@angular/material/icon';

import { Gateway } from '../../../../models/gateway/gateway.model';
import { GatewayService } from '../../../../services/gateway/gateway.service';
import { ApiError } from '../../../../models/api-error.model';
import { ActionMode } from '../../../../models/action-mode.model';
import { TenantService } from '../../../../services/tenant/tenant.service';

@Component({
  selector: 'app-gateway-commands',
  imports: [
    MatDialogModule,
    ReactiveFormsModule,
    MatFormFieldModule,
    MatInputModule,
    MatSelectModule,
    MatButtonModule,
    MatIcon,
  ],
  templateUrl: './gateway-commands.dialog.html',
  styleUrl: './gateway-commands.dialog.css',
})
export class GatewayCommandsDialog implements OnInit {
  private readonly dialogRef = inject(MatDialogRef<GatewayCommandsDialog>);
  private readonly formBuilder = inject(FormBuilder);
  private readonly gatewayService = inject(GatewayService);
  private readonly tenantService = inject(TenantService);
  protected readonly data = inject<{ gateway: Gateway; mode: ActionMode }>(MAT_DIALOG_DATA);

  protected readonly displayedTenants = this.tenantService.tenantList;

  private get mode(): ActionMode {
    return this.data.mode;
  }

  private get gateway(): Gateway {
    return this.data.gateway;
  }

  private get isDashboardMode(): boolean {
    return this.mode === 'dashboard';
  }

  private get isManageMode(): boolean {
    return this.mode === 'manage';
  }

  protected generalError = '';

  protected get showCommissionFields(): boolean {
    return !this.gateway.tenantId && this.commandForm.controls.command.value === 'commission';
  }

  // TENANT ADMIN NON FA NIENTE
  // Se gateway non commissionato, allora posso solo fare commission/reboot
  protected get commands(): { value: string; label: string }[] {
    if (this.isDashboardMode) {
      return [
        { value: 'restart', label: 'Restart' },
        { value: 'reboot', label: 'Reboot' },
      ];
    }

    if (this.gateway.tenantId) {
      return [
        { value: 'decommission', label: 'Decommission' },
        { value: 'restart', label: 'Restart' },
        { value: 'reboot', label: 'Reboot' },
      ];
    }

    return [
      { value: 'commission', label: 'Commission' },
      { value: 'decommission', label: 'Decommission' },
      { value: 'restart', label: 'Restart' },
      { value: 'reboot', label: 'Reboot' },
    ];
  }

  protected readonly commandForm = this.formBuilder.nonNullable.group({
    command: ['', Validators.required],
    tenantId: [''],
    token: [''],
  });

  ngOnInit(): void {
    this.tenantService.retrieveTenants();
    this.commandForm.controls.command.valueChanges.subscribe((command) => {
      const tenantIdCtrl = this.commandForm.controls.tenantId;
      const tokenCtrl = this.commandForm.controls.token;
      if (!this.gateway.tenantId && command === 'commission') {
        tenantIdCtrl.addValidators(Validators.required);
        tokenCtrl.addValidators(Validators.required);
      } else {
        tenantIdCtrl.removeValidators(Validators.required);
        tenantIdCtrl.setValue('');
        tokenCtrl.removeValidators(Validators.required);
        tokenCtrl.setValue('');
      }
      tenantIdCtrl.updateValueAndValidity();
      tokenCtrl.updateValueAndValidity();
    });
  }

  protected onConfirm(): void {
    if (!this.commandForm.valid) {
      this.commandForm.markAllAsTouched();
      return;
    }

    const command = this.commandForm.get('command')?.value;
    if (!command) {
      return;
    }

    switch (command) {
      case 'commission':
        this.gatewayService
          .commissionGateway(
            this.data.gateway.id,
            this.commandForm.controls.tenantId.value,
            this.commandForm.controls.token.value,
          )
          .subscribe({
            next: () => {
              this.dialogRef.close(true);
            },
            error: (err: ApiError) => {
              this.generalError = err.message ?? 'Failed to send command';
              this.dialogRef.close(false);
            },
          });
        break;
      case 'decommission':
        this.gatewayService.decommissionGateway(this.data.gateway.id).subscribe({
          next: () => {
            this.dialogRef.close(true);
          },
          error: (err: ApiError) => {
            this.generalError = err.message ?? 'Failed to send command';
            this.dialogRef.close(false);
          },
        });
        break;
      case 'restart':
        this.gatewayService.resetGateway(this.data.gateway.id).subscribe({
          next: () => {
            this.dialogRef.close(true);
          },
          error: (err: ApiError) => {
            this.generalError = err.message ?? 'Failed to send command';
            this.dialogRef.close(false);
          },
        });
        break;
      case 'reboot':
        this.gatewayService.rebootGateway(this.data.gateway.id).subscribe({
          next: () => {
            this.dialogRef.close(true);
          },
          error: (err: ApiError) => {
            this.generalError = err.message ?? 'Failed to send command';
            this.dialogRef.close(false);
          },
        });
        break;
      default:
        this.generalError = 'Unknown command';
        this.dialogRef.close(false);
    }
  }

  protected onCancel(): void {
    this.dialogRef.close(false);
  }
}
