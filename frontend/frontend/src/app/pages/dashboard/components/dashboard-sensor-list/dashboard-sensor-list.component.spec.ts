import { ComponentFixture, TestBed } from '@angular/core/testing';

import { DashboardSensorListComponent } from './dashboard-sensor-list.component';

describe('DashboardSensorListComponent', () => {
  let component: DashboardSensorListComponent;
  let fixture: ComponentFixture<DashboardSensorListComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [DashboardSensorListComponent],
    }).compileComponents();

    fixture = TestBed.createComponent(DashboardSensorListComponent);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
