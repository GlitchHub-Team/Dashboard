import { ComponentFixture, TestBed } from '@angular/core/testing';

import { CreateSensorDialog } from './create-sensor.dialog';

describe('CreateSensorDialog', () => {
  let component: CreateSensorDialog;
  let fixture: ComponentFixture<CreateSensorDialog>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [CreateSensorDialog],
    }).compileComponents();

    fixture = TestBed.createComponent(CreateSensorDialog);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
