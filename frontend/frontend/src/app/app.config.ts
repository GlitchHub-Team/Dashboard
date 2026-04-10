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
],
};