.PHONY: update, install,run,test,testv,lint. fmt

update:
	go mod tidy

upgrade:
	go get -u ./...

install:
	cd /go/src/github.com/howood/imagereductor/imagereductor && go install

run:
	go run ./imagereductor/imagereductor.go -v

test:
	go test ./...

testv:
	go test ./... -v

lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v2.0.2 &&  \
	./bin/golangci-lint run ./...

fmt:
	go install golang.org/x/tools/cmd/goimports@v0.28.0
	go install mvdan.cc/gofumpt@v0.7.0
	goimports -w .
	gofumpt -w .