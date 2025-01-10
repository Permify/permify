# Step 1: Modules caching
FROM golang:1.23.4-alpine3.20@sha256:2314d93ea3899b8118d11ec70714e96c4bb2d52ff46ee919622269a5c6207077 as modules
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

# Step 2: Tests
FROM golang:1.23.4-alpine3.20@sha256:2314d93ea3899b8118d11ec70714e96c4bb2d52ff46ee919622269a5c6207077
COPY --from=modules /go/pkg /go/pkg
COPY . /app
WORKDIR /app

RUN go env -w CGO_ENABLED=0
RUN go env -w GOOS=linux
RUN go env -w GOARCH=amd64

CMD ["go", "test", "-v", "./integration-test/..."]