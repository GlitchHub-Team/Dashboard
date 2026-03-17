import { ComponentFixture, TestBed } from '@angular/core/testing';
import { Router, ActivatedRoute } from '@angular/router';
import { of } from 'rxjs';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';

import { ResetPasswordPage } from './reset-password.page';
import { AuthActionsService } from '../../services/auth/auth-actions.service';
import { PasswordChange } from '../../models/password-change.model';
import { signal } from '@angular/core';

describe('ResetPasswordPage', () => {
  let component: ResetPasswordPage;
  let fixture: ComponentFixture<ResetPasswordPage>;

  const authActionsServiceMock = {
    confirmPasswordChange: vi.fn(),
    clearMessages: vi.fn(),
    loading: signal(false).asReadonly(),
    error: signal<string | null>(null).asReadonly(),
    passwordChangeResult: signal<boolean | null>(null).asReadonly(),
  };

  const routerMock = {
    navigate: vi.fn(),
  };

  const activatedRouteMock = {
    snapshot: {
      queryParamMap: {
        get: vi.fn().mockReturnValue(null),
      },
    },
  };

  beforeEach(async () => {
    vi.resetAllMocks();

    await TestBed.configureTestingModule({
      imports: [ResetPasswordPage],
      providers: [
        { provide: AuthActionsService, useValue: authActionsServiceMock },
        { provide: Router, useValue: routerMock },
        { provide: ActivatedRoute, useValue: activatedRouteMock },
      ],
      schemas: [CUSTOM_ELEMENTS_SCHEMA],
    }).compileComponents();

    fixture = TestBed.createComponent(ResetPasswordPage);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  describe('onSubmitReset', () => {
    it('should call confirmPasswordChange with password and token', () => {
      authActionsServiceMock.confirmPasswordChange.mockReturnValue(of(undefined));

      const expectedData: PasswordChange = {
        newPassword: 'newSecret123',
        token: 'TODO',
      };

      component['onSubmitReset']('newSecret123');

      expect(authActionsServiceMock.confirmPasswordChange).toHaveBeenCalledWith(expectedData);
    });

    it('should subscribe to the observable', () => {
      authActionsServiceMock.confirmPasswordChange.mockReturnValue(of(undefined));

      component['onSubmitReset']('newSecret123');

      expect(authActionsServiceMock.confirmPasswordChange).toHaveBeenCalledTimes(1);
    });
  });

  describe('onGoToLogin', () => {
    it('should navigate to /login', () => {
      component['onGoToLogin']();

      expect(routerMock.navigate).toHaveBeenCalledWith(['/login']);
    });
  });

  describe('onDismissError', () => {
    it('should call clearMessages', () => {
      component['onDismissError']();

      expect(authActionsServiceMock.clearMessages).toHaveBeenCalled();
    });
  });
});
