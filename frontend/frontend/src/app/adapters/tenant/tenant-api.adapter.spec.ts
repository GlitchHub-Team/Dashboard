import { describe, it, expect } from 'vitest';
import { TenantApiAdapter } from './tenant-api.adapter';
import { TenantBackend } from '../../models/tenant/tenant-backend.model';

describe('TenantApiAdapter', () => {
  const adapter = new TenantApiAdapter();

  const dto: TenantBackend = {
    tenant_id: 'tenant-01',
    tenant_name: 'Tenant 1',
    can_impersonate: true,
  };

  describe('fromDTO', () => {
    it.each([
      { field: 'id', expected: 'tenant-01' },
      { field: 'name', expected: 'Tenant 1' },
      { field: 'canImpersonate', expected: true },
    ] as const)('should map $field correctly', ({ field, expected }) => {
      expect(adapter.fromDTO(dto)[field]).toEqual(expected);
    });
  });

  describe('fromPaginatedDTO', () => {
    it('should map count, total and all tenants', () => {
      const response = {
        count: 2,
        total: 10,
        tenants: [dto, { ...dto, tenant_id: 'tenant-02', can_impersonate: false }],
      };

      const result = adapter.fromPaginatedDTO(response);

      expect(result.count).toBe(2);
      expect(result.total).toBe(10);
      expect(result.tenants).toHaveLength(2);
      expect(result.tenants[0].id).toBe('tenant-01');
      expect(result.tenants[1].canImpersonate).toBe(false);
    });

    it('should handle empty array', () => {
      const result = adapter.fromPaginatedDTO({ count: 0, total: 0, tenants: [] });
      expect(result.tenants).toEqual([]);
    });
  });
});
