.PHONY: update install run test testv coverage lint fmt

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

coverage:
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@go tool cover -func=coverage.out | grep total | awk '{print "Total coverage: " $$3}'

coverage-view:
	@if [ ! -f coverage.html ]; then \
		echo "Coverage report not found. Run 'make coverage' first."; \
		exit 1; \
	fi
	open coverage.html || xdg-open coverage.html || start coverage.html

lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v2.6.2 &&  \
	./bin/golangci-lint run ./...

fmt:
	go install golang.org/x/tools/cmd/goimports@v0.37.0
	go install mvdan.cc/gofumpt@v0.9.1
	goimports -w .
	gofumpt -w .

clean:
	rm -f coverage.out coverage.html
	find . -name '*.test' -delete
	find . -name '*.prof' -delete