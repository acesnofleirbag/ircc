BIN = ircc

all: run

.PHONY: build
build:
	@go build -o build/$(BIN) ./src

.PHONY: run
run: build
	@./build/$(BIN)

.PHONY: fmt
fmt:
	go fmt ./...

