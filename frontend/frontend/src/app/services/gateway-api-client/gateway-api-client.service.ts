// services/gateway-api-client/gateway-api-client.service.ts
import { HttpClient, HttpParams } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';

import { GatewayBackend } from '../../models/gateway/gateway-backend.model';
import { GatewayConfig } from '../../models/gateway/gateway-config.model';
import { PaginatedResponse } from '../../models/paginated-response.model';
import { environment } from '../../../environments/environment';

@Injectable({
  providedIn: 'root',
})
export class GatewayApiClientService {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = `${environment.apiUrl}/gateway`;

  public getGatewayListByTenant(
    tenantId: string,
    page: number,
    limit: number,
  ): Observable<PaginatedResponse<GatewayBackend>> {
    const params = new HttpParams().set('page', page.toString()).set('limit', limit.toString());

    return this.http.get<PaginatedResponse<GatewayBackend>>(`${this.apiUrl}/${tenantId}/list`, {
      params,
    });
  }

  public getGatewayList(
    page: number,
    limit: number,
  ): Observable<PaginatedResponse<GatewayBackend>> {
    const params = new HttpParams().set('page', page.toString()).set('limit', limit.toString());

    return this.http.get<PaginatedResponse<GatewayBackend>>(`${this.apiUrl}/list`, { params });
  }

  public addNewGateway(config: GatewayConfig): Observable<GatewayBackend> {
    return this.http.post<GatewayBackend>(`${this.apiUrl}/add`, config);
  }

  public deleteGateway(id: string): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/delete/${id}`);
  }

  public sendCommandToGateway(): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/command`, {});
  }
}
