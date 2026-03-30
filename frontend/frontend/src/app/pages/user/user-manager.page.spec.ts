import { ComponentFixture, TestBed } from '@angular/core/testing';
import { signal } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { MatDialog } from '@angular/material/dialog';
import { PageEvent } from '@angular/material/paginator';
import { Observable, of, Subject } from 'rxjs';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { UserSessionService } from '../../services/user-session/user-session.service';
import { UserManagerPage } from './user-manager.page';
import { UserService } from '../../services/user/user.service';
import { UserSessionService } from '../../services/user-session/user-session.service';
import { User } from '../../models/user/user.model';
import { UserRole } from '../../models/user/user-role.enum';
import { UserFormDialogComponent } from './dialogs/user-form/user-form.dialog';
import { ConfirmDeleteDialog } from '../gateway-sensor/dialogs/confirm-delete/confirm-delete.dialog';

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
  let userSessionServiceMock: {
    currentRole: ReturnType<typeof vi.fn>;
    currentTenant: ReturnType<typeof vi.fn>;
  };
  let activatedRouteMock: { data: Observable<unknown>; queryParams: Observable<unknown> };

  const routeContext = {
    title: 'Test Users',
    role: UserRole.TENANT_ADMIN,
  };

  const sessionTenantId = 'session-tenant';

  const userSessionMock = {
    currentUser: signal({
      userId: 'user-1',
      tenantId: sessionTenantId,
      role: UserRole.TENANT_ADMIN,
    }),
  };

  const userServiceMock = {
    userList: signal<User[]>([]),
    total: signal(0),
    pageIndex: signal(0),
    limit: signal(10),
    loading: signal(false),
    error: signal<string | null>(null),
    retrieveUser: vi.fn(),
    addNewUser: vi.fn(),
    removeUser: vi.fn().mockReturnValue(of(void 0)),
    changePage: vi.fn(),
  };

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

    userSessionServiceMock = {
      currentRole: vi.fn(),
      currentTenant: vi.fn(),
    };

    activatedRouteMock = {
      data: of({ userManagerContext: routeContext }),
      queryParams: of({}),
    };

    await TestBed.configureTestingModule({
      imports: [UserManagerPage, NoopAnimationsModule],
      providers: [
        { provide: UserService, useValue: userServiceMock },
        { provide: UserSessionService, useValue: userSessionMock },
        { provide: MatDialog, useValue: dialogMock },
        { provide: ActivatedRoute, useValue: activatedRouteMock },
        { provide: UserSessionService, useValue: userSessionServiceMock },
        { provide: Router, useValue: { navigate: vi.fn() } },
      ],
    })
      .overrideProvider(MatDialog, { useValue: dialogMock })
      .compileComponents();

    fixture = TestBed.createComponent(UserManagerPage);
    component = fixture.componentInstance;
    testApi = component as unknown as UserManagerPageTestApi;
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  describe('ngOnInit', () => {
    it('should set context for TENANT_ADMIN and retrieve their users', () => {
      userSessionServiceMock.currentRole.mockReturnValue(UserRole.TENANT_ADMIN);
      userSessionServiceMock.currentTenant.mockReturnValue('tenant-from-session');
      activatedRouteMock.queryParams = of({ tenantId: 'tenant-from-url-ignored' });

      fixture.detectChanges();

      const expectedContext = { ...routeContext, tenantId: 'tenant-from-session' };
      expect((component as any).context()).toEqual(expectedContext);
      expect(userServiceMock.retrieveUser).toHaveBeenCalledWith(
        expectedContext.role,
        expectedContext.tenantId,
      );
    });

    it('should set context for SUPER_ADMIN with tenantId from URL and retrieve users', () => {
      userSessionServiceMock.currentRole.mockReturnValue(UserRole.SUPER_ADMIN);
      activatedRouteMock.queryParams = of({ tenantId: 'tenant-from-url' });

      fixture.detectChanges();

      const expectedContext = { ...routeContext, tenantId: 'tenant-from-url' };
      expect((component as any).context()).toEqual(expectedContext);
      expect(userServiceMock.retrieveUser).toHaveBeenCalledWith(
        expectedContext.role,
        expectedContext.tenantId,
      );
    });

    it('should set context for SUPER_ADMIN without tenantId and retrieve users', () => {
      userSessionServiceMock.currentRole.mockReturnValue(UserRole.SUPER_ADMIN);
      activatedRouteMock.queryParams = of({});

      fixture.detectChanges();

      const expectedContext = { ...routeContext, tenantId: undefined };
      expect((component as any).context()).toEqual(expectedContext);
      expect(userServiceMock.retrieveUser).toHaveBeenCalledWith(
        expectedContext.role,
        expectedContext.tenantId,
      );
    });

    it('should not retrieve users if role is TENANT_USER and no tenantId is resolved', () => {
      const tenantUserRouteContext = {
        title: 'Test Tenant Users',
        role: UserRole.TENANT_USER,
      };
      activatedRouteMock.data = of({ userManagerContext: tenantUserRouteContext });
      userSessionServiceMock.currentRole.mockReturnValue(UserRole.TENANT_ADMIN);
      userSessionServiceMock.currentTenant.mockReturnValue(null);
      activatedRouteMock.queryParams = of({});

      fixture.detectChanges();

      const expectedContext = { ...tenantUserRouteContext, tenantId: undefined };
      expect((component as any).context()).toEqual(expectedContext);
      expect(userServiceMock.retrieveUser).not.toHaveBeenCalled();
    });
  it('should initialize context with session tenantId and retrieve users on init', () => {
    fixture.detectChanges();

    expect((component as unknown as { context: () => unknown }).context()).toEqual({
      title: 'Test Users',
      role: UserRole.TENANT_ADMIN,
      tenantId: sessionTenantId,
    });
    expect(userServiceMock.retrieveUser).toHaveBeenCalledWith(
      UserRole.TENANT_ADMIN,
      sessionTenantId,
    );
  });

  it('should open create dialog with correct config', () => {
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

  it('should refetch users after create dialog closes with true', () => {
    fixture.detectChanges();
    const callsBefore = (userServiceMock.retrieveUser as ReturnType<typeof vi.fn>).mock.calls
      .length;

    testApi.onCreateUser();
    afterClosedSubject.next(true);

    expect(userServiceMock.retrieveUser).toHaveBeenCalledTimes(callsBefore + 1);
  });

  it('should not refetch users after create dialog closes with false', () => {
    fixture.detectChanges();
    const callsBefore = (userServiceMock.retrieveUser as ReturnType<typeof vi.fn>).mock.calls
      .length;

    testApi.onCreateUser();
    afterClosedSubject.next(false);

    expect(userServiceMock.retrieveUser).toHaveBeenCalledTimes(callsBefore);
  });

  it('should open delete dialog with correct config', () => {
    const user: User = {
      id: '1',
      email: 'delete@user.com',
      username: 'deleteuser',
      role: UserRole.TENANT_ADMIN,
      tenantId: sessionTenantId,
    };

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
    userSessionServiceMock.currentRole.mockReturnValue(UserRole.SUPER_ADMIN);
    activatedRouteMock.queryParams = of({ tenantId: 'tenant-1' });
    fixture.detectChanges();
    const user: User = {
      id: '1',
      email: 'delete@user.com',
      username: 'deleteuser',
      role: UserRole.TENANT_ADMIN,
      tenantId: sessionTenantId,
    };

    userServiceMock.retrieveUser.mockClear();

    testApi.onDeleteUser(user);
    afterClosedSubject.next(confirmed);

    if (shouldDelete) {
      expect(userServiceMock.removeUser).toHaveBeenCalledWith(user);
      expect(userServiceMock.retrieveUser).toHaveBeenCalledWith(
        UserRole.TENANT_ADMIN,
        sessionTenantId,
      );
      return;
    }

    expect(userServiceMock.removeUser).not.toHaveBeenCalled();
    expect(userServiceMock.retrieveUser).not.toHaveBeenCalled();
  });

  it('should call changePage with context tenantId', () => {
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
