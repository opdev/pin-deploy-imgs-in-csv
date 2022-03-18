VERSION=$(shell git rev-parse HEAD)
BINARY=pin-deploy-imgs-in-csv
RELEASE_TAG ?= "unknown"

.PHONY: build
build:
	go build -o $(BINARY) -ldflags "-X github.com/opdev/pin-deploy-imgs-in-csv/cmd.version=$(RELEASE_TAG)" main.go
	@ls | grep pin-deploy-imgs-in-csv

.PHONY: tidy
tidy:
	go mod tidy -compat=1.17