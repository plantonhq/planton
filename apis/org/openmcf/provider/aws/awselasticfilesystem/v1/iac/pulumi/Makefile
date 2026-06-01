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
build:deps vet fmt

.PHONY: update-deps
update-deps:
	go get github.com/plantonhq/openmcf@latest

