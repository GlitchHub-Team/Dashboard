import { ComponentFixture, TestBed } from '@angular/core/testing';
import { signal, WritableSignal, Component, input, output } from '@angular/core';
import { Router, RouterOutlet } from '@angular/router';
import { MatDialog } from '@angular/material/dialog';
import { By } from '@angular/platform-browser';

import { AppShellPage } from './app-shell.page';
import { HeaderComponent } from './components/header/header.component';
import { SideBarComponent } from './components/side-bar/side-bar.component';
import { UserSessionService } from '../../services/user-session/user-session.service';
import { AuthSessionService } from '../../services/auth/auth-session.service';
import { PermissionService } from '../../services/permission/permission.service';
import { ChangePasswordDialog } from './dialogs/change-password/change-password.dialog';
import { UserRole } from '../../models/user/user-role.enum';
import { User } from '../../models/user/user.model';
import { NavItem } from '../../models/nav_items/nav-item.model';
import { Permission } from '../../models/permission.enum';

@Component({ selector: 'app-header', template: '', standalone: true })
class StubHeader {
  username = input<string | null>();
  currentTenant = input<string | null>();
  currentUserRole = input<UserRole | null>();
  logoutRequested = output();
  changePasswordRequested = output();
}

@Component({ selector: 'app-side-bar', template: '', standalone: true })
class StubSideBar {
  navItems = input<NavItem[]>();
}

@Component({ selector: 'router-outlet', template: '', standalone: true })
class StubRouterOutlet {}

describe('AppShellPage (Unit)', () => {
  let component: AppShellPage;
  let fixture: ComponentFixture<AppShellPage>;
  let router: Router;

  const mockUser: User = {
    id: '1',
    username: 'admin',
    email: 'admin@test.com',
    role: UserRole.SUPER_ADMIN,
    tenantId: 'tenant-1',
  };

  let currentUserSignal: WritableSignal<User | null>;
  let currentRoleSignal: WritableSignal<UserRole | null>;
  let currentTenantSignal: WritableSignal<string | null>;

  const authSessionServiceMock = { logout: vi.fn() };
  const permissionServiceMock = { canAny: vi.fn() };
  const dialogMock = { open: vi.fn() };

  const rebuildComponent = () => {
    fixture = TestBed.createComponent(AppShellPage);
    component = fixture.componentInstance;
    fixture.detectChanges();
  };

  beforeEach(async () => {
    vi.resetAllMocks();
    permissionServiceMock.canAny.mockReturnValue(true);

    currentUserSignal = signal<User | null>(mockUser);
    currentRoleSignal = signal<UserRole | null>(UserRole.SUPER_ADMIN);
    currentTenantSignal = signal<string | null>('tenant-1');

    await TestBed.configureTestingModule({
      imports: [AppShellPage],
      providers: [
        {
          provide: UserSessionService,
          useValue: {
            currentUser: currentUserSignal.asReadonly(),
            currentRole: currentRoleSignal.asReadonly(),
            currentTenant: currentTenantSignal.asReadonly(),
          },
        },
        { provide: AuthSessionService, useValue: authSessionServiceMock },
        { provide: PermissionService, useValue: permissionServiceMock },
        { provide: MatDialog, useValue: dialogMock },
      ],
    })
      .overrideComponent(AppShellPage, {
        remove: { imports: [HeaderComponent, SideBarComponent, RouterOutlet] },
        add: { imports: [StubHeader, StubSideBar, StubRouterOutlet] },
      })
      .compileComponents();

    router = TestBed.inject(Router);
    vi.spyOn(router, 'navigate').mockResolvedValue(true);

    rebuildComponent();
  });

  describe('rendering', () => {
    it('should create the component', () => {
      expect(component).toBeTruthy();
    });

    it('should render shell layout structure', () => {
      expect(fixture.debugElement.query(By.css('.shell-layout'))).toBeTruthy();
      expect(fixture.debugElement.query(By.css('.main-content'))).toBeTruthy();
      expect(fixture.debugElement.query(By.css('main'))).toBeTruthy();
    });

    it('should render header, sidebar, and router outlet', () => {
      expect(fixture.debugElement.query(By.directive(StubHeader))).toBeTruthy();
      expect(fixture.debugElement.query(By.directive(StubSideBar))).toBeTruthy();
      expect(fixture.debugElement.query(By.directive(StubRouterOutlet))).toBeTruthy();
    });
  });

  describe('input bindings', () => {
    const getHeader = () =>
      fixture.debugElement.query(By.directive(StubHeader)).componentInstance as StubHeader;

    it('should pass user email, role, and tenant to header', () => {
      const h = getHeader();
      expect(h.username()).toBe('admin@test.com');
      expect(h.currentUserRole()).toBe(UserRole.SUPER_ADMIN);
      expect(h.currentTenant()).toBe('tenant-1');
    });

    it('should pass null username when user is null', () => {
      currentUserSignal.set(null);
      fixture.detectChanges();
      expect(getHeader().username()).toBeNull();
    });

    it('should pass filtered nav items to sidebar', () => {
      const sidebar = fixture.debugElement.query(By.directive(StubSideBar))
        .componentInstance as StubSideBar;
      expect(sidebar.navItems()!.length).toBeGreaterThan(0);
    });

    it('should update header bindings when user changes', () => {
      currentUserSignal.set({
        id: '2',
        username: 'newuser',
        email: 'new@test.com',
        role: UserRole.TENANT_USER,
        tenantId: 'tenant-2',
      });
      currentRoleSignal.set(UserRole.TENANT_USER);
      currentTenantSignal.set('tenant-2');
      fixture.detectChanges();

      const h = getHeader();
      expect(h.username()).toBe('new@test.com');
      expect(h.currentUserRole()).toBe(UserRole.TENANT_USER);
      expect(h.currentTenant()).toBe('tenant-2');
    });
  });

  describe('navItems', () => {
    it('should include all 6 items when user has all permissions', () => {
      expect(component['navItems']().length).toBe(6);
    });

    it('should return empty list when user has no permissions', () => {
      permissionServiceMock.canAny.mockReturnValue(false);
      rebuildComponent();
      expect(component['navItems']()).toHaveLength(0);
    });

    it('should filter items based on specific permissions', () => {
      permissionServiceMock.canAny.mockImplementation((permissions: Permission[]) =>
        permissions.includes(Permission.DASHBOARD_ACCESS),
      );
      rebuildComponent();

      const items = component['navItems']();
      expect(items).toHaveLength(1);
      expect(items[0].label).toBe('Dashboard');
    });

    it('should call canAny with an array for each gated item', () => {
      component['navItems']();

      expect(permissionServiceMock.canAny).toHaveBeenCalledTimes(6);
      permissionServiceMock.canAny.mock.calls.forEach((call: unknown[][]) => {
        expect(Array.isArray(call[0])).toBe(true);
      });
    });

    it('should pass filtered items to sidebar', () => {
      permissionServiceMock.canAny.mockImplementation((permissions: Permission[]) =>
        permissions.includes(Permission.DASHBOARD_ACCESS),
      );
      rebuildComponent();

      const sidebar = fixture.debugElement.query(By.directive(StubSideBar))
        .componentInstance as StubSideBar;
      const items = sidebar.navItems()!;
      expect(items).toHaveLength(1);
      expect(items[0].label).toBe('Dashboard');
    });

    it('should show only Dashboard for TENANT_USER', () => {
      permissionServiceMock.canAny.mockImplementation((permissions: Permission[]) =>
        permissions.some((p) => [Permission.DASHBOARD_ACCESS].includes(p)),
      );
      rebuildComponent();

      const items = component['navItems']();
      expect(items).toHaveLength(1);
      expect(items[0].route).toBe('/dashboard');
    });

    it('should show Dashboard and Tenant User Management for TENANT_ADMIN', () => {
      const tenantAdminPerms = [
        Permission.DASHBOARD_ACCESS,
        Permission.GATEWAY_COMMANDS,
        Permission.TENANT_USER_MANAGEMENT,
      ];
      permissionServiceMock.canAny.mockImplementation((permissions: Permission[]) =>
        permissions.some((p) => tenantAdminPerms.includes(p)),
      );
      rebuildComponent();

      const items = component['navItems']();
      expect(items).toHaveLength(2);
      expect(items.map((i) => i.label)).toEqual(['Dashboard', 'Tenant User Management']);
    });
  });

  describe('onLogout', () => {
    const triggerLogout = (f: ComponentFixture<AppShellPage>) =>
      f.debugElement.query(By.directive(StubHeader)).triggerEventHandler('logoutRequested');

    it('should call logout and navigate to /login', () => {
      triggerLogout(fixture);
      expect(authSessionServiceMock.logout).toHaveBeenCalledOnce();
      expect(router.navigate).toHaveBeenCalledWith(['/login']);
    });

    it('should call logout before navigation', () => {
      const callOrder: string[] = [];
      authSessionServiceMock.logout.mockImplementation(() => callOrder.push('logout'));
      vi.spyOn(router, 'navigate').mockImplementation(() => {
        callOrder.push('navigate');
        return Promise.resolve(true);
      });

      triggerLogout(fixture);
      expect(callOrder).toEqual(['logout', 'navigate']);
    });
  });

  describe('onChangePassword', () => {
    it('should open ChangePasswordDialog with correct config', () => {
      fixture.debugElement
        .query(By.directive(StubHeader))
        .triggerEventHandler('changePasswordRequested');

      expect(dialogMock.open).toHaveBeenCalledWith(ChangePasswordDialog, {
        width: '800px',
        disableClose: true,
      });
    });
  });
});
