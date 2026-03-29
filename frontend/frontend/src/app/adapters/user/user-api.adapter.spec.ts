import { describe, it, expect } from 'vitest';
import { UserApiAdapter } from './user-api.adapter';
import { UserBackend } from '../../models/user/user-backend.model';
import { UserRole } from '../../models/user/user-role.enum';

describe('UserApiAdapter', () => {
  const adapter = new UserApiAdapter();

  const dto: UserBackend = {
    user_id: 1,
    username: 'john',
    email: 'john@example.com',
    user_role: 'tenant_admin',
    tenant_id: 'tenant-01',
  };

  describe('fromDTO', () => {
    it.each([
      { field: 'id', expected: '1' },
      { field: 'username', expected: 'john' },
      { field: 'email', expected: 'john@example.com' },
      { field: 'role', expected: UserRole.TENANT_ADMIN },
      { field: 'tenantId', expected: 'tenant-01' },
    ] as const)('should map $field correctly', ({ field, expected }) => {
      expect(adapter.fromDTO(dto)[field]).toEqual(expected);
    });

    it.each([
      ['super_admin', UserRole.SUPER_ADMIN],
      ['tenant_admin', UserRole.TENANT_ADMIN],
      ['tenant_user', UserRole.TENANT_USER],
    ])('should map role "%s" correctly', (backendRole, expected) => {
      expect(adapter.fromDTO({ ...dto, user_role: backendRole }).role).toBe(expected);
    });

    it('should map empty tenant_id to empty string', () => {
      expect(adapter.fromDTO({ ...dto, tenant_id: '' }).tenantId).toBe('');
      expect(adapter.fromDTO({ ...dto, tenant_id: undefined }).tenantId).toBe('');
    });
  });

  describe('fromPaginatedDTO', () => {
    it('should map count, total and all users', () => {
      const response = {
        count: 2,
        total: 10,
        users: [dto, { ...dto, user_id: 2, user_role: 'tenant_user' }],
      };

      const result = adapter.fromPaginatedDTO(response);

      expect(result.count).toBe(2);
      expect(result.total).toBe(10);
      expect(result.users).toHaveLength(2);
      expect(result.users[0].id).toBe('1');
      expect(result.users[1].role).toBe(UserRole.TENANT_USER);
    });

    it('should handle empty array', () => {
      const result = adapter.fromPaginatedDTO({ count: 0, total: 0, users: [] });
      expect(result.users).toEqual([]);
    });
  });
});
