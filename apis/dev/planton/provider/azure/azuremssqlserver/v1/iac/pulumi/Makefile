COMPONENT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
MODULE_DIR := $(COMPONENT_DIR)/module

.PHONY: deps build test clean

deps:
	cd $(MODULE_DIR) && go mod tidy

build:
	cd $(MODULE_DIR) && go build ./...
	go build ./...

test:
	cd $(MODULE_DIR) && go test -v ./...

clean:
	rm -f $(COMPONENT_DIR)/azure-mssql-server

run:
	cd $(COMPONENT_DIR) && go run main.go
