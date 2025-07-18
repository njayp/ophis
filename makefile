.PHONY: all
all: up gen lint test build

.PHONY: up
up:
	go get -u ./...
	go mod tidy

.PHONY: gen
gen:
	go generate ./...

.PHONY: lint
lint: 
	golangci-lint fmt ./...
	golangci-lint run ./...

.PHONY: test
test:
	go test ./...

.PHONY: build
build:
	go build -o bin/ ./...

