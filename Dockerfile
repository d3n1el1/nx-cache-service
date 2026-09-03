FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

ARG TARGETOS
ARG TARGETARCH

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /out/nx-cache-service ./cmd/nx-cache-service

FROM alpine:3

RUN apk add --no-cache ca-certificates && \
    adduser -D -u 10001 nxcache

WORKDIR /app
RUN mkdir -p /app/.nx-cache-service && chown -R nxcache:nxcache /app

COPY --from=build /out/nx-cache-service /usr/local/bin/nx-cache-service

USER nxcache

ENV ADDR=:8080

EXPOSE 8080

ENTRYPOINT ["nx-cache-service"]
