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

  // RESET
  public resetGateway(gatewayId: string): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/gateway/${gatewayId}/reset`, {});
  }

  // RIAVVIO
  public rebootGateway(gatewayId: string): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/gateway/${gatewayId}/reboot`, {});
  }

  // STOP INVIO DATI
  public interruptGateway(gatewayId: string): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/gateway/${gatewayId}/interrupt`, {});
  }

  // RIPRESA INVIO DATI
  public resumeGateway(gatewayId: string): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/gateway/${gatewayId}/resume`, {});
  }
}
