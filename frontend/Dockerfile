FROM node:20.19.0-bookworm AS build-stage

ARG ANGULAR_CLI_VERSION="21.1.0"
WORKDIR /app

COPY package*.json ./
RUN npm ci
RUN npm install -g @angular/cli@${ANGULAR_CLI_VERSION}

COPY . .
RUN ng build --configuration production

FROM nginx:stable-alpine

RUN rm /etc/nginx/conf.d/default.conf

COPY default.conf.template /etc/nginx/templates/default.conf.template

COPY --from=build-stage /app/dist/frontend/browser /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]