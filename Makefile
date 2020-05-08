SHELL   := /bin/bash
VERSION := v0.1.0

.PHONY: all
all: build

.PHONY: .build
build:
	go build -ldflags "-X main.version=$(VERSION)" ./cmd/qrn

.PHONY: clean
clean:
	rm -f qrn
