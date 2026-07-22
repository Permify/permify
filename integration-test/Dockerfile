# Step 1: Modules caching
FROM golang:1.27rc2-alpine@sha256:dcbb18cc5fa1082364dc6aa95224b6b55429d09cbb9631a053d8064c1c367300 AS modules
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

# Step 2: Tests
FROM golang:1.27rc2-alpine@sha256:dcbb18cc5fa1082364dc6aa95224b6b55429d09cbb9631a053d8064c1c367300
COPY --from=modules /go/pkg /go/pkg
COPY . /app
WORKDIR /app

RUN go env -w CGO_ENABLED=0
RUN go env -w GOOS=linux
RUN go env -w GOARCH=amd64

CMD ["go", "test", "-v", "./integration-test/..."]