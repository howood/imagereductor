FROM golang:1.25 AS build-env

WORKDIR /go/src/github.com/howood/imagereductor

# Copy go.mod and go.sum first for better layer caching
COPY go.mod /go/src/github.com/howood/imagereductor/go.mod
COPY go.sum /go/src/github.com/howood/imagereductor/go.sum
RUN go mod download

# Copy source code
COPY application /go/src/github.com/howood/imagereductor/application
COPY di /go/src/github.com/howood/imagereductor/di
COPY domain /go/src/github.com/howood/imagereductor/domain
COPY imagereductor /go/src/github.com/howood/imagereductor/imagereductor
COPY infrastructure /go/src/github.com/howood/imagereductor/infrastructure
COPY interfaces /go/src/github.com/howood/imagereductor/interfaces
COPY library /go/src/github.com/howood/imagereductor/library

# Build with optimizations
RUN cd /go/src/github.com/howood/imagereductor/imagereductor && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags "-s -w" -o /go/bin/imagereductor


FROM busybox:latest

# Create non-root user
RUN adduser -D -u 1000 appuser

# Copy SSL certificates and binary
COPY --from=build-env /etc/ssl/certs /etc/ssl/certs
COPY --from=build-env /go/bin/imagereductor /usr/local/bin/imagereductor

# Change ownership and switch to non-root user
RUN chown appuser:appuser /usr/local/bin/imagereductor
USER appuser

# Expose default port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8080/info?key=healthcheck || exit 1

ENTRYPOINT ["/usr/local/bin/imagereductor"]
