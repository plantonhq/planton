.PHONY: deps
deps:
	go mod download
	go mod tidy

.PHONY: vet
vet:
	go vet ./...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: build
build: deps vet fmt

.PHONY: update-deps
update-deps:
	go get github.com/plantonhq/planton@latest

.PHONY: install-pulumi-plugins
install-pulumi-plugins:
	pulumi plugin install resource auth0

.PHONY: test
test: install-pulumi-plugins
	@echo "Running test with manifest..."
	STACK_INPUT_FILE=../hack/manifest.yaml pulumi preview --stack local


