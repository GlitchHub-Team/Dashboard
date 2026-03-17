import { ComponentFixture, TestBed } from '@angular/core/testing';
import { signal } from '@angular/core';
import { of } from 'rxjs';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';

import { TenantManagerPage } from './tenant-manager.page';
import { TenantService } from '../../services/tenant/tenant.service';
import { Tenant } from '../../models/tenant.model';
import { TenantFormDialog } from './dialogs/tenant-form.dialog';

class MockTenantService {
  tenantList = signal<Tenant[]>([]);
  loading = signal(false);
  
  retrieveTenantCalled = false;
  removeTenantCalledWith = '';

  retrieveTenant() {
    this.retrieveTenantCalled = true;
  }

  removeTenant(name: string) {
    this.removeTenantCalledWith = name;
    return of(void 0);
  }

  reset() {
    this.retrieveTenantCalled = false;
    this.removeTenantCalledWith = '';
  }
}

class MockDialog {
  openCalled = false;
  openArgs: { component?: unknown; config?: { width?: string; data?: unknown } } | null = null;

  open(component: unknown, config: { width?: string; data?: unknown }) {
    this.openCalled = true;
    this.openArgs = { component, config };
    return {
      afterClosed: () => of(true) // Simula la conferma di default
    };
  }

  reset() {
    this.openCalled = false;
    this.openArgs = null;
  }
}

describe('TenantManagerPage', () => {
  let component: TenantManagerPage;
  let fixture: ComponentFixture<TenantManagerPage>;
  let tenantService: MockTenantService;
  let dialog: MockDialog;

  beforeEach(async () => {
    tenantService = new MockTenantService();
    dialog = new MockDialog();

    await TestBed.configureTestingModule({
      imports: [TenantManagerPage, MatDialogModule, NoopAnimationsModule],
      providers: [
        { provide: TenantService, useValue: tenantService },
        { provide: MatDialog, useValue: dialog },
      ],
    })
    .overrideProvider(MatDialog, { useValue: dialog })
    .compileComponents();

    fixture = TestBed.createComponent(TenantManagerPage);
    component = fixture.componentInstance;

    // Reset dello stato dei mock
    tenantService.reset();
    dialog.reset();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should call retrieveTenant on init', () => {
    fixture.detectChanges(); // ngOnInit viene chiamato qui
    expect(tenantService.retrieveTenantCalled).toBe(true);
  });

  it('should open the TenantFormDialog on create', () => {
    component.onCreateTenant();
    expect(dialog.openCalled).toBe(true);
    expect(dialog.openArgs?.component).toBe(TenantFormDialog);
    expect(dialog.openArgs?.config?.width).toBe('400px');
    // Verifica che la lista venga ricaricata dopo la chiusura del dialog
    expect(tenantService.retrieveTenantCalled).toBe(true);
  });

  it('should open confirm dialog and remove tenant on confirmed delete', () => {
    const tenantToDelete: Tenant = { name: 'Test Tenant' };
    component.onDeleteTenant(tenantToDelete);
    expect(dialog.openCalled).toBe(true);
    expect(tenantService.removeTenantCalledWith).toBe(tenantToDelete.name);
  });
});