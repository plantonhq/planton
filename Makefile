name=openmcf
name_local=openmcf
pkg=github.com/plantonhq/openmcf
build_dir=build
version?=$(shell python3 tools/ci/release/next_version.py patch 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X ${pkg}/internal/cli/version.Version=${version}"

# bump: major, minor, or patch (default)
bump ?= patch

# Detect if version was explicitly provided on command line
ifeq ($(origin version),command line)
VERSION_EXPLICIT := true
else
VERSION_EXPLICIT := false
endif

BAZEL?=./bazelw

# If OPENMCF_BUILDBUDDY_API_KEY is set, enable the :bb config and inject only the header.
ifneq ($(strip $(OPENMCF_BUILDBUDDY_API_KEY)),)
BAZEL_REMOTE_FLAGS=--config=bb --remote_header=x-buildbuddy-api-key=$$OPENMCF_BUILDBUDDY_API_KEY
else
BAZEL_REMOTE_FLAGS=
endif

build_cmd=go build -v ${LDFLAGS}

PARALLEL?=$(shell getconf _NPROCESSORS_ONLN 2>/dev/null || sysctl -n hw.ncpu)

clean-bazel:
	rm -rf .bazelbsp bazel-bin bazel-out bazel-testlogs bazel-openmcf

reset-ide: clean-bazel
	rm -rf .idea

.PHONY: deps
deps:
	go mod download
	go mod tidy

.PHONY: build_darwin
build_darwin:
	GOOS=darwin ${build_cmd} -o ${build_dir}/${name}-darwin .

.PHONY: buf-generate
buf-generate: protos

.PHONY: protos
protos:
	pushd apis;make build;popd
	@echo "Verifying generated Java stubs compile..."
	${BAZEL} build ${BAZEL_REMOTE_FLAGS} //apis/generated/stubs/java:java
	${BAZEL} run //:gazelle

.PHONY: buf-lint
buf-lint:
	$(MAKE) -C apis buf-lint

.PHONY: bazel-mod-tidy
bazel-mod-tidy:
	${BAZEL} mod tidy

.PHONY: gazelle
gazelle: bazel-gazelle

.PHONY: bazel-gazelle
bazel-gazelle:
	${BAZEL} run ${BAZEL_REMOTE_FLAGS} //:gazelle

.PHONY: clean-gazelle
clean-gazelle:
	@echo "Cleaning all BUILD.bazel files (excluding root)..."
	@find . -mindepth 2 -name "BUILD.bazel" -type f -delete
	@echo "✅ All BUILD.bazel files removed (root preserved)."

.PHONY: reset-gazelle
reset-gazelle: clean-gazelle gazelle
	@echo "✅ Gazelle reset complete. BUILD.bazel files regenerated."

.PHONY: bazel-build-cli
bazel-build-cli:
	${BAZEL} build ${BAZEL_REMOTE_FLAGS} //:openmcf

.PHONY: bazel-test
bazel-test:
	${BAZEL} test ${BAZEL_REMOTE_FLAGS} --test_output=errors //...

# Generates kind_map_gen.go containing ToMessageMap.
# The "-tags codegen" flag is REQUIRED to avoid chicken-and-egg compilation errors.
# See pkg/crkreflect/new_instance.go and pkg/crkreflect/codegen/main.go for details.
.PHONY: generate-cloud-resource-kind-map
generate-cloud-resource-kind-map:
	rm -f pkg/crkreflect/kind_map_gen.go
	go run -tags codegen ./pkg/crkreflect/codegen

.PHONY: generate-kubernetes-types
generate-kubernetes-types:
	pushd pkg/kubernetes/kubernetestypes;make build;popd

.PHONY: build-go
build-go: fmt deps vet
	GOOS=darwin GOARCH=amd64 ${build_cmd} -o ${build_dir}/${name}-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 ${build_cmd} -o ${build_dir}/${name}-darwin-arm64 .
	GOOS=linux GOARCH=amd64 ${build_cmd} -o ${build_dir}/${name}-linux .
	openssl dgst -sha256 ${build_dir}/${name}-darwin-arm64
	openssl dgst -sha256 ${build_dir}/${name}-linux

.PHONY: build-cli
build-cli: build-go

.PHONY: build-backend
build-backend:
	$(MAKE) -C app/backend build

.PHONY: build-frontend
build-frontend:
	$(MAKE) -C app/frontend build

.PHONY: build
build: protos generate-cloud-resource-kind-map bazel-mod-tidy bazel-gazelle bazel-build-cli build-cli build-backend build-frontend

${build_dir}/${name}: build-go

# ── Docker (Unified Image) ─────────────────────────────────────────────────────
DOCKER_IMAGE?=ghcr.io/plantonhq/openmcf
DOCKER_TAG?=latest
DOCKERFILE_UNIFIED=app/Dockerfile.unified

.PHONY: docker-build
docker-build:
	@echo "Building Docker image: $(DOCKER_IMAGE):$(DOCKER_TAG)"
	docker build -f $(DOCKERFILE_UNIFIED) -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

.PHONY: docker-build-multiarch
docker-build-multiarch:
	@echo "Building multi-architecture Docker image: $(DOCKER_IMAGE):$(DOCKER_TAG)"
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		-f $(DOCKERFILE_UNIFIED) \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG) \
		--push \
		.

.PHONY: docker-run
docker-run:
	@echo "Running Docker container from $(DOCKER_IMAGE):$(DOCKER_TAG)"
	docker run -d \
		--name openmcf-webapp \
		-p 3000:3000 \
		-p 50051:50051 \
		-v openmcf-mongodb:/data/db \
		-v openmcf-pulumi:/home/appuser/.pulumi \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: docker-stop
docker-stop:
	@echo "Stopping and removing container..."
	docker stop openmcf-webapp || true
	docker rm openmcf-webapp || true

.PHONY: docker-logs
docker-logs:
	docker logs -f openmcf-webapp

.PHONY: docker-shell
docker-shell:
	docker exec -it openmcf-webapp /bin/bash

# ──────────────────────────────────────────────────────────────────────────────

.PHONY: test
test:
	go test -race -v -count=1 -p $(PARALLEL) ./...

.PHONY: run
run: build
	${build_dir}/${name}

.PHONY: vet
vet:
	go vet ./cmd/...
	go vet ./internal/...
	go vet ./pkg/...
	go vet ./app/backend/...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: clean
clean:
	rm -rf ${build_dir}

.PHONY: checksum_darwin
checksum_darwin:
	@openssl dgst -sha256 ${build_dir}/${name}-darwin

.PHONY: checksum_linux
checksum_linux:
	openssl dgst -sha256 ${build_dir}/${name}-linux

.PHONY: checksum
checksum: checksum_darwin checksum_linux

.PHONY: local
local: build_darwin
	rm -f ${HOME}/.local/bin/${name_local}
	cp ./${build_dir}/${name}-darwin ${HOME}/.local/bin/${name_local}
	chmod +x ${HOME}/.local/bin/${name_local}

.PHONY: show-todo
show-todo:
	grep -r "TODO:" cmd internal

.PHONY: release-buf
release-buf:
	pushd apis;buf push;buf push --label ${version};popd

.PHONY: next-version
next-version:  ## show what the next version would be
	@python3 tools/ci/release/next_version.py $(bump)

.PHONY: snapshot
snapshot: deps  ## build a local snapshot using GoReleaser
	goreleaser release --snapshot --clean --skip=publish

.PHONY: release
release: test  ## auto-bump version, tag & push (bump=major|minor|patch, default: patch). Override with version=vX.Y.Z
	@if [ "$(VERSION_EXPLICIT)" = "true" ]; then \
		rel_version="$(version)"; \
		echo "Releasing: $$rel_version (explicit version)"; \
	else \
		rel_version=$$(python3 tools/ci/release/next_version.py $(bump)); \
		echo "Releasing: $$rel_version ($(bump) bump)"; \
	fi; \
	git tag -a $$rel_version -m "$$rel_version"; \
	git push origin $$rel_version

.PHONY: run-docs
run-docs:
	pushd docs;make run;popd

.PHONY: build-docs
build-docs:
	pushd docs;make build;popd

# ── website (site/) ────────────────────────────────────────────────────────────
.PHONY: run-site
run-site:
	$(MAKE) -C site dev

.PHONY: build-site
build-site:
	$(MAKE) -C site build

.PHONY: preview-site
preview-site:
	$(MAKE) -C site preview-site

# ── Base Images ───────────────────────────────────────────────────────────────
.PHONY: build-iac-runner-base-image
build-iac-runner-base-image:
	$(MAKE) -C base-images/iac-runner build-image
