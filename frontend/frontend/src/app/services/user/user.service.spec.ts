import { TestBed } from '@angular/core/testing';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { of, throwError } from 'rxjs';

import { UserService } from './user.service';
import { UserApiClientService, UserConfig } from './user-api-client.service';
import { UserRole } from '../../models/user-role.enum';
import { User } from '../../models/user.model';

describe('UserService', () => {
  let service: UserService;

  const mockUsers: User[] = [
    { id: '1', email: 'admin@test.com', role: UserRole.TENANT_ADMIN, tenantId: 'tenant-1' },
    { id: '2', email: 'user@test.com', role: UserRole.TENANT_USER, tenantId: 'tenant-1' },
  ];
  const newUser: User = {
    id: '3',
    email: 'new@test.com',
    role: UserRole.TENANT_USER,
    tenantId: 'tenant-1',
  };
  const newUserConfig: UserConfig = {
    email: 'new@test.com',
    role: UserRole.TENANT_USER,
  };

  const userApiMock = {
    getUsers: vi.fn(),
    createUser: vi.fn(),
    deleteUser: vi.fn(),
  };

  beforeEach(() => {
    vi.resetAllMocks();
    TestBed.configureTestingModule({
      providers: [UserService, { provide: UserApiClientService, useValue: userApiMock }],
    });
    service = TestBed.inject(UserService);
  });

  it('should be created with default state', () => {
    expect(service).toBeTruthy();
    expect(service.loading()).toBe(false);
    expect(service.error()).toBeNull();
    expect(service.userList()).toEqual([]);
    expect(service.total()).toBe(0);
    expect(service.pageIndex()).toBe(0);
    expect(service.limit()).toBe(10);
  });

  describe('retrieveUser', () => {
    it('should retrieve users and update the list', () => {
      userApiMock.getUsers.mockReturnValue(of({ items: mockUsers, totalCount: 2 }));

      service.retrieveUser(UserRole.TENANT_ADMIN);

      expect(userApiMock.getUsers).toHaveBeenCalledWith(UserRole.TENANT_ADMIN, undefined, 0, 10);
      expect(service.loading()).toBe(false);
      expect(service.userList()).toEqual(mockUsers);
      expect(service.total()).toBe(2);
      expect(service.error()).toBeNull();
    });

    it('should retrieve users with tenantId when provided', () => {
      userApiMock.getUsers.mockReturnValue(of({ items: mockUsers, totalCount: 2 }));

      service.retrieveUser(UserRole.TENANT_USER, 'tenant-1');

      expect(userApiMock.getUsers).toHaveBeenCalledWith(UserRole.TENANT_USER, 'tenant-1', 0, 10);
    });

    it.each([
      { error: new Error('Failed to fetch'), expected: 'Failed to fetch' },
      { error: { message: '' } as Error, expected: 'Errore nel recupero degli utenti' },
    ])('should handle retrieval errors', ({ error, expected }) => {
      userApiMock.getUsers.mockReturnValue(throwError(() => error));

      service.retrieveUser(UserRole.TENANT_ADMIN);

      expect(service.loading()).toBe(false);
      expect(service.userList()).toEqual([]);
      expect(service.error()).toBe(expected);
    });
  });

  describe('changePage', () => {
    it('should update pagination and retrieve users with new page values', () => {
      userApiMock.getUsers.mockReturnValue(of({ items: mockUsers, totalCount: 2 }));

      service.changePage(2, 25, UserRole.TENANT_USER, 'tenant-1');

      expect(service.pageIndex()).toBe(2);
      expect(service.limit()).toBe(25);
      expect(userApiMock.getUsers).toHaveBeenCalledWith(UserRole.TENANT_USER, 'tenant-1', 2, 25);
    });
  });

  describe('addNewUser', () => {
    it('should call createUser and set loading false on success', () => {
      userApiMock.createUser.mockReturnValue(of(newUser));

      let result: User | undefined;
      service.addNewUser(newUserConfig, 'tenant-1').subscribe((user) => {
        result = user;
      });

      expect(userApiMock.createUser).toHaveBeenCalledWith(newUserConfig, 'tenant-1');
      expect(result).toEqual(newUser);
      expect(service.loading()).toBe(false);
      expect(service.error()).toBeNull();
    });

    it('should set loading false on create error', () => {
      const error = new Error('Error creating');
      userApiMock.createUser.mockReturnValue(throwError(() => error));

      let thrownError: unknown;
      service.addNewUser(newUserConfig, 'tenant-1').subscribe({
        error: (err) => {
          thrownError = err;
        },
      });

      expect(userApiMock.createUser).toHaveBeenCalledWith(newUserConfig, 'tenant-1');
      expect(thrownError).toBe(error);
      expect(service.loading()).toBe(false);
    });
  });

  describe('removeUser', () => {
    it('should call deleteUser with id, role, and tenantId and set loading false on success', () => {
      const user: User = {
        id: '1',
        email: 'test@test.com',
        role: UserRole.TENANT_USER,
        tenantId: 't1',
      };
      userApiMock.deleteUser.mockReturnValue(of(void 0));

      let completed = false;
      service.removeUser(user).subscribe({
        complete: () => {
          completed = true;
        },
      });

      expect(userApiMock.deleteUser).toHaveBeenCalledWith('1', UserRole.TENANT_USER, 't1');
      expect(service.loading()).toBe(false);
      expect(completed).toBe(true);
    });

    it('should set loading false on delete error', () => {
      const error = new Error('Error deleting');
      const user: User = {
        id: '1',
        email: 'test@test.com',
        role: UserRole.TENANT_USER,
        tenantId: 't1',
      };
      userApiMock.deleteUser.mockReturnValue(throwError(() => error));

      let thrownError: unknown;
      service.removeUser(user).subscribe({
        error: (err) => {
          thrownError = err;
        },
      });

      expect(userApiMock.deleteUser).toHaveBeenCalledWith('1', UserRole.TENANT_USER, 't1');
      expect(thrownError).toBe(error);
      expect(service.loading()).toBe(false);
    });
  });
});
