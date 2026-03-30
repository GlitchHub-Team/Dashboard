import { ComponentFixture, TestBed } from '@angular/core/testing';
import { signal } from '@angular/core';
import { MatDialog } from '@angular/material/dialog';
import { Router } from '@angular/router';
import { PageEvent } from '@angular/material/paginator';
import { of, Subject } from 'rxjs';
import { describe, it, expect, beforeEach, vi } from 'vitest';

import { TenantManagerPage } from './tenant-manager.page';
import { TenantService } from '../../services/tenant/tenant.service';
import { Tenant } from '../../models/tenant/tenant.model';
import { TenantFormDialog } from './dialogs/tenant-form/tenant-form.dialog';
import { ConfirmDeleteDialog } from '../gateway-sensor/dialogs/confirm-delete/confirm-delete.dialog';

interface TenantManagerPageTestApi {
  onCreateTenant: () => void;
  onDeleteTenant: (tenant: Tenant) => void;
  onPageChange: (event: PageEvent) => void;
  onGoToDashboard: (tenant: Tenant) => void;
  onGoToTenantUserManagement: (tenant: Tenant) => void;
}

describe('TenantManagerPage', () => {
  let component: TenantManagerPage;
  let fixture: ComponentFixture<TenantManagerPage>;
  let testApi: TenantManagerPageTestApi;

  let afterClosedSubject: Subject<unknown>;
  let dialogMock: { open: ReturnType<typeof vi.fn> };

  const mockTenantService = {
    tenantList: signal<Tenant[]>([]),
    total: signal(0),
    pageIndex: signal(0),
    limit: signal(10),
    loading: signal(false),
    retrieveTenant: vi.fn(),
    removeTenant: vi.fn(),
    changePage: vi.fn(),
  };

  const mockRouter = {
    navigate: vi.fn(),
  };

  const mockTenants: Tenant[] = [
    { id: 'tenant-01', name: 'Tenant 1', canImpersonate: false },
    { id: 'tenant-02', name: 'Tenant 2', canImpersonate: true },
  ];

  beforeEach(async () => {
    vi.clearAllMocks();

    afterClosedSubject = new Subject();
    dialogMock = {
      open: vi.fn().mockReturnValue({
        afterClosed: () => afterClosedSubject.asObservable(),
      }),
    };

    await TestBed.configureTestingModule({
      imports: [TenantManagerPage],
      providers: [
        { provide: TenantService, useValue: mockTenantService },
        { provide: Router, useValue: mockRouter },
        { provide: MatDialog, useValue: dialogMock },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(TenantManagerPage);
    component = fixture.componentInstance;
    testApi = component as unknown as TenantManagerPageTestApi;
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should fetch tenants on init', () => {
    fixture.detectChanges();
    expect(mockTenantService.retrieveTenant).toHaveBeenCalledTimes(1);
  });

  it.each([
    {
      name: 'page title',
      selector: 'h1',
      expectedText: 'Gestione Tenant',
      setup: () => undefined,
    },
    {
      name: 'create button',
      selector: 'button[mat-raised-button]',
      expectedText: 'Aggiungi Nuovo Tenant',
      setup: () => undefined,
    },
    {
      name: 'tenant table',
      selector: 'app-tenant-table',
      expectedText: null,
      setup: () => mockTenantService.tenantList.set(mockTenants),
    },
    {
      name: 'paginator',
      selector: 'mat-paginator',
      expectedText: null,
      setup: () => undefined,
    },
  ])('should render $name', ({ selector, expectedText, setup }) => {
    setup();
    fixture.detectChanges();

    const element = fixture.nativeElement.querySelector(selector);
    expect(element).toBeTruthy();

    if (expectedText) {
      expect(element.textContent).toContain(expectedText);
    }
  });

  it('should open create dialog with correct config', () => {
    testApi.onCreateTenant();

    expect(dialogMock.open).toHaveBeenCalledWith(TenantFormDialog, {
      width: '500px',
      data: null,
    });
  });

  it('should refetch tenants after create dialog closes', () => {
    testApi.onCreateTenant();
    afterClosedSubject.next(true);

    expect(mockTenantService.retrieveTenant).toHaveBeenCalled();
  });

  it('should open delete dialog with correct config', () => {
    const tenant = mockTenants[0];

    testApi.onDeleteTenant(tenant);

    expect(dialogMock.open).toHaveBeenCalledWith(ConfirmDeleteDialog, {
      width: '400px',
      data: {
        title: 'Delete Tenant',
        message: `Sei sicuro di voler eliminare il tenant "${tenant.name}"?`,
      },
    });
  });

  it.each([
    { confirmed: true, shouldDelete: true },
    { confirmed: false, shouldDelete: false },
  ])('should handle delete confirmation: $confirmed', ({ confirmed, shouldDelete }) => {
    mockTenantService.removeTenant.mockReturnValue(of(void 0));
    const tenant = mockTenants[0];

    testApi.onDeleteTenant(tenant);
    afterClosedSubject.next(confirmed);

    if (shouldDelete) {
      expect(mockTenantService.removeTenant).toHaveBeenCalledWith(tenant.id);
      return;
    }

    expect(mockTenantService.removeTenant).not.toHaveBeenCalled();
  });

  it('should call changePage on page change', () => {
    const event: PageEvent = { pageIndex: 2, pageSize: 25, length: 100 };

    testApi.onPageChange(event);

    expect(mockTenantService.changePage).toHaveBeenCalledWith(2, 25);
  });

  it.each([
    { tenant: mockTenants[0], tenantId: 'tenant-01' },
    { tenant: mockTenants[1], tenantId: 'tenant-02' },
  ])('should navigate to dashboard for $tenantId', ({ tenant, tenantId }) => {
    testApi.onGoToDashboard(tenant);

    expect(mockRouter.navigate).toHaveBeenCalledWith(['/dashboard'], {
      queryParams: { tenantId },
    });
  });

    it.each([
    { tenant: mockTenants[0], tenantId: 'tenant-01' },
    { tenant: mockTenants[1], tenantId: 'tenant-02' },
  ])('should navigate to tenant user management for $tenantId', ({ tenant, tenantId }) => {
    testApi.onGoToTenantUserManagement(tenant);

    expect(mockRouter.navigate).toHaveBeenCalledWith(['/user-management/tenant-users'], {
      queryParams: { tenantId },
    });
  });
});
