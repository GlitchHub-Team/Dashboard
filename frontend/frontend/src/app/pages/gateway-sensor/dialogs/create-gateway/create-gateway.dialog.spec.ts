import { ComponentFixture, TestBed } from '@angular/core/testing';

import { CreateGatewayDialog } from './create-gateway.dialog';

describe('CreateGatewayDialog', () => {
  let component: CreateGatewayDialog;
  let fixture: ComponentFixture<CreateGatewayDialog>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [CreateGatewayDialog],
    }).compileComponents();

    fixture = TestBed.createComponent(CreateGatewayDialog);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
