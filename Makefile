.PHONY: build build-mac build-linux run dev test generate install

ADDR ?= :8080
ARCH_MAC ?= arm64
ARCH_LINUX ?= amd64
DIST ?= dist

# Native binary for the machine you're on.
build:
	mkdir -p $(DIST)
	go build -o $(DIST)/life-dots ./cmd/life-dots

# Cross-compile macOS (default arm64). Override: make build-mac ARCH_MAC=amd64
build-mac:
	mkdir -p $(DIST)
	GOOS=darwin GOARCH=$(ARCH_MAC) go build -o $(DIST)/life-dots-mac ./cmd/life-dots

# Cross-compile Linux (default amd64). Override: make build-linux ARCH_LINUX=arm64
build-linux:
	mkdir -p $(DIST)
	GOOS=linux GOARCH=$(ARCH_LINUX) go build -o $(DIST)/life-dots-linux ./cmd/life-dots

run: build
	./$(DIST)/life-dots serve --addr $(ADDR)

# Local studio with live reload (Go/HTML/CSS/JS). ADDR still works: make dev ADDR=:9090
dev:
	go run github.com/air-verse/air@v1.67.4 -build.args_bin=--addr -build.args_bin=$(ADDR)

test:
	go test ./e2e/ ./internal/... -count=1 -timeout 120s

generate: build
	./$(DIST)/life-dots generate --mode life --output wallpaper.png

install: build
	./$(DIST)/life-dots service install --addr $(ADDR)
