import { ComponentFixture, TestBed } from '@angular/core/testing';
import { WritableSignal, signal } from '@angular/core';
import { Router } from '@angular/router';
import { MatDialog } from '@angular/material/dialog';
import { MatSnackBar } from '@angular/material/snack-bar';
import { By } from '@angular/platform-browser';
import { of, Subject } from 'rxjs';
import { describe, expect, it, vi, afterEach } from 'vitest';

import { TenantManagerPage } from './tenant-manager.page';
import { TenantTableComponent } from './components/tenant-table/tenant-table.component';
import { TenantService } from '../../services/tenant/tenant.service';
import { TenantFormDialog } from './dialogs/tenant-form/tenant-form.dialog';
import { ConfirmDeleteDialog } from '../shared/dialogs/confirm-delete/confirm-delete.dialog';
import { Tenant } from '../../models/tenant/tenant.model';

const mockTenants: Tenant[] = [
  { id: 'tenant-1', name: 'Acme Corp', canImpersonate: true },
  { id: 'tenant-2', name: 'Globex Inc', canImpersonate: true },
  { id: 'tenant-3', name: 'Restricted Co', canImpersonate: false },
];

function createTenantServiceMock() {
  return {
    tenantList: signal<Tenant[]>([]),
    total: signal(0),
    pageIndex: signal(0),
    limit: signal(10),
    loading: signal(false),
    error: signal<string | null>(null),
    retrieveTenants: vi.fn(),
    changePage: vi.fn(),
    removeTenant: vi.fn().mockReturnValue(of(void 0)),
  };
}

function setupTestBed() {
  const afterClosedSubject = new Subject<unknown>();
  const tenantServiceMock = createTenantServiceMock();
  const dialogMock = {
    open: vi.fn().mockReturnValue({ afterClosed: () => afterClosedSubject.asObservable() }),
  };
  const snackBarMock = { open: vi.fn() };
  const routerMock = { navigate: vi.fn() };

  TestBed.configureTestingModule({
    imports: [TenantManagerPage, TenantTableComponent],
    providers: [
      { provide: TenantService, useValue: tenantServiceMock },
      { provide: Router, useValue: routerMock },
    ],
  })
    .overrideProvider(MatDialog, { useValue: dialogMock })
    .overrideProvider(MatSnackBar, { useValue: snackBarMock });

  const fixture = TestBed.createComponent(TenantManagerPage);
  return { fixture, tenantServiceMock, dialogMock, snackBarMock, routerMock, afterClosedSubject };
}

function getTableRows(fixture: ComponentFixture<TenantManagerPage>): HTMLElement[] {
  return Array.from(fixture.nativeElement.querySelectorAll('mat-row'));
}

function getHeaderCells(fixture: ComponentFixture<TenantManagerPage>): string[] {
  return Array.from<HTMLElement>(fixture.nativeElement.querySelectorAll('mat-header-cell')).map(
    (h) => h.textContent?.trim() ?? '',
  );
}

function getDeleteButtons(fixture: ComponentFixture<TenantManagerPage>): HTMLButtonElement[] {
  return Array.from(fixture.nativeElement.querySelectorAll('mat-cell button[color="warn"]'));
}

function getDashboardButtons(fixture: ComponentFixture<TenantManagerPage>): HTMLButtonElement[] {
  return Array.from<HTMLButtonElement>(
    fixture.nativeElement.querySelectorAll('mat-cell button[color="primary"]'),
  ).filter((btn) => btn.querySelector('mat-icon')?.textContent?.trim() === 'dashboard');
}

function getTenantUserButtons(fixture: ComponentFixture<TenantManagerPage>): HTMLButtonElement[] {
  return Array.from<HTMLButtonElement>(
    fixture.nativeElement.querySelectorAll('mat-cell button[color="primary"]'),
  ).filter((btn) => btn.querySelector('mat-icon')?.textContent?.trim() === 'people');
}

describe('TenantManagerPage (Integration)', () => {
  afterEach(() => {
    TestBed.resetTestingModule();
    vi.resetAllMocks();
  });

  describe('Initialization', () => {
    it('should render page title and call retrieveTenants on init', () => {
      const { fixture, tenantServiceMock } = setupTestBed();
      fixture.detectChanges();

      expect(fixture.nativeElement.querySelector('h1').textContent).toContain('Gestione Tenant');
      expect(tenantServiceMock.retrieveTenants).toHaveBeenCalledOnce();
    });
  });

  describe('Page -> Table: Input Bindings', () => {
    it('should render tenants, display correct data and column headers', () => {
      const { fixture, tenantServiceMock } = setupTestBed();
      (tenantServiceMock.tenantList as WritableSignal<Tenant[]>).set(mockTenants);
      fixture.detectChanges();

      expect(getTableRows(fixture).length).toBe(3);

      const headers = getHeaderCells(fixture);
      expect(headers).toContain('ID');
      expect(headers).toContain('Nome');
      expect(headers).toContain('Azioni');

      (tenantServiceMock.tenantList as WritableSignal<Tenant[]>).set([mockTenants[0]]);
      fixture.detectChanges();

      const cellTexts = Array.from<HTMLElement>(
        fixture.nativeElement.querySelectorAll('mat-row mat-cell'),
      ).map((c) => c.textContent?.trim());
      expect(cellTexts).toContain('tenant-1');
      expect(cellTexts).toContain('Acme Corp');
    });

    it('should show spinner when loading and empty state when idle with no tenants', () => {
      const { fixture, tenantServiceMock } = setupTestBed();
      (tenantServiceMock.loading as WritableSignal<boolean>).set(true);
      fixture.detectChanges();

      expect(fixture.nativeElement.querySelector('mat-spinner')).toBeTruthy();
      expect(getTableRows(fixture).length).toBe(0);

      (tenantServiceMock.loading as WritableSignal<boolean>).set(false);
      fixture.detectChanges();

      const emptyState = fixture.nativeElement.querySelector('.empty-state');
      expect(emptyState).toBeTruthy();
      expect(emptyState.textContent).toContain('Nessun tenant disponibile');
    });

    it('should show error banner when error signal has value', () => {
      const { fixture, tenantServiceMock } = setupTestBed();
      (tenantServiceMock.error as WritableSignal<string | null>).set('Failed to load tenants');
      fixture.detectChanges();

      const errorBanner = fixture.nativeElement.querySelector('.error-banner');
      expect(errorBanner).toBeTruthy();
      expect(errorBanner.textContent).toContain('Failed to load tenants');
    });

    it('should render paginator', () => {
      const { fixture, tenantServiceMock } = setupTestBed();
      (tenantServiceMock.tenantList as WritableSignal<Tenant[]>).set(mockTenants);
      (tenantServiceMock.total as WritableSignal<number>).set(50);
      fixture.detectChanges();

      expect(fixture.nativeElement.querySelector('mat-paginator')).toBeTruthy();
    });
  });

  describe('Action Buttons: canImpersonate', () => {
    it.each([
      ['with canImpersonate=true', [mockTenants[0]], 1, 1, 1],
      ['with canImpersonate=false', [mockTenants[2]], 0, 0, 1],
      ['with mixed tenants (2 true, 1 false)', mockTenants, 2, 2, 3],
    ] as const)(
      'should show correct action buttons %s',
      (_label, tenants, dashCount, userCount, deleteCount) => {
        const { fixture, tenantServiceMock } = setupTestBed();
        (tenantServiceMock.tenantList as WritableSignal<Tenant[]>).set(tenants as Tenant[]);
        fixture.detectChanges();

        expect(getDashboardButtons(fixture).length).toBe(dashCount);
        expect(getTenantUserButtons(fixture).length).toBe(userCount);
        expect(getDeleteButtons(fixture).length).toBe(deleteCount);
      },
    );
  });

  describe('Table -> Page: Navigation Events', () => {
    it.each([
      [
        'dashboard',
        (f: ComponentFixture<TenantManagerPage>) => getDashboardButtons(f)[0],
        ['/dashboard'],
        { tenantId: 'tenant-1' },
      ],
      [
        'tenant user management',
        (f: ComponentFixture<TenantManagerPage>) => getTenantUserButtons(f)[0],
        ['/user-management/tenant-users'],
        { tenantId: 'tenant-1' },
      ],
      [
        'dashboard (second tenant)',
        (f: ComponentFixture<TenantManagerPage>) => getDashboardButtons(f)[1],
        ['/dashboard'],
        { tenantId: 'tenant-2' },
      ],
    ] as const)(
      'should navigate to %s when button is clicked',
      (_label, getBtn, path, queryParams) => {
        const { fixture, tenantServiceMock, routerMock } = setupTestBed();
        (tenantServiceMock.tenantList as WritableSignal<Tenant[]>).set(mockTenants);
        fixture.detectChanges();

        getBtn(fixture).click();
        fixture.detectChanges();

        expect(routerMock.navigate).toHaveBeenCalledWith(path, { queryParams });
      },
    );
  });

  describe('Table -> Page: Pagination', () => {
    it('should call changePage when page event is emitted', () => {
      const { fixture, tenantServiceMock } = setupTestBed();
      (tenantServiceMock.tenantList as WritableSignal<Tenant[]>).set(mockTenants);
      (tenantServiceMock.total as WritableSignal<number>).set(50);
      fixture.detectChanges();

      expect(fixture.debugElement.query(By.css('mat-paginator'))).toBeTruthy();

      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      (fixture.componentInstance as any).onPageChange({ pageIndex: 2, pageSize: 25, length: 50 });

      expect(tenantServiceMock.changePage).toHaveBeenCalledWith(2, 25);
    });
  });

  describe('Full Delete Flow', () => {
    it('should call removeTenant for the correct tenant, show snackbar when confirmed', () => {
      const { fixture, tenantServiceMock, dialogMock, snackBarMock, afterClosedSubject } =
        setupTestBed();

      (tenantServiceMock.tenantList as WritableSignal<Tenant[]>).set(mockTenants);
      fixture.detectChanges();

      getDeleteButtons(fixture)[1].click();
      fixture.detectChanges();

      expect(dialogMock.open).toHaveBeenCalledWith(ConfirmDeleteDialog, {
        width: '400px',
        data: {
          title: 'Delete Tenant',
          message: 'Sei sicuro di voler eliminare il tenant "Globex Inc"?',
        },
      });

      afterClosedSubject.next(true);

      expect(tenantServiceMock.removeTenant).toHaveBeenCalledWith('tenant-2');
      expect(snackBarMock.open).toHaveBeenCalledWith('Tenant eliminato con successo', 'Close', {
        duration: 3000,
      });
    });

    it('should NOT call removeTenant when dialog is cancelled', () => {
      const { fixture, tenantServiceMock, afterClosedSubject } = setupTestBed();

      (tenantServiceMock.tenantList as WritableSignal<Tenant[]>).set([mockTenants[0]]);
      fixture.detectChanges();

      getDeleteButtons(fixture)[0].click();
      fixture.detectChanges();
      afterClosedSubject.next(false);

      expect(tenantServiceMock.removeTenant).not.toHaveBeenCalled();
    });
  });

  describe('Full Create Flow', () => {
    it('should open create dialog, refresh and show snackbar after success', () => {
      const { fixture, tenantServiceMock, dialogMock, snackBarMock, afterClosedSubject } =
        setupTestBed();

      fixture.detectChanges();
      tenantServiceMock.retrieveTenants.mockClear();

      fixture.nativeElement.querySelector('button[mat-raised-button]').click();
      fixture.detectChanges();

      expect(dialogMock.open).toHaveBeenCalledWith(TenantFormDialog, {
        width: '500px',
        data: null,
      });

      afterClosedSubject.next(true);

      expect(tenantServiceMock.retrieveTenants).toHaveBeenCalledOnce();
      expect(snackBarMock.open).toHaveBeenCalledWith('Tenant creato con successo', 'Close', {
        duration: 3000,
      });
    });

    it('should NOT refresh when create dialog is cancelled', () => {
      const { fixture, tenantServiceMock, afterClosedSubject } = setupTestBed();

      fixture.detectChanges();
      tenantServiceMock.retrieveTenants.mockClear();

      fixture.nativeElement.querySelector('button[mat-raised-button]').click();
      fixture.detectChanges();
      afterClosedSubject.next(false);

      expect(tenantServiceMock.retrieveTenants).not.toHaveBeenCalled();
    });
  });
});
