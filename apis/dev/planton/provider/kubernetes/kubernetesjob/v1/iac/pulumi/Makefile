.PHONY: build test clean install-pulumi-plugins

# Build the Pulumi module
build:
	go build -o bin/pulumi-resource-kubernetes-job .

# Install required Pulumi plugins
install-pulumi-plugins:
	pulumi plugin install resource kubernetes v4.18.1

# Run the module with the test manifest
test: install-pulumi-plugins
	@echo "Running KubernetesJob Pulumi module test..."
	@export STACK_INPUT=$$(cat ../hack/manifest.yaml | base64) && \
	pulumi preview --non-interactive

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf .pulumi/

# Format Go code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run ./...
