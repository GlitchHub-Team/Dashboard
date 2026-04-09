import { ComponentFixture, TestBed } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { RouterLink, RouterLinkActive, provideRouter } from '@angular/router';

import { SideBarComponent } from './side-bar.component';
import { NavItem } from '../../../../models/nav_items/nav-item.model';
import { Permission } from '../../../../models/permission.enum';

describe('SideBarComponent (Unit)', () => {
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

  const setItems = (items: NavItem[]) => {
    fixture.componentRef.setInput('navItems', items);
    fixture.detectChanges();
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [SideBarComponent],
      providers: [provideRouter([])],
    }).compileComponents();

    fixture = TestBed.createComponent(SideBarComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  describe('initial state', () => {
    it('should create with empty navItems and render no items', () => {
      expect(component).toBeTruthy();
      expect(component.navItems()).toEqual([]);
      expect(fixture.debugElement.queryAll(By.css('.nav-item'))).toHaveLength(0);
    });
  });

  describe('rendering', () => {
    beforeEach(() => setItems(mockNavItems));

    it('should render correct number of nav items', () => {
      expect(fixture.debugElement.queryAll(By.css('.nav-item'))).toHaveLength(2);
    });

    it('should render labels and icons', () => {
      const spans = fixture.debugElement.queryAll(By.css('.nav-item span'));
      const icons = fixture.debugElement.queryAll(By.css('.nav-item mat-icon'));

      expect(spans[0].nativeElement.textContent).toContain('Dashboard');
      expect(spans[1].nativeElement.textContent).toContain('Settings');
      expect(icons[0].nativeElement.textContent).toContain('dashboard');
      expect(icons[1].nativeElement.textContent).toContain('settings');
    });

    it('should render correct routerLinks and routerLinkActive directives', () => {
      const links = fixture.debugElement.queryAll(By.directive(RouterLink));
      expect(links).toHaveLength(2);
      expect(links[0].nativeElement.getAttribute('href')).toBe('/dashboard');
      expect(links[1].nativeElement.getAttribute('href')).toBe('/settings');
      expect(fixture.debugElement.queryAll(By.directive(RouterLinkActive))).toHaveLength(2);
    });
  });

  describe('edge cases', () => {
    it('should update when navItems changes', () => {
      setItems([mockNavItems[0]]);
      expect(fixture.debugElement.queryAll(By.css('.nav-item'))).toHaveLength(1);

      setItems(mockNavItems);
      expect(fixture.debugElement.queryAll(By.css('.nav-item'))).toHaveLength(2);
    });
  });
});
