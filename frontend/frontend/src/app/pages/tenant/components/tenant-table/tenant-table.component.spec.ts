import { ComponentFixture, TestBed } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { TenantTableComponent } from './tenant-table.component';
import { Tenant } from '../../../../models/tenant/tenant.model';

interface TenantTableTestApi {
  displayedColumns: () => string[];
}

describe('TenantTableComponent (Unit)', () => {
  let component: TenantTableComponent;
  let fixture: ComponentFixture<TenantTableComponent>;
  let testApi: TenantTableTestApi;

  const mockTenants: Tenant[] = [
    { id: 'tenant-1', name: 'Tenant One', canImpersonate: true },
    { id: 'tenant-2', name: 'Tenant Two', canImpersonate: false },
  ];

  const setInput = (key: string, value: unknown) => {
    fixture.componentRef.setInput(key, value);
    fixture.detectChanges();
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [TenantTableComponent],
    }).compileComponents();

    fixture = TestBed.createComponent(TenantTableComponent);
    component = fixture.componentInstance;
    testApi = component as unknown as TenantTableTestApi;

    fixture.componentRef.setInput('tenants', mockTenants);
    fixture.componentRef.setInput('loading', false);
    fixture.detectChanges();
  });

  describe('initial state', () => {
    it('should create and expose displayed columns', () => {
      expect(component).toBeTruthy();
      expect(testApi.displayedColumns()).toEqual(['id', 'name', 'actions']);
    });
  });

  describe('loading state', () => {
    it('should render only spinner when loading is true', () => {
      setInput('loading', true);

      expect(fixture.debugElement.query(By.css('mat-spinner'))).toBeTruthy();
      expect(fixture.debugElement.query(By.css('.empty-state'))).toBeFalsy();
      expect(fixture.debugElement.query(By.css('mat-table'))).toBeFalsy();
    });
  });

  describe('empty state', () => {
    it('should render empty state when tenant list is empty', () => {
      setInput('tenants', []);
      setInput('loading', false);

      const empty = fixture.debugElement.query(By.css('.empty-state'));
      expect(empty).toBeTruthy();
      expect(empty.query(By.css('mat-icon')).nativeElement.textContent).toContain('business');
      expect(empty.query(By.css('p')).nativeElement.textContent).toContain('Nessun tenant disponibile');
      expect(fixture.debugElement.query(By.css('mat-spinner'))).toBeFalsy();
      expect(fixture.debugElement.query(By.css('mat-table'))).toBeFalsy();
    });
  });

  describe('table with data', () => {
    it('should render table with rows and tenant data', () => {
      setInput('loading', false);

      expect(fixture.debugElement.query(By.css('mat-table'))).toBeTruthy();
      expect(fixture.debugElement.query(By.css('mat-header-row'))).toBeTruthy();
      expect(fixture.debugElement.queryAll(By.css('mat-row')).length).toBe(2);

      const cellTexts = fixture.debugElement
        .queryAll(By.css('mat-cell'))
        .map((cell) => cell.nativeElement.textContent.trim());

      expect(cellTexts).toEqual(expect.arrayContaining(['tenant-1', 'tenant-2']));
      expect(cellTexts).toEqual(expect.arrayContaining(['Tenant One', 'Tenant Two']));
    });

    it('should show dashboard button only for tenants that can impersonate', () => {
      const dashboardButtons = fixture.debugElement
        .queryAll(By.css('button mat-icon'))
        .filter((icon) => icon.nativeElement.textContent.trim() === 'dashboard');

      expect(dashboardButtons.length).toBe(1);
    });

    it('should render delete button for each tenant row', () => {
      const deleteButtons = fixture.debugElement
        .queryAll(By.css('button mat-icon'))
        .filter((icon) => icon.nativeElement.textContent.trim() === 'delete');

      expect(deleteButtons.length).toBe(2);
    });
  });

  describe('outputs', () => {
    it('should emit dashboardRequested when dashboard action is clicked', () => {
      const spy = vi.fn();
      component.dashboardRequested.subscribe(spy);

      const dashboardButton = fixture.debugElement
        .queryAll(By.css('button'))
        .find((btn) => btn.nativeElement.textContent.includes('dashboard'));

      expect(dashboardButton).toBeTruthy();

      const stopPropagation = vi.fn();
      dashboardButton!.triggerEventHandler('click', { stopPropagation });

      expect(spy).toHaveBeenCalledWith(mockTenants[0]);
      expect(stopPropagation).toHaveBeenCalledTimes(1);
    });

    it('should emit deleteRequested when delete action is clicked', () => {
      const spy = vi.fn();
      component.deleteRequested.subscribe(spy);

      const deleteButton = fixture.debugElement
        .queryAll(By.css('button'))
        .find((btn) => btn.nativeElement.textContent.includes('delete'));

      expect(deleteButton).toBeTruthy();

      const stopPropagation = vi.fn();
      deleteButton!.triggerEventHandler('click', { stopPropagation });

      expect(spy).toHaveBeenCalledWith(mockTenants[0]);
      expect(stopPropagation).toHaveBeenCalledTimes(1);
    });
  });

  describe('disabled actions while loading', () => {
    it('should hide action buttons and show spinner when loading is true', () => {
      setInput('loading', true);
      setInput('tenants', mockTenants);

      const buttons = fixture.debugElement.queryAll(By.css('button'));
      expect(buttons.length).toBe(0);
      expect(fixture.debugElement.query(By.css('mat-spinner'))).toBeTruthy();
      expect(fixture.debugElement.query(By.css('mat-table'))).toBeFalsy();
    });
  });
});
