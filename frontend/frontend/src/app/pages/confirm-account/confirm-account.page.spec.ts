import { ComponentFixture, TestBed } from '@angular/core/testing';
import { NO_ERRORS_SCHEMA, signal, WritableSignal } from '@angular/core';
import { Router, ActivatedRoute } from '@angular/router';
import { By } from '@angular/platform-browser';
import { of, EMPTY } from 'rxjs';

import { ConfirmAccountPage } from './confirm-account.page';
import { AuthActionsService } from '../../services/auth/auth-actions.service';
import { ConfirmAccountResponse } from '../../models/auth/confirm-account.model';

describe('ConfirmAccountPage', () => {
  let component: ConfirmAccountPage;
  let fixture: ComponentFixture<ConfirmAccountPage>;
  // ESLint whining
  let confirmFormDebug: any;

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
        get: vi.fn().mockReturnValue(null),
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
      schemas: [NO_ERRORS_SCHEMA],
    }).compileComponents();

    fixture = TestBed.createComponent(ConfirmAccountPage);
    component = fixture.componentInstance;
    fixture.detectChanges();

    confirmFormDebug = fixture.debugElement.query(By.css('app-confirm-account-form'));
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

    it('should call confirmAccount with the emitted request and empty token when route has no token', () => {
      authActionsServiceMock.confirmAccount.mockReturnValue(of(undefined));

      confirmFormDebug.triggerEventHandler('submitConfirmAccount', mockRequest);
      fixture.detectChanges();

      expect(authActionsServiceMock.confirmAccount).toHaveBeenCalledWith({
        ...mockRequest,
        token: '',
      });
    });

    it('should call confirmAccount with the token from the route merged in', () => {
      activatedRouteMock.snapshot.queryParamMap.get.mockReturnValue('route-token');
      authActionsServiceMock.confirmAccount.mockReturnValue(of(undefined));

      const tokenFixture = TestBed.createComponent(ConfirmAccountPage);
      tokenFixture.detectChanges();
      const tokenFormDebug = tokenFixture.debugElement.query(By.css('app-confirm-account-form'));

      tokenFormDebug.triggerEventHandler('submitConfirmAccount', mockRequest);
      tokenFixture.detectChanges();

      expect(authActionsServiceMock.confirmAccount).toHaveBeenCalledWith({
        ...mockRequest,
        token: 'route-token',
      });
    });

    it('should navigate to /login on success', () => {
      authActionsServiceMock.confirmAccount.mockReturnValue(of(undefined));

      confirmFormDebug.triggerEventHandler('submitConfirmAccount', mockRequest);
      fixture.detectChanges();

      expect(routerMock.navigate).toHaveBeenCalledWith(['/login']);
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
