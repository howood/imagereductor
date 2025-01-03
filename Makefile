.PHONY: update, install,run,test,testv,lint. fmt

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
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.62.2 &&  \
	./bin/golangci-lint run ./...

fmt:
	go install golang.org/x/tools/cmd/goimports@v0.28.0
	go install mvdan.cc/gofumpt@v0.7.0
	goimports -w .
	gofumpt -w .