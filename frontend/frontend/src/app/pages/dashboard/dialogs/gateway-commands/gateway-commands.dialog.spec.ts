import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';
import { By } from '@angular/platform-browser';
import { of, throwError } from 'rxjs';
import { signal } from '@angular/core';

import { GatewayCommandsDialog } from './gateway-commands.dialog';
import { GatewayService } from '../../../../services/gateway/gateway.service';
import { TenantService } from '../../../../services/tenant/tenant.service';
import { Gateway } from '../../../../models/gateway/gateway.model';
import { Status } from '../../../../models/gateway-sensor-status.enum';
import { ApiError } from '../../../../models/api-error.model';

// Mappa tipo di comando agli args che il form ritorna
const COMMAND_CASES: [
  string,
  'commissionGateway' | 'decommissionGateway' | 'resetGateway' | 'rebootGateway',
  string[],
][] = [
  ['commission', 'commissionGateway', ['gw-1', 'tenant-1', 'commission-token']],
  ['decommission', 'decommissionGateway', ['gw-1']],
  ['reset', 'resetGateway', ['gw-1']],
  ['reboot', 'rebootGateway', ['gw-1']],
];

describe('GatewayCommandsDialog (Unit)', () => {
  let fixture: ComponentFixture<GatewayCommandsDialog>;
  let component: GatewayCommandsDialog;
  let dialogRefMock: { close: ReturnType<typeof vi.fn> };
  let gatewayServiceMock: {
    commissionGateway: ReturnType<typeof vi.fn>;
    decommissionGateway: ReturnType<typeof vi.fn>;
    resetGateway: ReturnType<typeof vi.fn>;
    rebootGateway: ReturnType<typeof vi.fn>;
  };

  const mockGateway: Gateway = {
    id: 'gw-1',
    tenantId: undefined,
    name: 'Main Lobby Gateway',
    status: Status.ACTIVE,
    interval: 60,
  };

  const sendBtn = () => fixture.debugElement.query(By.css('button[color="primary"]'));
  const cancelBtn = () =>
    fixture.debugElement
      .queryAll(By.css('button'))
      .find((btn) => btn.nativeElement.textContent.includes('Annulla'))!;
  const selectCommand = (value: string) => {
    component['commandForm'].controls.command.setValue(value);

    if (value === 'commission') {
      component['commandForm'].controls.tenantId.setValue('tenant-1');
      component['commandForm'].controls.token.setValue('commission-token');
    }

    fixture.detectChanges();
  };

  beforeEach(async () => {
    vi.resetAllMocks();
    dialogRefMock = { close: vi.fn() };
    gatewayServiceMock = {
      commissionGateway: vi.fn(),
      decommissionGateway: vi.fn(),
      resetGateway: vi.fn(),
      rebootGateway: vi.fn(),
    };

    const tenantServiceMock = {
      tenantList: signal([]),
      retrieveTenants: vi.fn(),
    };

    await TestBed.configureTestingModule({
      imports: [GatewayCommandsDialog],
      providers: [
        { provide: MatDialogRef, useValue: dialogRefMock },
        { provide: MAT_DIALOG_DATA, useValue: { gateway: mockGateway, mode: 'manage' } },
        { provide: GatewayService, useValue: gatewayServiceMock },
        { provide: TenantService, useValue: tenantServiceMock },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(GatewayCommandsDialog);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  describe('initial state', () => {
    it('should create and display title, gateway name, and disabled Send button', () => {
      expect(component).toBeTruthy();
      expect(
        fixture.debugElement.query(By.css('[mat-dialog-title]')).nativeElement.textContent,
      ).toContain('Comando Gateway');
      const input: HTMLInputElement = fixture.debugElement.query(
        By.css('input[disabled]'),
      ).nativeElement;
      expect(input.value).toBe('Main Lobby Gateway');
      expect(component['commandForm'].controls.command.value).toBe('');
      expect(sendBtn().nativeElement.disabled).toBe(true);
    });
  });

  describe('form validation', () => {
    it('should enable Send button once a command is selected', () => {
      selectCommand('commission');
      expect(sendBtn().nativeElement.disabled).toBe(false);
    });

    it('should show required error when command is touched without selection', () => {
      component['commandForm'].controls.command.markAsTouched();
      fixture.detectChanges();
      expect(fixture.debugElement.query(By.css('mat-error')).nativeElement.textContent).toContain(
        'Campo obbligatorio',
      );
    });

    it('should mark command as touched and not call any service when form is invalid', () => {
      component['onConfirm']();
      fixture.detectChanges();
      expect(component['commandForm'].controls.command.touched).toBe(true);
      for (const [, method] of COMMAND_CASES) {
        expect(gatewayServiceMock[method]).not.toHaveBeenCalled();
      }
    });
  });

  describe('cancel', () => {
    it('should close with false when Cancel is clicked', () => {
      cancelBtn().nativeElement.click();
      expect(dialogRefMock.close).toHaveBeenCalledWith(false);
    });
  });

  describe('command execution', () => {
    it.each(COMMAND_CASES)(
      '%s: should call %s with gateway id and close with true on success',
      (command, method, args) => {
        gatewayServiceMock[method].mockReturnValue(of(void 0));
        selectCommand(command);
        sendBtn().nativeElement.click();
        expect(gatewayServiceMock[method]).toHaveBeenCalledWith(...args);
        expect(dialogRefMock.close).toHaveBeenCalledWith(true);
      },
    );

    it.each(COMMAND_CASES)(
      '%s: should set generalError and keep dialog open on error',
      (command, method) => {
        gatewayServiceMock[method].mockReturnValue(
          throwError(() => ({ message: `${command} failed` }) as ApiError),
        );
        selectCommand(command);
        sendBtn().nativeElement.click();
        expect(component['generalError']()).toBe(`${command} failed`);
        expect(dialogRefMock.close).not.toHaveBeenCalled();
      },
    );

    it('should use fallback error message when API error has no message', () => {
      gatewayServiceMock.commissionGateway.mockReturnValue(
        throwError(() => ({ status: 500 }) as ApiError),
      );
      selectCommand('commission');
      sendBtn().nativeElement.click();
      expect(component['generalError']()).toBe('Failed to send command');
    });
  });
});
