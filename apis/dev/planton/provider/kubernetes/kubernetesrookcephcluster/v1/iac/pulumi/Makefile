.PHONY: build install-pulumi-plugins test

build:
	go build -o pulumi-kubernetes-rook-ceph-cluster .

install-pulumi-plugins:
	pulumi plugin install resource kubernetes 4.18.3

test: build install-pulumi-plugins
	@echo "Running Pulumi preview with hack manifest..."
	export PLANTON_CLOUD_RESOURCE_MANIFEST_PATH=../hack/manifest.yaml && \
	pulumi preview --stack dev --non-interactive
