import { ComponentFixture, TestBed } from '@angular/core/testing';

import { DashboardSensorRowComponent } from './dashboard-sensor-row.component';

describe('DashboardSensorRowComponent', () => {
  let component: DashboardSensorRowComponent;
  let fixture: ComponentFixture<DashboardSensorRowComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [DashboardSensorRowComponent],
    }).compileComponents();

    fixture = TestBed.createComponent(DashboardSensorRowComponent);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
