import { ComponentFixture, TestBed } from '@angular/core/testing';
import { NO_ERRORS_SCHEMA, signal, WritableSignal } from '@angular/core';
import { Router, ActivatedRoute } from '@angular/router';
import { By } from '@angular/platform-browser';
import { of, EMPTY } from 'rxjs';

import { ResetPasswordPage } from './reset-password.page';
import { AuthActionsService } from '../../services/auth/auth-actions.service';
import { ForgotPasswordResponse } from '../../models/auth/forgot-password.model';

describe('ResetPasswordPage', () => {
  let component: ResetPasswordPage;
  let fixture: ComponentFixture<ResetPasswordPage>;
  // ESLint whining
  let resetFormDebug: any;

  let loadingSignal: WritableSignal<boolean>;
  let errorSignal: WritableSignal<string | null>;
  let passwordChangeResultSignal: WritableSignal<boolean | null>;

  let authActionsServiceMock: {
    confirmPasswordReset: ReturnType<typeof vi.fn>;
    clearMessages: ReturnType<typeof vi.fn>;
    loading: ReturnType<WritableSignal<boolean>['asReadonly']>;
    error: ReturnType<WritableSignal<string | null>['asReadonly']>;
    passwordChangeResult: ReturnType<WritableSignal<boolean | null>['asReadonly']>;
  };

  const routerMock = {
    navigate: vi.fn(),
  };

  const activatedRouteMock = {
    snapshot: {
      queryParamMap: {
        get: vi.fn().mockReturnValue('reset-token'),
      },
    },
  };

  beforeEach(async () => {
    vi.resetAllMocks();
    activatedRouteMock.snapshot.queryParamMap.get.mockReturnValue('reset-token');

    loadingSignal = signal(false);
    errorSignal = signal<string | null>(null);
    passwordChangeResultSignal = signal<boolean | null>(null);

    authActionsServiceMock = {
      confirmPasswordReset: vi.fn(),
      clearMessages: vi.fn(),
      loading: loadingSignal.asReadonly(),
      error: errorSignal.asReadonly(),
      passwordChangeResult: passwordChangeResultSignal.asReadonly(),
    };

    await TestBed.configureTestingModule({
      imports: [ResetPasswordPage],
      providers: [
        { provide: AuthActionsService, useValue: authActionsServiceMock },
        { provide: Router, useValue: routerMock },
        { provide: ActivatedRoute, useValue: activatedRouteMock },
      ],
      schemas: [NO_ERRORS_SCHEMA],
    }).compileComponents();

    fixture = TestBed.createComponent(ResetPasswordPage);
    component = fixture.componentInstance;
    fixture.detectChanges();

    resetFormDebug = fixture.debugElement.query(By.css('app-reset-password-form'));
  });

  describe('initial state', () => {
    it('should create', () => {
      expect(component).toBeTruthy();
    });

    it('should render the reset container', () => {
      const container = fixture.debugElement.query(By.css('.reset-container'));
      expect(container).toBeTruthy();
    });

    it('should render the heading', () => {
      const heading = fixture.debugElement.query(By.css('h1'));
      expect(heading.nativeElement.textContent).toContain('Reset Password');
    });

    it('should render the reset password form', () => {
      expect(resetFormDebug).toBeTruthy();
    });
  });

  describe('onSubmitReset', () => {
    const mockResponse: ForgotPasswordResponse = {
      token: 'reset-token',
      newPassword: 'newSecret123',
    };

    it('should call confirmPasswordReset with the emitted ForgotPasswordResponse', () => {
      authActionsServiceMock.confirmPasswordReset.mockReturnValue(of(undefined));

      resetFormDebug.triggerEventHandler('submitReset', mockResponse);
      fixture.detectChanges();

      expect(authActionsServiceMock.confirmPasswordReset).toHaveBeenCalledWith(mockResponse);
    });

    it('should subscribe to the observable', () => {
      authActionsServiceMock.confirmPasswordReset.mockReturnValue(of(undefined));

      resetFormDebug.triggerEventHandler('submitReset', mockResponse);
      fixture.detectChanges();

      expect(authActionsServiceMock.confirmPasswordReset).toHaveBeenCalledTimes(1);
    });

    it('should not error when service returns EMPTY', () => {
      authActionsServiceMock.confirmPasswordReset.mockReturnValue(EMPTY);

      resetFormDebug.triggerEventHandler('submitReset', mockResponse);
      fixture.detectChanges();

      expect(authActionsServiceMock.confirmPasswordReset).toHaveBeenCalledTimes(1);
    });
  });

  describe('onGoToLogin', () => {
    it('should navigate to /login', () => {
      resetFormDebug.triggerEventHandler('goToLogin');
      fixture.detectChanges();

      expect(routerMock.navigate).toHaveBeenCalledWith(['/login']);
    });
  });

  describe('onDismissError', () => {
    it('should call clearMessages', () => {
      resetFormDebug.triggerEventHandler('dismissError');
      fixture.detectChanges();

      expect(authActionsServiceMock.clearMessages).toHaveBeenCalled();
    });
  });
});
