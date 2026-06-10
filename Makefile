GO  := /Users/arvinz/sdk/go1.25.4/bin/go
EXE := $(shell pwd)/bin/qt

.PHONY: build init download python-install clean

build:
	$(GO) build -o bin/qt ./cmd/qt

init: build
	$(EXE) init

download: build
	$(EXE) download --all

python-install:
	cd python && uv sync

clean:
	rm -rf bin/
	rm -rf data/
	rm -rf logs/
