# Step 1: Modules caching
FROM golang:1.24rc3-alpine3.20@sha256:3cf3c4170cea58c587c2178d4c6ecb8e421df7ff29f9df409a7daaab5b24b6cb as modules
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

# Step 2: Tests
FROM golang:1.24rc3-alpine3.20@sha256:3cf3c4170cea58c587c2178d4c6ecb8e421df7ff29f9df409a7daaab5b24b6cb
COPY --from=modules /go/pkg /go/pkg
COPY . /app
WORKDIR /app

RUN go env -w CGO_ENABLED=0
RUN go env -w GOOS=linux
RUN go env -w GOARCH=amd64

CMD ["go", "test", "-v", "./integration-test/..."]