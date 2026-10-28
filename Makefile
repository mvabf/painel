BIN := $(HOME)/.local/bin/painel

.PHONY: build test install

build:
	go build -o painel .

test:
	go test ./...

install:
	go build -o $(BIN) .
	@echo "instalado em $(BIN)"
