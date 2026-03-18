import { ComponentFixture, TestBed } from '@angular/core/testing';
import { NO_ERRORS_SCHEMA, signal, WritableSignal } from '@angular/core';
import { Router, RouterModule } from '@angular/router';
import { MatDialog } from '@angular/material/dialog';
import { By } from '@angular/platform-browser';

import { AppShellPage } from './app-shell.page';
import { UserSessionService } from '../../services/user-session/user-session.service';
import { AuthSessionService } from '../../services/auth/auth-session.service';
import { PermissionService } from '../../services/permission/permission.service';
import { ChangePasswordDialog } from './dialogs/change-password/change-password.dialog';
import { UserRole } from '../../models/user-role.enum';
import { User } from '../../models/user.model';

describe('AppShellPage', () => {
  let component: AppShellPage;
  let fixture: ComponentFixture<AppShellPage>;
  // ESLint whining
  let headerDebug: any;
  let sideBarDebug: any;
  let router: Router;

  const mockUser: User = {
    id: '1',
    email: 'admin@test.com',
    role: UserRole.SUPER_ADMIN,
    tenantId: 'tenant-1',
  };

  let currentUserSignal: WritableSignal<User | null>;
  let currentRoleSignal: WritableSignal<UserRole | null>;
  let currentTenantSignal: WritableSignal<string | null>;

  let userSessionMock: {
    currentUser: ReturnType<WritableSignal<User | null>['asReadonly']>;
    currentRole: ReturnType<WritableSignal<UserRole | null>['asReadonly']>;
    currentTenant: ReturnType<WritableSignal<string | null>['asReadonly']>;
  };

  const authSessionServiceMock = {
    logout: vi.fn(),
  };

  const permissionServiceMock = {
    canAny: vi.fn().mockReturnValue(true),
  };

  const dialogMock = {
    open: vi.fn(),
  };

  beforeEach(async () => {
    vi.resetAllMocks();
    permissionServiceMock.canAny.mockReturnValue(true);

    currentUserSignal = signal<User | null>(mockUser);
    currentRoleSignal = signal<UserRole | null>(UserRole.SUPER_ADMIN);
    currentTenantSignal = signal<string | null>('tenant-1');

    userSessionMock = {
      currentUser: currentUserSignal.asReadonly(),
      currentRole: currentRoleSignal.asReadonly(),
      currentTenant: currentTenantSignal.asReadonly(),
    };

    await TestBed.configureTestingModule({
      imports: [AppShellPage, RouterModule.forRoot([])],
      providers: [
        { provide: UserSessionService, useValue: userSessionMock },
        { provide: AuthSessionService, useValue: authSessionServiceMock },
        { provide: PermissionService, useValue: permissionServiceMock },
        { provide: MatDialog, useValue: dialogMock },
      ],
      schemas: [NO_ERRORS_SCHEMA],
    }).compileComponents();

    router = TestBed.inject(Router);
    vi.spyOn(router, 'navigate').mockResolvedValue(true);

    fixture = TestBed.createComponent(AppShellPage);
    component = fixture.componentInstance;
    fixture.detectChanges();

    headerDebug = fixture.debugElement.query(By.css('app-header'));
    sideBarDebug = fixture.debugElement.query(By.css('app-side-bar'));
  });

  describe('initial state', () => {
    it('should create', () => {
      expect(component).toBeTruthy();
    });

    it('should render the shell layout', () => {
      const layout = fixture.debugElement.query(By.css('.shell-layout'));
      expect(layout).toBeTruthy();
    });

    it('should render the main content area', () => {
      const mainContent = fixture.debugElement.query(By.css('.main-content'));
      expect(mainContent).toBeTruthy();
    });

    it('should render the main element', () => {
      const main = fixture.debugElement.query(By.css('main'));
      expect(main).toBeTruthy();
    });

    it('should render the router outlet', () => {
      const routerOutlet = fixture.debugElement.query(By.css('router-outlet'));
      expect(routerOutlet).toBeTruthy();
    });

    it('should render the side bar', () => {
      expect(sideBarDebug).toBeTruthy();
    });

    it('should render the header', () => {
      expect(headerDebug).toBeTruthy();
    });

    it('should expose current user', () => {
      expect(component['currentUser']()).toEqual(mockUser);
    });

    it('should expose current role', () => {
      expect(component['currentUserRole']()).toBe(UserRole.SUPER_ADMIN);
    });

    it('should expose current tenant', () => {
      expect(component['currentTenant']()).toBe('tenant-1');
    });
  });

  describe('navItems', () => {
    it('should include items when user has permission', () => {
      permissionServiceMock.canAny.mockReturnValue(true);

      const items = component['navItems']();

      expect(items.length).toBeGreaterThan(0);
    });

    it('should return empty list when user has no permissions', () => {
      permissionServiceMock.canAny.mockReturnValue(false);

      fixture = TestBed.createComponent(AppShellPage);
      component = fixture.componentInstance;
      fixture.detectChanges();

      const items = component['navItems']();

      expect(items).toEqual([]);
    });

    it('should filter out items user has no permission for', () => {
      permissionServiceMock.canAny.mockReturnValue(false);

      fixture = TestBed.createComponent(AppShellPage);
      component = fixture.componentInstance;
      fixture.detectChanges();

      const items = component['navItems']();
      const itemsWithoutPermission = items.filter((item) => !item.permission);

      expect(items.length).toBe(itemsWithoutPermission.length);
    });
  });

  describe('template bindings', () => {
    it('should handle null user email gracefully', () => {
      currentUserSignal.set(null);
      fixture.detectChanges();

      expect(component['currentUser']()).toBeNull();
    });
  });

  describe('onLogout', () => {
    it('should call authSessionService.logout and navigate to /login', () => {
      headerDebug.triggerEventHandler('logoutRequested');
      fixture.detectChanges();

      expect(authSessionServiceMock.logout).toHaveBeenCalled();
      expect(router.navigate).toHaveBeenCalledWith(['/login']);
    });
  });

  describe('onChangePassword', () => {
    it('should open ChangePasswordDialog', () => {
      headerDebug.triggerEventHandler('changePasswordRequested');
      fixture.detectChanges();

      expect(dialogMock.open).toHaveBeenCalledWith(ChangePasswordDialog, {
        width: '800px',
        disableClose: true,
      });
    });
  });
});
