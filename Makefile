SHELL   := /bin/bash
VERSION := v1.0.1

.PHONY: all
all: build

.PHONY: .build
build:
	go build -ldflags "-X main.version=$(VERSION)" ./cmd/qrn

.PHONY: clean
clean:
	rm -f qrn
