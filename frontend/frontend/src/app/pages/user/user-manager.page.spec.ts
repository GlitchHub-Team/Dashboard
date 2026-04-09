import { ComponentFixture, TestBed } from '@angular/core/testing';
import { WritableSignal, signal } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { MatDialog } from '@angular/material/dialog';
import { MatSnackBar } from '@angular/material/snack-bar';
import { PageEvent } from '@angular/material/paginator';
import { Observable, of, Subject } from 'rxjs';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { UserSessionService } from '../../services/user-session/user-session.service';
import { TenantService } from '../../services/tenant/tenant.service';
import { UserManagerPage } from './user-manager.page';
import { UserService } from '../../services/user/user.service';
import { User } from '../../models/user/user.model';
import { UserRole } from '../../models/user/user-role.enum';
import { UserSession } from '../../models/auth/user-session.model';
import { UserFormDialogComponent } from './dialogs/user-form/user-form.dialog';
import { ConfirmDeleteDialog } from '../shared/dialogs/confirm-delete/confirm-delete.dialog';

interface UserManagerPageTestApi {
  onCreateUser: () => void;
  onDeleteUser: (user: User) => void;
  onPageChange: (event: PageEvent) => void;
}

describe('UserManagerPage', () => {
  let component: UserManagerPage;
  let fixture: ComponentFixture<UserManagerPage>;
  let testApi: UserManagerPageTestApi;

  let afterClosedSubject: Subject<unknown>;
  let dialogMock: { open: ReturnType<typeof vi.fn> };
  let currentUserSignal: WritableSignal<UserSession | null>;
  let activatedRouteMock: { data: Observable<unknown>; queryParams: Observable<unknown> };
  let tenantServiceMock: { getTenant: ReturnType<typeof vi.fn> };

  const routeContext = {
    title: 'Test Users',
    role: UserRole.TENANT_ADMIN,
  };

  const sessionTenantId = 'session-tenant';

  const userServiceMock = {
    userList: signal<User[]>([]),
    total: signal(0),
    pageIndex: signal(0),
    limit: signal(10),
    loading: signal(false),
    error: signal<string | null>(null),
    retrieveUsers: vi.fn(),
    addNewUser: vi.fn(),
    removeUser: vi.fn().mockReturnValue(of(void 0)),
    changePage: vi.fn(),
  };

  function createComponent(): void {
    fixture = TestBed.createComponent(UserManagerPage);
    component = fixture.componentInstance;
    testApi = component as unknown as UserManagerPageTestApi;
  }

  beforeEach(async () => {
    vi.resetAllMocks();

    afterClosedSubject = new Subject();
    userServiceMock.addNewUser.mockReturnValue(of(void 0));
    userServiceMock.removeUser.mockReturnValue(of(void 0));

    dialogMock = {
      open: vi.fn().mockReturnValue({
        afterClosed: () => afterClosedSubject.asObservable(),
      }),
    };

    currentUserSignal = signal<UserSession | null>({
      userId: 'user-1',
      tenantId: sessionTenantId,
      role: UserRole.TENANT_ADMIN,
    });

    activatedRouteMock = {
      data: of({ userManagerContext: routeContext }),
      queryParams: of({}),
    };

    tenantServiceMock = {
      getTenant: vi.fn().mockReturnValue(of({ id: 'mock', name: 'Mock', canImpersonate: true })),
    };

    await TestBed.configureTestingModule({
      imports: [UserManagerPage],
      providers: [
        { provide: UserService, useValue: userServiceMock },
        { provide: UserSessionService, useValue: { currentUser: currentUserSignal } },
        { provide: ActivatedRoute, useValue: activatedRouteMock },
        { provide: Router, useValue: { navigate: vi.fn() } },
        { provide: TenantService, useValue: tenantServiceMock },
      ],
    })
      .overrideProvider(MatDialog, { useValue: dialogMock })
      .overrideProvider(MatSnackBar, { useValue: { open: vi.fn() } })
      .compileComponents();
  });

  it('should create, initialize context with session tenantId, and retrieve users on init', () => {
    createComponent();
    fixture.detectChanges();

    expect(component).toBeTruthy();
    expect((component as any).context()).toEqual({
      title: 'Test Users',
      role: UserRole.TENANT_ADMIN,
      tenantId: sessionTenantId,
    });
    expect(userServiceMock.retrieveUsers).toHaveBeenCalledWith(
      UserRole.TENANT_ADMIN,
      sessionTenantId,
    );
  });

  describe('ngOnInit', () => {
    it('should set context for TENANT_ADMIN and retrieve their users', () => {
      currentUserSignal.set({
        userId: 'user-1',
        tenantId: 'tenant-from-session',
        role: UserRole.TENANT_ADMIN,
      });
      activatedRouteMock.queryParams = of({ tenantId: 'tenant-from-url-ignored' });

      createComponent();
      fixture.detectChanges();

      const expectedContext = { ...routeContext, tenantId: 'tenant-from-session' };
      expect((component as any).context()).toEqual(expectedContext);
      expect(userServiceMock.retrieveUsers).toHaveBeenCalledWith(
        expectedContext.role,
        expectedContext.tenantId,
      );
    });

    it('should set context for SUPER_ADMIN with tenantId from URL and retrieve users', () => {
      currentUserSignal.set({ userId: 'user-1', role: UserRole.SUPER_ADMIN });
      activatedRouteMock.queryParams = of({ tenantId: 'tenant-from-url' });

      createComponent();
      fixture.detectChanges();

      expect(tenantServiceMock.getTenant).toHaveBeenCalledWith('tenant-from-url');
      const expectedContext = { ...routeContext, tenantId: 'tenant-from-url' };
      expect((component as any).context()).toEqual(expectedContext);
      expect(userServiceMock.retrieveUsers).toHaveBeenCalledWith(
        expectedContext.role,
        expectedContext.tenantId,
      );
    });

    it('should navigate to /user-management/tenant-users when tenant canImpersonate is false', () => {
      const routerSpy = vi.fn();
      TestBed.overrideProvider(Router, { useValue: { navigate: routerSpy } });

      tenantServiceMock.getTenant.mockReturnValue(
        of({ id: 'restricted', name: 'Restricted', canImpersonate: false }),
      );
      currentUserSignal.set({ userId: 'user-1', role: UserRole.SUPER_ADMIN });
      activatedRouteMock.queryParams = of({ tenantId: 'restricted-tenant' });

      createComponent();
      fixture.detectChanges();

      expect(tenantServiceMock.getTenant).toHaveBeenCalledWith('restricted-tenant');
      expect(routerSpy).toHaveBeenCalledWith(['/user-management/tenant-users']);
    });

    it('should set context for SUPER_ADMIN without tenantId and retrieve users', () => {
      currentUserSignal.set({ userId: 'user-1', role: UserRole.SUPER_ADMIN });
      activatedRouteMock.queryParams = of({});

      createComponent();
      fixture.detectChanges();

      const expectedContext = { ...routeContext, tenantId: undefined };
      expect((component as any).context()).toEqual(expectedContext);
      expect(userServiceMock.retrieveUsers).toHaveBeenCalledWith(
        expectedContext.role,
        expectedContext.tenantId,
      );
    });

    it('should not retrieve users if role is TENANT_USER and no tenantId is resolved', () => {
      const tenantUserRouteContext = {
        title: 'Test Tenant Users',
        role: UserRole.TENANT_USER,
      };
      currentUserSignal.set({ userId: 'user-1', role: UserRole.TENANT_ADMIN });
      activatedRouteMock.data = of({ userManagerContext: tenantUserRouteContext });
      activatedRouteMock.queryParams = of({});

      createComponent();
      fixture.detectChanges();

      const expectedContext = { ...tenantUserRouteContext, tenantId: undefined };
      expect((component as any).context()).toEqual(expectedContext);
      expect(userServiceMock.retrieveUsers).not.toHaveBeenCalled();
    });

    it('should initialize context with session tenantId and retrieve users on init', () => {
      createComponent();
      fixture.detectChanges();

      // covered by the top-level create test; kept as a final sanity check
      expect(userServiceMock.retrieveUsers).toHaveBeenCalledWith(
        UserRole.TENANT_ADMIN,
        sessionTenantId,
      );
    });
  });

  describe('onCreateUser', () => {
    it('should open create dialog with correct config', () => {
      createComponent();
      fixture.detectChanges();
      testApi.onCreateUser();

      expect(dialogMock.open).toHaveBeenCalledWith(UserFormDialogComponent, {
        width: '400px',
        data: {
          role: UserRole.TENANT_ADMIN,
          tenantId: sessionTenantId,
        },
      });
    });

    it.each([
      { result: true, shouldRefetch: true },
      { result: false, shouldRefetch: false },
    ])('should refetch=$shouldRefetch after dialog closes with $result', ({ result, shouldRefetch }) => {
      createComponent();
      fixture.detectChanges();
      const callsBefore = (userServiceMock.retrieveUsers as ReturnType<typeof vi.fn>).mock.calls.length;

      testApi.onCreateUser();
      afterClosedSubject.next(result);

      expect(userServiceMock.retrieveUsers).toHaveBeenCalledTimes(
        shouldRefetch ? callsBefore + 1 : callsBefore,
      );
    });
  });

  describe('onDeleteUser', () => {
    it('should open delete dialog with correct config', () => {
      const user: User = {
        id: '1',
        email: 'delete@user.com',
        username: 'deleteuser',
        role: UserRole.TENANT_ADMIN,
        tenantId: sessionTenantId,
      };

      createComponent();
      fixture.detectChanges();
      testApi.onDeleteUser(user);

      expect(dialogMock.open).toHaveBeenCalledWith(ConfirmDeleteDialog, {
        width: '400px',
        data: {
          title: 'Elimina Utente',
          message: `Sei sicuro di voler eliminare "${user.email}"?`,
        },
      });
    });

    it.each([
      { confirmed: true, shouldDelete: true },
      { confirmed: false, shouldDelete: false },
    ])('should handle delete confirmation: $confirmed', ({ confirmed, shouldDelete }) => {
      currentUserSignal.set({ userId: 'user-1', role: UserRole.SUPER_ADMIN });
      activatedRouteMock.queryParams = of({ tenantId: 'tenant-1' });

      createComponent();
      fixture.detectChanges();

      const user: User = {
        id: '1',
        email: 'delete@user.com',
        username: 'deleteuser',
        role: UserRole.TENANT_ADMIN,
        tenantId: sessionTenantId,
      };

      userServiceMock.retrieveUsers.mockClear();

      testApi.onDeleteUser(user);
      afterClosedSubject.next(confirmed);

      if (shouldDelete) {
        expect(userServiceMock.removeUser).toHaveBeenCalledWith(user);
        expect(userServiceMock.retrieveUsers).toHaveBeenCalledWith(
          UserRole.TENANT_ADMIN,
          'tenant-1',
        );
        return;
      }

      expect(userServiceMock.removeUser).not.toHaveBeenCalled();
      expect(userServiceMock.retrieveUsers).not.toHaveBeenCalled();
    });
  });

  describe('onPageChange', () => {
    it('should call changePage with context tenantId', () => {
      createComponent();
      fixture.detectChanges();
      const event: PageEvent = { pageIndex: 2, pageSize: 25, length: 100 };

      testApi.onPageChange(event);

      expect(userServiceMock.changePage).toHaveBeenCalledWith(
        2,
        25,
        UserRole.TENANT_ADMIN,
        sessionTenantId,
      );
    });
  });
});
