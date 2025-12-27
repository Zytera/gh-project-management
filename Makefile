BINARY_NAME=gh-project-managment

## build: Build the binary
build:
	go build -o $(BINARY_NAME) .

## install: Install as gh extension
install: build
	gh extension remove $(BINARY_NAME) 2>/dev/null || true
	gh extension install .

## fmt: Format code
fmt:
	go fmt ./...

## vet: Run go vet
vet:
	go vet ./...

## dev: Build and install for development
dev: fmt vet build install
	@echo "✓ Development build installed"
