import { ApplicationConfig, importProvidersFrom } from '@angular/core';
import { provideRouter } from '@angular/router';
import { provideHttpClient, withInterceptors } from '@angular/common/http';
import { provideNativeDateAdapter } from '@angular/material/core';
import { MatDialogModule } from '@angular/material/dialog';
import { routes } from './app.routes';
import { authInterceptor } from './interceptors/auth/auth.interceptor';
import { httpErrorInterceptor } from './interceptors/error/http-error.interceptor';
import { GatewayAdapter } from './adapters/gateway.adapter';
import { GatewayApiAdapter } from './adapters/gateway-api.adapter';
import { SensorAdapter } from './adapters/sensor.adapter';
import { SensorApiAdapter } from './adapters/sensor-api.adapter';
// TODO: solo per testing per ora, da rimuovere quando avremo un backend funzionante
import { AuthApiClientService } from './services/auth-api-client/auth-api-client.service';
import { AuthServiceMock } from './mocks/auth.service.mock';
import { TenantApiClientService } from './services/tenant/tenant-api-client.service';
import { TenantApiClientMockService } from './services/tenant/tenant-api-client.mock';
import { UserApiClientService } from './services/user/user-api-client.service';
import { UserApiClientMockService } from './services/user/user-api-client.mock';
import { SensorApiClientServiceMock } from './mocks/sensor.service.mock';
import { GatewayApiClientServiceMock } from './mocks/gateway.service.mock';
import { SensorApiClientService } from './services/sensor-api-client/sensor-api-client.service';
import { GatewayApiClientService } from './services/gateway-api-client/gateway-api-client.service';

export const appConfig: ApplicationConfig = {
  providers: [
    provideRouter(routes),
    provideNativeDateAdapter(),
    provideHttpClient(withInterceptors([authInterceptor, httpErrorInterceptor])),
    importProvidersFrom(MatDialogModule),
    { provide: GatewayAdapter, useClass: GatewayApiAdapter },
    { provide: SensorAdapter, useClass: SensorApiAdapter },
    // TODO: solo per testing per ora, da rimuovere quando avremo un backend funzionante
    { provide: AuthApiClientService, useClass: AuthServiceMock },
    { provide: TenantApiClientService, useClass: TenantApiClientMockService },
    { provide: UserApiClientService, useClass: UserApiClientMockService },
    { provide: SensorApiClientService, useClass: SensorApiClientServiceMock },
    { provide: GatewayApiClientService, useClass: GatewayApiClientServiceMock },
  ],
};
