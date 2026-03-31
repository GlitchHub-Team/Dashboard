import { ComponentFixture, TestBed } from '@angular/core/testing';
import { signal, WritableSignal, Component, input, output } from '@angular/core';
import { Router, ActivatedRoute } from '@angular/router';
import { By } from '@angular/platform-browser';
import { DebugElement } from '@angular/core';
import { of, EMPTY } from 'rxjs';

import { ConfirmAccountPage } from './confirm-account.page';
import { ConfirmAccountFormComponent } from './components/confirm-account-form/confirm-account-form.component';
import { AuthActionsService } from '../../services/auth/auth-actions.service';
import { ConfirmAccountResponse } from '../../models/auth/confirm-account.model';

@Component({ selector: 'app-confirm-account-form', template: '', standalone: true })
class StubConfirmAccountForm {
  loading = input(false);
  generalError = input<string | null>(null);
  submitConfirmAccount = output<ConfirmAccountResponse>();
  dismissError = output<void>();
}

describe('ConfirmAccountPage', () => {
  let component: ConfirmAccountPage;
  let fixture: ComponentFixture<ConfirmAccountPage>;
  let confirmFormDebug: DebugElement;

  let loadingSignal: WritableSignal<boolean>;
  let errorSignal: WritableSignal<string | null>;

  let authActionsServiceMock: {
    confirmAccount: ReturnType<typeof vi.fn>;
    clearMessages: ReturnType<typeof vi.fn>;
    loading: ReturnType<WritableSignal<boolean>['asReadonly']>;
    error: ReturnType<WritableSignal<string | null>['asReadonly']>;
  };

  const routerMock = { navigate: vi.fn() };

  const activatedRouteMock = {
    snapshot: {
      queryParamMap: {
        get: vi.fn().mockImplementation((key: string) => {
          if (key === 'token') return null;
          if (key === 'tenant_id') return null;
          return null;
        }),
      },
    },
  };

  beforeEach(async () => {
    vi.resetAllMocks();

    loadingSignal = signal(false);
    errorSignal = signal<string | null>(null);

    authActionsServiceMock = {
      confirmAccount: vi.fn(),
      clearMessages: vi.fn(),
      loading: loadingSignal.asReadonly(),
      error: errorSignal.asReadonly(),
    };

    await TestBed.configureTestingModule({
      imports: [ConfirmAccountPage],
      providers: [
        { provide: AuthActionsService, useValue: authActionsServiceMock },
        { provide: Router, useValue: routerMock },
        { provide: ActivatedRoute, useValue: activatedRouteMock },
      ],
    })
      .overrideComponent(ConfirmAccountPage, {
        remove: { imports: [ConfirmAccountFormComponent] },
        add: { imports: [StubConfirmAccountForm] },
      })
      .compileComponents();

    fixture = TestBed.createComponent(ConfirmAccountPage);
    component = fixture.componentInstance;
    fixture.detectChanges();

    confirmFormDebug = fixture.debugElement.query(By.directive(StubConfirmAccountForm));
  });

  describe('initial state', () => {
    it('should create', () => {
      expect(component).toBeTruthy();
    });

    it('should render the confirm-account container', () => {
      const container = fixture.debugElement.query(By.css('.confirm-account-container'));
      expect(container).toBeTruthy();
    });

    it('should render the heading', () => {
      const heading = fixture.debugElement.query(By.css('h1'));
      expect(heading.nativeElement.textContent).toContain('Conferma Account');
    });

    it('should render the confirm account form', () => {
      expect(confirmFormDebug).toBeTruthy();
    });
  });

  describe('onConfirmAccount', () => {
    const mockRequest: ConfirmAccountResponse = { token: '', newPassword: 'newSecret123' };

    it('should call confirmAccount with empty token and undefined tenant_id when route has none', () => {
      authActionsServiceMock.confirmAccount.mockReturnValue(of(undefined));

      confirmFormDebug.triggerEventHandler('submitConfirmAccount', mockRequest);
      fixture.detectChanges();

      expect(authActionsServiceMock.confirmAccount).toHaveBeenCalledWith({
        ...mockRequest,
        token: '',
        tenant_id: undefined,
      });
    });

    it('should merge token from route into the request', () => {
      activatedRouteMock.snapshot.queryParamMap.get.mockImplementation((key: string) => {
        if (key === 'token') return 'route-token';
        return null;
      });
      authActionsServiceMock.confirmAccount.mockReturnValue(of(undefined));

      const tokenFixture = TestBed.createComponent(ConfirmAccountPage);
      tokenFixture.detectChanges();
      const tokenFormDebug = tokenFixture.debugElement.query(By.directive(StubConfirmAccountForm));

      tokenFormDebug.triggerEventHandler('submitConfirmAccount', mockRequest);
      tokenFixture.detectChanges();

      expect(authActionsServiceMock.confirmAccount).toHaveBeenCalledWith({
        ...mockRequest,
        token: 'route-token',
        tenant_id: undefined,
      });
    });

    it('should merge tenant_id from route into the request', () => {
      activatedRouteMock.snapshot.queryParamMap.get.mockImplementation((key: string) => {
        if (key === 'token') return 'route-token';
        if (key === 'tenant_id') return 'tenant-01';
        return null;
      });
      authActionsServiceMock.confirmAccount.mockReturnValue(of(undefined));

      const tenantFixture = TestBed.createComponent(ConfirmAccountPage);
      tenantFixture.detectChanges();
      const tenantFormDebug = tenantFixture.debugElement.query(
        By.directive(StubConfirmAccountForm),
      );

      tenantFormDebug.triggerEventHandler('submitConfirmAccount', mockRequest);
      tenantFixture.detectChanges();

      expect(authActionsServiceMock.confirmAccount).toHaveBeenCalledWith({
        ...mockRequest,
        token: 'route-token',
        tenantId: 'tenant-01',
      });
    });

    it('should navigate to /dashboard on success', () => {
      authActionsServiceMock.confirmAccount.mockReturnValue(of(undefined));

      confirmFormDebug.triggerEventHandler('submitConfirmAccount', mockRequest);
      fixture.detectChanges();

      expect(routerMock.navigate).toHaveBeenCalledWith(['/dashboard']);
    });

    it('should not navigate when service returns EMPTY', () => {
      authActionsServiceMock.confirmAccount.mockReturnValue(EMPTY);

      confirmFormDebug.triggerEventHandler('submitConfirmAccount', mockRequest);
      fixture.detectChanges();

      expect(authActionsServiceMock.confirmAccount).toHaveBeenCalledTimes(1);
      expect(routerMock.navigate).not.toHaveBeenCalled();
    });
  });

  describe('onDismissError', () => {
    it('should call clearMessages', () => {
      confirmFormDebug.triggerEventHandler('dismissError');
      fixture.detectChanges();

      expect(authActionsServiceMock.clearMessages).toHaveBeenCalled();
    });
  });
});
