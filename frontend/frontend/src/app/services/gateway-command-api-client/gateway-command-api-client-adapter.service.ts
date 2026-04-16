import { Observable } from 'rxjs';

import { Gateway } from '../../models/gateway/gateway.model';

export abstract class GatewayCommandApiClientAdapter {
  abstract commissionGateway(
    gatewayId: string,
    tenantId: string,
    token: string,
  ): Observable<Gateway>;

  abstract decommissionGateway(gatewayId: string): Observable<void>;
  abstract resetGateway(gatewayId: string): Observable<void>;
  abstract rebootGateway(gatewayId: string): Observable<void>;
  abstract interruptGateway(gatewayId: string): Observable<void>;
  abstract resumeGateway(gatewayId: string): Observable<void>;
}