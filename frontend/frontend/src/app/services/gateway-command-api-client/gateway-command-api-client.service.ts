import { HttpClient } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable, map } from 'rxjs';

import { environment } from '../../../environments/environment';
import { GatewayCommandApiClientAdapter } from './gateway-command-api-client-adapter.service';
import { GatewayApiAdapter } from '../../adapters/gateway/gateway-api.adapter';
import { GatewayBackend } from '../../models/gateway/gateway-backend.model';
import { Gateway } from '../../models/gateway/gateway.model';

@Injectable({
  providedIn: 'root',
})
export class GatewayCommandApiClientService extends GatewayCommandApiClientAdapter {
  private readonly http = inject(HttpClient);
  private readonly mapper = inject(GatewayApiAdapter);
  private readonly apiUrl = `${environment.apiUrl}`;

  commissionGateway(
    gatewayId: string,
    tenantId: string,
    token: string,
  ): Observable<Gateway> {
    return this.http
      .post<GatewayBackend>(`${this.apiUrl}/gateway/${gatewayId}/commission`, {
        tenant_id: tenantId,
        commission_token: token,
      })
      .pipe(map((dto) => this.mapper.fromDTO(dto)));
  }

  decommissionGateway(gatewayId: string): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/gateway/${gatewayId}/decommission`, {});
  }

  resetGateway(gatewayId: string): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/gateway/${gatewayId}/reset`, {});
  }

  rebootGateway(gatewayId: string): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/gateway/${gatewayId}/reboot`, {});
  }

  interruptGateway(gatewayId: string): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/gateway/${gatewayId}/interrupt`, {});
  }

  resumeGateway(gatewayId: string): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/gateway/${gatewayId}/resume`, {});
  }
}