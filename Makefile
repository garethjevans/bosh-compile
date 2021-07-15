SHELL := /bin/bash
GO := GO111MODULE=on GO15VENDOREXPERIMENT=1 go
GO_NOMOD := GO111MODULE=off go
GO_VERSION := $(shell $(GO) version | sed -e 's/^[^0-9.]*\([0-9.]*\).*/\1/')
GO_DEPENDENCIES := $(shell find . -type f -name '*.go')
PACKAGE_DIRS := $(shell $(GO) list ./... | grep -v /vendor/ | grep -v e2e)

CGO_ENABLED = 0
BUILDTAGS :=

build:
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(BUILDTAGS) $(BUILDFLAGS) -o build/bc cmd/main.go

linux:
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 $(GO) build $(BUILDTAGS) $(BUILDFLAGS) -o build/linux/bc cmd/main.go

docker: linux
	docker build -t garethjevans/bosh-compile:latest .
	docker push garethjevans/bosh-compile:latest

docs: build
	./build/bc docs

clean:
	rm -fr build

lint:
	golangci-lint run

test:
	$(GO) test -cover ./...

.PHONY: build linux

