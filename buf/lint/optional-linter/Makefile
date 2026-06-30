.PHONY: build
build:
	@echo "Building local plugin binary..."
	@go build -o $(shell go env GOPATH)/bin/optional-linter ./cmd/optional-linter

.PHONY: build-wasm
build-wasm:
	@echo "Building WebAssembly binary..."
	@GOOS=wasip1 GOARCH=wasm go build -o optional-linter.wasm ./cmd/optional-linter

.PHONY: publish
publish: build-wasm
	@if [ -z "$(version)" ]; then \
		echo "Error: version is required. Usage: make publish version=v0.1.0"; \
		exit 1; \
	fi
	@echo "Publishing plugin version $(version)..."
	@MAJOR=$$(echo $(version) | cut -d. -f1); \
	MINOR=$$(echo $(version) | cut -d. -f1-2); \
	buf plugin push buf.build/planton/optional-linter \
		--binary=optional-linter.wasm \
		--create \
		--create-type=check \
		--create-visibility=public \
		--label $(version) \
		--label $$MINOR \
		--label $$MAJOR \
		--label main
	@echo "Successfully published $(version)"

.PHONY: clean
clean:
	@rm -f optional-linter.wasm

