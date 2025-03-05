# Step 1: Modules caching
FROM golang:1.24-alpine3.20@sha256:3d9132b88a6317b846b55aa8e821821301906fe799932ecbc4f814468c6977a5 as modules
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

# Step 2: Tests
FROM golang:1.24-alpine3.20@sha256:3d9132b88a6317b846b55aa8e821821301906fe799932ecbc4f814468c6977a5
COPY --from=modules /go/pkg /go/pkg
COPY . /app
WORKDIR /app

RUN go env -w CGO_ENABLED=0
RUN go env -w GOOS=linux
RUN go env -w GOARCH=amd64

CMD ["go", "test", "-v", "./integration-test/..."]