import { ComponentFixture, TestBed } from '@angular/core/testing';

import { RealTimeChartComponent } from './real-time-chart.component';

describe('RealTimeChartComponent', () => {
  let component: RealTimeChartComponent;
  let fixture: ComponentFixture<RealTimeChartComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [RealTimeChartComponent],
    }).compileComponents();

    fixture = TestBed.createComponent(RealTimeChartComponent);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
