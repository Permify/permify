# Step 1: Modules caching
FROM golang:1.24-alpine3.20@sha256:9fed4022a220fb64327baa90cddfd98607f3b816cb4f5769187500571f73072d as modules
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

# Step 2: Tests
FROM golang:1.24-alpine3.20@sha256:9fed4022a220fb64327baa90cddfd98607f3b816cb4f5769187500571f73072d
COPY --from=modules /go/pkg /go/pkg
COPY . /app
WORKDIR /app

RUN go env -w CGO_ENABLED=0
RUN go env -w GOOS=linux
RUN go env -w GOARCH=amd64

CMD ["go", "test", "-v", "./integration-test/..."]