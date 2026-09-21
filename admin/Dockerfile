# Az admin felület statikus SPA: buildeljük, majd nginx szolgálja ki.
# Ugyanez a kép megy majd a VPS-re is, ezért a Keycloak/API címeket NEM
# égetjük be - build-argumentként jönnek (a Vite fordításkor helyettesíti be
# az import.meta.env értékeket, futásidőben már nem lehet módosítani).

FROM node:20-alpine AS build
WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . .

ARG VITE_KEYCLOAK_URL=http://localhost:8081
ARG VITE_KEYCLOAK_REALM=snitt
ARG VITE_KEYCLOAK_CLIENT_ID=snitt-admin
ARG VITE_API_BASE_URL=http://localhost:8090
# Opcionális hibajelentés (GlitchTip). Üres DSN = kikapcsolva: az SDK ilyenkor a
# csomagba sem kerül be. Az environment üresen a Vite módját (production) kapja.
ARG VITE_SENTRY_DSN=
ARG VITE_SENTRY_ENVIRONMENT=
ENV VITE_KEYCLOAK_URL=$VITE_KEYCLOAK_URL \
    VITE_KEYCLOAK_REALM=$VITE_KEYCLOAK_REALM \
    VITE_KEYCLOAK_CLIENT_ID=$VITE_KEYCLOAK_CLIENT_ID \
    VITE_API_BASE_URL=$VITE_API_BASE_URL \
    VITE_SENTRY_DSN=$VITE_SENTRY_DSN \
    VITE_SENTRY_ENVIRONMENT=$VITE_SENTRY_ENVIRONMENT
RUN npm run build

FROM nginx:1.27-alpine
COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
# 127.0.0.1 és nem localhost: a konténerben a localhost előbb ::1-re oldódik
# fel, az nginx viszont alapból csak IPv4-en figyel - a healthcheck így
# "connection refused"-ot kapna egy tökéletesen működő kiszolgálótól.
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s \
  CMD wget -qO- http://127.0.0.1/ >/dev/null || exit 1
