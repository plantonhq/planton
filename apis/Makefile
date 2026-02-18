.PHONY: build
build: lint fmt
	rm -rf generated/stubs
	mkdir -p generated/stubs
	buf generate
	cp -R generated/stubs/go/github.com/plantonhq/openmcf/apis/. .
	# Go stubs are copied to their destinations, so we only clean those
	rm -rf generated/stubs/go

.PHONY: protos
protos: build

.PHONY: fmt
fmt:
	buf format -w

.PHONY: lint
lint: buf-lint

.PHONY: build-optional-linter-plugin
build-optional-linter-plugin:
	@echo "Building buf lint plugin..."
	@cd ../buf/lint/optional-linter && $(MAKE) build

.PHONY: buf-lint
buf-lint: build-optional-linter-plugin
	buf lint
