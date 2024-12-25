# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=etcd-k8s-extract
DB_FILE=./data

# All target
all: test build-linux build-windows build-darwin

# Build the project
build:
	$(GOBUILD) -o bin/$(BINARY_NAME) -v

# Build for linux
build-linux:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o bin/$(BINARY_NAME) -v

# Build for windows
build-windows:
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o bin/$(BINARY_NAME).exe -v

# Build for darwin
build-darwin:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o bin/$(BINARY_NAME)-darwin -v

# Run tests
test: 
	$(GOTEST) -v ./...

# Clean the build files
clean: 
	$(GOCLEAN)
	rm -f bin/*

# Go mod tidy
mod-tidy: 
	$(GOCMD) mod tidy

.PHONY: all build clean test deps build-linux run mod-tidy