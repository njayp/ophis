.PHONY: all
all: up lint test build

.PHONY: up
up:
	# do not update all, breaks builds
	# go get -u ./...
	go get -u github.com/argoproj/argo-cd/v3
	go mod tidy

.PHONY: lint
lint: 
	golangci-lint fmt ./...
	golangci-lint run --allow-parallel-runners ./...

.PHONY: test
test:
	go test ./...

.PHONY: build
build:
	go build -o ../../bin/ ./...