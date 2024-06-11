FROM golang:1.22-alpine3.18@sha256:d1a601b64de09e2fa38c95e55838961811d5ca11062a8f4230a5c434b3ae2a34 as permify-builder
WORKDIR /go/src/app
RUN apk update && apk add --no-cache git
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg/mod CGO_ENABLED=0 go build -v ./cmd/permify/

FROM cgr.dev/chainguard/static:latest@sha256:288b818c1b3dd89776d176f07f5f671b118fe836c4d80ec2cc3299b596fe71b7
COPY --from=ghcr.io/grpc-ecosystem/grpc-health-probe:v0.4.25 /ko-app/grpc-health-probe /usr/local/bin/grpc_health_probe
COPY --from=permify-builder /go/src/app/permify /usr/local/bin/permify
ENV PATH="$PATH:/usr/local/bin"
ENTRYPOINT ["permify"]