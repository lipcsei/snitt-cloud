FROM golang:1.26-alpine AS build
WORKDIR /src

# A függőségek külön rétegben: forráskód-változás nem invalidálja a modul cache-t.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/account ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/account /account
EXPOSE 8090
USER nonroot:nonroot
ENTRYPOINT ["/account"]
