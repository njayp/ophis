.PHONY: all
all: up lint test build

.PHONY: up
up:
	go get -u ./...
	go mod tidy

.PHONY: lint
lint: 
	golangci-lint fmt ./...
	golangci-lint run ./...

.PHONY: test
test:
	go test -count=1 ./...

.PHONY: build
build:
	go build -o ../../bin/ ./...