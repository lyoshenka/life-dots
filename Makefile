.PHONY: build run dev test generate install

ADDR ?= :8080

build:
	go build -o life-dots ./cmd/life-dots

run: build
	./life-dots serve --addr $(ADDR)

# Local studio — no config dir; everything lives in the URL.
dev:
	go run ./cmd/life-dots serve --addr $(ADDR)

test:
	go test ./e2e/ ./internal/... -count=1 -timeout 120s

generate: build
	./life-dots generate --mode life --output wallpaper.png

install: build
	./life-dots service install --addr $(ADDR)
