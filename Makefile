SHELL   := /bin/bash
VERSION := v1.3.4
GOOS      := $(shell go env GOOS)
GOARCH    := $(shell go env GOARCH)

.PHONY: all
all: build

.PHONY: build
build:
	go build -ldflags "-X main.version=$(VERSION)" ./cmd/qrn

.PHONY: package
package: clean build
	gzip qrn -c > qrn_$(VERSION)_$(GOOS)_$(GOARCH).gz
	sha1sum qrn_$(VERSION)_$(GOOS)_$(GOARCH).gz > qrn_$(VERSION)_$(GOOS)_$(GOARCH).gz.sha1sum

.PHONY: clean
clean:
	rm -f qrn
