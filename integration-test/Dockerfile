# Step 1: Modules caching
FROM golang:1.24rc3-alpine3.20@sha256:57ab2a012e4452805e14da7819a17aa810d476a80a3db63f8507682d29177f0b as modules
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

# Step 2: Tests
FROM golang:1.24rc3-alpine3.20@sha256:57ab2a012e4452805e14da7819a17aa810d476a80a3db63f8507682d29177f0b
COPY --from=modules /go/pkg /go/pkg
COPY . /app
WORKDIR /app

RUN go env -w CGO_ENABLED=0
RUN go env -w GOOS=linux
RUN go env -w GOARCH=amd64

CMD ["go", "test", "-v", "./integration-test/..."]