import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';

import { UserApiClientService, UserConfig } from './user-api-client.service';
import { environment } from '../../../environments/environment';
import { UserDataAdapter, RawPaginatedResponse } from '../../adapters/user-data.adapter';
import { User } from '../../models/user.model';
import { RawUserConfig } from '../../models/raw-user-config.model';
import { UserRole } from '../../models/user-role.enum';

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

  const adapterMock = {
    adaptPaginated: vi.fn(),
    adapt: vi.fn(),
  };

  const rawPaginatedResponse: RawPaginatedResponse = {
    items: [
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
    ],
    totalCount: 2,
  };

  const adaptedPaginatedResponse: { items: User[]; totalCount: number } = {
    items: [
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
    ],
    totalCount: 2,
  };

  const rawUser: RawUserConfig = {
    id: '3',
    username: 'newuser',
    email: 'new@test.com',
    role: UserRole.TENANT_USER,
    tenantId: 'tenant-1',
  };

  const adaptedUser: User = {
    id: '3',
    email: 'new@test.com',
    username: 'newuser',
    role: UserRole.TENANT_USER,
    tenantId: 'tenant-1',
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
        { provide: UserDataAdapter, useValue: adapterMock },
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
      'should send GET request for $role to $url with page=$page and size=$size',
      ({ role, tenantId, page, size, url }) => {
        adapterMock.adaptPaginated.mockReturnValue(adaptedPaginatedResponse);

        service.getUsers(role, tenantId, page, size).subscribe((response) => {
          expect(response).toEqual(adaptedPaginatedResponse);
        });

        const req = httpMock.expectOne(
          (request) =>
            request.url === `${apiUrl}/${url}` &&
            request.params.get('page') === `${page}` &&
            request.params.get('size') === `${size}`,
        );
        expect(req.request.method).toBe('GET');
        req.flush(rawPaginatedResponse);
      },
    );

    it('should map the paginated response through UserDataAdapter', () => {
      adapterMock.adaptPaginated.mockReturnValue(adaptedPaginatedResponse);

      let result: { items: User[]; totalCount: number } | undefined;
      service.getUsers(UserRole.SUPER_ADMIN).subscribe((response) => {
        result = response;
      });

      const req = httpMock.expectOne(`${apiUrl}/super_admins?page=0&size=10`);
      req.flush(rawPaginatedResponse);

      expect(adapterMock.adaptPaginated).toHaveBeenCalledWith(rawPaginatedResponse);
      expect(result).toEqual(adaptedPaginatedResponse);
    });

    it.each(TENANT_ID_ERROR_CASES)(
      'should throw when tenantId is missing for $role',
      ({ role, error }) => {
        expect(() => service.getUsers(role)).toThrow(error);
      },
    );
  });

  describe('getUser', () => {
    it.each([
      {
        id: '1',
        role: UserRole.SUPER_ADMIN,
        tenantId: undefined,
        url: 'super_admin/1',
        raw: {
          id: '1',
          email: 'super@test.com',
          username: 'super',
          role: UserRole.SUPER_ADMIN,
          tenantId: '',
        } satisfies RawUserConfig,
        adapted: {
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
        tenantId: 'tenant-1',
        url: 'tenant/tenant-1/tenant_user/3',
        raw: rawUser,
        adapted: adaptedUser,
      },
    ])('should GET $url and map the response', ({ id, role, tenantId, url, raw, adapted }) => {
      adapterMock.adapt.mockReturnValue(adapted);

      let result: User | undefined;
      service.getUser(id, role, tenantId).subscribe((user) => {
        result = user;
      });

      const req = expectRequest('GET', `${apiUrl}/${url}`);
      req.flush(raw);

      expect(adapterMock.adapt).toHaveBeenCalledWith(raw);
      expect(result).toEqual(adapted);
    });
  });

  describe('createUser', () => {
    it.each([
      {
        config: {
          email: 'super@test.com',
          role: UserRole.SUPER_ADMIN,
        } satisfies UserConfig,
        tenantId: undefined,
        url: 'super_admin',
        raw: {
          id: '10',
          email: 'super@test.com',
          username: 'super',
          role: UserRole.SUPER_ADMIN,
          tenantId: '',
        } satisfies RawUserConfig,
        adapted: {
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
          role: UserRole.TENANT_USER,
        } satisfies UserConfig,
        tenantId: 'tenant-1',
        url: 'tenant/tenant-1/tenant_user',
        raw: rawUser,
        adapted: adaptedUser,
      },
    ])('should POST to $url and map the response', ({ config, tenantId, url, raw, adapted }) => {
      adapterMock.adapt.mockReturnValue(adapted);

      let result: User | undefined;
      service.createUser(config, tenantId).subscribe((user) => {
        result = user;
      });

      const req = expectRequest('POST', `${apiUrl}/${url}`);
      expect(req.request.body).toEqual(config);
      req.flush(raw);

      expect(adapterMock.adapt).toHaveBeenCalledWith(raw);
      expect(result).toEqual(adapted);
    });

    it('should throw when tenantId is missing for tenant-scoped creation', () => {
      const config: UserConfig = {
        email: 'new@test.com',
        role: UserRole.TENANT_ADMIN,
      };

      expect(() => service.createUser(config)).toThrow('tenantId is required for TENANT_ADMIN');
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
      expect(() => service.deleteUser('1', UserRole.TENANT_USER)).toThrow(
        'tenantId is required for TENANT_USER',
      );
    });
  });
});
