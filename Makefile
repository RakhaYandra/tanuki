.PHONY: build run vet test race lint clean install

PREFIX ?= $(HOME)/.local
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X main.version=$(VERSION)

build:
	go build -ldflags "$(LDFLAGS)" -o bin/tanuki ./cmd/tanuki

release:
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/tanuki-linux-amd64 ./cmd/tanuki
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/tanuki-linux-arm64 ./cmd/tanuki
	sha256sum dist/tanuki-linux-* > dist/SHA256SUMS

lint:
	test -z "$$(gofmt -l cmd internal)" || (gofmt -l cmd internal && exit 1)
	go vet ./...

install: build
	install -Dm755 bin/tanuki $(PREFIX)/bin/tanuki

run: build
	./bin/tanuki --help

vet:
	go vet ./...

test:
	go test ./...

race:
	go test -race ./...

clean:
	rm -rf bin/ dist/
