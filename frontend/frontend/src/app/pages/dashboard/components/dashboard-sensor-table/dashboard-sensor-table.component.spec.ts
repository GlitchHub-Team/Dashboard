import { ComponentFixture, TestBed } from '@angular/core/testing';

import { DashboardSensorTableComponent } from './dashboard-sensor-table.component';

describe('DashboardSensorTableComponent', () => {
  let component: DashboardSensorTableComponent;
  let fixture: ComponentFixture<DashboardSensorTableComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [DashboardSensorTableComponent],
    }).compileComponents();

    fixture = TestBed.createComponent(DashboardSensorTableComponent);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
