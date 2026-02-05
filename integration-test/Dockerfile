# Step 1: Modules caching
FROM golang:1.26rc3-alpine@sha256:343c20fd6876bfb5ba9f46b0a452008b7dced3804e424ff7ada0ceadafad5c55 AS modules
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

# Step 2: Tests
FROM golang:1.26rc3-alpine@sha256:343c20fd6876bfb5ba9f46b0a452008b7dced3804e424ff7ada0ceadafad5c55
COPY --from=modules /go/pkg /go/pkg
COPY . /app
WORKDIR /app

RUN go env -w CGO_ENABLED=0
RUN go env -w GOOS=linux
RUN go env -w GOARCH=amd64

CMD ["go", "test", "-v", "./integration-test/..."]