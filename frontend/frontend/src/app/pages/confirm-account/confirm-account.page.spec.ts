import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ConfirmAccountPage } from './confirm-account.page';

describe('ConfirmAccountPage', () => {
  let component: ConfirmAccountPage;
  let fixture: ComponentFixture<ConfirmAccountPage>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ConfirmAccountPage],
    }).compileComponents();

    fixture = TestBed.createComponent(ConfirmAccountPage);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
