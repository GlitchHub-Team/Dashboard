import { Injectable } from '@angular/core';
import { User } from '../../models/user.model';
import { RawUserConfig } from '../../models/raw-user-config.model';

@Injectable({ providedIn: 'root'})
export class UserDataAdapter {
  public adapt(input: RawUserConfig): User {
    return {
      id: input.id || '',
      email: input.email || '',
      role: input.role || '',
      tenantId: (input as { tenantId?: string }).tenantId || '',
    };
  }

  public adaptArray(items: RawUserConfig[]): User[] {
    return items ? items.map(item => this.adapt(item)) : [];
  }
}
