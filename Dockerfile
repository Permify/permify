FROM golang:1.25.7-alpine@sha256:f6751d823c26342f9506c03797d2527668d095b0a15f1862cddb4d927a7a4ced AS permify-builder
WORKDIR /go/src/app
RUN apk update && apk add --no-cache git
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg/mod CGO_ENABLED=0 go build -v ./cmd/permify/

FROM golang:1.25.7-alpine@sha256:f6751d823c26342f9506c03797d2527668d095b0a15f1862cddb4d927a7a4ced AS health-probe-builder
WORKDIR /go/src/app
RUN apk update && apk add --no-cache git
RUN git clone https://github.com/grpc-ecosystem/grpc-health-probe.git
WORKDIR /go/src/app/grpc-health-probe
RUN git checkout master
ENV GOTOOLCHAIN=local
RUN CGO_ENABLED=0 go install -a -tags netgo -ldflags=-w

FROM cgr.dev/chainguard/static:latest@sha256:11ec91f0372630a2ca3764cea6325bebb0189a514084463cbb3724e5bb350d14
COPY --from=health-probe-builder /go/bin/grpc-health-probe /usr/local/bin/grpc_health_probe
COPY --from=permify-builder /go/src/app/permify /usr/local/bin/permify
ENV PATH="$PATH:/usr/local/bin"
ENTRYPOINT ["permify"]
CMD ["serve"]
