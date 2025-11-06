# Step 1: Modules caching
FROM golang:1.25.4-alpine@sha256:d2ede9f3341a67413127cf5366bb25bbad9b0a66e8173cae3a900ab00e84861f AS modules
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

# Step 2: Tests
FROM golang:1.25.4-alpine@sha256:d2ede9f3341a67413127cf5366bb25bbad9b0a66e8173cae3a900ab00e84861f
COPY --from=modules /go/pkg /go/pkg
COPY . /app
WORKDIR /app

RUN go env -w CGO_ENABLED=0
RUN go env -w GOOS=linux
RUN go env -w GOARCH=amd64

CMD ["go", "test", "-v", "./integration-test/..."]