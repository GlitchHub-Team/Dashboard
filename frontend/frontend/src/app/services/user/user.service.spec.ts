import { TestBed } from '@angular/core/testing';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { of, throwError } from 'rxjs';

import { UserService } from './user.service';
import { UserApiClientService } from '../user-api-client/user-api-client.service';
import { UserRole } from '../../models/user/user-role.enum';
import { User } from '../../models/user/user.model';
import { UserConfig } from '../../models/user/user-config.model';
import { UserAdapter } from '../../adapters/user/user.adapter';
import { ApiError } from '../../models/api-error.model';

describe('UserService', () => {
  let service: UserService;

  const mockUsers: User[] = [
    {
      id: '1',
      username: 'admin',
      email: 'admin@test.com',
      role: UserRole.TENANT_ADMIN,
      tenantId: 'tenant-1',
    },
    {
      id: '2',
      username: 'user',
      email: 'user@test.com',
      role: UserRole.TENANT_USER,
      tenantId: 'tenant-1',
    },
  ];
  const newUser: User = {
    id: '3',
    username: 'newuser',
    email: 'new@test.com',
    role: UserRole.TENANT_USER,
    tenantId: 'tenant-1',
  };
  const newUserConfig: UserConfig = {
    email: 'new@test.com',
    username: 'newuser',
  };

  const userApiMock = {
    getUser: vi.fn(),
    getUsers: vi.fn(),
    createUser: vi.fn(),
    deleteUser: vi.fn(),
  };

  const userAdapterMock = {
    fromDTO: vi.fn(),
    fromPaginatedDTO: vi.fn(),
  };

  const rawPaginatedResponse = {
    count: 2,
    total: 2,
    users: mockUsers,
  };

  beforeEach(() => {
    vi.resetAllMocks();
    TestBed.configureTestingModule({
      providers: [
        UserService,
        { provide: UserApiClientService, useValue: userApiMock },
        { provide: UserAdapter, useValue: userAdapterMock },
      ],
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

  describe('getUser', () => {
    const rawDto = { id: '1', username: 'admin', email: 'admin@test.com' };
    const adaptedUser: User = mockUsers[0];

    beforeEach(() => {
      userApiMock.getUser = vi.fn();
    });

    it.each([
      ['with tenantId', 'tenant-1' as string | undefined],
      ['without tenantId', undefined as string | undefined],
    ])('should call API, adapt DTO, and return user (%s)', (_label, tenantId) => {
      userApiMock.getUser.mockReturnValue(of(rawDto));
      userAdapterMock.fromDTO.mockReturnValue(adaptedUser);

      let result: User | undefined;
      service.getUser('1', UserRole.TENANT_ADMIN, tenantId).subscribe((user) => {
        result = user;
      });

      expect(userApiMock.getUser).toHaveBeenCalledWith('1', UserRole.TENANT_ADMIN, tenantId);
      expect(userAdapterMock.fromDTO).toHaveBeenCalledWith(rawDto);
      expect(result).toEqual(adaptedUser);
      expect(service.loading()).toBe(false);
      expect(service.error()).toBeNull();
    });

    it.each([
      { error: { status: 500, message: 'Server error' } as ApiError, expected: 'Server error' },
      { error: { status: 500 } as ApiError, expected: 'Failed to load user' },
    ])('should set error "$expected", reset loading and propagate error', ({ error, expected }) => {
      userApiMock.getUser.mockReturnValue(throwError(() => error));

      let propagatedError: ApiError | undefined;
      service
        .getUser('1', UserRole.TENANT_ADMIN)
        .subscribe({ error: (err) => (propagatedError = err) });

      expect(service.error()).toBe(expected);
      expect(service.loading()).toBe(false);
      expect(propagatedError).toEqual(error);
    });
  });

  describe('retrieveUsers', () => {
    it.each([
      ['without tenantId', undefined as string | undefined, UserRole.TENANT_ADMIN],
      ['with tenantId', 'tenant-1' as string | undefined, UserRole.TENANT_USER],
    ])('should retrieve users and update state (%s)', (_label, tenantId, role) => {
      userApiMock.getUsers.mockReturnValue(of(rawPaginatedResponse));
      userAdapterMock.fromPaginatedDTO.mockReturnValue(rawPaginatedResponse);

      service.retrieveUsers(role, tenantId);

      expect(userApiMock.getUsers).toHaveBeenCalledWith(role, 0, 10, tenantId);
      expect(userAdapterMock.fromPaginatedDTO).toHaveBeenCalledWith(rawPaginatedResponse);
      expect(service.loading()).toBe(false);
      expect(service.userList()).toEqual(mockUsers);
      expect(service.total()).toBe(2);
      expect(service.error()).toBeNull();
    });

    it.each([
      { error: new Error('Failed to fetch'), expected: 'Failed to fetch' },
      { error: {} as Error, expected: 'Failed to load users' },
    ])('should handle retrieval errors', ({ error, expected }) => {
      userApiMock.getUsers.mockReturnValue(throwError(() => error));

      service.retrieveUsers(UserRole.TENANT_ADMIN);

      expect(service.loading()).toBe(false);
      expect(service.userList()).toEqual([]);
      expect(service.error()).toBe(expected);
    });
  });

  describe('changePage', () => {
    it('should update pagination and retrieve users with new page values', () => {
      userApiMock.getUsers.mockReturnValue(of(rawPaginatedResponse));
      userAdapterMock.fromPaginatedDTO.mockReturnValue(rawPaginatedResponse);

      service.changePage(2, 25, UserRole.TENANT_USER, 'tenant-1');

      expect(service.pageIndex()).toBe(2);
      expect(service.limit()).toBe(25);
      expect(userApiMock.getUsers).toHaveBeenCalledWith(UserRole.TENANT_USER, 2, 25, 'tenant-1');
    });
  });

  describe('addNewUser', () => {
    it('should create a user and emit adapted user', () => {
      userApiMock.createUser.mockReturnValue(of(newUser));
      userAdapterMock.fromDTO.mockReturnValue(newUser);

      let result: User | undefined;
      service.addNewUser(newUserConfig, UserRole.TENANT_ADMIN, 'tenant-1').subscribe((user) => {
        result = user;
      });

      expect(userApiMock.createUser).toHaveBeenCalledWith(
        newUserConfig,
        UserRole.TENANT_ADMIN,
        'tenant-1',
      );
      expect(userAdapterMock.fromDTO).toHaveBeenCalledWith(newUser);
      expect(result).toEqual(newUser);
      expect(service.loading()).toBe(false);
      expect(service.error()).toBeNull();
    });
  });

  describe('removeUser', () => {
    const user: User = {
      id: '1',
      username: 'testuser',
      email: 'test@test.com',
      role: UserRole.TENANT_USER,
      tenantId: 't1',
    };

    it('should delete a user and refetch current page', () => {
      userApiMock.deleteUser.mockReturnValue(of(void 0));
      userApiMock.getUsers.mockReturnValue(of(rawPaginatedResponse));
      userAdapterMock.fromPaginatedDTO.mockReturnValue(rawPaginatedResponse);

      let completed = false;
      service.removeUser(user).subscribe({
        complete: () => {
          completed = true;
        },
      });

      expect(userApiMock.deleteUser).toHaveBeenCalledWith('1', UserRole.TENANT_USER, 't1');
      expect(userApiMock.getUsers).toHaveBeenCalledWith(UserRole.TENANT_USER, 0, 10, 't1');
      expect(service.userList()).toEqual(mockUsers);
      expect(service.loading()).toBe(false);
      expect(service.error()).toBeNull();
      expect(completed).toBe(true);
    });

    it.each([
      {
        error: { status: 500, message: 'Failed to delete' } as ApiError,
        expected: 'Failed to delete',
      },
      { error: { status: 500 } as ApiError, expected: 'Failed to delete user' },
    ])('should handle delete errors', ({ error, expected }) => {
      userApiMock.deleteUser.mockReturnValue(throwError(() => error));

      service.removeUser(user).subscribe();

      expect(userApiMock.deleteUser).toHaveBeenCalledWith('1', UserRole.TENANT_USER, 't1');
      expect(service.loading()).toBe(false);
      expect(service.error()).toBe(expected);
    });
  });
});
