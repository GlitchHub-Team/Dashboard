import { ComponentFixture, TestBed } from '@angular/core/testing';
import { signal } from '@angular/core';
import { of } from 'rxjs';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';
import { ActivatedRoute } from '@angular/router';

import { UserManagerPage } from './user-manager.page';
import { UserService } from '../../services/user/user.service';
import { User } from '../../models/user.model';
import { UserRole } from '../../models/user-role.enum';
import { UserFormDialogComponent } from './dialogs/user-form.dialog';

class MockUserService {
  userList = signal<User[]>([]);
  loading = signal(false);

  retrieveUserCalledWith: UserRole | null = null;
  addNewUserCalledWith: { email: string; role: UserRole } | null = null;
  removeUserCalledWith: string | null = null;

  retrieveUser(role: UserRole) {
    this.retrieveUserCalledWith = role;
  }

  addNewUser(config: { email: string; role: UserRole }) {
    this.addNewUserCalledWith = config;
    return of(void 0);
  }

  removeUser(email: string) {
    this.removeUserCalledWith = email;
    return of(void 0);
  }

  reset() {
    this.retrieveUserCalledWith = null;
    this.addNewUserCalledWith = null;
    this.removeUserCalledWith = null;
  }
}

class MockDialog {
  openCalled = false;
  openArgs: { component?: unknown; config?: { width?: string; data?: unknown } } | null = null;
  returnValue: unknown = true; // Valore di default simulato alla chiusura

  open(component: unknown, config: { width?: string; data?: unknown }) {
    this.openCalled = true;
    this.openArgs = { component, config };
    return {
      afterClosed: () => of(this.returnValue),
    };
  }

  reset() {
    this.openCalled = false;
    this.openArgs = null;
    this.returnValue = true;
  }
}

describe('UserManagerPage', () => {
  let component: UserManagerPage;
  let fixture: ComponentFixture<UserManagerPage>;
  let userService: MockUserService;
  let dialog: MockDialog;

  beforeEach(async () => {
    userService = new MockUserService();
    dialog = new MockDialog();

    await TestBed.configureTestingModule({
      imports: [UserManagerPage, MatDialogModule, NoopAnimationsModule],
      providers: [
        { provide: UserService, useValue: userService },
        { provide: MatDialog, useValue: dialog },
        {
          provide: ActivatedRoute,
          useValue: {
            data: of({ userManagerContext: { title: 'Test Users', role: UserRole.TENANT_ADMIN } }),
          },
        },
      ],
    })
      .overrideProvider(MatDialog, { useValue: dialog })
      .compileComponents();

    fixture = TestBed.createComponent(UserManagerPage);
    component = fixture.componentInstance;

    // Reset dello stato dei mock
    userService.reset();
    dialog.reset();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should call retrieveUser on init with role from ActivatedRoute', () => {
    fixture.detectChanges();
    expect(userService.retrieveUserCalledWith).toBe(UserRole.TENANT_ADMIN);
  });

  it('should open the UserFormDialog on create and add the user if closed with data', () => {
    fixture.detectChanges(); // Inizializza ngOnInit e il context
    dialog.returnValue = { email: 'new@user.com' }; // Dati simulati dal form al suo salvataggio

    component.onCreateUser();

    expect(dialog.openCalled).toBe(true);
    expect(dialog.openArgs?.component).toBe(UserFormDialogComponent);
    expect(dialog.openArgs?.config?.width).toBe('400px');
    
    // Verifica se ha chiamato il servizio per creare l'utente e poi lo ha ricaricato
    expect(userService.addNewUserCalledWith).toEqual({ email: 'new@user.com', role: UserRole.TENANT_ADMIN });
    expect(userService.retrieveUserCalledWith).toBe(UserRole.TENANT_ADMIN);
  });

  it('should open confirm dialog and remove user on confirmed delete', () => {
    fixture.detectChanges();
    const userToDelete: User = { id: '1', email: 'delete@user.com', role: UserRole.TENANT_ADMIN, tenantId: 'tenant-1' };
    dialog.returnValue = true; // Simula la conferma dell'eliminazione

    component.onDeleteUser(userToDelete);

    expect(dialog.openCalled).toBe(true);
    expect(userService.removeUserCalledWith).toBe('delete@user.com');
    
    // Verifica ricaricamento della lista dopo la cancellazione
    expect(userService.retrieveUserCalledWith).toBe(UserRole.TENANT_ADMIN);
  });
});
