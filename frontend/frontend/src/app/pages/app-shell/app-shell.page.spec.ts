import { ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA, signal } from '@angular/core';
import { Router, RouterOutlet } from '@angular/router';
import { MatDialog } from '@angular/material/dialog';

import { AppShellPage } from './app-shell.page';
import { UserSessionService } from '../../services/user-session/user-session.service';
import { AuthSessionService } from '../../services/auth/auth-session.service';
import { PermissionService } from '../../services/permission/permission.service';
import { ChangePasswordDialog } from './dialogs/change-password/change-password.dialog';
import { HeaderComponent } from './components/header/header.component';
import { SideBarComponent } from './components/side-bar/side-bar.component';
import { UserRole } from '../../models/user-role.enum';
import { User } from '../../models/user.model';

describe('AppShellPage', () => {
  let component: AppShellPage;
  let fixture: ComponentFixture<AppShellPage>;

  const mockUser: User = {
    id: '1',
    email: 'admin@test.com',
    role: UserRole.SUPER_ADMIN,
    tenantId: 'tenant-1',
  };

  const userSessionMock = {
    currentUser: signal<User | null>(mockUser).asReadonly(),
    currentRole: signal<UserRole | null>(UserRole.SUPER_ADMIN).asReadonly(),
    currentTenant: signal<string | null>('tenant-1').asReadonly(),
  };

  const authSessionServiceMock = {
    logout: vi.fn(),
  };

  const permissionServiceMock = {
    canAny: vi.fn().mockReturnValue(true),
  };

  const routerMock = {
    navigate: vi.fn(),
  };

  const dialogMock = {
    open: vi.fn(),
  };

  beforeEach(async () => {
    vi.resetAllMocks();
    permissionServiceMock.canAny.mockReturnValue(true);

    await TestBed.configureTestingModule({
      imports: [AppShellPage],
      providers: [
        { provide: UserSessionService, useValue: userSessionMock },
        { provide: AuthSessionService, useValue: authSessionServiceMock },
        { provide: PermissionService, useValue: permissionServiceMock },
        { provide: Router, useValue: routerMock },
        { provide: MatDialog, useValue: dialogMock },
      ],
    })
      .overrideComponent(AppShellPage, {
        remove: { imports: [RouterOutlet, SideBarComponent, HeaderComponent] },
        add: { schemas: [CUSTOM_ELEMENTS_SCHEMA] },
      })
      .compileComponents();

    fixture = TestBed.createComponent(AppShellPage);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  describe('initial state', () => {
    it('should create', () => {
      expect(component).toBeTruthy();
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

  describe('onLogout', () => {
    it('should call authSessionService.logout', () => {
      component['onLogout']();

      expect(authSessionServiceMock.logout).toHaveBeenCalled();
    });

    it('should navigate to /login', () => {
      component['onLogout']();

      expect(routerMock.navigate).toHaveBeenCalledWith(['/login']);
    });
  });

  describe('onChangePassword', () => {
    it('should open ChangePasswordDialog', () => {
      component['onChangePassword']();

      expect(dialogMock.open).toHaveBeenCalledWith(ChangePasswordDialog, {
        width: '800px',
        disableClose: true,
      });
    });
  });
});
