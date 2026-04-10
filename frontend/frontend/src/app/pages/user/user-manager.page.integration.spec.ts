import { ComponentFixture, TestBed } from '@angular/core/testing';
import { WritableSignal, signal } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { MatDialog } from '@angular/material/dialog';
import { MatSnackBar } from '@angular/material/snack-bar';
import { By } from '@angular/platform-browser';
import { of, Subject } from 'rxjs';
import { describe, expect, it, vi } from 'vitest';

import { UserManagerPage } from './user-manager.page';
import { UserTableComponent } from './components/user-table/user-table.component';
import { UserService } from '../../services/user/user.service';
import { UserSessionService } from '../../services/user-session/user-session.service';
import { TenantService } from '../../services/tenant/tenant.service';
import { User } from '../../models/user/user.model';
import { UserRole } from '../../models/user/user-role.enum';
import { UserSession } from '../../models/auth/user-session.model';

const mockUsers: User[] = [
  {
    id: '1',
    email: 'alice@test.com',
    username: 'alice',
    role: UserRole.TENANT_ADMIN,
    tenantId: 'tenant-1',
  },
  {
    id: '2',
    email: 'bob@test.com',
    username: 'bob',
    role: UserRole.TENANT_ADMIN,
    tenantId: 'tenant-1',
  },
  {
    id: '3',
    email: 'charlie@test.com',
    username: 'charlie',
    role: UserRole.TENANT_ADMIN,
    tenantId: 'tenant-1',
  },
];

const tenantAdminSession: UserSession = {
  userId: '1',
  role: UserRole.TENANT_ADMIN,
  tenantId: 'tenant-1',
};
const superAdminSession: UserSession = { userId: '1', role: UserRole.SUPER_ADMIN };
const tenantAdminContext = { title: 'Gestione Tenant Admin', role: UserRole.TENANT_ADMIN };
const tenantUserContext = { title: 'Gestione Tenant User', role: UserRole.TENANT_USER };

function createUserServiceMock() {
  return {
    userList: signal<User[]>([]),
    total: signal(0),
    pageIndex: signal(0),
    limit: signal(10),
    loading: signal(false),
    error: signal<string | null>(null),
    retrieveUsers: vi.fn(),
    changePage: vi.fn(),
    removeUser: vi.fn().mockReturnValue(of(void 0)),
  };
}

function setupTestBed(options: {
  session: UserSession;
  routeContext: { title: string; role: UserRole };
  queryParams?: Record<string, string>;
  canImpersonate?: boolean;
}) {
  const afterClosedSubject = new Subject<unknown>();
  const userServiceMock = createUserServiceMock();
  const dialogMock = {
    open: vi.fn().mockReturnValue({ afterClosed: () => afterClosedSubject.asObservable() }),
  };
  const snackBarMock = { open: vi.fn() };
  const routerMock = { navigate: vi.fn() };
  const tenantServiceMock = {
    getTenant: vi.fn().mockReturnValue(of({ id: 'mock', name: 'Mock', canImpersonate: options.canImpersonate ?? true })),
    getAllTenants: vi.fn().mockReturnValue(of([])),
  };

  TestBed.configureTestingModule({
    imports: [UserManagerPage, UserTableComponent],
    providers: [
      { provide: UserService, useValue: userServiceMock },
      { provide: UserSessionService, useValue: { currentUser: signal(options.session) } },
      {
        provide: ActivatedRoute,
        useValue: {
          data: of({ userManagerContext: options.routeContext }),
          queryParams: of(options.queryParams ?? {}),
        },
      },
      { provide: Router, useValue: routerMock },
      { provide: TenantService, useValue: tenantServiceMock },
    ],
  })
    .overrideProvider(MatDialog, { useValue: dialogMock })
    .overrideProvider(MatSnackBar, { useValue: snackBarMock });

  const fixture = TestBed.createComponent(UserManagerPage);
  return { fixture, userServiceMock, dialogMock, snackBarMock, routerMock, afterClosedSubject, tenantServiceMock };
}

function getTable(fixture: ComponentFixture<UserManagerPage>) {
  return fixture.debugElement.query(By.directive(UserTableComponent));
}

function getTableRows(fixture: ComponentFixture<UserManagerPage>): HTMLElement[] {
  return fixture.nativeElement.querySelectorAll('mat-row');
}

function getHeaderCells(fixture: ComponentFixture<UserManagerPage>): string[] {
  return Array.from<HTMLElement>(fixture.nativeElement.querySelectorAll('mat-header-cell')).map(
    (h) => h.textContent?.trim() ?? '',
  );
}

function getDeleteButtons(fixture: ComponentFixture<UserManagerPage>): HTMLButtonElement[] {
  return Array.from(fixture.nativeElement.querySelectorAll('mat-cell button[color="warn"]'));
}

describe('UserManagerPage (Integration)', () => {
  afterEach(() => {
    TestBed.resetTestingModule();
    vi.resetAllMocks();
  });

  describe('Page -> Table: Input Bindings', () => {
    it('should render users and display correct data in table cells', () => {
      const { fixture, userServiceMock } = setupTestBed({
        session: tenantAdminSession,
        routeContext: tenantAdminContext,
      });

      (userServiceMock.userList as WritableSignal<User[]>).set(mockUsers);
      fixture.detectChanges();

      expect(getTableRows(fixture).length).toBe(3);

      (userServiceMock.userList as WritableSignal<User[]>).set([mockUsers[0]]);
      fixture.detectChanges();

      const cellTexts = Array.from<HTMLElement>(
        fixture.nativeElement.querySelectorAll('mat-row mat-cell'),
      ).map((c) => c.textContent?.trim());
      expect(cellTexts).toContain('alice');
      expect(cellTexts).toContain('alice@test.com');
      expect(cellTexts).toContain('tenant-1');
    });

    it('should show spinner when loading and empty state when idle with no users', () => {
      const { fixture, userServiceMock } = setupTestBed({
        session: tenantAdminSession,
        routeContext: tenantAdminContext,
      });

      (userServiceMock.loading as WritableSignal<boolean>).set(true);
      fixture.detectChanges();

      expect(fixture.nativeElement.querySelector('mat-spinner')).toBeTruthy();
      expect(getTableRows(fixture).length).toBe(0);

      (userServiceMock.loading as WritableSignal<boolean>).set(false);
      fixture.detectChanges();

      const emptyState = fixture.nativeElement.querySelector('.empty-state');
      expect(emptyState).toBeTruthy();
      expect(emptyState.textContent).toContain('Nessun utente disponibile');
    });

    it.each([
      ['show', tenantAdminSession, tenantAdminContext, true],
      [
        'hide',
        superAdminSession,
        { title: 'Gestione Super Admin', role: UserRole.SUPER_ADMIN },
        false,
      ],
    ] as const)(
      'should %s tenantId column based on role',
      (_label, session, routeContext, visible) => {
        const { fixture, userServiceMock } = setupTestBed({ session, routeContext });
        (userServiceMock.userList as WritableSignal<User[]>).set(mockUsers);
        fixture.detectChanges();

        const headers = getHeaderCells(fixture);
        if (visible) {
          expect(headers).toContain('Tenant ID');
        } else {
          expect(headers).not.toContain('Tenant ID');
        }
      },
    );

    it('should pass pagination signals to table and render paginator', () => {
      const { fixture, userServiceMock } = setupTestBed({
        session: tenantAdminSession,
        routeContext: tenantAdminContext,
      });

      (userServiceMock.userList as WritableSignal<User[]>).set(mockUsers);
      (userServiceMock.total as WritableSignal<number>).set(50);
      (userServiceMock.pageIndex as WritableSignal<number>).set(0);
      (userServiceMock.limit as WritableSignal<number>).set(10);
      fixture.detectChanges();

      const table = getTable(fixture);
      expect(table.componentInstance.total()).toBe(50);
      expect(table.componentInstance.pageIndex()).toBe(0);
      expect(table.componentInstance.limit()).toBe(10);
      expect(fixture.nativeElement.querySelector('mat-paginator')).toBeTruthy();
    });
  });

  describe('Table -> Page: Output Events', () => {
    it('should open delete dialog with correct user data when delete button is clicked', () => {
      const { fixture, userServiceMock, dialogMock } = setupTestBed({
        session: tenantAdminSession,
        routeContext: tenantAdminContext,
      });

      (userServiceMock.userList as WritableSignal<User[]>).set([mockUsers[1]]);
      fixture.detectChanges();

      getDeleteButtons(fixture)[0].click();
      fixture.detectChanges();

      expect(dialogMock.open).toHaveBeenCalledWith(
        expect.anything(),
        expect.objectContaining({
          data: {
            title: 'Elimina Utente',
            message: 'Sei sicuro di voler eliminare "bob@test.com"?',
          },
        }),
      );
    });

    it('should call changePage when paginator emits pageChange', () => {
      const { fixture, userServiceMock } = setupTestBed({
        session: tenantAdminSession,
        routeContext: tenantAdminContext,
      });

      (userServiceMock.userList as WritableSignal<User[]>).set(mockUsers);
      (userServiceMock.total as WritableSignal<number>).set(50);
      fixture.detectChanges();

      getTable(fixture).componentInstance.pageChange.emit({
        pageIndex: 1,
        pageSize: 10,
        length: 50,
      });

      expect(userServiceMock.changePage).toHaveBeenCalledWith(
        1,
        10,
        UserRole.TENANT_ADMIN,
        'tenant-1',
      );
    });
  });

  describe('Full Delete Flow', () => {
    it('should call removeUser for the correct user, refresh, and show snackbar when confirmed', () => {
      const { fixture, userServiceMock, dialogMock, snackBarMock, afterClosedSubject } =
        setupTestBed({
          session: tenantAdminSession,
          routeContext: tenantAdminContext,
        });

      (userServiceMock.userList as WritableSignal<User[]>).set(mockUsers);
      fixture.detectChanges();
      userServiceMock.retrieveUsers.mockClear();

      getDeleteButtons(fixture)[0].click();
      fixture.detectChanges();

      expect(dialogMock.open).toHaveBeenCalledWith(
        expect.anything(),
        expect.objectContaining({
          data: {
            title: 'Elimina Utente',
            message: 'Sei sicuro di voler eliminare "bob@test.com"?',
          },
        }),
      );

      afterClosedSubject.next(true);

      expect(userServiceMock.removeUser).toHaveBeenCalledWith(mockUsers[1]);
      expect(userServiceMock.retrieveUsers).toHaveBeenCalledWith(UserRole.TENANT_ADMIN, 'tenant-1');
      expect(snackBarMock.open).toHaveBeenCalledWith('Utente eliminato con successo', 'Chiudi', {
        duration: 3000,
      });
    });

    it('should NOT call removeUser when dialog is cancelled', () => {
      const { fixture, userServiceMock, afterClosedSubject } = setupTestBed({
        session: tenantAdminSession,
        routeContext: tenantAdminContext,
      });

      (userServiceMock.userList as WritableSignal<User[]>).set([mockUsers[1]]);
      fixture.detectChanges();
      userServiceMock.retrieveUsers.mockClear();

      getDeleteButtons(fixture)[0].click();
      fixture.detectChanges();
      afterClosedSubject.next(false);

      expect(userServiceMock.removeUser).not.toHaveBeenCalled();
      expect(userServiceMock.retrieveUsers).not.toHaveBeenCalled();
    });
  });

  describe('Full Create Flow', () => {
    it('should refresh table and show snackbar after successful creation', () => {
      const { fixture, userServiceMock, snackBarMock, afterClosedSubject } = setupTestBed({
        session: tenantAdminSession,
        routeContext: tenantAdminContext,
      });

      (userServiceMock.userList as WritableSignal<User[]>).set(mockUsers);
      fixture.detectChanges();
      userServiceMock.retrieveUsers.mockClear();

      fixture.nativeElement.querySelector('button[mat-raised-button]').click();
      fixture.detectChanges();
      afterClosedSubject.next(true);

      expect(userServiceMock.retrieveUsers).toHaveBeenCalledWith(UserRole.TENANT_ADMIN, 'tenant-1');
      expect(snackBarMock.open).toHaveBeenCalledWith('Utente creato con successo', 'Chiudi', {
        duration: 3000,
      });
    });
  });

  describe('Template: Conditional Rendering', () => {
    it('should hide table, show warning, and disable create button for TENANT_USER without tenantId', () => {
      const { fixture } = setupTestBed({
        session: superAdminSession,
        routeContext: tenantUserContext,
        queryParams: {},
      });
      fixture.detectChanges();

      expect(getTable(fixture)).toBeFalsy();
      expect(fixture.nativeElement.querySelector('.warning-banner')).toBeTruthy();
      expect(fixture.nativeElement.querySelector('button[mat-raised-button]').disabled).toBe(true);
    });

    it('should hide table, show warning, and disable create button for SUPER_ADMIN with TENANT_ADMIN context and no tenantId', () => {
      const { fixture } = setupTestBed({
        session: superAdminSession,
        routeContext: tenantAdminContext,
        queryParams: {},
      });
      fixture.detectChanges();

      expect(getTable(fixture)).toBeFalsy();
      expect(fixture.nativeElement.querySelector('.warning-banner')).toBeTruthy();
      expect(fixture.nativeElement.querySelector('button[mat-raised-button]').disabled).toBe(true);
    });

    it('should show table for TENANT_USER when tenantId is provided', () => {
      const { fixture, userServiceMock } = setupTestBed({
        session: superAdminSession,
        routeContext: tenantUserContext,
        queryParams: { tenantId: 'tenant-1' },
      });
      (userServiceMock.userList as WritableSignal<User[]>).set(mockUsers);
      fixture.detectChanges();

      expect(getTable(fixture)).toBeTruthy();
    });

    it('should show table for TENANT_ADMIN when tenantId is provided via URL', () => {
      const { fixture, userServiceMock } = setupTestBed({
        session: superAdminSession,
        routeContext: tenantAdminContext,
        queryParams: { tenantId: 'tenant-1' },
      });
      (userServiceMock.userList as WritableSignal<User[]>).set(mockUsers);
      fixture.detectChanges();

      expect(getTable(fixture)).toBeTruthy();
    });

    it('should show tenant banner with 2 buttons for SUPER_ADMIN viewing TENANT_USER', () => {
      const { fixture } = setupTestBed({
        session: superAdminSession,
        routeContext: tenantUserContext,
        queryParams: { tenantId: 'tenant-xyz' },
      });
      fixture.detectChanges();

      const banner = fixture.nativeElement.querySelector('.tenant-banner');
      expect(banner).toBeTruthy();
      expect(banner.textContent).toContain('tenant-xyz');
      expect(banner.querySelectorAll('button').length).toBe(2);
    });

    it('should NOT show tenant banner for SUPER_ADMIN viewing TENANT_ADMIN', () => {
      const { fixture } = setupTestBed({
        session: superAdminSession,
        routeContext: tenantAdminContext,
        queryParams: { tenantId: 'tenant-xyz' },
      });
      fixture.detectChanges();

      expect(fixture.nativeElement.querySelector('.tenant-banner')).toBeFalsy();
    });

    it('should NOT show tenant banner for non-SUPER_ADMIN session', () => {
      const { fixture } = setupTestBed({
        session: tenantAdminSession,
        routeContext: tenantAdminContext,
      });
      fixture.detectChanges();

      expect(fixture.nativeElement.querySelector('.tenant-banner')).toBeFalsy();
    });

    it('should show error banner when error signal has value', () => {
      const { fixture, userServiceMock } = setupTestBed({
        session: tenantAdminSession,
        routeContext: tenantAdminContext,
      });

      (userServiceMock.error as WritableSignal<string | null>).set('Failed to load users');
      fixture.detectChanges();

      const errorBanner = fixture.nativeElement.querySelector('.error-banner');
      expect(errorBanner).toBeTruthy();
      expect(errorBanner.textContent).toContain('Failed to load users');
    });

    it('should show tenant dropdown for SUPER_ADMIN with TENANT_ADMIN context', () => {
      const { fixture } = setupTestBed({
        session: superAdminSession,
        routeContext: tenantAdminContext,
        queryParams: {},
      });
      fixture.detectChanges();

      expect(fixture.nativeElement.querySelector('.tenant-select-bar')).toBeTruthy();
      expect(fixture.nativeElement.querySelector('mat-select')).toBeTruthy();
    });

    it('should NOT show tenant dropdown for SUPER_ADMIN with TENANT_USER context', () => {
      const { fixture } = setupTestBed({
        session: superAdminSession,
        routeContext: tenantUserContext,
        queryParams: {},
      });
      fixture.detectChanges();

      expect(fixture.nativeElement.querySelector('.tenant-select-bar')).toBeFalsy();
    });

    it('should NOT show tenant dropdown for TENANT_ADMIN session', () => {
      const { fixture } = setupTestBed({
        session: tenantAdminSession,
        routeContext: tenantAdminContext,
      });
      fixture.detectChanges();

      expect(fixture.nativeElement.querySelector('.tenant-select-bar')).toBeFalsy();
    });

    it('should show tab group for TENANT_ADMIN session', () => {
      const { fixture } = setupTestBed({
        session: tenantAdminSession,
        routeContext: tenantAdminContext,
      });
      fixture.detectChanges();

      expect(fixture.nativeElement.querySelector('mat-tab-group')).toBeTruthy();
    });

    it('should NOT show tab group for SUPER_ADMIN session', () => {
      const { fixture } = setupTestBed({
        session: superAdminSession,
        routeContext: tenantAdminContext,
        queryParams: { tenantId: 'tenant-1' },
      });
      fixture.detectChanges();

      expect(fixture.nativeElement.querySelector('mat-tab-group')).toBeFalsy();
    });

    it('should hide the delete button for the current user\'s row', () => {
      const { fixture, userServiceMock } = setupTestBed({
        session: tenantAdminSession,
        routeContext: tenantAdminContext,
      });
      // mockUsers[0] has id='1' which matches tenantAdminSession.userId='1'
      (userServiceMock.userList as WritableSignal<User[]>).set(mockUsers);
      fixture.detectChanges();

      const deleteButtons = getDeleteButtons(fixture);
      // alice (id='1') has no button; bob and charlie do
      expect(deleteButtons.length).toBe(mockUsers.length - 1);
    });

    describe('Navigation', () => {
      it.each([
        [
          'dashboard with tenantId from dashboard button',
          tenantUserContext,
          { tenantId: 'tenant-123' },
          '.tenant-banner .banner-actions button:first-child',
          [['/dashboard'], { queryParams: { tenantId: 'tenant-123' } }],
        ],
        [
          'tenant-management from warning banner',
          tenantUserContext,
          {},
          '.warning-banner button',
          [['/tenant-management']],
        ],
      ] as const)(
        'should navigate to %s',
        (_label, routeContext, queryParams, selector, expectedArgs) => {
          const { fixture, routerMock } = setupTestBed({
            session: superAdminSession,
            routeContext,
            queryParams,
          });
          fixture.detectChanges();

          fixture.nativeElement.querySelector(selector).click();

          expect(routerMock.navigate).toHaveBeenCalledWith(...expectedArgs);
        },
      );

      it('should navigate to /user-management/tenant-users when tenant canImpersonate is false', () => {
        const { fixture, routerMock } = setupTestBed({
          session: superAdminSession,
          routeContext: tenantAdminContext,
          queryParams: { tenantId: 'restricted-tenant' },
          canImpersonate: false,
        });
        fixture.detectChanges();

        expect(routerMock.navigate).toHaveBeenCalledWith(['/user-management/tenant-users']);
      });
    });
  });
});
