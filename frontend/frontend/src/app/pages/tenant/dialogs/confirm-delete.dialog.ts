import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatDialogModule, MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';
import { MatButtonModule } from '@angular/material/button';

export interface ConfirmDeleteDialogData {
  title?: string;
  message?: string;
}

@Component({
  selector: 'app-confirm-delete-dialog',
  standalone: true,
  imports: [CommonModule, MatDialogModule, MatButtonModule],
  template: `
    <h2 mat-dialog-title>{{ data?.title || 'Conferma Eliminazione' }}</h2>
    <mat-dialog-content>
      <p>{{ data?.message || 'Sei sicuro di voler procedere?' }}</p>
    </mat-dialog-content>
    <mat-dialog-actions align="end">
      <button mat-button (click)="onCancel()">Annulla</button>
      <button mat-raised-button color="warn" (click)="onConfirm()">Elimina</button>
    </mat-dialog-actions>
  `,
})
export class ConfirmDeleteDialog {
  public dialogRef = inject(MatDialogRef<ConfirmDeleteDialog>);
  public data = inject<ConfirmDeleteDialogData | null>(MAT_DIALOG_DATA);

  onConfirm(): void {
    this.dialogRef.close(true);
  }

  onCancel(): void {
    this.dialogRef.close(false);
  }
}