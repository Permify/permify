# Step 1: Modules caching
FROM golang:1.23.4-alpine3.20@sha256:6a8532e5441593becc88664617107ed567cb6862cb8b2d87eb33b7ee750f653c as modules
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

# Step 2: Tests
FROM golang:1.23.4-alpine3.20@sha256:6a8532e5441593becc88664617107ed567cb6862cb8b2d87eb33b7ee750f653c
COPY --from=modules /go/pkg /go/pkg
COPY . /app
WORKDIR /app

RUN go env -w CGO_ENABLED=0
RUN go env -w GOOS=linux
RUN go env -w GOARCH=amd64

CMD ["go", "test", "-v", "./integration-test/..."]