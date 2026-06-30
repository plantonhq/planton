.PHONY: build
build: lint fmt
	rm -rf generated/stubs
	mkdir -p generated/stubs
	buf generate
	cp -R generated/stubs/go/github.com/plantonhq/planton/apis/. .
	rm -rf generated/stubs/go
	# Generate a BUILD.bazel for the Java stubs so Bazel can compile them.
	# This serves as a build-time gate: if any proto generates invalid Java
	# (e.g., reserved keyword conflicts), the build fails here instead of in
	# downstream consumers.
	@printf '# gazelle:ignore\njava_library(\n    name = "java",\n    srcs = glob(["**/*.java"]),\n    visibility = ["//visibility:public"],\n    deps = [\n        "@maven//:build_buf_protovalidate",\n        "@maven//:com_google_guava_guava",\n        "@maven//:com_google_protobuf_protobuf_java",\n        "@maven//:io_grpc_grpc_api",\n        "@maven//:io_grpc_grpc_protobuf",\n        "@maven//:io_grpc_grpc_stub",\n        "@maven//:javax_annotation_javax_annotation_api",\n    ],\n)\n' > generated/stubs/java/BUILD.bazel

.PHONY: protos
protos: build

.PHONY: fmt
fmt:
	buf format -w

.PHONY: lint
lint: buf-lint

.PHONY: build-optional-linter-plugin
build-optional-linter-plugin:
	@echo "Building buf lint plugin..."
	@cd ../buf/lint/optional-linter && $(MAKE) build

.PHONY: buf-lint
buf-lint: build-optional-linter-plugin
	buf lint
