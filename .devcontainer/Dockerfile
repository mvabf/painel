# Ambiente de desenvolvimento: Go, Python com Pillow e as fontes DejaVu (para
# o GIF do README). Usado pelo Dev Container e pelos alvos docker-* do Makefile.
FROM golang:1.24-bookworm

RUN apt-get update \
    && apt-get install -y --no-install-recommends make python3 python3-pil fonts-dejavu-core \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download && chmod -R a+rwX /go

ENV GOCACHE=/tmp/go-build
