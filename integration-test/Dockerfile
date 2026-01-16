# Step 1: Modules caching
FROM golang:1.26rc2-alpine@sha256:ebf4414ac1ebce5d728a1b9446037ba9aee1b46228bbc8972fb57819a91e164e AS modules
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

# Step 2: Tests
FROM golang:1.26rc2-alpine@sha256:ebf4414ac1ebce5d728a1b9446037ba9aee1b46228bbc8972fb57819a91e164e
COPY --from=modules /go/pkg /go/pkg
COPY . /app
WORKDIR /app

RUN go env -w CGO_ENABLED=0
RUN go env -w GOOS=linux
RUN go env -w GOARCH=amd64

CMD ["go", "test", "-v", "./integration-test/..."]