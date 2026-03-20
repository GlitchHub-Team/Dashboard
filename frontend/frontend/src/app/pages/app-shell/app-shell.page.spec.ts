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

  const authSessionServiceMock = { logout: vi.fn() };
  const permissionServiceMock = { canAny: vi.fn().mockReturnValue(true) };
  const dialogMock = { open: vi.fn() };

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
  });

  describe('initial state', () => {
    it('should create and render the full shell layout with header, sidebar, and router outlet', () => {
      expect(component).toBeTruthy();
      expect(fixture.debugElement.query(By.css('.shell-layout'))).toBeTruthy();
      expect(fixture.debugElement.query(By.css('.main-content'))).toBeTruthy();
      expect(fixture.debugElement.query(By.css('main'))).toBeTruthy();
      expect(fixture.debugElement.query(By.css('router-outlet'))).toBeTruthy();
      expect(fixture.debugElement.query(By.css('app-side-bar'))).toBeTruthy();
      expect(headerDebug).toBeTruthy();
    });

    it('should expose current user, role, and tenant from session', () => {
      expect(component['currentUser']()).toEqual(mockUser);
      expect(component['currentUserRole']()).toBe(UserRole.SUPER_ADMIN);
      expect(component['currentTenant']()).toBe('tenant-1');
    });
  });

  describe('navItems', () => {
    it('should include items when user has permission', () => {
      expect(component['navItems']().length).toBeGreaterThan(0);
    });

    it('should return only items without a permission gate when canAny returns false', () => {
      permissionServiceMock.canAny.mockReturnValue(false);

      fixture = TestBed.createComponent(AppShellPage);
      component = fixture.componentInstance;
      fixture.detectChanges();

      const items = component['navItems']();
      expect(items.every((item) => !item.permission)).toBe(true);
    });
  });

  describe('template bindings', () => {
    it('should handle null user gracefully', () => {
      currentUserSignal.set(null);
      fixture.detectChanges();
      expect(component['currentUser']()).toBeNull();
    });
  });

  describe('onLogout', () => {
    it('should call logout and navigate to /login', () => {
      headerDebug.triggerEventHandler('logoutRequested');
      fixture.detectChanges();

      expect(authSessionServiceMock.logout).toHaveBeenCalled();
      expect(router.navigate).toHaveBeenCalledWith(['/login']);
    });
  });

  describe('onChangePassword', () => {
    it('should open ChangePasswordDialog with correct config', () => {
      headerDebug.triggerEventHandler('changePasswordRequested');
      fixture.detectChanges();

      expect(dialogMock.open).toHaveBeenCalledWith(ChangePasswordDialog, {
        width: '800px',
        disableClose: true,
      });
    });
  });
});
