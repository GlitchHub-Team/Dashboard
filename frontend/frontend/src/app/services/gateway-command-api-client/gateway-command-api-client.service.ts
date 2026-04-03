import { HttpClient } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs/internal/Observable';

import { environment } from '../../../environments/environment';
import { GatewayBackend } from '../../models/gateway/gateway-backend.model';

@Injectable({
  providedIn: 'root',
})
export class GatewayCommandApiClientService {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = `${environment.apiUrl}`;

  // TODO: Manca l'invio del commission_token, da dove lo prendo ???
  public commissionGateway(
    gatewayId: string,
    tenantId: string,
    token: string,
  ): Observable<GatewayBackend> {
    return this.http.post<GatewayBackend>(`${this.apiUrl}/gateway/${gatewayId}/commission`, {
      tenant_id: tenantId,
      commission_token: token,
    });
  }

  public decommissionGateway(gatewayId: string): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/gateway/${gatewayId}/decommission`, {});
  }

  public resetGateway(gatewayId: string): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/gateway/${gatewayId}/reset`, {});
  }

  public rebootGateway(gatewayId: string): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/gateway/${gatewayId}/reboot`, {});
  }
}
