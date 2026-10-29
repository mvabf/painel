BIN := $(HOME)/.local/bin/painel
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X main.version=$(VERSION)

.PHONY: build test install demo-gif docker-image docker-test docker-demo docker-demo-gif

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

# Os mesmos comandos dentro do Docker, sem Go nem Python na máquina. Se o
# build não achar a internet (VPN, DNS), use DOCKER_FLAGS=--network=host.
DOCKER_IMAGE := painel-dev
DOCKER_FLAGS ?=
DOCKER_RUN := docker run --rm $(DOCKER_FLAGS) -u $(shell id -u):$(shell id -g) -e HOME=/tmp -v $(CURDIR):/src -w /src

docker-image:
	docker build $(DOCKER_FLAGS) -t $(DOCKER_IMAGE) -f .devcontainer/Dockerfile .

docker-test: docker-image
	$(DOCKER_RUN) $(DOCKER_IMAGE) sh -c 'test -z "$$(gofmt -l .)" && go vet ./... && go test ./...'

docker-demo: docker-image
	$(DOCKER_RUN) -it $(DOCKER_IMAGE) go run . demo

docker-demo-gif: docker-image
	$(DOCKER_RUN) $(DOCKER_IMAGE) make demo-gif
