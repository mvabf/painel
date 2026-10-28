BIN := $(HOME)/.local/bin/painel
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X main.version=$(VERSION)

.PHONY: build test install demo-gif

build:
	go build -ldflags "$(LDFLAGS)" -o painel .

test:
	go test ./...

install:
	go build -ldflags "$(LDFLAGS)" -o $(BIN) .
	@echo "instalado em $(BIN)"

# Regenera o GIF do README com dados fictícios (precisa de Python 3 + Pillow).
demo-gif:
	rm -rf .demo-frames
	PAINEL_DEMO_DIR=$(CURDIR)/.demo-frames go test -count=1 -run TestDemoFrames .
	python3 scripts/demo_gif.py .demo-frames docs/demo.gif
	rm -rf .demo-frames
