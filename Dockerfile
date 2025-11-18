FROM golang:1.25 AS build-env

WORKDIR /go/src/github.com/howood/imagereductor

COPY application /go/src/github.com/howood/imagereductor/application
COPY di /go/src/github.com/howood/imagereductor/di
COPY domain /go/src/github.com/howood/imagereductor/domain
COPY imagereductor /go/src/github.com/howood/imagereductor/imagereductor
COPY infrastructure /go/src/github.com/howood/imagereductor/infrastructure
COPY interfaces /go/src/github.com/howood/imagereductor/interfaces
COPY library /go/src/github.com/howood/imagereductor/library
COPY go.mod /go/src/github.com/howood/imagereductor/go.mod
COPY go.sum /go/src/github.com/howood/imagereductor/go.sum


RUN \
    cd /go/src/github.com/howood/imagereductor/imagereductor &&  \
    CGO_ENABLED=0 go install


FROM busybox
COPY --from=build-env /etc/ssl/certs /etc/ssl/certs
COPY --from=build-env /go/bin/imagereductor /usr/local/bin/imagereductor
ENTRYPOINT ["/usr/local/bin/imagereductor"]