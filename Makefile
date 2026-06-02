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
	$(MAKE) -C apis build
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
	$(MAKE) -C pkg/kubernetes/kubernetestypes build

.PHONY: build-go
build-go: fmt deps vet
	GOOS=darwin GOARCH=amd64 ${build_cmd} -o ${build_dir}/${name}-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 ${build_cmd} -o ${build_dir}/${name}-darwin-arm64 .
	GOOS=linux GOARCH=amd64 ${build_cmd} -o ${build_dir}/${name}-linux .
	openssl dgst -sha256 ${build_dir}/${name}-darwin-arm64
	openssl dgst -sha256 ${build_dir}/${name}-linux

.PHONY: build-cli
build-cli: build-go

.PHONY: build
build: protos generate-cloud-resource-kind-map bazel-mod-tidy bazel-gazelle bazel-build-cli build-cli e2e-matrix

${build_dir}/${name}: build-go

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

.PHONY: package-content
package-content:  ## Package all content zips (presets, iac-source, catalog-pages, proto-source)
	bash tools/ci/release/package_content.sh ${version}

.PHONY: release-buf
release-buf:
	cd apis && buf push && buf push --label ${version}

.PHONY: next-version
next-version:  ## show what the next version would be
	@python3 tools/ci/release/next_version.py $(bump)

.PHONY: snapshot
snapshot: deps  ## build a local snapshot using GoReleaser
	goreleaser release --snapshot --clean --skip=publish

.PHONY: release
release:  ## auto-bump version, tag & push (bump=major|minor|patch, default: patch). Override with version=vX.Y.Z
	@if [ "$(VERSION_EXPLICIT)" = "true" ]; then \
		rel_version="$(version)"; \
		echo "Releasing: $$rel_version (explicit version)"; \
	else \
		rel_version=$$(python3 tools/ci/release/next_version.py $(bump)); \
		echo "Releasing: $$rel_version ($(bump) bump)"; \
	fi; \
	git tag -a $$rel_version -m "$$rel_version"; \
	git push origin $$rel_version

.PHONY: test-and-release
test-and-release: test release

.PHONY: run-docs
run-docs:
	$(MAKE) -C docs run

.PHONY: build-docs
build-docs:
	$(MAKE) -C docs build

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

# ── E2E Tests ─────────────────────────────────────────────────────────────────
.PHONY: e2e-test-kubernetes
e2e-test-kubernetes:  ## Run all Kubernetes E2E tests -- Tier 1 + Tier 2 + Tier 3 + Tier 4 (requires kind, pulumi, kubectl, Docker)
	go test -tags=e2e -timeout=360m -v -count=1 ./e2e/...

.PHONY: e2e-test-kubernetes-tier1
e2e-test-kubernetes-tier1:  ## Run Kubernetes Tier 1 (native K8s) E2E tests only
	go test -tags=e2e -timeout=60m -v -count=1 -run "Test(KubernetesNamespace|KubernetesDeployment|KubernetesStatefulSet|KubernetesSecret|KubernetesService|KubernetesCronJob|KubernetesJob|KubernetesDaemonSet|KubernetesManifest)_" ./e2e/...

.PHONY: e2e-test-kubernetes-tier2
e2e-test-kubernetes-tier2:  ## Run Kubernetes Tier 2 (Helm-based) E2E tests only
	go test -tags=e2e -timeout=120m -v -count=1 -run "Test(KubernetesRedis|KubernetesGrafana|KubernetesOpenBao|KubernetesArgoCD|KubernetesLocust|KubernetesNats|KubernetesNeo4j|KubernetesJenkins|KubernetesSolrOperator|KubernetesPerconaMongoOperator|KubernetesPerconaMysqlOperator|KubernetesPerconaPostgresOperator|KubernetesGitlab)_" ./e2e/...

.PHONY: e2e-test-kubernetes-tier3
e2e-test-kubernetes-tier3:  ## Run Kubernetes Tier 3 (operator-dependent) E2E tests -- fixtures deployed automatically
	go test -tags=e2e -timeout=120m -v -count=1 -run "Test(KubernetesPostgres|KubernetesKafka|KubernetesElasticsearch|KubernetesMongodb|KubernetesSolr|KubernetesClickHouse)_" ./e2e/...

.PHONY: e2e-test-kubernetes-tier4
e2e-test-kubernetes-tier4:  ## Run Kubernetes Tier 4 (operators, addons, cluster infra) E2E tests
	go test -tags=e2e -timeout=150m -v -count=1 -run "Test(KubernetesZalandoPostgresOperator|KubernetesStrimziKafkaOperator|KubernetesElasticOperator|KubernetesAltinityOperator|KubernetesGatewayApiCrds|KubernetesGhaRunnerScaleSetController|KubernetesRookCephOperator|KubernetesExternalSecrets|KubernetesIngressNginx|KubernetesTekton|KubernetesTektonOperator|KubernetesIstio|KubernetesIstioBaseCrds|KubernetesGatewayClass|KubernetesGateway|KubernetesHttpRoute|KubernetesGrpcRoute|KubernetesTcpRoute|KubernetesTlsRoute|KubernetesReferenceGrant|KubernetesPeerAuthentication|KubernetesRequestAuthentication|KubernetesAuthorizationPolicy|KubernetesServiceEntry|KubernetesEnvoyFilter)_" ./e2e/...

# ── Terraform-only E2E targets (requires kind, tofu/terraform, kubectl, Docker) ──

.PHONY: e2e-test-kubernetes-terraform-tier1
e2e-test-kubernetes-terraform-tier1:  ## Run Kubernetes Tier 1 Terraform E2E tests only
	go test -tags=e2e -timeout=60m -v -count=1 -run "Test(KubernetesNamespace|KubernetesDeployment|KubernetesStatefulSet|KubernetesSecret|KubernetesService|KubernetesCronJob|KubernetesJob|KubernetesDaemonSet|KubernetesManifest)_Terraform" ./e2e/...

.PHONY: e2e-test-kubernetes-terraform-tier2
e2e-test-kubernetes-terraform-tier2:  ## Run Kubernetes Tier 2 Terraform (Helm-based) E2E tests only
	go test -tags=e2e -timeout=120m -v -count=1 -run "Test(KubernetesRedis|KubernetesGrafana|KubernetesArgoCD|KubernetesLocust|KubernetesNats|KubernetesSolrOperator|KubernetesPerconaMongoOperator|KubernetesPerconaMysqlOperator|KubernetesPerconaPostgresOperator)_Terraform" ./e2e/...

.PHONY: e2e-test-kubernetes-terraform-tier3
e2e-test-kubernetes-terraform-tier3:  ## Run Kubernetes Tier 3 Terraform (operator-dependent) E2E tests
	go test -tags=e2e -timeout=120m -v -count=1 -run "Test(KubernetesPostgres|KubernetesKafka|KubernetesElasticsearch|KubernetesMongodb|KubernetesSolr|KubernetesClickHouse)_Terraform" ./e2e/...

.PHONY: e2e-test-kubernetes-terraform-tier4
e2e-test-kubernetes-terraform-tier4:  ## Run Kubernetes Tier 4 Terraform (operators, addons) E2E tests
	go test -tags=e2e -timeout=150m -v -count=1 -run "Test(KubernetesZalandoPostgresOperator|KubernetesStrimziKafkaOperator|KubernetesElasticOperator|KubernetesAltinityOperator|KubernetesGatewayApiCrds|KubernetesGhaRunnerScaleSetController|KubernetesRookCephOperator|KubernetesExternalSecrets|KubernetesTekton|KubernetesIstioBaseCrds|KubernetesGatewayClass|KubernetesGateway|KubernetesHttpRoute|KubernetesGrpcRoute|KubernetesTcpRoute|KubernetesTlsRoute|KubernetesReferenceGrant|KubernetesPeerAuthentication|KubernetesRequestAuthentication|KubernetesAuthorizationPolicy|KubernetesServiceEntry|KubernetesEnvoyFilter)_Terraform" ./e2e/...

# ── Auth0 E2E targets ────────────────────────────────────────────────────────

.PHONY: e2e-test-auth0
e2e-test-auth0:  ## Run all Auth0 E2E tests (requires AUTH0_DOMAIN, AUTH0_CLIENT_ID, AUTH0_CLIENT_SECRET)
	go test -tags=e2e -timeout=20m -v -count=1 ./e2e/auth0/...

.PHONY: e2e-test-auth0-pulumi
e2e-test-auth0-pulumi:  ## Run Auth0 Pulumi E2E tests only
	go test -tags=e2e -timeout=20m -v -count=1 -run ".*_Pulumi" ./e2e/auth0/...

.PHONY: e2e-test-auth0-terraform
e2e-test-auth0-terraform:  ## Run Auth0 Terraform E2E tests only
	go test -tags=e2e -timeout=20m -v -count=1 -run ".*_Terraform" ./e2e/auth0/...

# ── Generic component E2E targets ────────────────────────────────────────────

.PHONY: e2e-test-component
e2e-test-component:  ## Single component E2E test (usage: make e2e-test-component component=KubernetesNamespace)
	go test -tags=e2e -timeout=15m -v -count=1 -run "Test.*$(component)" ./e2e/...

.PHONY: e2e-matrix
e2e-matrix:  ## Regenerate E2E GitHub Actions matrix JSON from profiles
	go run . e2e discover \
		--provider kubernetes --status green --output github-matrix \
		> .github/e2e-matrix-kubernetes.json

.PHONY: e2e-build
e2e-build:  ## Compile E2E tests without running them
	go build -tags=e2e ./e2e/...

.PHONY: e2e-vet
e2e-vet:  ## Run go vet on E2E packages
	go vet ./e2e/framework/...
	go vet -tags=e2e ./e2e/...

# ── Base Images ───────────────────────────────────────────────────────────────
.PHONY: build-iac-runner-base-image
build-iac-runner-base-image:
	$(MAKE) -C base-images/iac-runner build-image
