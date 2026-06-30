# Optional Linter

A custom [Buf](https://buf.build) lint plugin that validates scalar proto fields with defaults are marked as optional.

## Overview

This plugin validates that scalar fields with `(dev.planton.shared.options.default)` are marked as `optional` to enable proper field presence tracking in Protocol Buffer definitions.

## Rules

### DEFAULT_REQUIRES_OPTIONAL

**Purpose**: Ensures scalar fields with default values are marked as `optional` to enable proper field presence tracking.

**Background**: Proto3's implicit presence tracking cannot distinguish between "field not set" and "field set to zero value" for scalar types. This caused a critical bug where user-provided zero values were incorrectly replaced with defaults. By requiring the `optional` keyword on fields with defaults, we enable explicit presence tracking via pointers in generated code.

**Violation Example**:
```protobuf
syntax = "proto3";

message Example {
  // VIOLATION: Has default but not optional
  string namespace = 1 [(dev.planton.shared.options.default) = "external-dns"];
  int32 port = 2 [(dev.planton.shared.options.default) = "443"];
}
```

**Correct Usage**:
```protobuf
syntax = "proto3";

message Example {
  // CORRECT: Has default and is optional
  optional string namespace = 1 [(dev.planton.shared.options.default) = "external-dns"];
  optional int32 port = 2 [(dev.planton.shared.options.default) = "443"];
}
```

**What is checked**:
- ✅ Scalar fields with `(dev.planton.shared.options.default)` must be `optional`
- ✅ Fields without defaults can be optional or non-optional (no validation)
- ✅ Message fields are always implicitly optional (skip validation)
- ✅ Repeated fields (lists) cannot have defaults (skip validation)
- ✅ Map fields cannot have defaults (skip validation)
- ✅ Fields with `recommended_default` do NOT require optional (different extension)

## Installation

The plugin is automatically built and used when you run `make protos` in the `apis/` directory.

### Manual Installation

To install the plugin manually for local development:

```bash
cd buf/lint/planton
make build
```

## Usage

The plugin is automatically configured in `apis/buf.yaml` and runs as part of `buf lint`:

```yaml
version: v2
lint:
  use:
    - STANDARD
    - DEFAULT_REQUIRES_OPTIONAL  # Custom rule from this plugin
plugins:
  - plugin: buf.build/planton/optional-linter:v0.1.0
```

### Running Lint

```bash
cd apis
make lint
```

This will:
1. Build the plugin (if not already built)
2. Run `buf lint` which includes the custom rules

## Development

### Project Structure

```
buf/lint/planton/
├── cmd/
│   └── optional-linter/
│       └── main.go              # Plugin entry point
├── rules/
│   └── default_requires_optional.go  # Rule implementation
├── go.mod                       # Plugin dependencies
├── go.sum
├── Makefile                     # Build and publish automation
├── buf.plugin.yaml              # Plugin metadata
└── README.md                    # This file
```

### Adding New Rules

To add a new custom lint rule:

1. Create a new file in `rules/` (e.g., `my_new_rule.go`)
2. Define your rule using `check.RuleSpec`:
   ```go
   package rules

   import (
       "buf.build/go/bufplugin/check"
       "buf.build/go/bufplugin/check/checkutil"
   )

   var MyNewRule = &check.RuleSpec{
       ID:      "MY_NEW_RULE",
       Purpose: "Checks that something is correct.",
       Type:    check.RuleTypeLint,
       Handler: checkutil.NewFieldRuleHandler(checkMyNewRule),
   }

   func checkMyNewRule(_ context.Context, responseWriter check.ResponseWriter, _ check.Request, field protoreflect.FieldDescriptor) error {
       // Implement your validation logic here
       return nil
   }
   ```

3. Register the rule in `cmd/optional-linter/main.go`:
   ```go
   check.Main(&check.Spec{
       Rules: []*check.RuleSpec{
           rules.DefaultRequiresOptionalRule,
           rules.MyNewRule,  // Add your new rule
       },
   })
   ```

4. Add the rule ID to `apis/buf.yaml` under `lint.use`:
   ```yaml
   lint:
     use:
       - STANDARD
       - DEFAULT_REQUIRES_OPTIONAL
       - MY_NEW_RULE
   ```

### Dependencies

The plugin uses:
- `buf.build/go/bufplugin` v0.9.0 - Buf plugin framework
- `google.golang.org/protobuf` - Protobuf reflection APIs
- Local `planton` module for accessing custom options

### Building

```bash
cd buf/lint/planton
make build        # Local binary
make build-wasm   # WebAssembly binary for BSR
```

### Testing

The plugin is tested by running it against the actual proto files in the repository:

```bash
cd apis
make lint
```

All existing proto files should pass validation (no violations).

## Related Documentation

- [Buf Custom Lint Plugins](https://buf.build/docs/cli/buf-plugins/)
- [PluginRPC Protocol](https://buf.build/pluginrpc/pluginrpc/docs)
- [Proto Field Presence](https://protobuf.dev/programming-guides/field_presence/)
- [Changelog: Fix Proto Field Presence](../../../changelog/2025-10-18-fix-proto-field-presence.md)

## Troubleshooting

### Plugin not found

If `buf lint` reports that it cannot find `optional-linter`:

1. Ensure the plugin is built:
   ```bash
   cd buf/lint/planton
   make build
   ```

2. Verify `$(go env GOPATH)/bin` is in your `$PATH`

3. Check the plugin works:
   ```bash
   optional-linter --protocol
   ```
   Should output: `1`

### Build errors

If the plugin fails to build, ensure you have the correct Go version (1.24.0+) and all dependencies:

```bash
cd buf/lint/planton
go mod tidy
make build
```

## Maintenance

- **Upgrade bufplugin**: Update `buf.build/go/bufplugin` in `go.mod` and test thoroughly
- **Upgrade protovalidate**: Keep in sync with main project's version
- **Go workspace**: The `go.work` file at the root handles local module resolution

