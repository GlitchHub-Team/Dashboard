import { HttpClient, HttpParams } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable, map } from 'rxjs';

import { environment } from '../../../environments/environment';
import { GatewayApiClientAdapter } from './gateway-api-client-adapter.service';
import { GatewayApiAdapter } from '../../adapters/gateway/gateway-api.adapter';
import { GatewayBackend } from '../../models/gateway/gateway-backend.model';
import { GatewayConfig } from '../../models/gateway/gateway-config.model';
import { PaginatedGatewayResponse } from '../../models/gateway/paginated-gateway-response.model';
import { Gateway } from '../../models/gateway/gateway.model';

@Injectable({
  providedIn: 'root',
})
export class GatewayApiClientService extends GatewayApiClientAdapter {
  private readonly http = inject(HttpClient);
  private readonly mapper = inject(GatewayApiAdapter);
  private readonly apiUrl = `${environment.apiUrl}`;

  getGatewayListByTenant(
    tenantId: string,
    page: number,
    limit: number,
  ): Observable<PaginatedGatewayResponse<Gateway>> {
    const params = new HttpParams().set('page', page).set('limit', limit);

    return this.http
      .get<PaginatedGatewayResponse<GatewayBackend>>(
        `${this.apiUrl}/tenant/${tenantId}/gateways`,
        { params },
      )
      .pipe(map((response) => this.mapper.fromPaginatedDTO(response)));
  }

  getGatewayList(
    page: number,
    limit: number,
  ): Observable<PaginatedGatewayResponse<Gateway>> {
    const params = new HttpParams().set('page', page).set('limit', limit);

    return this.http
      .get<PaginatedGatewayResponse<GatewayBackend>>(
        `${this.apiUrl}/gateways`,
        { params },
      )
      .pipe(map((response) => this.mapper.fromPaginatedDTO(response)));
  }

  addNewGateway(config: GatewayConfig): Observable<Gateway> {
    return this.http
      .post<GatewayBackend>(`${this.apiUrl}/gateway`, {
        name: config.name,
        interval: config.interval,
      })
      .pipe(map((dto) => this.mapper.fromDTO(dto)));
  }

  deleteGateway(gatewayId: string): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/gateway/${gatewayId}`);
  }
}