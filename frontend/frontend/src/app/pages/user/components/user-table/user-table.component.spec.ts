import { ComponentFixture, TestBed } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { PageEvent } from '@angular/material/paginator';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { UserTableComponent } from './user-table.component';
import { User } from '../../../../models/user/user.model';
import { UserRole } from '../../../../models/user/user-role.enum';

describe('UserTableComponent (Unit)', () => {
  let component: UserTableComponent;
  let fixture: ComponentFixture<UserTableComponent>;

  const mockUsers: User[] = [
    {
      id: 'user-1',
      username: 'alice',
      email: 'alice@example.com',
      role: UserRole.TENANT_USER,
      tenantId: 'tenant-1',
    },
    {
      id: 'user-2',
      username: 'bob',
      email: 'bob@example.com',
      role: UserRole.TENANT_ADMIN,
      tenantId: 'tenant-2',
    },
  ];

  const setInput = (key: string, value: unknown) => {
    fixture.componentRef.setInput(key, value);
    fixture.detectChanges();
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [UserTableComponent],
    }).compileComponents();

    fixture = TestBed.createComponent(UserTableComponent);
    component = fixture.componentInstance;

    fixture.componentRef.setInput('users', mockUsers);
    fixture.componentRef.setInput('total', mockUsers.length);
    fixture.componentRef.setInput('pageIndex', 0);
    fixture.componentRef.setInput('limit', 10);
    fixture.detectChanges();
  });

  describe('initial state', () => {
    it('should create with provided inputs', () => {
      expect(component).toBeTruthy();
      expect(component.users()).toEqual(mockUsers);
      expect(component.total()).toBe(2);
      expect(component.pageIndex()).toBe(0);
      expect(component.limit()).toBe(10);
    });
  });

  describe('loading state', () => {
    it('should render only spinner when loading is true', () => {
      setInput('loading', true);

      expect(fixture.debugElement.query(By.css('mat-spinner'))).toBeTruthy();
      expect(fixture.debugElement.query(By.css('.empty-state'))).toBeFalsy();
      expect(fixture.debugElement.query(By.css('mat-table'))).toBeFalsy();
      expect(fixture.debugElement.query(By.css('mat-paginator'))).toBeFalsy();
    });
  });

  describe('empty state', () => {
    it('should render empty state when users array is empty', () => {
      setInput('loading', false);
      setInput('users', []);
      setInput('total', 0);

      const empty = fixture.debugElement.query(By.css('.empty-state'));
      expect(empty).toBeTruthy();
      expect(empty.query(By.css('mat-icon')).nativeElement.textContent).toContain('group');
      expect(empty.query(By.css('p')).nativeElement.textContent).toContain('Nessun utente disponibile');
      expect(fixture.debugElement.query(By.css('mat-spinner'))).toBeFalsy();
      expect(fixture.debugElement.query(By.css('mat-table'))).toBeFalsy();
      expect(fixture.debugElement.query(By.css('mat-paginator'))).toBeFalsy();
    });
  });

  describe('table with data', () => {
    it('should render table and paginator when users are present', () => {
      setInput('loading', false);

      expect(fixture.debugElement.query(By.css('mat-spinner'))).toBeFalsy();
      expect(fixture.debugElement.query(By.css('.empty-state'))).toBeFalsy();
      expect(fixture.debugElement.query(By.css('mat-table'))).toBeTruthy();
      expect(fixture.debugElement.query(By.css('mat-header-row'))).toBeTruthy();
      expect(fixture.debugElement.queryAll(By.css('mat-row')).length).toBe(2);
      expect(fixture.debugElement.query(By.css('mat-paginator'))).toBeTruthy();
    });

    it('should render user values in table cells', () => {
      setInput('loading', false);
      setInput('targetRole', UserRole.TENANT_ADMIN);

      const cellTexts = fixture.debugElement
        .queryAll(By.css('mat-cell'))
        .map((cell) => cell.nativeElement.textContent.trim());

      expect(cellTexts).toEqual(expect.arrayContaining(['alice', 'bob']));
      expect(cellTexts).toEqual(expect.arrayContaining(['alice@example.com', 'bob@example.com']));
      expect(cellTexts).toEqual(expect.arrayContaining(['tenant-1', 'tenant-2']));
    });
  });

  describe('outputs', () => {
    it('should emit deleteRequested when delete button is clicked', () => {
      setInput('loading', false);
      const spy = vi.fn();
      component.deleteRequested.subscribe(spy);

      const deleteButtons = fixture.debugElement.queryAll(By.css('mat-cell button'));
      expect(deleteButtons.length).toBe(2);

      const stopPropagation = vi.fn();
      deleteButtons[0].triggerEventHandler('click', { stopPropagation });

      expect(spy).toHaveBeenCalledWith(mockUsers[0]);
      expect(stopPropagation).toHaveBeenCalledTimes(1);
    });

    it('should hide delete button for the row matching currentUserId and currentUserRole', () => {
      setInput('loading', false);
      setInput('currentUserId', 'user-1');
      setInput('currentUserRole', UserRole.TENANT_USER);
      fixture.detectChanges();

      const deleteButtons = fixture.debugElement.queryAll(By.css('mat-cell button'));
      expect(deleteButtons.length).toBe(1);

      const spy = vi.fn();
      component.deleteRequested.subscribe(spy);
      const stopPropagation = vi.fn();
      deleteButtons[0].triggerEventHandler('click', { stopPropagation });
      expect(spy).toHaveBeenCalledWith(mockUsers[1]);
    });

    it('should show delete button when id matches but role differs', () => {
      setInput('loading', false);
      setInput('currentUserId', 'user-1');
      setInput('currentUserRole', UserRole.TENANT_ADMIN); // same id, different role
      fixture.detectChanges();

      const deleteButtons = fixture.debugElement.queryAll(By.css('mat-cell button'));
      expect(deleteButtons.length).toBe(2); // both buttons visible
    });

    it('should emit pageChange on paginator page event', () => {
      setInput('loading', false);
      const spy = vi.fn();
      component.pageChange.subscribe(spy);

      const event: PageEvent = { pageIndex: 1, pageSize: 25, length: 100 };
      fixture.debugElement.query(By.css('mat-paginator')).triggerEventHandler('page', event);

      expect(spy).toHaveBeenCalledWith(event);
    });
  });
});
