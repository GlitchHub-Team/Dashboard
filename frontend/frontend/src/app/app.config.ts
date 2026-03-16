import { ApplicationConfig } from '@angular/core';
import { provideRouter } from '@angular/router';
import { provideHttpClient, withInterceptors } from '@angular/common/http';
import { provideNativeDateAdapter } from '@angular/material/core';

import { routes } from './app.routes';
import { authInterceptor } from './interceptors/auth/auth.interceptor';
import { httpErrorInterceptor } from './interceptors/error/http-error.interceptor';
import { AuthApiClientService } from './services/auth-api-client/auth-api-client.service';
import { AuthServiceMock } from './mocks/auth.service.mock';

export const appConfig: ApplicationConfig = {
  providers: [
    provideRouter(routes),
    provideNativeDateAdapter(),
    provideHttpClient(withInterceptors([authInterceptor, httpErrorInterceptor])),
    // TODO: solo per testing per ora, da rimuovere quando avremo un backend funzionante
    { provide: AuthApiClientService, useClass: AuthServiceMock },
  ],
};
