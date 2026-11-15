.PHONY: all build test clean run arena ui-build ui-dev

BINARY_NAME=evalarena
BIN_DIR=bin

all: build

ui-build:
	cd web && npm install && npm run build

ui-dev:
	cd web && npm run dev

build:
	go build -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/evalarena

test:
	go test -v -race ./...

test-short:
	go test -v ./...

clean:
	rm -rf $(BIN_DIR)
	rm -rf web/dist

run: build
	./$(BIN_DIR)/$(BINARY_NAME) run --config examples/evalarena.yaml

arena: build
	./$(BIN_DIR)/$(BINARY_NAME) arena --port 7070
