import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ConfirmAccountFormComponent } from './confirm-account-form.component';

describe('ConfirmAccountFormComponent', () => {
  let component: ConfirmAccountFormComponent;
  let fixture: ComponentFixture<ConfirmAccountFormComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ConfirmAccountFormComponent],
    }).compileComponents();

    fixture = TestBed.createComponent(ConfirmAccountFormComponent);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
