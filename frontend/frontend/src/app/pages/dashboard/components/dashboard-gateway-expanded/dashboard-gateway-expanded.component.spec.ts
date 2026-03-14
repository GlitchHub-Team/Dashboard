import { ComponentFixture, TestBed } from '@angular/core/testing';

import { DashboardGatewayExpandedComponent } from './dashboard-gateway-expanded.component';

describe('DashboardGatewayExpandedComponent', () => {
  let component: DashboardGatewayExpandedComponent;
  let fixture: ComponentFixture<DashboardGatewayExpandedComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [DashboardGatewayExpandedComponent],
    }).compileComponents();

    fixture = TestBed.createComponent(DashboardGatewayExpandedComponent);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
