FROM golang:1.22-alpine3.18@sha256:ede158fb846dd8689c757a7795f1f884f3f1fb7fb04cad31de69870ab4a93067 as permify-builder
WORKDIR /go/src/app
RUN apk update && apk add --no-cache git
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg/mod CGO_ENABLED=0 go build -v ./cmd/permify/

FROM cgr.dev/chainguard/static:latest@sha256:739aaf25ce9c6ba75c3752d7fde4a94de386c6f44eba27239d2b98e91752e3ff
COPY --from=ghcr.io/grpc-ecosystem/grpc-health-probe:v0.4.25 /ko-app/grpc-health-probe /usr/local/bin/grpc_health_probe
COPY --from=permify-builder /go/src/app/permify /usr/local/bin/permify
ENV PATH="$PATH:/usr/local/bin"
ENTRYPOINT ["permify"]