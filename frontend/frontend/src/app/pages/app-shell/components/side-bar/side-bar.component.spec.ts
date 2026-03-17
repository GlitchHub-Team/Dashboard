import { ComponentFixture, TestBed } from '@angular/core/testing';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { By } from '@angular/platform-browser';
import { RouterModule } from '@angular/router';

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
      imports: [SideBarComponent, RouterModule.forRoot([])],
      schemas: [NO_ERRORS_SCHEMA],
    }).compileComponents();

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

    it('should render no nav items by default', () => {
      const navLinks = fixture.debugElement.queryAll(By.css('.nav-item'));
      expect(navLinks.length).toBe(0);
    });
  });

  describe('inputs', () => {
    it('should accept navItems and render them', () => {
      fixture.componentRef.setInput('navItems', mockNavItems);
      fixture.detectChanges();

      expect(component.navItems()).toEqual(mockNavItems);

      const navLinks = fixture.debugElement.queryAll(By.css('.nav-item'));
      expect(navLinks.length).toBe(2);
    });

    it('should render correct labels', () => {
      fixture.componentRef.setInput('navItems', mockNavItems);
      fixture.detectChanges();

      const spans = fixture.debugElement.queryAll(By.css('.nav-item span'));
      expect(spans[0].nativeElement.textContent).toContain('Dashboard');
      expect(spans[1].nativeElement.textContent).toContain('Settings');
    });

    it('should render correct icons', () => {
      fixture.componentRef.setInput('navItems', mockNavItems);
      fixture.detectChanges();

      const icons = fixture.debugElement.queryAll(By.css('.nav-item mat-icon'));
      expect(icons[0].nativeElement.textContent).toContain('dashboard');
      expect(icons[1].nativeElement.textContent).toContain('settings');
    });

    it('should accept empty array', () => {
      fixture.componentRef.setInput('navItems', []);
      fixture.detectChanges();

      expect(component.navItems()).toEqual([]);

      const navLinks = fixture.debugElement.queryAll(By.css('.nav-item'));
      expect(navLinks.length).toBe(0);
    });
  });
});
