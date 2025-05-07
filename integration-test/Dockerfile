# Step 1: Modules caching
FROM golang:1.24-alpine3.20@sha256:9f98e9893fbc798c710f3432baa1e0ac6127799127c3101d2c263c3a954f0abe as modules
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

# Step 2: Tests
FROM golang:1.24-alpine3.20@sha256:9f98e9893fbc798c710f3432baa1e0ac6127799127c3101d2c263c3a954f0abe
COPY --from=modules /go/pkg /go/pkg
COPY . /app
WORKDIR /app

RUN go env -w CGO_ENABLED=0
RUN go env -w GOOS=linux
RUN go env -w GOARCH=amd64

CMD ["go", "test", "-v", "./integration-test/..."]