import { HttpClient, HttpParams } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';

import { GatewayBackend } from '../../models/gateway/gateway-backend.model';
import { GatewayConfig } from '../../models/gateway/gateway-config.model';
import { PaginatedGatewayResponse } from '../../models/gateway/paginated-gateway-response.model';
import { environment } from '../../../environments/environment';

@Injectable({
  providedIn: 'root',
})
export class GatewayApiClientService {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = `${environment.apiUrl}`;

  public getGatewayListByTenant(
    tenantId: string,
    page: number,
    limit: number,
  ): Observable<PaginatedGatewayResponse<GatewayBackend>> {
    const params = new HttpParams().set('page', page.toString()).set('limit', limit.toString());

    return this.http.get<PaginatedGatewayResponse<GatewayBackend>>(
      `${this.apiUrl}/tenant/${tenantId}/gateways`,
      {
        params,
      },
    );
  }

  public getGatewayList(
    page: number,
    limit: number,
  ): Observable<PaginatedGatewayResponse<GatewayBackend>> {
    const params = new HttpParams().set('page', page.toString()).set('limit', limit.toString());

    return this.http.get<PaginatedGatewayResponse<GatewayBackend>>(`${this.apiUrl}/gateways`, {
      params,
    });
  }

  public addNewGateway(config: GatewayConfig): Observable<GatewayBackend> {
    return this.http.post<GatewayBackend>(`${this.apiUrl}/gateway`, config);
  }

  public deleteGateway(gatewayId: string): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/gateway/${gatewayId}`);
  }
}
