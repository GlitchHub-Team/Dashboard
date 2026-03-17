import { HttpClient } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';

import { Gateway } from '../../models/gateway.model';
import { GatewayConfig } from '../../models/gateway-config.model';
import { environment } from '../../../environments/environment';

@Injectable({
  providedIn: 'root',
})
export class GatewayApiClientService {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = `${environment.apiUrl}/gateway`;

  public getGatewayListByTenant(tenantId: string): Observable<Gateway[]> {
    return this.http.get<Gateway[]>(`${this.apiUrl}/list/`, { params: { tenantId } });
  }

  public getGatewayList(): Observable<Gateway[]> {
    return this.http.get<Gateway[]>(`${this.apiUrl}/list`);
  }

  public addNewGateway(config: GatewayConfig): Observable<Gateway> {
    return this.http.post<Gateway>(`${this.apiUrl}/add`, config);
  }

  public deleteGateway(id: string): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/delete/${id}`);
  }

  // TODO: models per GatewayCommand e CommandResult
  public sendCommandToGateway(): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/command`, {});
  }
}
