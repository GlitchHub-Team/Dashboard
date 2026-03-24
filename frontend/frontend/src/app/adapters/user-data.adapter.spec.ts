import { UserDataAdapter } from './user-data.adapter';
import { User } from '../models/user.model';
import { UserRole } from '../models/user-role.enum';
import { RawUserConfig } from '../models/raw-user-config.model';

describe('UserDataAdapter', () => {
  let adapter: UserDataAdapter;

  beforeEach(() => {
    adapter = new UserDataAdapter();
  });

  it('should adapt a raw user config to a user', () => {
    const raw = { id: '1', email: 'admin@test.com', role: UserRole.TENANT_ADMIN, tenantId: 'tenant-1' } as unknown as RawUserConfig;
    const expected: User = { id: '1', email: 'admin@test.com', role: UserRole.TENANT_ADMIN, tenantId: 'tenant-1' };
    
    const result = adapter.adapt(raw);
    
    expect(result).toEqual(expected);
  });

  it('should handle missing or null fields when adapting', () => {
    // Simuliamo un oggetto con dati mancanti in arrivo dal backend
    const raw = { id: '2', email: null } as unknown as RawUserConfig; 
    // Adatta i valori expected ai default che hai effettivamente impostato nel tuo UserDataAdapter
    const expected: User = { id: '2', email: '', role: undefined as unknown as UserRole, tenantId: '' }; 
    
    const result = adapter.adapt(raw);
    
    expect(result.id).toEqual(expected.id);
    expect(result.email).toEqual(expected.email);
  });

  it('should adapt an array of raw user configs', () => {
    const rawArray = [
      { id: '1', email: 'admin@test.com', role: UserRole.TENANT_ADMIN, tenantId: 'tenant-1' },
      { id: '2', email: 'user@test.com', role: UserRole.TENANT_USER, tenantId: 'tenant-1' }
    ] as unknown as RawUserConfig[];
    
    const result = adapter.adaptArray(rawArray);
    
    expect(result.length).toBe(2);
    expect(result[0].email).toBe('admin@test.com');
    expect(result[1].email).toBe('user@test.com');
  });
});
