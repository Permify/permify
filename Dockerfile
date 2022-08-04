# Step 1: Builder
FROM golang:1.17.1-alpine3.14 as permify-builder
WORKDIR /go/src/app
RUN apk update && apk add --no-cache git
COPY . .
RUN go build -v ./cmd/permify/

# Step 2: Final
FROM alpine:3.16.1
EXPOSE 3476 3476
COPY --from=permify-builder /go/src/app/permify /usr/local/bin/permify
COPY --from=permify-builder /go/src/app/default.config.yaml /default.config.yaml
ENTRYPOINT ["permify"]