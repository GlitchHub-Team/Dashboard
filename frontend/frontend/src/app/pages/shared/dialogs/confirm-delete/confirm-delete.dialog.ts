import { Component, inject } from '@angular/core';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatButtonModule } from '@angular/material/button';

@Component({
  selector: 'app-confirm-delete',
  imports: [MatDialogModule, MatButtonModule],
  templateUrl: './confirm-delete.dialog.html',
  styleUrl: './confirm-delete.dialog.css',
})
export class ConfirmDeleteDialog {
  private readonly dialogRef = inject(MatDialogRef<ConfirmDeleteDialog>);
  protected readonly data = inject<{ title: string; message: string }>(MAT_DIALOG_DATA);

  protected onConfirm(): void {
    this.dialogRef.close(true);
  }

  protected onCancel(): void {
    this.dialogRef.close(false);
  }
}
