# Step 1: Builder
FROM golang:1.19.5-alpine3.16 as permify-builder
WORKDIR /go/src/app
RUN apk update && apk add --no-cache git
COPY . .
RUN go build -v ./cmd/permify/

# Step 2: Final
FROM alpine:latest as final
EXPOSE 3476
EXPOSE 3478
COPY --from=permify-builder /go/src/app/permify /usr/local/bin/permify
ENTRYPOINT ["permify"]