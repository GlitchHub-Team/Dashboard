import { ComponentFixture, TestBed } from '@angular/core/testing';

import { DashboardGatewayTableComponent } from './dashboard-gateway-table.component';

describe('DashboardGatewayTableComponent', () => {
  let component: DashboardGatewayTableComponent;
  let fixture: ComponentFixture<DashboardGatewayTableComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [DashboardGatewayTableComponent],
    }).compileComponents();

    fixture = TestBed.createComponent(DashboardGatewayTableComponent);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
