import { ComponentFixture, TestBed } from '@angular/core/testing';

import { HeaderComponent } from './header.component';

describe('HeaderComponent', () => {
  let component: HeaderComponent;
  let fixture: ComponentFixture<HeaderComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [HeaderComponent],
    }).compileComponents();

    fixture = TestBed.createComponent(HeaderComponent);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  describe('initial state', () => {
    it('should create', () => {
      expect(component).toBeTruthy();
    });

    it('should have default input values', () => {
      expect(component.username()).toBeNull();
      expect(component.currentTenant()).toBeNull();
      expect(component.currentUserRole()).toBeNull();
    });
  });

  describe('inputs', () => {
    it('should accept username', () => {
      fixture.componentRef.setInput('username', 'admin@test.com');
      fixture.detectChanges();

      expect(component.username()).toBe('admin@test.com');
    });

    it('should accept currentTenant', () => {
      fixture.componentRef.setInput('currentTenant', 'tenant-1');
      fixture.detectChanges();

      expect(component.currentTenant()).toBe('tenant-1');
    });

    it('should accept currentUserRole', () => {
      fixture.componentRef.setInput('currentUserRole', 'SUPER_ADMIN');
      fixture.detectChanges();

      expect(component.currentUserRole()).toBe('SUPER_ADMIN');
    });

    it('should accept null values', () => {
      fixture.componentRef.setInput('username', null);
      fixture.componentRef.setInput('currentTenant', null);
      fixture.componentRef.setInput('currentUserRole', null);
      fixture.detectChanges();

      expect(component.username()).toBeNull();
      expect(component.currentTenant()).toBeNull();
      expect(component.currentUserRole()).toBeNull();
    });
  });

  describe('logoutRequested output', () => {
    it('should emit when logoutRequested is called', () => {
      const spy = vi.fn();
      component.logoutRequested.subscribe(spy);

      component.logoutRequested.emit();

      expect(spy).toHaveBeenCalled();
    });
  });

  describe('changePasswordRequested output', () => {
    it('should emit when changePasswordRequested is called', () => {
      const spy = vi.fn();
      component.changePasswordRequested.subscribe(spy);

      component.changePasswordRequested.emit();

      expect(spy).toHaveBeenCalled();
    });
  });
});
