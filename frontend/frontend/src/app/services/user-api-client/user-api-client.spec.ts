import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';

import { UserApiClientService } from './user-api-client.service';
import { UserApiAdapter } from '../../adapters/user/user-api.adapter';
import { environment } from '../../../environments/environment';
import { UserBackend } from '../../models/user/user-backend.model';
import { UserConfig } from '../../models/user/user-config.model';
import { UserRole } from '../../models/user/user-role.enum';
import { PaginatedUserResponse } from '../../models/user/paginated-user-response.model';
import { User } from '../../models/user/user.model';

const GET_USERS_CASES = [
  {
    role: UserRole.SUPER_ADMIN,
    tenantId: undefined,
    page: 0,
    limit: 10,
    url: 'super_admins',
  },
  {
    role: UserRole.TENANT_ADMIN,
    tenantId: 'tenant-1',
    page: 2,
    limit: 25,
    url: 'tenant/tenant-1/tenant_admins',
  },
  {
    role: UserRole.TENANT_USER,
    tenantId: 'tenant-1',
    page: 1,
    limit: 20,
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

  const mockBackendUsers: UserBackend[] = [
    {
      user_id: 1,
      username: 'admin',
      email: 'admin@test.com',
      user_role: UserRole.TENANT_ADMIN,
      tenant_id: 'tenant-1',
    },
    {
      user_id: 2,
      username: 'user',
      email: 'user@test.com',
      user_role: UserRole.TENANT_USER,
      tenant_id: 'tenant-1',
    },
  ];

  const mockBackendPaginatedResponse: PaginatedUserResponse<UserBackend> = {
    count: 2,
    total: 2,
    users: mockBackendUsers,
  };

  const mockMappedUsers: User[] = [
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

  const mockMappedPaginatedResponse: PaginatedUserResponse<User> = {
    count: 2,
    total: 2,
    users: mockMappedUsers,
  };

  const userBackend: UserBackend = {
    user_id: 3,
    username: 'newuser',
    email: 'new@test.com',
    user_role: UserRole.TENANT_USER,
    tenant_id: 'tenant-1',
  };

  const mappedUser: User = {
    id: '3',
    username: 'newuser',
    email: 'new@test.com',
    role: UserRole.TENANT_USER,
    tenantId: 'tenant-1',
  };

  const mapperMock = {
    fromPaginatedDTO: vi.fn(),
    fromDTO: vi.fn(),
  };

  const expectRequest = (method: string, url: string) => {
    const req = httpMock.expectOne(url);
    expect(req.request.method).toBe(method);
    return req;
  };

  beforeEach(() => {
    vi.resetAllMocks();

    TestBed.configureTestingModule({
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        UserApiClientService,
        { provide: UserApiAdapter, useValue: mapperMock },
      ],
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
      'should send GET request for $role to $url, map through adapter, and return mapped response',
      ({ role, tenantId, page, limit, url }) => {
        mapperMock.fromPaginatedDTO.mockReturnValue(mockMappedPaginatedResponse);

        service.getUsers(role, page, limit, tenantId).subscribe((response) => {
          expect(response).toEqual(mockMappedPaginatedResponse);
          expect(response.users[0].id).toBe('1');
          expect(response.users[1].id).toBe('2');
        });

        const req = httpMock.expectOne(
          (request) =>
            request.url === `${apiUrl}/${url}` &&
            request.params.get('page') === `${page}` &&
            request.params.get('limit') === `${limit}`,
        );
        expect(req.request.method).toBe('GET');
        req.flush(mockBackendPaginatedResponse);

        expect(mapperMock.fromPaginatedDTO).toHaveBeenCalledWith(mockBackendPaginatedResponse);
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
        backendResponse: {
          user_id: 1,
          email: 'super@test.com',
          username: 'super',
          user_role: UserRole.SUPER_ADMIN,
          tenant_id: '',
        } satisfies UserBackend,
        mappedResponse: {
          id: '1',
          email: 'super@test.com',
          username: 'super',
          role: UserRole.SUPER_ADMIN,
          tenantId: '',
        } satisfies User,
      },
      {
        id: '3',
        role: UserRole.TENANT_USER,
        tenant_id: 'tenant-1',
        url: 'tenant/tenant-1/tenant_user/3',
        backendResponse: userBackend,
        mappedResponse: mappedUser,
      },
    ])('should GET $url, map through adapter, and return domain model', ({ id, role, tenant_id, url, backendResponse, mappedResponse }) => {
      mapperMock.fromDTO.mockReturnValue(mappedResponse);

      let result: User | undefined;
      service.getUser(id, role, tenant_id).subscribe((user) => {
        result = user;
      });

      const req = expectRequest('GET', `${apiUrl}/${url}`);
      req.flush(backendResponse);

      expect(mapperMock.fromDTO).toHaveBeenCalledWith(backendResponse);
      expect(result).toEqual(mappedResponse);
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
        backendResponse: {
          user_id: 10,
          email: 'super@test.com',
          username: 'super',
          user_role: UserRole.SUPER_ADMIN,
          tenant_id: '',
        } satisfies UserBackend,
        mappedResponse: {
          id: '10',
          email: 'super@test.com',
          username: 'super',
          role: UserRole.SUPER_ADMIN,
          tenantId: '',
        } satisfies User,
      },
      {
        config: {
          email: 'new@test.com',
          username: 'newuser',
        } satisfies UserConfig,
        role: UserRole.TENANT_USER,
        tenantId: 'tenant-1',
        url: 'tenant/tenant-1/tenant_user',
        backendResponse: userBackend,
        mappedResponse: mappedUser,
      },
    ])(
      'should POST to $url, map through adapter, and return domain model',
      ({ config, role, tenantId, url, backendResponse, mappedResponse }) => {
        mapperMock.fromDTO.mockReturnValue(mappedResponse);

        let result: User | undefined;
        service.createUser(config, role, tenantId).subscribe((user) => {
          result = user;
        });

        const req = expectRequest('POST', `${apiUrl}/${url}`);
        expect(req.request.body).toEqual(config);
        req.flush(backendResponse);

        expect(mapperMock.fromDTO).toHaveBeenCalledWith(backendResponse);
        expect(result).toEqual(mappedResponse);
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