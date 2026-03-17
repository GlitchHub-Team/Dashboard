import { ComponentFixture, TestBed } from '@angular/core/testing';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { RouterLink, RouterLinkActive } from '@angular/router';

import { SideBarComponent } from './side-bar.component';
import { NavItem } from '../../../../models/nav-item.model';
import { Permission } from '../../../../models/permission.enum';

describe('SideBarComponent', () => {
  let component: SideBarComponent;
  let fixture: ComponentFixture<SideBarComponent>;

  const mockNavItems: NavItem[] = [
    {
      label: 'Dashboard',
      icon: 'dashboard',
      route: '/dashboard',
      permission: Permission.DASHBOARD_ACCESS,
    },
    {
      label: 'Settings',
      icon: 'settings',
      route: '/settings',
      permission: Permission.GATEWAY_MANAGEMENT,
    },
  ];

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [SideBarComponent],
    })
      .overrideComponent(SideBarComponent, {
        remove: { imports: [RouterLink, RouterLinkActive] },
        add: { schemas: [NO_ERRORS_SCHEMA] },
      })
      .compileComponents();

    fixture = TestBed.createComponent(SideBarComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  describe('initial state', () => {
    it('should create', () => {
      expect(component).toBeTruthy();
    });

    it('should default to empty navItems', () => {
      expect(component.navItems()).toEqual([]);
    });
  });

  describe('inputs', () => {
    it('should accept navItems', () => {
      fixture.componentRef.setInput('navItems', mockNavItems);
      fixture.detectChanges();

      expect(component.navItems()).toEqual(mockNavItems);
    });

    it('should accept empty array', () => {
      fixture.componentRef.setInput('navItems', []);
      fixture.detectChanges();

      expect(component.navItems()).toEqual([]);
    });
  });
});
