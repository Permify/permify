# Step 1: Modules caching
FROM golang:1.24rc2-alpine3.20@sha256:3c8ed31d9b4e74f61ce30e60c9b6a71a8dff7044c6c871764d706468917f5376 as modules
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

# Step 2: Tests
FROM golang:1.24rc2-alpine3.20@sha256:3c8ed31d9b4e74f61ce30e60c9b6a71a8dff7044c6c871764d706468917f5376
COPY --from=modules /go/pkg /go/pkg
COPY . /app
WORKDIR /app

RUN go env -w CGO_ENABLED=0
RUN go env -w GOOS=linux
RUN go env -w GOARCH=amd64

CMD ["go", "test", "-v", "./integration-test/..."]