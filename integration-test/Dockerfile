# Step 1: Modules caching
FROM golang:1.23.4-alpine3.20@sha256:9bfa54e3a85920df66ecf1c4979dfdf88baa80ca9752adda442713863087ef18 as modules
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

# Step 2: Tests
FROM golang:1.23.4-alpine3.20@sha256:9bfa54e3a85920df66ecf1c4979dfdf88baa80ca9752adda442713863087ef18
COPY --from=modules /go/pkg /go/pkg
COPY . /app
WORKDIR /app

RUN go env -w CGO_ENABLED=0
RUN go env -w GOOS=linux
RUN go env -w GOARCH=amd64

CMD ["go", "test", "-v", "./integration-test/..."]