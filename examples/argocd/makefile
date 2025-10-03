.PHONY: all
all: up lint test build

.PHONY: up
up:
	# do not update, breaks builds
	# go get -u ./...
	# go mod tidy

.PHONY: lint
lint: 
	golangci-lint fmt ./...
	golangci-lint run --allow-parallel-runners ./...

.PHONY: test
test:
	go test -count=1 ./...

.PHONY: build
build:
	go build -o ../../bin/ ./...