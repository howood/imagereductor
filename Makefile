.PHONY: update, install,run,test,testv,lint

update:
	go mod tidy

install:
	cd /go/src/github.com/howood/imagereductor/imagereductor && export GO111MODULE=on && go install

run:
	export GO111MODULE=on && go run ./imagereductor/imagereductor.go -v

test:
	export GO111MODULE=on && go test ./...

testv:
	export GO111MODULE=on && go test ./... -v

lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.53.3 &&  \
	cd /go/src/github.com/howood/imagereductor &&  \
	./bin/golangci-lint run ./...
