FROM golang:1.23.4-alpine3.20@sha256:9a31ef0803e6afdf564edc8ba4b4e17caed22a0b1ecd2c55e3c8fdd8d8f68f98 as permify-builder
WORKDIR /go/src/app
RUN apk update && apk add --no-cache git
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg/mod CGO_ENABLED=0 go build -v ./cmd/permify/

FROM cgr.dev/chainguard/static:latest@sha256:5ff428f8a48241b93a4174dbbc135a4ffb2381a9e10bdbbc5b9db145645886d5
COPY --from=ghcr.io/grpc-ecosystem/grpc-health-probe:v0.4.28 /ko-app/grpc-health-probe /usr/local/bin/grpc_health_probe
COPY --from=permify-builder /go/src/app/permify /usr/local/bin/permify
ENV PATH="$PATH:/usr/local/bin"
ENTRYPOINT ["permify"]
CMD ["serve"]
