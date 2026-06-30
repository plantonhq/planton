module github.com/plantonhq/planton/buf/lint/optional-linter

go 1.25.0

require (
	buf.build/go/bufplugin v0.9.0
	github.com/plantonhq/planton v0.0.0
	google.golang.org/protobuf v1.36.10
)

require (
	buf.build/gen/go/bufbuild/bufplugin/protocolbuffers/go v1.36.3-20250121211742-6d880cc6cc8d.1 // indirect
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.10-20251209175733-2a1774d88802.1 // indirect
	buf.build/gen/go/pluginrpc/pluginrpc/protocolbuffers/go v1.36.3-20241007202033-cf42259fcbfc.1 // indirect
	buf.build/go/protovalidate v1.1.0 // indirect
	buf.build/go/spdx v0.2.0 // indirect
	cel.dev/expr v0.24.0 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/google/cel-go v0.26.1 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/stoewer/go-strcase v1.3.1 // indirect
	golang.org/x/exp v0.0.0-20250813145105-42675adae3e6 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250811230008-5f3141c8851a // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250811230008-5f3141c8851a // indirect
	pluginrpc.com/pluginrpc v0.5.0 // indirect
)

replace (
	github.com/bufbuild/protovalidate-go => buf.build/go/protovalidate v1.0.0
	github.com/plantonhq/planton => ../../..
)
