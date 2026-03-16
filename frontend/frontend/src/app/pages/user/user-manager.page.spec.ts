import { ComponentFixture, TestBed } from '@angular/core/testing';
import { UserManagerPage } from './user-manager.page';

describe('UserManagerPage', () => {
  let component: UserManagerPage;
  let fixture: ComponentFixture<UserManagerPage>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [UserManagerPage]
    })
    .compileComponents();

    fixture = TestBed.createComponent(UserManagerPage);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
