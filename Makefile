BIN := $(HOME)/.local/bin/painel
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X main.version=$(VERSION)

.PHONY: build test install

build:
	go build -ldflags "$(LDFLAGS)" -o painel .

test:
	go test ./...

install:
	go build -ldflags "$(LDFLAGS)" -o $(BIN) .
	@echo "instalado em $(BIN)"
