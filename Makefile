GO  := /Users/arvinz/sdk/go1.25.4/bin/go
EXE := $(shell pwd)/bin/qt

.PHONY: build init download list python-install clean

build:
	$(GO) build -o bin/qt ./cmd/qt

init: build
	$(EXE) init

download: build
	$(EXE) download --all

list: build
	$(EXE) list

python-install:
	cd python && uv sync

clean:
	rm -rf bin/
	rm -rf data/
	rm -rf logs/
