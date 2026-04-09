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

  describe('empty state', () => {
    it('should render empty state when tenant list is empty', () => {
      setInput('tenants', []);
      setInput('loading', false);

      const empty = fixture.debugElement.query(By.css('.empty-state'));
      expect(empty).toBeTruthy();
      expect(empty.query(By.css('mat-icon')).nativeElement.textContent).toContain('business');
      expect(empty.query(By.css('p')).nativeElement.textContent).toContain(
        'Nessun tenant disponibile',
      );
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
    it.each([
      ['dashboardRequested', 'dashboard', (c: TenantTableComponent) => c.dashboardRequested],
      ['deleteRequested', 'delete', (c: TenantTableComponent) => c.deleteRequested],
    ] as const)('should emit %s when %s button is clicked', (_outputName, iconText, getOutput) => {
      const spy = vi.fn();
      getOutput(component).subscribe(spy);

      const btn = fixture.debugElement
        .queryAll(By.css('button'))
        .find((b) => b.nativeElement.textContent.includes(iconText));
      expect(btn).toBeTruthy();

      const stopPropagation = vi.fn();
      btn!.triggerEventHandler('click', { stopPropagation });

      expect(spy).toHaveBeenCalledWith(mockTenants[0]);
      expect(stopPropagation).toHaveBeenCalledTimes(1);
    });
  });

  describe('loading state', () => {
    it('should show only spinner, hide table and buttons when loading', () => {
      setInput('loading', true);
      setInput('tenants', mockTenants);

      expect(fixture.debugElement.query(By.css('mat-spinner'))).toBeTruthy();
      expect(fixture.debugElement.query(By.css('mat-table'))).toBeFalsy();
      expect(fixture.debugElement.query(By.css('.empty-state'))).toBeFalsy();
      expect(fixture.debugElement.queryAll(By.css('button')).length).toBe(0);
    });
  });

  describe('paginator', () => {
    it('should render paginator when tenants are present', () => {
      expect(fixture.debugElement.query(By.css('mat-paginator'))).toBeTruthy();
    });

    it('should not render paginator when tenants list is empty', () => {
      setInput('tenants', []);
      expect(fixture.debugElement.query(By.css('mat-paginator'))).toBeFalsy();
    });

    it('should emit pageChange when paginator emits a page event', () => {
      const spy = vi.fn();
      component.pageChange.subscribe(spy);

      const event = { pageIndex: 1, pageSize: 25, length: 50 };
      fixture.debugElement.query(By.css('mat-paginator')).componentInstance.page.emit(event);

      expect(spy).toHaveBeenCalledWith(event);
    });
  });
});
