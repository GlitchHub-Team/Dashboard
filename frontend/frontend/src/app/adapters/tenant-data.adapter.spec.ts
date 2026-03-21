import { TenantDataAdapter } from '../services/tenant/tenant-data.adapter';
import { RawTenantConfig } from '../models/raw-tenant-config.model';
import { Tenant } from '../models/tenant.model';

describe('TenantDataAdapter', () => {
  let adapter: TenantDataAdapter;

  beforeEach(() => {
    adapter = new TenantDataAdapter();
  });

  it('should adapt a raw tenant config to a tenant', () => {
    const raw: RawTenantConfig = { name: 'Test Tenant' };
    const expected: Tenant = { name: 'Test Tenant' };
    const result = adapter.adapt(raw);
    expect(result).toEqual(expected);
  });

  it('should handle null or empty name when adapting', () => {
    const raw = { name: null } as unknown as RawTenantConfig;
    const expected: Tenant = { name: '' };
    const result = adapter.adapt(raw);
    expect(result).toEqual(expected);
  });

  it('should adapt an array of raw tenant configs', () => {
    const rawArray: RawTenantConfig[] = [{ name: 'Tenant 1' }, { name: 'Tenant 2' }];
    const expectedArray: Tenant[] = [{ name: 'Tenant 1' }, { name: 'Tenant 2' }];
    const result = adapter.adaptArray(rawArray);
    expect(result).toEqual(expectedArray);
  });
});
