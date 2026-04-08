# --- STAGE 1: Build ---
FROM node:20.19.0-bookworm AS build-stage
ARG ANGULAR_CLI_VERSION="21.1.0"
WORKDIR /app
COPY package*.json ./
RUN npm ci
RUN npm install -g @angular/cli@${ANGULAR_CLI_VERSION}
COPY . .
RUN ng build --configuration production

# --- STAGE 2: Runtime ---
FROM nginx:stable-alpine

# PULIZIA TOTALE: Rimuoviamo i conf di default E la cartella templates
RUN rm -rf /etc/nginx/conf.d/* /etc/nginx/templates/*

# Copiamo il nostro file statico
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copiamo i file compilati di Angular
COPY --from=build-stage /app/dist/frontend/browser /usr/share/nginx/html

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]