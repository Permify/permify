# Step 1: Modules caching
FROM golang:1.24-alpine3.20@sha256:79f7ffeff943577c82791c9e125ab806f133ae3d8af8ad8916704922ddcc9dd8 as modules
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

# Step 2: Tests
FROM golang:1.24-alpine3.20@sha256:79f7ffeff943577c82791c9e125ab806f133ae3d8af8ad8916704922ddcc9dd8
COPY --from=modules /go/pkg /go/pkg
COPY . /app
WORKDIR /app

RUN go env -w CGO_ENABLED=0
RUN go env -w GOOS=linux
RUN go env -w GOARCH=amd64

CMD ["go", "test", "-v", "./integration-test/..."]