import { ApplicationConfig, importProvidersFrom } from '@angular/core';
import { provideRouter } from '@angular/router';
import { provideHttpClient, withInterceptors } from '@angular/common/http';
import { provideNativeDateAdapter } from '@angular/material/core';
import { MatDialogModule } from '@angular/material/dialog';
import { routes } from './app.routes';
import { authInterceptor } from './interceptors/auth/auth.interceptor';
import { httpErrorInterceptor } from './interceptors/error/http-error.interceptor';
import { GatewayAdapter } from './adapters/gateway/gateway.adapter';
import { GatewayApiAdapter } from './adapters/gateway/gateway-api.adapter';
import { SensorAdapter } from './adapters/sensor/sensor.adapter';
import { SensorApiAdapter } from './adapters/sensor/sensor-api.adapter';
import { UserAdapter } from './adapters/user/user.adapter';
import { UserApiAdapter } from './adapters/user/user-api.adapter';
import { TenantApiAdapter } from './adapters/tenant/tenant-api.adapter';
import { TenantAdapter } from './adapters/tenant/tenant.adapter';

import { AuthApiClientService } from './services/auth-api-client/auth-api-client.service';
import { AuthApiClientAdapter } from './services/auth-api-client/auth-api-client-adapter.service';
import { GatewayApiClientAdapter } from './services/gateway-api-client/gateway-api-client-adapter.service';
import { GatewayCommandApiClientAdapter } from './services/gateway-command-api-client/gateway-command-api-client-adapter.service';
import { GatewayApiClientService } from './services/gateway-api-client/gateway-api-client.service';
import { GatewayCommandApiClientService } from './services/gateway-command-api-client/gateway-command-api-client.service';
import { SensorHistoricApiAdapter } from './services/sensor-historic-api/sensor-historic-api-adapter.service';
import { SensorLiveReadingsApiAdapter } from './services/sensor-live-api/sensor-live-readings-api-adapter.service';
import { SensorHistoricApiService } from './services/sensor-historic-api/sensor-historic-api.service';
import { SensorLiveReadingsApiService } from './services/sensor-live-api/sensor-live-readings-api.service';
import { UserApiClientService } from './services/user-api-client/user-api-client.service';
import { TenantApiClientService } from './services/tenant-api-client/tenant-api-client.service';
import { UserApiClientAdapter } from './services/user-api-client/user-api-client-adapter.service';
import { TenantApiClientAdapter } from './services/tenant-api-client/tenant-api-client-adapter.service';
import { SensorApiClientAdapter } from './services/sensor-api-client/sensor-api-client-adapter.service';
import { SensorApiClientService } from './services/sensor-api-client/sensor-api-client.service';
import { SensorCommandApiClientAdapter } from './services/sensor-command-api-client/sensor-command-api-client-adapter.service';
import { SensorCommandApiClientService } from './services/sensor-command-api-client/sensor-command-api-client.service';

export const appConfig: ApplicationConfig = {
  providers: [
    provideRouter(routes),
    provideNativeDateAdapter(),
    provideHttpClient(withInterceptors([httpErrorInterceptor, authInterceptor])),
    importProvidersFrom(MatDialogModule),
    { provide: GatewayAdapter, useClass: GatewayApiAdapter },
    { provide: SensorAdapter, useClass: SensorApiAdapter },
    { provide: UserAdapter, useClass: UserApiAdapter },
    { provide: TenantAdapter, useClass: TenantApiAdapter },
    { provide: AuthApiClientAdapter, useClass: AuthApiClientService },
    { provide: TenantApiClientAdapter, useClass: TenantApiClientService },
    { provide: UserApiClientAdapter, useClass: UserApiClientService },
    { provide: SensorApiClientAdapter, useClass: SensorApiClientService },
    { provide: SensorCommandApiClientAdapter, useClass: SensorCommandApiClientService },
    { provide: GatewayApiClientAdapter, useClass: GatewayApiClientService },
    { provide: GatewayCommandApiClientAdapter, useClass: GatewayCommandApiClientService },
    { provide: SensorHistoricApiAdapter, useClass: SensorHistoricApiService },
    { provide: SensorLiveReadingsApiAdapter, useClass: SensorLiveReadingsApiService },
  ],
};