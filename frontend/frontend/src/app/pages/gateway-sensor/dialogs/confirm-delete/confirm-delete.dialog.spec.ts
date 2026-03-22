import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';
import { By } from '@angular/platform-browser';

import { ConfirmDeleteDialog } from './confirm-delete.dialog';

describe('ConfirmDeleteDialog (Unit)', () => {
  let fixture: ComponentFixture<ConfirmDeleteDialog>;
  let component: ConfirmDeleteDialog;
  let dialogRefMock: { close: ReturnType<typeof vi.fn> };

  const mockData = {
    title: 'Delete Gateway',
    message: 'Are you sure you want to delete the gateway "Gateway Alpha"?',
  };

  beforeEach(async () => {
    vi.resetAllMocks();
    dialogRefMock = { close: vi.fn() };

    await TestBed.configureTestingModule({
      imports: [ConfirmDeleteDialog],
      providers: [
        { provide: MatDialogRef, useValue: dialogRefMock },
        { provide: MAT_DIALOG_DATA, useValue: mockData },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(ConfirmDeleteDialog);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  const deleteBtn = () => fixture.debugElement.query(By.css('button[color="warn"]'));
  const cancelBtn = () =>
    fixture.debugElement
      .queryAll(By.css('button'))
      .find((btn) => btn.nativeElement.textContent.includes('Cancel'))!;

  describe('rendering', () => {
    it('should create and display title, message, and both buttons', () => {
      expect(component).toBeTruthy();
      expect(
        fixture.debugElement.query(By.css('[mat-dialog-title]')).nativeElement.textContent,
      ).toContain('Delete Gateway');
      expect(
        fixture.debugElement.query(By.css('mat-dialog-content p')).nativeElement.textContent,
      ).toContain('Are you sure you want to delete the gateway "Gateway Alpha"?');
      expect(cancelBtn()).toBeTruthy();
      expect(deleteBtn()).toBeTruthy();
      expect(deleteBtn().nativeElement.textContent).toContain('Delete');
    });
  });

  describe('fallback values', () => {
    it('should show fallback title and message when data is empty', async () => {
      await TestBed.resetTestingModule();
      await TestBed.configureTestingModule({
        imports: [ConfirmDeleteDialog],
        providers: [
          { provide: MatDialogRef, useValue: dialogRefMock },
          { provide: MAT_DIALOG_DATA, useValue: { title: '', message: '' } },
        ],
      }).compileComponents();
      fixture = TestBed.createComponent(ConfirmDeleteDialog);
      fixture.detectChanges();
      expect(
        fixture.debugElement.query(By.css('[mat-dialog-title]')).nativeElement.textContent,
      ).toContain('Confirm Deletion');
      expect(
        fixture.debugElement.query(By.css('mat-dialog-content p')).nativeElement.textContent,
      ).toContain('Are you sure you want to proceed?');
    });
  });

  describe('actions', () => {
    it('should close with true when Delete is clicked and false when Cancel is clicked', () => {
      deleteBtn().nativeElement.click();
      expect(dialogRefMock.close).toHaveBeenCalledWith(true);

      cancelBtn().nativeElement.click();
      expect(dialogRefMock.close).toHaveBeenCalledWith(false);
    });
  });
});
