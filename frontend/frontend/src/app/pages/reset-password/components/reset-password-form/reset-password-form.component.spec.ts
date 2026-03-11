import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ResetPasswordFormComponent } from './reset-password-form.component';

describe('ResetPasswordFormComponent', () => {
  let component: ResetPasswordFormComponent;
  let fixture: ComponentFixture<ResetPasswordFormComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ResetPasswordFormComponent],
    }).compileComponents();

    fixture = TestBed.createComponent(ResetPasswordFormComponent);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
