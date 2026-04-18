import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';
import { By } from '@angular/platform-browser';
import { of, throwError } from 'rxjs';

import { GatewayCommandsDialog } from './gateway-commands.dialog';
import { GatewayService } from '../../../../services/gateway/gateway.service';
import { TenantService } from '../../../../services/tenant/tenant.service';
import { Gateway } from '../../../../models/gateway/gateway.model';
import { GatewayStatus } from '../../../../models/gateway-status.enum';
import { ApiError } from '../../../../models/api-error.model';
import { Tenant } from '../../../../models/tenant/tenant.model';

// Mappa tipo di comando agli args che il form ritorna
const COMMAND_CASES: [
  string,
  'commissionGateway' | 'decommissionGateway' | 'resetGateway' | 'rebootGateway' | 'interruptGateway' | 'resumeGateway',
  string[],
][] = [
  ['commission', 'commissionGateway', ['gw-1', 'tenant-1', 'commission-token']],
  ['decommission', 'decommissionGateway', ['gw-1']],
  ['reset', 'resetGateway', ['gw-1']],
  ['reboot', 'rebootGateway', ['gw-1']],
  ['interrupt', 'interruptGateway', ['gw-1']],
  ['resume', 'resumeGateway', ['gw-1']],
];

describe('GatewayCommandsDialog (Unit)', () => {
  let fixture: ComponentFixture<GatewayCommandsDialog>;
  let component: GatewayCommandsDialog;
  let dialogRefMock: { close: ReturnType<typeof vi.fn> };
  let tenantServiceMock: { getAllTenants: ReturnType<typeof vi.fn> };
  let gatewayServiceMock: {
    commissionGateway: ReturnType<typeof vi.fn>;
    decommissionGateway: ReturnType<typeof vi.fn>;
    resetGateway: ReturnType<typeof vi.fn>;
    rebootGateway: ReturnType<typeof vi.fn>;
    interruptGateway: ReturnType<typeof vi.fn>;
    resumeGateway: ReturnType<typeof vi.fn>;
  };

  const mockTenants: Tenant[] = [
    { id: 'tenant-01', name: 'Tenant 1', canImpersonate: false },
    { id: 'tenant-02', name: 'Tenant 2', canImpersonate: true },
  ];

  const mockGateway: Gateway = {
    id: 'gw-1',
    tenantId: undefined,
    name: 'Main Lobby Gateway',
    status: GatewayStatus.ACTIVE,
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
    tenantServiceMock = { getAllTenants: vi.fn().mockReturnValue(of(mockTenants)) };
    gatewayServiceMock = {
      commissionGateway: vi.fn(),
      decommissionGateway: vi.fn(),
      resetGateway: vi.fn(),
      rebootGateway: vi.fn(),
      interruptGateway: vi.fn(),
      resumeGateway: vi.fn(),
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

  describe('ngOnInit', () => {
    it('should call getAllTenants and populate displayedTenants signal', () => {
      expect(component['displayedTenants']()).toEqual(mockTenants);
    });

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

  describe('commands getter', () => {
    it('should return commission/reset/reboot for DECOMMISSIONED gateway in manage mode', () => {
      (component as unknown as Record<string, unknown>)['data'] = { ...component['data'], gateway: { ...mockGateway, status: GatewayStatus.DECOMMISSIONED } };
      expect(component['commands'].map((c) => c.value)).toEqual(['commission', 'reset', 'reboot']);
    });

    it('should return decommission/reset/reboot/interrupt for ACTIVE gateway in manage mode', () => {
      expect(component['commands'].map((c) => c.value)).toEqual(['decommission', 'reset', 'reboot', 'interrupt']);
    });

    it('should return decommission/reset/reboot/resume for INACTIVE gateway in manage mode', () => {
      (component as unknown as Record<string, unknown>)['data'] = { ...component['data'], gateway: { ...mockGateway, status: GatewayStatus.INACTIVE } };
      expect(component['commands'].map((c) => c.value)).toEqual(['decommission', 'reset', 'reboot', 'resume']);
    });

    it('should return reset/reboot for DECOMMISSIONED gateway in dashboard mode', () => {
      (component as unknown as Record<string, unknown>)['data'] = { ...component['data'], gateway: { ...mockGateway, status: GatewayStatus.DECOMMISSIONED }, mode: 'dashboard' };
      expect(component['commands'].map((c) => c.value)).toEqual(['reset', 'reboot']);
    });

    it('should return reset/reboot/interrupt for ACTIVE gateway in dashboard mode', () => {
      (component as unknown as Record<string, unknown>)['data'] = { ...component['data'], mode: 'dashboard' };
      expect(component['commands'].map((c) => c.value)).toEqual(['reset', 'reboot', 'interrupt']);
    });

    it('should return reset/reboot/resume for INACTIVE gateway in dashboard mode', () => {
      (component as unknown as Record<string, unknown>)['data'] = { ...component['data'], gateway: { ...mockGateway, status: GatewayStatus.INACTIVE }, mode: 'dashboard' };
      expect(component['commands'].map((c) => c.value)).toEqual(['reset', 'reboot', 'resume']);
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

describe('GatewayCommandsDialog - getAllTenants error', () => {
  const mockGatewayForError: Gateway = {
    id: 'gw-1',
    tenantId: undefined,
    name: 'Main Lobby Gateway',
    status: GatewayStatus.ACTIVE,
    interval: 60,
  };

  it.each([
    [{ status: 500, message: 'Server error' } as ApiError, 'Server error'],
    [{ status: 500 } as ApiError, 'Failed to fetch tenants'],
  ])('should set generalError when getAllTenants fails', async (error, expected) => {
    const errorTenantServiceMock = { getAllTenants: vi.fn().mockReturnValue(throwError(() => error)) };
    const errorDialogRefMock = { close: vi.fn() };
    const errorGatewayServiceMock = {
      commissionGateway: vi.fn(),
      decommissionGateway: vi.fn(),
      resetGateway: vi.fn(),
      rebootGateway: vi.fn(),
      interruptGateway: vi.fn(),
      resumeGateway: vi.fn(),
    };

    await TestBed.configureTestingModule({
      imports: [GatewayCommandsDialog],
      providers: [
        { provide: MatDialogRef, useValue: errorDialogRefMock },
        { provide: MAT_DIALOG_DATA, useValue: { gateway: mockGatewayForError, mode: 'manage' } },
        { provide: GatewayService, useValue: errorGatewayServiceMock },
        { provide: TenantService, useValue: errorTenantServiceMock },
      ],
    }).compileComponents();

    const fixture = TestBed.createComponent(GatewayCommandsDialog);
    const component = fixture.componentInstance;
    fixture.detectChanges();

    expect(component['generalError']()).toBe(expected);
    expect(component['displayedTenants']()).toEqual([]);
  });
});
