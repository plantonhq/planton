.PHONY: build test install-pulumi-plugins

build:
	go build -o pulumi-azure-virtual-machine .

install-pulumi-plugins:
	pulumi plugin install resource azure-native

test: install-pulumi-plugins
	@if [ -z "$$STACK_INPUT" ]; then \
		echo "Error: STACK_INPUT environment variable is not set"; \
		exit 1; \
	fi
	pulumi login --local
	pulumi stack init test --non-interactive || true
	pulumi preview --non-interactive
