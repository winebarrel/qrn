.PHONY: all
all: vet build

.PHONY: build
build:
	go build ./cmd/qrn

.PHONY: vet
vet:
	go vet ./...

.PHONY: clean
clean:
	rm -f qrn qrn.exe
