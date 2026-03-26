import { TestBed } from '@angular/core/testing';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { of, throwError } from 'rxjs';

import { UserService } from './user.service';
import { UserApiClientService } from '../user-api-client/user-api-client.service';
import { UserRole } from '../../models/user/user-role.enum';
import { User } from '../../models/user/user.model';
import { UserConfig } from '../../models/user/user-config.model';
import { UserAdapter } from '../../adapters/user.adapter';

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

  describe('retrieveUser', () => {
    it('should retrieve users and update the list', () => {
      userApiMock.getUsers.mockReturnValue(of(rawPaginatedResponse));
      userAdapterMock.fromPaginatedDTO.mockReturnValue(rawPaginatedResponse);

      service.retrieveUser(UserRole.TENANT_ADMIN);

      expect(userApiMock.getUsers).toHaveBeenCalledWith(UserRole.TENANT_ADMIN, 0, 10, undefined);
      expect(userAdapterMock.fromPaginatedDTO).toHaveBeenCalledWith(rawPaginatedResponse);
      expect(service.loading()).toBe(false);
      expect(service.userList()).toEqual(mockUsers);
      expect(service.total()).toBe(2);
      expect(service.error()).toBeNull();
    });

    it('should retrieve users with tenantId when provided', () => {
      userApiMock.getUsers.mockReturnValue(of(rawPaginatedResponse));
      userAdapterMock.fromPaginatedDTO.mockReturnValue(rawPaginatedResponse);

      service.retrieveUser(UserRole.TENANT_USER, 'tenant-1');

      expect(userApiMock.getUsers).toHaveBeenCalledWith(UserRole.TENANT_USER, 0, 10, 'tenant-1');
    });

    it.each([
      { error: new Error('Failed to fetch'), expected: 'Failed to fetch' },
      { error: {} as Error, expected: 'Failed to load users' },
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
      userApiMock.getUsers.mockReturnValue(of(rawPaginatedResponse));
      userAdapterMock.fromPaginatedDTO.mockReturnValue(rawPaginatedResponse);

      service.changePage(2, 25, UserRole.TENANT_USER, 'tenant-1');

      expect(service.pageIndex()).toBe(2);
      expect(service.limit()).toBe(25);
      expect(userApiMock.getUsers).toHaveBeenCalledWith(UserRole.TENANT_USER, 2, 25, 'tenant-1');
    });
  });

  describe('addNewUser', () => {
    it('should call createUser and set loading false on success', () => {
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

    it('should set loading false on create error', () => {
      const error = new Error('Error creating');
      userApiMock.createUser.mockReturnValue(throwError(() => error));

      let thrownError: unknown;
      service.addNewUser(newUserConfig, UserRole.TENANT_ADMIN, 'tenant-1').subscribe({
        error: (err) => {
          thrownError = err;
        },
      });

      expect(userApiMock.createUser).toHaveBeenCalledWith(
        newUserConfig,
        UserRole.TENANT_ADMIN,
        'tenant-1',
      );
      expect(thrownError).toBe(error);
      expect(service.loading()).toBe(false);
    });
  });

  describe('removeUser', () => {
    it('should call deleteUser with id, role, and tenantId and set loading false on success', () => {
      const user: User = {
        id: '1',
        username: 'testuser',
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
        username: 'testuser',
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
