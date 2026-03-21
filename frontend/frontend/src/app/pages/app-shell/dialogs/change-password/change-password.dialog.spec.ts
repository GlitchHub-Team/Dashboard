import { ComponentFixture, TestBed } from '@angular/core/testing';
import { MatDialogRef } from '@angular/material/dialog';
import { signal } from '@angular/core';

import { ChangePasswordDialog } from './change-password.dialog';
import { AuthActionsService } from '../../../../services/auth/auth-actions.service';
import { TokenStorageService } from '../../../../services/token-storage/token-storage.service';

describe('ChangePasswordDialog', () => {
  let component: ChangePasswordDialog;
  let fixture: ComponentFixture<ChangePasswordDialog>;

  const authActionsServiceMock = {
    confirmPasswordChange: vi.fn(),
    clearMessages: vi.fn(),
    loading: signal(false).asReadonly(),
    error: signal<string | null>(null).asReadonly(),
    passwordChangeResult: signal<boolean | null>(null).asReadonly(),
  };

  const dialogRefMock = {
    close: vi.fn(),
  };

  const tokenStorageMock = {
    getToken: vi.fn().mockReturnValue('mock-token'),
  };

  beforeEach(async () => {
    vi.resetAllMocks();
    tokenStorageMock.getToken.mockReturnValue('mock-token');

    await TestBed.configureTestingModule({
      imports: [ChangePasswordDialog],
      providers: [
        { provide: AuthActionsService, useValue: authActionsServiceMock },
        { provide: MatDialogRef, useValue: dialogRefMock },
        { provide: TokenStorageService, useValue: tokenStorageMock },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(ChangePasswordDialog);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
