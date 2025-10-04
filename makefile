.PHONY: all
all: up lint test
	go run github.com/magefile/mage@latest all

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
	go test ./...

.PHONY: build
build:
	go run github.com/magefile/mage@latest build

.PHONY: release
release: all
	@echo "Creating release..."
	@if ! git diff-index --quiet HEAD --; then \
		echo "Error: Working directory is not clean. Please commit or stash changes."; \
		exit 1; \
	fi
	@LATEST_TAG=$$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"); \
	echo "Latest tag: $$LATEST_TAG"; \
	VERSION=$$(echo $$LATEST_TAG | sed 's/^v//' | awk -F. '{print $$1"."$$2"."$$3+1}'); \
	NEW_TAG="v$$VERSION"; \
	echo "Creating new tag: $$NEW_TAG"; \
	git tag -a $$NEW_TAG -m "Release $$NEW_TAG"; \
	git push origin $$NEW_TAG; \
	echo "Successfully created and pushed tag: $$NEW_TAG"

