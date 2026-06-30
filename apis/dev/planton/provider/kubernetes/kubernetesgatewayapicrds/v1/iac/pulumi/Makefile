.PHONY: build install-pulumi-plugins test

build:
	go build -o bin/pulumi-kubernetes-gateway-api-crds .

install-pulumi-plugins:
	pulumi plugin install resource kubernetes v4.18.0

test:
	@echo "Running Pulumi preview with test manifest..."
	STACK_INPUT_FILE_PATH=../hack/manifest.yaml pulumi preview --non-interactive
