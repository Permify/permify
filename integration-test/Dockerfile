# Step 1: Modules caching
FROM golang:1.27rc3-alpine@sha256:c5aca77a4d16cb6688dbf3ccade67eff6f05ee208bc854d060e6947f5c27e23c AS modules
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

# Step 2: Tests
FROM golang:1.27rc3-alpine@sha256:c5aca77a4d16cb6688dbf3ccade67eff6f05ee208bc854d060e6947f5c27e23c
COPY --from=modules /go/pkg /go/pkg
COPY . /app
WORKDIR /app

RUN go env -w CGO_ENABLED=0
RUN go env -w GOOS=linux
RUN go env -w GOARCH=amd64

CMD ["go", "test", "-v", "./integration-test/..."]