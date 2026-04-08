import { ComponentFixture, TestBed } from '@angular/core/testing';
import { signal } from '@angular/core';
import { provideRouter } from '@angular/router';
import { MatDialog } from '@angular/material/dialog';
import { MatSnackBar } from '@angular/material/snack-bar';
import { By } from '@angular/platform-browser';
import { of, EMPTY, Subject } from 'rxjs';
import { describe, expect, it, vi, afterEach } from 'vitest';
import { TestbedHarnessEnvironment } from '@angular/cdk/testing/testbed';
import { MatMenuHarness } from '@angular/material/menu/testing';

import { AppShellPage } from './app-shell.page';
import { HeaderComponent } from './components/header/header.component';
import { SideBarComponent } from './components/side-bar/side-bar.component';
import { ChangePasswordDialog } from './dialogs/change-password/change-password.dialog';
import { UserSessionService } from '../../services/user-session/user-session.service';
import { AuthSessionService } from '../../services/auth/auth-session.service';
import { UserService } from '../../services/user/user.service';
import { TenantService } from '../../services/tenant/tenant.service';
import { PermissionService } from '../../services/permission/permission.service';
import { UserRole } from '../../models/user/user-role.enum';
import { UserSession } from '../../models/auth/user-session.model';
import { User } from '../../models/user/user.model';
import { Tenant } from '../../models/tenant/tenant.model';
import { Permission } from '../../models/permission.enum';

const mockUser: User = {
  id: '1',
  username: 'admin',
  email: 'admin@test.com',
  role: UserRole.SUPER_ADMIN,
  tenantId: 'tenant-1',
};

const mockTenant: Tenant = { id: 'tenant-1', name: 'Tenant Alpha', canImpersonate: false };

const superAdminSession: UserSession = { userId: '1', role: UserRole.SUPER_ADMIN, tenantId: 'tenant-1' };
const tenantAdminSession: UserSession = { userId: '2', role: UserRole.TENANT_ADMIN, tenantId: 'tenant-1' };

function setupTestBed(options?: { session?: UserSession; user?: User; userResult$?: typeof EMPTY; tenant?: Tenant | null; canAny?: boolean | ((perms: Permission[]) => boolean) }) {
  const session = options?.session ?? superAdminSession;
  const user = options?.user ?? mockUser;
  const tenant = options?.tenant !== undefined ? options.tenant : mockTenant;
  const canAnyFn = typeof options?.canAny === 'function' ? options.canAny : () => (options?.canAny ?? true);

  const afterClosedSubject = new Subject<unknown>();
  const dialogMock = { open: vi.fn().mockReturnValue({ afterClosed: () => afterClosedSubject.asObservable() }) };
  const snackBarMock = { open: vi.fn() };
  const authSessionMock = { logout: vi.fn() };
  const userServiceMock = { getUser: vi.fn().mockReturnValue(options?.userResult$ ?? of(user)) };
  const tenantServiceMock = { getTenant: vi.fn().mockReturnValue(tenant ? of(tenant) : EMPTY) };
  const permissionServiceMock = { canAny: vi.fn().mockImplementation(canAnyFn) };

  TestBed.configureTestingModule({
    imports: [AppShellPage, HeaderComponent, SideBarComponent],
    providers: [
      provideRouter([]),
      { provide: UserSessionService, useValue: { currentUser: signal(session) } },
      { provide: AuthSessionService, useValue: authSessionMock },
      { provide: UserService, useValue: userServiceMock },
      { provide: TenantService, useValue: tenantServiceMock },
      { provide: PermissionService, useValue: permissionServiceMock },
    ],
  })
    .overrideProvider(MatDialog, { useValue: dialogMock })
    .overrideProvider(MatSnackBar, { useValue: snackBarMock });

  const fixture = TestBed.createComponent(AppShellPage);
  return { fixture, dialogMock, snackBarMock, authSessionMock, userServiceMock, tenantServiceMock, permissionServiceMock, afterClosedSubject };
}

function getSidebar(fixture: ComponentFixture<AppShellPage>) {
  return fixture.debugElement.query(By.directive(SideBarComponent));
}

function getHeader(fixture: ComponentFixture<AppShellPage>) {
  return fixture.debugElement.query(By.directive(HeaderComponent));
}

describe('AppShellPage (Integration)', () => {
  afterEach(() => {
    TestBed.resetTestingModule();
    vi.resetAllMocks();
  });

  describe('Initialization', () => {
    it('should render shell layout with sidebar, header, main, and router-outlet', () => {
      const { fixture } = setupTestBed();
      fixture.detectChanges();

      expect(fixture.debugElement.query(By.css('.shell-layout'))).toBeTruthy();
      expect(fixture.debugElement.query(By.css('.main-content'))).toBeTruthy();
      expect(fixture.debugElement.query(By.css('main'))).toBeTruthy();
      expect(getSidebar(fixture)).toBeTruthy();
      expect(getHeader(fixture)).toBeTruthy();
      expect(fixture.debugElement.query(By.css('router-outlet'))).toBeTruthy();
    });

    it('should call getUser and getTenant on init', () => {
      const { fixture, userServiceMock, tenantServiceMock } = setupTestBed();
      fixture.detectChanges();

      expect(userServiceMock.getUser).toHaveBeenCalledWith('1', UserRole.SUPER_ADMIN, 'tenant-1');
      expect(tenantServiceMock.getTenant).toHaveBeenCalledWith('tenant-1');
    });
  });

  describe('Page -> Header: Input Bindings', () => {
    it('should pass username, tenant name, and role to header', () => {
      const { fixture } = setupTestBed();
      fixture.detectChanges();

      const header = getHeader(fixture).componentInstance as HeaderComponent;
      expect(header.username()).toBe('admin');
      expect(header.currentTenant()).toBe('Tenant Alpha');
      expect(header.currentUserRole()).toBe(UserRole.SUPER_ADMIN);
    });

    it('should render tenant badge and role badge in header', () => {
      const { fixture } = setupTestBed();
      fixture.detectChanges();

      expect(fixture.nativeElement.querySelector('.tenant-badge').textContent).toContain('Tenant Alpha');
      expect(fixture.nativeElement.querySelector('.role-badge').textContent).toContain('SUPER ADMIN');
    });

    it('should pass null username when user service returns EMPTY', () => {
      const { fixture } = setupTestBed({ userResult$: EMPTY });
      fixture.detectChanges();

      const header = getHeader(fixture).componentInstance as HeaderComponent;
      expect(header.username()).toBeNull();
    });

    it('should not show tenant badge when session has no tenantId', () => {
      const noTenantSession: UserSession = { userId: '1', role: UserRole.SUPER_ADMIN };
      const { fixture } = setupTestBed({ session: noTenantSession, tenant: null });
      fixture.detectChanges();

      expect(fixture.nativeElement.querySelector('.tenant-badge')).toBeFalsy();
    });
  });

  describe('Page -> Sidebar: NavItems', () => {
    it('should render all 7 nav items when user has all permissions', () => {
      const { fixture } = setupTestBed();
      fixture.detectChanges();

      const navLinks = fixture.debugElement.queryAll(By.css('.nav-item'));
      expect(navLinks.length).toBe(7);
    });

    it('should filter nav items based on permissions', () => {
      const { fixture } = setupTestBed({
        canAny: (perms: Permission[]) => perms.includes(Permission.DASHBOARD_ACCESS),
      });
      fixture.detectChanges();

      const navLinks = fixture.debugElement.queryAll(By.css('.nav-item'));
      expect(navLinks.length).toBe(1);
      expect(navLinks[0].nativeElement.textContent).toContain('Dashboard');
    });

    it('should render no nav items when user has no permissions', () => {
      const { fixture } = setupTestBed({ canAny: false });
      fixture.detectChanges();

      expect(fixture.debugElement.queryAll(By.css('.nav-item')).length).toBe(0);
    });

    it('should render divider for Dashboard item when user is SUPER_ADMIN', () => {
      const { fixture } = setupTestBed();
      fixture.detectChanges();

      expect(fixture.nativeElement.querySelector('mat-divider')).toBeTruthy();
    });

    it('should not render divider for non-SUPER_ADMIN', () => {
      const tenantUser: User = { ...mockUser, id: '2', role: UserRole.TENANT_ADMIN };
      const { fixture } = setupTestBed({ session: tenantAdminSession, user: tenantUser });
      fixture.detectChanges();

      expect(fixture.nativeElement.querySelector('mat-divider')).toBeFalsy();
    });
  });

  describe('Header -> Page: User Menu Actions', () => {
    it('should emit logout and call authSessionService.logout', async () => {
      const { fixture, authSessionMock } = setupTestBed();
      fixture.detectChanges();

      const loader = TestbedHarnessEnvironment.loader(fixture);
      const menu = await loader.getHarness(MatMenuHarness);
      await menu.open();
      const items = await menu.getItems({ text: /Esci/i });
      await items[0].click();

      expect(authSessionMock.logout).toHaveBeenCalled();
    });

    it('should open ChangePasswordDialog from menu', async () => {
      const { fixture, dialogMock } = setupTestBed();
      fixture.detectChanges();

      const loader = TestbedHarnessEnvironment.loader(fixture);
      const menu = await loader.getHarness(MatMenuHarness);
      await menu.open();
      const items = await menu.getItems({ text: /Cambia Password/i });
      await items[0].click();

      expect(dialogMock.open).toHaveBeenCalledWith(ChangePasswordDialog, {
        width: '800px',
        disableClose: true,
      });
    });

    it('should display username in menu header', async () => {
      const { fixture } = setupTestBed();
      fixture.detectChanges();

      const loader = TestbedHarnessEnvironment.loader(fixture);
      const menu = await loader.getHarness(MatMenuHarness);
      await menu.open();

      const menuHeader = document.querySelector('.menu-header');
      expect(menuHeader).toBeTruthy();
      expect(menuHeader?.textContent).toContain('admin');
    });
  });

  describe('Change Password Flow', () => {
    it('should show snackbar when dialog closes with true', async () => {
      const { fixture, dialogMock, snackBarMock, afterClosedSubject } = setupTestBed();
      fixture.detectChanges();

      const loader = TestbedHarnessEnvironment.loader(fixture);
      const menu = await loader.getHarness(MatMenuHarness);
      await menu.open();
      const items = await menu.getItems({ text: /Cambia Password/i });
      await items[0].click();

      expect(dialogMock.open).toHaveBeenCalled();
      afterClosedSubject.next(true);

      expect(snackBarMock.open).toHaveBeenCalledWith('Password modificata con successo', 'Close', { duration: 3000 });
    });

    it('should not show snackbar when dialog closes with falsy', async () => {
      const { fixture, snackBarMock, afterClosedSubject } = setupTestBed();
      fixture.detectChanges();

      const loader = TestbedHarnessEnvironment.loader(fixture);
      const menu = await loader.getHarness(MatMenuHarness);
      await menu.open();
      const items = await menu.getItems({ text: /Cambia Password/i });
      await items[0].click();

      afterClosedSubject.next(false);
      expect(snackBarMock.open).not.toHaveBeenCalled();
    });
  });
});
