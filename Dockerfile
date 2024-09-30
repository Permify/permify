FROM golang:1.23.1-alpine3.20@sha256:ac67716dd016429be8d4c2c53a248d7bcdf06d34127d3dc451bda6aa5a87bc06 as permify-builder
WORKDIR /go/src/app
RUN apk update && apk add --no-cache git
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg/mod CGO_ENABLED=0 go build -v ./cmd/permify/

FROM cgr.dev/chainguard/static:latest@sha256:d2a76860057c1260ea5dc8ae4e18beff5ccfb1b67004295c9ab6951833e93de7
COPY --from=ghcr.io/grpc-ecosystem/grpc-health-probe:v0.4.28 /ko-app/grpc-health-probe /usr/local/bin/grpc_health_probe
COPY --from=permify-builder /go/src/app/permify /usr/local/bin/permify
ENV PATH="$PATH:/usr/local/bin"
ENTRYPOINT ["permify"]
CMD ["serve"]
