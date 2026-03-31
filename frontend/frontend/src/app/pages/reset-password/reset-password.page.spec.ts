import { ComponentFixture, TestBed } from '@angular/core/testing';
import { signal, WritableSignal, DebugElement, Component, input, output } from '@angular/core';
import { Router, ActivatedRoute } from '@angular/router';
import { By } from '@angular/platform-browser';
import { of, EMPTY } from 'rxjs';

import { ResetPasswordPage } from './reset-password.page';
import { ResetPasswordFormComponent } from './components/reset-password-form/reset-password-form.component';
import { AuthActionsService } from '../../services/auth/auth-actions.service';
import { ForgotPasswordResponse } from '../../models/auth/forgot-password.model';

@Component({ selector: 'app-reset-password-form', template: '', standalone: true })
class StubResetPasswordForm {
  loading = input(false);
  generalError = input<string | null>(null);
  success = input(false);
  submitReset = output<ForgotPasswordResponse>();
  goToLogin = output<void>();
  dismissError = output<void>();
}

describe('ResetPasswordPage', () => {
  let component: ResetPasswordPage;
  let fixture: ComponentFixture<ResetPasswordPage>;
  let resetFormDebug: DebugElement;

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
        get: vi.fn().mockImplementation((key: string) => {
          if (key === 'token') return 'reset-token';
          if (key === 'tenant_id') return null;
          return null;
        }),
      },
    },
  };

  beforeEach(async () => {
    vi.resetAllMocks();
    activatedRouteMock.snapshot.queryParamMap.get.mockImplementation((key: string) => {
      if (key === 'token') return 'reset-token';
      if (key === 'tenant_id') return null;
      return null;
    });

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
    })
      .overrideComponent(ResetPasswordPage, {
        remove: { imports: [ResetPasswordFormComponent] },
        add: { imports: [StubResetPasswordForm] },
      })
      .compileComponents();

    fixture = TestBed.createComponent(ResetPasswordPage);
    component = fixture.componentInstance;
    fixture.detectChanges();

    resetFormDebug = fixture.debugElement.query(By.directive(StubResetPasswordForm));
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
      expect(heading.nativeElement.textContent).toContain('Reimposta Password');
    });

    it('should render the reset password form', () => {
      expect(resetFormDebug).toBeTruthy();
    });
  });

  describe('onSubmitReset', () => {
    const mockResponse: ForgotPasswordResponse = {
      token: '',
      newPassword: 'newSecret123',
    };

    it('should merge token from route and undefined tenant_id into the request', () => {
      authActionsServiceMock.confirmPasswordReset.mockReturnValue(of(undefined));

      resetFormDebug.triggerEventHandler('submitReset', mockResponse);
      fixture.detectChanges();

      expect(authActionsServiceMock.confirmPasswordReset).toHaveBeenCalledWith({
        ...mockResponse,
        token: 'reset-token',
        tenant_id: undefined,
      });
    });

    it('should merge tenant_id from route into the request', () => {
      activatedRouteMock.snapshot.queryParamMap.get.mockImplementation((key: string) => {
        if (key === 'token') return 'reset-token';
        if (key === 'tenant_id') return 'tenant-01';
        return null;
      });
      authActionsServiceMock.confirmPasswordReset.mockReturnValue(of(undefined));

      const tenantFixture = TestBed.createComponent(ResetPasswordPage);
      tenantFixture.detectChanges();
      const tenantFormDebug = tenantFixture.debugElement.query(By.directive(StubResetPasswordForm));

      tenantFormDebug.triggerEventHandler('submitReset', mockResponse);
      tenantFixture.detectChanges();

      expect(authActionsServiceMock.confirmPasswordReset).toHaveBeenCalledWith({
        ...mockResponse,
        token: 'reset-token',
        tenantId: 'tenant-01',
      });
    });

    it('should not error when service returns EMPTY', () => {
      authActionsServiceMock.confirmPasswordReset.mockReturnValue(EMPTY);

      resetFormDebug.triggerEventHandler('submitReset', mockResponse);
      fixture.detectChanges();

      expect(authActionsServiceMock.confirmPasswordReset).toHaveBeenCalledTimes(1);
    });
  });

  describe('onGoToLogin', () => {
    it('should call clearMessages and navigate to /login', () => {
      resetFormDebug.triggerEventHandler('goToLogin');
      fixture.detectChanges();

      expect(authActionsServiceMock.clearMessages).toHaveBeenCalled();
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
