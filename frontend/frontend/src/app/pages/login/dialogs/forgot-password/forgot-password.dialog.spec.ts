import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ForgotPasswordDialog } from './forgot-password.dialog';

describe('ForgotPasswordDialog', () => {
  let component: ForgotPasswordDialog;
  let fixture: ComponentFixture<ForgotPasswordDialog>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ForgotPasswordDialog]
    })
    .compileComponents();

    fixture = TestBed.createComponent(ForgotPasswordDialog);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
