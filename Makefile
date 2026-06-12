GO  ?= go
EXE := $(shell pwd)/bin/qt

.PHONY: build init download backtest run stats python-install clean

build:
	$(GO) build -o bin/qt ./cmd/qt

init: build
	$(EXE) init

download: build
	$(EXE) download --all

backtest: build
	$(EXE) backtest --strategy ma_cross --code 600519.SH

run: build
	$(EXE) run

stats: build
	$(EXE) stats

python-install:
	cd python && uv sync

# 只清理构建产物；行情数据库不随 clean 删除，如需重置请手动 rm -rf data/
clean:
	rm -rf bin/
