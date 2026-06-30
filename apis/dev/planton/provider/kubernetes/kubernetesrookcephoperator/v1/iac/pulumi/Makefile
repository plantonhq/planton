.PHONY: build test install-pulumi-plugins

build:
	go build -o /dev/null .

install-pulumi-plugins:
	pulumi plugin install resource kubernetes v4.18.3

test: install-pulumi-plugins
	pulumi preview --stack local --non-interactive
