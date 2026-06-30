.PHONY: build install-pulumi-plugins test

build:
	go build -o bin/pulumi-cloudflare-dns-record .

install-pulumi-plugins:
	pulumi plugin install resource cloudflare v6.0.0

test: install-pulumi-plugins
	@echo "Running Pulumi preview with test manifest..."
	@export STACK_INPUT=$$(cat ../hack/manifest.yaml | base64) && \
	pulumi preview --stack test --non-interactive
