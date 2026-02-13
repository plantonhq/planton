SHELL := /bin/bash

.PHONY: deps dev build start run lint clean preview-site export-static copy-docs generate-structure

PORT ?= 3000

deps:
	yarn install

generate-structure: copy-docs
	yarn tsx scripts/generate-docs-structure.ts

dev: generate-structure
	yarn dev --port $(PORT)

run: dev

copy-docs:
	yarn copy-docs

build: deps copy-docs
	yarn build

start:
	yarn start

lint:
	yarn lint

export-static: build

preview-site: export-static
	yarn serve -s out -l $(PORT)

clean:
	rm -rf .next out public/docs-structure.json
