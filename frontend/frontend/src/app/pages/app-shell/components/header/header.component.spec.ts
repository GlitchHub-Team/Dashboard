import { ComponentFixture, TestBed } from '@angular/core/testing';
import { TestbedHarnessEnvironment } from '@angular/cdk/testing/testbed';
import { HarnessLoader } from '@angular/cdk/testing';
import { MatMenuHarness } from '@angular/material/menu/testing';
import { HeaderComponent } from './header.component';
import { vi, describe, it, expect, beforeEach } from 'vitest';

describe('HeaderComponent', () => {
  let component: HeaderComponent;
  let fixture: ComponentFixture<HeaderComponent>;
  let loader: HarnessLoader;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [HeaderComponent],
    }).compileComponents();

    fixture = TestBed.createComponent(HeaderComponent);
    component = fixture.componentInstance;
    loader = TestbedHarnessEnvironment.loader(fixture);
    fixture.detectChanges();
  });

  describe('initial state', () => {
    it('should create', () => {
      expect(component).toBeTruthy();
    });

    it('should have null defaults and update when inputs change', () => {
      expect(component.username()).toBeNull();
      expect(component.currentTenant()).toBeNull();
      expect(component.currentUserRole()).toBeNull();

      fixture.componentRef.setInput('username', 'test@user.com');
      fixture.componentRef.setInput('currentTenant', 'Alpha');
      fixture.componentRef.setInput('currentUserRole', 'admin');

      expect(component.username()).toBe('test@user.com');
      expect(component.currentTenant()).toBe('Alpha');
      expect(component.currentUserRole()).toBe('admin');
    });
  });

  describe('template rendering', () => {
    it('should show/hide the tenant badge based on currentTenant input', () => {
      fixture.componentRef.setInput('currentTenant', null);
      fixture.detectChanges();
      let badge = fixture.nativeElement.querySelector('.tenant-badge');
      expect(badge).toBeNull();

      fixture.componentRef.setInput('currentTenant', 'Acme Corp');
      fixture.detectChanges();
      badge = fixture.nativeElement.querySelector('.tenant-badge');
      expect(badge.textContent).toContain('Acme Corp');
    });

    it('should render the role badge in UPPERCASE', () => {
      fixture.componentRef.setInput('currentUserRole', 'manager');
      fixture.detectChanges();

      const badge = fixture.nativeElement.querySelector('.role-badge');
      expect(badge.textContent).toBe('MANAGER');
    });
  });

  describe('user menu and outputs', () => {
    it('should display username inside the menu when it exists', async () => {
      fixture.componentRef.setInput('username', 'John Doe');
      fixture.detectChanges();

      const menu = await loader.getHarness(MatMenuHarness);
      await menu.open();

      const menuHeader = document.querySelector('.menu-header');
      expect(menuHeader).toBeTruthy();
      expect(menuHeader?.textContent).toContain('John Doe');
    });

    it('should emit changePasswordRequested when the menu button is clicked', async () => {
      const spy = vi.fn();
      component.changePasswordRequested.subscribe(spy);

      const menu = await loader.getHarness(MatMenuHarness);
      await menu.open();

      const items = await menu.getItems({ text: /Cambia Password/i });
      await items[0].click();

      expect(spy).toHaveBeenCalledOnce();
    });

    it('should emit logoutRequested when the logout button is clicked', async () => {
      const spy = vi.fn();
      component.logoutRequested.subscribe(spy);

      const menu = await loader.getHarness(MatMenuHarness);
      await menu.open();

      const items = await menu.getItems({ text: /Logout/i });
      await items[0].click();

      expect(spy).toHaveBeenCalledOnce();
    });
  });
});
