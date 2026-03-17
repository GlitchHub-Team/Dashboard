import { ComponentFixture, TestBed } from '@angular/core/testing';
import { UserManagerPage } from './user-manager.page';
import { ActivatedRoute } from '@angular/router';
import { of } from 'rxjs';
import { UserService } from '../../services/user/user.service';
import { MatDialog } from '@angular/material/dialog';
import { signal } from '@angular/core';

describe('UserManagerPage', () => {
  let component: UserManagerPage;
  let fixture: ComponentFixture<UserManagerPage>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [UserManagerPage],
      providers: [
        {
          provide: ActivatedRoute,
          useValue: {
            data: of({ userManagerContext: { title: 'Test', role: 'Tenant Admin' } })
          }
        }
      ]
    })
    .overrideProvider(UserService, { useValue: { userList: signal([]), loading: signal(false), retrieveUser: () => {} } })
    .overrideProvider(MatDialog, { useValue: { open: () => ({ afterClosed: () => of(true) }) } })
    .compileComponents();

    fixture = TestBed.createComponent(UserManagerPage);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
