import { ComponentFixture, TestBed } from '@angular/core/testing';
import { TenantManagerPage } from './tenant-manager.page';

describe('TenantManagerPage', () => {
  let component: TenantManagerPage;
  let fixture: ComponentFixture<TenantManagerPage>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [TenantManagerPage],
    }).compileComponents();

    fixture = TestBed.createComponent(TenantManagerPage);
    component = fixture.componentInstance;
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});