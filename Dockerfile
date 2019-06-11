FROM golang:1.12-alpine AS builder

# Install git and gcc.
RUN apk add --no-cache git gcc musl-dev

# Setup directory.
WORKDIR /go/src/github.com/googleapis/grpc-fallback-go
COPY . .

# Compile for Linux.
ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOARCH amd64
ENV GO111MODULE on
ENV GOPROXY https://proxy.golang.org

# Install showcase.
RUN go mod download
RUN go build -installsuffix cgo \
  -ldflags="-w -s" \
  -o /go/bin/fallback-server \
  ./cmd/fallback-server

# Start a fresh image, and only copy the built binary.
FROM scratch
COPY --from=builder /go/bin/fallback-server /go/bin/fallback-server

# Expose port
EXPOSE 1337

# Run the server.
ENTRYPOINT ["/go/bin/fallback-server"]
