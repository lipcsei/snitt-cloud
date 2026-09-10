# A landing statikus oldal: buildeljük, majd nginx szolgálja ki. Élesben a
# GitHub Pages szolgálja ki ugyanezt a kimenetet - ez a kép a helyi
# megtekintéshez van (a Vite dev szerverrel szemben itt a VALÓDI production
# build fut, az előrendelt meta-fájlokkal együtt).

FROM node:20-alpine AS build
WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . .

ARG VITE_KEYCLOAK_URL=http://localhost:8081
ARG VITE_KEYCLOAK_REALM=snitt
ARG VITE_KEYCLOAK_CLIENT_ID=snitt-landing
# A kanonikus URL-ek és a hreflang ebből épülnek (lásd scripts/prerender-meta.mjs).
ARG SITE_URL=http://localhost:8088
ENV VITE_KEYCLOAK_URL=$VITE_KEYCLOAK_URL \
    VITE_KEYCLOAK_REALM=$VITE_KEYCLOAK_REALM \
    VITE_KEYCLOAK_CLIENT_ID=$VITE_KEYCLOAK_CLIENT_ID \
    SITE_URL=$SITE_URL
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
