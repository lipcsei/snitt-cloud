# syntax=docker/dockerfile:1
FROM golang:1.26-alpine AS build
WORKDIR /src

# Az account service a privát github.com/lipcsei/commons modult is importálja (errtrack,
# keycloak): a "go mod download" ezt git+ssh-n éri el, a build --mount=type=ssh-je
# (docker-compose.yml "ssh: [default]") által idehozott ügynök-socketen. A known_hosts csak a
# build idejére él, a képbe nem kerül.
RUN apk add --no-cache git openssh-client && mkdir -p -m 0700 /root/.ssh \
    && ssh-keyscan -t ed25519 github.com >> /root/.ssh/known_hosts
ENV GOPRIVATE=github.com/lipcsei/*
RUN git config --global url."ssh://git@github.com/".insteadOf "https://github.com/"

# A függőségek külön rétegben: forráskód-változás nem invalidálja a modul cache-t.
COPY go.mod go.sum ./
RUN --mount=type=ssh go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/account ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/account /account
EXPOSE 8090
USER nonroot:nonroot
ENTRYPOINT ["/account"]
