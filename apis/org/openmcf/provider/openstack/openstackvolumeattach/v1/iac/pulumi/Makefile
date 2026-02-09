.PHONY: build install-pulumi-plugins test

build:
	go build -o bin/pulumi-openstack-volume-attach .

install-pulumi-plugins:
	pulumi plugin install resource openstack v5.4.0

test: install-pulumi-plugins
	@echo "Running Pulumi preview with test manifest..."
	@export STACK_INPUT=$$(cat ../hack/manifest.yaml | base64) && \
	pulumi preview --stack test --non-interactive
