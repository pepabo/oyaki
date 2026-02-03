.PHONY: all
all: format test build lint

.PHONY: format
format:
	go fmt ./...

.PHONY: deps
deps:
	go get -d -t ./...

.PHONY: test
test: deps
	go test -v ./...

.PHONY: bench
bench: deps
	go test -bench . -benchmem -benchtime 5s -count 10

.PHONY: build
build: deps
	CGO_ENABLED=0 go build

.PHONY: lint
lint:
	golangci-lint run
	go vet ./...
