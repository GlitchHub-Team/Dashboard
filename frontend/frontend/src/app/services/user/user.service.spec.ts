import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { UserService } from './user.service';
import { UserApiClientService, UserConfig } from './user-api-client.service';
import { UserRole } from '../../models/user-role.enum';
import { User } from '../../models/user.model';

class MockUserApiClientService {
  getUsersResult = of<User[]>([
    { id: '1', email: 'admin@test.com', role: UserRole.TENANT_ADMIN, tenantId: 'tenant-1' },
    { id: '2', email: 'user@test.com', role: UserRole.TENANT_USER, tenantId: 'tenant-1' },
  ]);
  createUserResult = of<User>({ id: '3', email: 'new@test.com', role: UserRole.TENANT_USER, tenantId: 'tenant-1' });
  deleteUserResult = of<void>(undefined);

  getUsersCalled = false;
  getUsersRole: UserRole | undefined;

  getUsers(role?: UserRole) {
    this.getUsersCalled = true;
    this.getUsersRole = role;
    return this.getUsersResult;
  }

  createUser() {
    return this.createUserResult;
  }

  deleteUser() {
    return this.deleteUserResult;
  }
}

describe('UserService', () => {
  let service: UserService;
  let apiClient: MockUserApiClientService;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [
        UserService,
        { provide: UserApiClientService, useClass: MockUserApiClientService },
      ],
    });
    service = TestBed.inject(UserService);
    apiClient = TestBed.inject(UserApiClientService) as unknown as MockUserApiClientService;
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  describe('retrieveUser', () => {
    it('should retrieve users and update the list', () => {
      const mockUsers: User[] = [
        { id: '1', email: 'admin@test.com', role: UserRole.TENANT_ADMIN, tenantId: 'tenant-1' },
      ];
      apiClient.getUsersResult = of(mockUsers);

      service.retrieveUser(UserRole.TENANT_ADMIN);

      expect(apiClient.getUsersCalled).toBe(true);
      expect(apiClient.getUsersRole).toBe(UserRole.TENANT_ADMIN);
      expect(service.loading()).toBe(false);
      expect(service.userList().length).toBe(1);
      expect(service.userList()).toEqual(mockUsers);
      expect(service.error()).toBeNull();
    });

    it('should handle errors when retrieving users', () => {
      const error = new Error('Failed to fetch');
      apiClient.getUsersResult = throwError(() => error);

      service.retrieveUser();

      expect(apiClient.getUsersCalled).toBe(true);
      expect(service.loading()).toBe(false);
      expect(service.userList()).toEqual([]);
      expect(service.error()).toBe('Failed to fetch');
    });
  });

  describe('addNewUser', () => {
    it('should call createUser and manage loading state on success', () => {
      const newUserConfig: UserConfig = { email: 'new@test.com', role: UserRole.TENANT_USER };
      
      service.addNewUser(newUserConfig).subscribe({
        next: (user) => {
          expect(user.id).toBe('3');
          expect(service.loading()).toBe(false);
        },
      });
    });

    it('should set loading to false on error', () => {
      const newUserConfig: UserConfig = { email: 'new@test.com', role: UserRole.TENANT_USER };
      apiClient.createUserResult = throwError(() => new Error('Error creating'));

      service.addNewUser(newUserConfig).subscribe({
        error: () => {
          expect(service.loading()).toBe(false);
        },
      });
    });
  });

  describe('removeUser', () => {
    it('should call deleteUser and manage loading state on success', () => {
      service.removeUser('test@test.com').subscribe({
        next: () => {
          expect(service.loading()).toBe(false);
        },
      });
    });

    it('should set loading to false on error', () => {
      apiClient.deleteUserResult = throwError(() => new Error('Error deleting'));

      service.removeUser('test@test.com').subscribe({
        error: () => {
          expect(service.loading()).toBe(false);
        },
      });
    });
  });
});
