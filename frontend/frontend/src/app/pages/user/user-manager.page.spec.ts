import { ComponentFixture, TestBed } from '@angular/core/testing';
import { signal } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { MatDialog } from '@angular/material/dialog';
import { PageEvent } from '@angular/material/paginator';
import { of, Subject } from 'rxjs';
import { beforeEach, describe, expect, it, vi } from 'vitest';

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
    addNewUser: vi.fn().mockReturnValue(of(void 0)),
    removeUser: vi.fn().mockReturnValue(of(void 0)),
    changePage: vi.fn(),
  };

  beforeEach(async () => {
    vi.clearAllMocks();

    afterClosedSubject = new Subject();
    dialogMock = {
      open: vi.fn().mockReturnValue({
        afterClosed: () => afterClosedSubject.asObservable(),
      }),
    };

    await TestBed.configureTestingModule({
      imports: [UserManagerPage],
      providers: [
        { provide: UserService, useValue: userServiceMock },
        { provide: UserSessionService, useValue: userSessionMock },
        { provide: MatDialog, useValue: dialogMock },
        {
          provide: ActivatedRoute,
          useValue: { data: of({ userManagerContext: routeContext }) },
        },
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
        user: null,
        role: UserRole.TENANT_ADMIN,
      },
    });
  });

  it('should create user with dialog tenantId when provided', () => {
    fixture.detectChanges();

    testApi.onCreateUser();
    afterClosedSubject.next({
      email: 'new@user.com',
      username: 'newuser',
      tenantId: 'custom-tenant',
    });

    expect(userServiceMock.addNewUser).toHaveBeenCalledWith(
      { email: 'new@user.com', username: 'newuser' },
      UserRole.TENANT_ADMIN,
      'custom-tenant',
    );
    expect(userServiceMock.retrieveUser).toHaveBeenCalledWith(
      UserRole.TENANT_ADMIN,
      sessionTenantId,
    );
  });

  it('should create user with context tenantId when dialog provides none', () => {
    fixture.detectChanges();

    testApi.onCreateUser();
    afterClosedSubject.next({ email: 'new@user.com', username: 'newuser' });

    expect(userServiceMock.addNewUser).toHaveBeenCalledWith(
      { email: 'new@user.com', username: 'newuser' },
      UserRole.TENANT_ADMIN,
      sessionTenantId,
    );
    expect(userServiceMock.retrieveUser).toHaveBeenCalledWith(
      UserRole.TENANT_ADMIN,
      sessionTenantId,
    );
  });

  it('should not create user when create dialog is cancelled', () => {
    fixture.detectChanges();

    testApi.onCreateUser();
    afterClosedSubject.next(null);

    expect(userServiceMock.addNewUser).not.toHaveBeenCalled();
  });

  it('should open delete dialog with correct config', () => {
    const user: User = {
      id: '1',
      email: 'delete@user.com',
      username: 'deleteuser',
      role: UserRole.TENANT_ADMIN,
      tenantId: sessionTenantId,
    };

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
    fixture.detectChanges();
    const user: User = {
      id: '1',
      email: 'delete@user.com',
      username: 'deleteuser',
      role: UserRole.TENANT_ADMIN,
      tenantId: sessionTenantId,
    };

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
