.PHONY: build test install-pulumi-plugins

build:
	go build -o main .

install-pulumi-plugins:
	pulumi plugin install resource kubernetes v4.0.0

test: install-pulumi-plugins
	@echo "Running test deployment..."
	@export STACK_INPUT_JSON=$$(cat ../hack/manifest.yaml | yq -o=json) && \
		pulumi preview --stack dev

