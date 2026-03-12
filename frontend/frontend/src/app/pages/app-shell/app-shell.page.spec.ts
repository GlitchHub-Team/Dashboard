import { ComponentFixture, TestBed } from '@angular/core/testing';

import { AppShellPage } from './app-shell.page';

describe('AppShellPage', () => {
  let component: AppShellPage;
  let fixture: ComponentFixture<AppShellPage>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [AppShellPage],
    }).compileComponents();

    fixture = TestBed.createComponent(AppShellPage);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
