# Step 1: Builder
FROM golang:1.20-alpine3.16 as permify-builder
WORKDIR /go/src/app
RUN apk update && apk add --no-cache git
COPY . .
RUN CGO_ENABLED=0 go build -v ./cmd/permify/

# Step 2: Final
FROM cgr.dev/chainguard/static:latest
EXPOSE 3476
EXPOSE 3478
COPY --from=ghcr.io/grpc-ecosystem/grpc-health-probe:v0.4.12 /ko-app/grpc-health-probe /usr/local/bin/grpc_health_probe
COPY --from=permify-builder /go/src/app/permify /usr/local/bin/permify
ENTRYPOINT ["permify"]