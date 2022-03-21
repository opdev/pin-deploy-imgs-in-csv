VERSION=$(shell git rev-parse HEAD)
BINARY=pin-deploy-imgs-in-csv
RELEASE_TAG ?= "unknown"
# This is a test comment - it will be removed.

# build for your system
.PHONY: build
build:
	go build -o $(BINARY) -ldflags "-X github.com/opdev/pin-deploy-imgs-in-csv/cmd.version=$(RELEASE_TAG)" main.go
	@ls | grep pin-deploy-imgs-in-csv

.PHONY: tidy
tidy:
	go mod tidy -compat=1.17

.PHONY: build-cross
build-cross:
	make build-linux
	make build-darwin

.PHONY: build-linux
build-linux:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY)-linux-amd64 -ldflags "-X github.com/opdev/pin-deploy-imgs-in-csv/cmd.version=$(RELEASE_TAG)" main.go
	@ls | grep $(BINARY)-linux-amd64

.PHONY: build-darwin
build-darwin:
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY)-darwin-amd64 -ldflags "-X github.com/opdev/pin-deploy-imgs-in-csv/cmd.version=$(RELEASE_TAG)" main.go
	@ls | grep $(BINARY)-darwin-amd64

.PHONY: test
test: build
	./test/test.sh