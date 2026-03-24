import { ComponentFixture, TestBed } from '@angular/core/testing';
import { signal } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { MatDialog } from '@angular/material/dialog';
import { PageEvent } from '@angular/material/paginator';
import { of, Subject } from 'rxjs';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { UserManagerPage } from './user-manager.page';
import { UserService } from '../../services/user/user.service';
import { User } from '../../models/user.model';
import { UserRole } from '../../models/user-role.enum';
import { UserFormDialogComponent } from './dialogs/user-form.dialog';
import { ConfirmDeleteDialog } from '../tenant/dialogs/confirm-delete.dialog';

describe('UserManagerPage', () => {
  let component: UserManagerPage;
  let fixture: ComponentFixture<UserManagerPage>;

  let afterClosedSubject: Subject<unknown>;
  let dialogMock: { open: ReturnType<typeof vi.fn> };

  const routeContext = {
    title: 'Test Users',
    role: UserRole.TENANT_ADMIN,
    tenantId: 'tenant-1',
  };

  const userServiceMock = {
    userList: signal<User[]>([]),
    total: signal(0),
    pageIndex: signal(0),
    limit: signal(10),
    loading: signal(false),
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
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should initialize context and retrieve users on init', () => {
    fixture.detectChanges();

    expect(component.context).toEqual(routeContext);
    expect(userServiceMock.retrieveUser).toHaveBeenCalledWith(UserRole.TENANT_ADMIN, 'tenant-1');
  });

  it('should open create dialog with correct config', () => {
    component.onCreateUser();

    expect(dialogMock.open).toHaveBeenCalledWith(UserFormDialogComponent, {
      width: '400px',
      data: {
        user: null,
        role: UserRole.TENANT_ADMIN,
      },
    });
  });

  it('should create and refetch users when create dialog returns data', () => {
    fixture.detectChanges();

    component.onCreateUser();
    afterClosedSubject.next({ email: 'new@user.com' });

    expect(userServiceMock.addNewUser).toHaveBeenCalledWith(
      { email: 'new@user.com', role: UserRole.TENANT_ADMIN },
      'tenant-1',
    );
    expect(userServiceMock.retrieveUser).toHaveBeenCalledWith(UserRole.TENANT_ADMIN, 'tenant-1');
  });

  it('should not create user when create dialog is cancelled', () => {
    fixture.detectChanges();

    component.onCreateUser();
    afterClosedSubject.next(null);

    expect(userServiceMock.addNewUser).not.toHaveBeenCalled();
  });

  it('should open delete dialog with correct config', () => {
    const user: User = {
      id: '1',
      email: 'delete@user.com',
      username: 'deleteuser',
      role: UserRole.TENANT_ADMIN,
      tenantId: 'tenant-1',
    };

    component.onDeleteUser(user);

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
      tenantId: 'tenant-1',
    };

    component.onDeleteUser(user);
    afterClosedSubject.next(confirmed);

    if (shouldDelete) {
      expect(userServiceMock.removeUser).toHaveBeenCalledWith(user);
      expect(userServiceMock.retrieveUser).toHaveBeenCalledWith(UserRole.TENANT_ADMIN, 'tenant-1');
      return;
    }

    expect(userServiceMock.removeUser).not.toHaveBeenCalled();
  });

  it('should call changePage with route context', () => {
    fixture.detectChanges();
    const event: PageEvent = { pageIndex: 2, pageSize: 25, length: 100 };

    component.onPageChange(event);

    expect(userServiceMock.changePage).toHaveBeenCalledWith(
      2,
      25,
      UserRole.TENANT_ADMIN,
      'tenant-1',
    );
  });
});
