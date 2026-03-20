import { Injectable } from '@angular/core';
import { User } from '../../models/user.model';
import { RawUserConfig } from '../../models/raw-user-config.model';

export interface RawPaginatedResponse {
  content?: RawUserConfig[];
  items?: RawUserConfig[];
  data?: RawUserConfig[];
  totalElements?: number;
  totalCount?: number;
  total?: number;
}

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

  // Adatta la risposta paginata dal backend
  public adaptPaginated(response: RawPaginatedResponse): { items: User[]; totalCount: number } {
    // NOTA: Sostituisci "content" e "totalElements" con le chiavi esatte restituite dal tuo backend
    return {
      items: this.adaptArray(response.content || response.items || response.data || []),
      totalCount: response.totalElements || response.totalCount || response.total || 0,
    };
  }
}
