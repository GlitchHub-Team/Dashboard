import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';

import { UserApiClientService } from './user-api-client.service';
import { environment } from '../../../environments/environment';
import { UserBackend } from '../../models/user/user-backend.model';
import { UserConfig } from '../../models/user/user-config.model';
import { UserRole } from '../../models/user/user-role.enum';
import { PaginatedUserResponse } from '../../models/user/paginated-user-response.model';

const GET_USERS_CASES = [
  {
    role: UserRole.SUPER_ADMIN,
    tenantId: undefined,
    page: 0,
    size: 10,
    url: 'super_admins',
  },
  {
    role: UserRole.TENANT_ADMIN,
    tenantId: 'tenant-1',
    page: 2,
    size: 25,
    url: 'tenant/tenant-1/tenant_admins',
  },
  {
    role: UserRole.TENANT_USER,
    tenantId: 'tenant-1',
    page: 1,
    size: 20,
    url: 'tenant/tenant-1/tenant_users',
  },
] as const;

const TENANT_ID_ERROR_CASES = [
  {
    role: UserRole.TENANT_ADMIN,
    error: 'tenantId is required for TENANT_ADMIN',
  },
  {
    role: UserRole.TENANT_USER,
    error: 'tenantId is required for TENANT_USER',
  },
] as const;

describe('UserApiClientService', () => {
  let service: UserApiClientService;
  let httpMock: HttpTestingController;

  const apiUrl = `${environment.apiUrl}`;

  const paginatedUserResponse: PaginatedUserResponse<UserBackend> = {
    count: 2,
    total: 2,
    users: [
      {
        user_id: '1',
        username: 'admin',
        email: 'admin@test.com',
        user_role: UserRole.TENANT_ADMIN,
        tenant_id: 'tenant-1',
      },
      {
        user_id: '2',
        username: 'user',
        email: 'user@test.com',
        user_role: UserRole.TENANT_USER,
        tenant_id: 'tenant-1',
      },
    ],
  };

  const userBackend: UserBackend = {
    user_id: '3',
    username: 'newuser',
    email: 'new@test.com',
    user_role: UserRole.TENANT_USER,
    tenant_id: 'tenant-1',
  };

  const expectRequest = (method: string, url: string) => {
    const req = httpMock.expectOne(url);
    expect(req.request.method).toBe(method);
    return req;
  };

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });

    service = TestBed.inject(UserApiClientService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  describe('getUsers', () => {
    it.each(GET_USERS_CASES)(
      'should send GET request for $role to $url',
      ({ role, tenantId, page, size, url }) => {
        service.getUsers(role, page, size, tenantId).subscribe((response) => {
          expect(response).toEqual(paginatedUserResponse);
        });

        const req = httpMock.expectOne(
          (request) =>
            request.url === `${apiUrl}/${url}` &&
            request.params.get('page') === `${page}` &&
            request.params.get('size') === `${size}`,
        );
        expect(req.request.method).toBe('GET');
        req.flush(paginatedUserResponse);
      },
    );

    it.each(TENANT_ID_ERROR_CASES)(
      'should throw when tenantId is missing for $role',
      ({ role, error }) => {
        expect(() => service.getUsers(role, 0, 10)).toThrow(error);
      },
    );
  });

  describe('getUser', () => {
    it.each([
      {
        id: '1',
        role: UserRole.SUPER_ADMIN,
        tenant_id: undefined,
        url: 'super_admin/1',
        response: {
          user_id: '1',
          email: 'super@test.com',
          username: 'super',
          user_role: UserRole.SUPER_ADMIN,
          tenant_id: '',
        } satisfies UserBackend,
      },
      {
        id: '3',
        role: UserRole.TENANT_USER,
        tenant_id: 'tenant-1',
        url: 'tenant/tenant-1/tenant_user/3',
        response: userBackend,
      },
    ])('should GET $url and return dto response', ({ id, role, tenant_id, url, response }) => {
      let result: UserBackend | undefined;
      service.getUser(id, role, tenant_id).subscribe((user) => {
        result = user;
      });

      const req = expectRequest('GET', `${apiUrl}/${url}`);
      req.flush(response);

      expect(result).toEqual(response);
    });
  });

  describe('createUser', () => {
    it.each([
      {
        config: {
          email: 'super@test.com',
          username: 'super',
        } satisfies UserConfig,
        role: UserRole.SUPER_ADMIN,
        tenantId: undefined,
        url: 'super_admin',
        response: {
          user_id: '10',
          email: 'super@test.com',
          username: 'super',
          user_role: UserRole.SUPER_ADMIN,
          tenant_id: '',
        } satisfies UserBackend,
      },
      {
        config: {
          email: 'new@test.com',
          username: 'newuser',
        } satisfies UserConfig,
        role: UserRole.TENANT_USER,
        tenantId: 'tenant-1',
        url: 'tenant/tenant-1/tenant_user',
        response: userBackend,
      },
    ])(
      'should POST to $url and return dto response',
      ({ config, role, tenantId, url, response }) => {
        let result: UserBackend | undefined;
        service.createUser(config, role, tenantId).subscribe((user) => {
          result = user;
        });

        const req = expectRequest('POST', `${apiUrl}/${url}`);
        expect(req.request.body).toEqual(config);
        req.flush(response);

        expect(result).toEqual(response);
      },
    );

    it('should throw when tenantId is missing for tenant-scoped creation', () => {
      const config: UserConfig = {
        email: 'new@test.com',
        username: 'newuser',
      };

      expect(() => service.createUser(config, UserRole.TENANT_ADMIN, undefined)).toThrow(
        'tenantId is required for TENANT_ADMIN',
      );
    });
  });

  describe('deleteUser', () => {
    it.each([
      {
        id: '1',
        role: UserRole.SUPER_ADMIN,
        tenantId: undefined,
        url: 'super_admin/1',
      },
      {
        id: '1',
        role: UserRole.TENANT_ADMIN,
        tenantId: 'tenant-1',
        url: 'tenant/tenant-1/tenant_admin/1',
      },
    ])('should send DELETE request to $url', ({ id, role, tenantId, url }) => {
      let result: void | null | undefined;

      service.deleteUser(id, role, tenantId).subscribe((response) => {
        result = response;
      });

      const req = expectRequest('DELETE', `${apiUrl}/${url}`);
      req.flush(null);

      expect(result).toBeNull();
    });

    it('should throw when tenantId is missing for tenant-scoped deletion', () => {
      expect(() => service.deleteUser('1', UserRole.TENANT_USER, undefined)).toThrow(
        'tenantId is required for TENANT_USER',
      );
    });
  });
});
