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
import { SensorHistoricAdapter } from './adapters/sensor-historic.adapter';
import { SensorLiveReadingAdapter } from './adapters/sensor-live-reading.adapter';
import { SensorHistoricApiAdapter } from './adapters/sensor-historic-api.adapter';
import { SensorLiveReadingApiAdapter } from './adapters/sensor-livel-reading-api.adapter';
// TODO: solo per testing per ora, da rimuovere quando avremo un backend funzionante
import { AuthApiClientService } from './services/auth-api-client/auth-api-client.service';
import { AuthServiceMock } from './mocks/auth.service.mock';
import { TenantApiClientService } from './services/tenant-api-client/tenant-api-client.service';
import { TenantApiClientMockService } from './mocks/tenant-api-client.mock';
import { UserApiClientService } from './services/user-api-client/user-api-client.service';
import { UserApiClientMockService } from './mocks/user-api-client.mock';
import { SensorApiClientServiceMock } from './mocks/sensor.service.mock';
import { GatewayApiClientServiceMock } from './mocks/gateway.service.mock';
import { SensorApiClientService } from './services/sensor-api-client/sensor-api-client.service';
import { GatewayApiClientService } from './services/gateway-api-client/gateway-api-client.service';
import { SensorHistoricMockService } from './mocks/historic.service.mock';
import { SensorRealTimeMockService } from './mocks/live.service.mock';
import { SensorLiveReadingsApiService } from './services/sensor-live-api/sensor-live-readings-api.service';
import { SensorHistoricApiService } from './services/sensor-historic-api/sensor-historic-api.service';
import { GatewayCommandApiClientMockService } from './mocks/gateway-command-api-client.mock';
import { GatewayCommandApiClientService } from './services/gateway-command-api-client/gateway-command-api-client.service';

export const appConfig: ApplicationConfig = {
  providers: [
    provideRouter(routes),
    provideNativeDateAdapter(),
    provideHttpClient(withInterceptors([authInterceptor, httpErrorInterceptor])),
    importProvidersFrom(MatDialogModule),
    { provide: GatewayAdapter, useClass: GatewayApiAdapter },
    { provide: SensorAdapter, useClass: SensorApiAdapter },
    { provide: SensorHistoricAdapter, useClass: SensorHistoricApiAdapter },
    { provide: SensorLiveReadingAdapter, useClass: SensorLiveReadingApiAdapter },
    // TODO: solo per testing per ora, da rimuovere quando avremo un backend funzionante
    { provide: AuthApiClientService, useClass: AuthServiceMock },
    { provide: TenantApiClientService, useClass: TenantApiClientMockService },
    { provide: UserApiClientService, useClass: UserApiClientMockService },
    { provide: SensorApiClientService, useClass: SensorApiClientServiceMock },
    { provide: GatewayApiClientService, useClass: GatewayApiClientServiceMock },
    { provide: SensorLiveReadingsApiService, useClass: SensorRealTimeMockService },
    { provide: SensorHistoricApiService, useClass: SensorHistoricMockService },
    { provide: GatewayCommandApiClientService, useClass: GatewayCommandApiClientMockService },
  ],
};
