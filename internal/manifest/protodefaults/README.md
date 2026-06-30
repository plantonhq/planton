# Proto Field Defaults

Automatic application of default values from protobuf field options when loading manifests.

## What Are Proto Field Defaults?

Proto field defaults allow you to define sensible default values directly in your protobuf definitions, which are then automatically applied when manifests are loaded. This keeps manifests concise while ensuring fields get appropriate values even when not explicitly specified.

**Without defaults** (verbose):
```yaml
apiVersion: kubernetes.planton.dev/v1
kind: ExternalDnsKubernetes
metadata:
  name: external-dns
spec:
  namespace: external-dns      # Must specify
  externalDnsVersion: v0.19.0  # Must specify
  helmChartVersion: 1.19.0     # Must specify
  targetCluster:
    kubernetesProviderConfigId: k8s-cluster-01
```

**With defaults** (concise):
```yaml
apiVersion: kubernetes.planton.dev/v1
kind: ExternalDnsKubernetes
metadata:
  name: external-dns
spec:
  # namespace, versions omitted - automatically get defaults
  targetCluster:
    kubernetesProviderConfigId: k8s-cluster-01
```

The defaults are defined once in the proto file and honored everywhere manifests are loaded.

## How It Works

When you load a manifest, defaults are applied in this pipeline:

```
1. Read YAML file
2. Convert YAML → JSON
3. Unmarshal JSON → Proto message
4. Apply defaults ← This package
5. Apply CLI --set overrides
6. Return loaded manifest
```

Defaults are applied **after** unmarshaling (so explicit manifest values take precedence) but **before** CLI overrides (so flags can override defaults).

## Usage

### Defining Defaults in Proto Files

Add the `(dev.planton.shared.options.default)` extension to fields that should have defaults:

```protobuf
syntax = "proto3";

import "dev/planton/shared/options/options.proto";

message ExternalDnsKubernetesSpec {
  // CRITICAL: Use 'optional' keyword for all fields with defaults
  optional string namespace = 2 [(dev.planton.shared.options.default) = "external-dns"];
  optional string external_dns_version = 3 [(dev.planton.shared.options.default) = "v0.19.0"];
  optional string helm_chart_version = 4 [(dev.planton.shared.options.default) = "1.19.0"];
  optional int32 port = 5 [(dev.planton.shared.options.default) = "443"];
  optional bool enabled = 6 [(dev.planton.shared.options.default) = "true"];
}
```

**Supported types**: All proto scalar types (string, int32, int64, uint32, uint64, float, double, bool, enums).

**Not supported**: Lists (`repeated`), maps, and nested messages cannot have defaults.

### The `optional` Requirement

**Every field with a default MUST be marked as `optional`.** This is not negotiable—it's enforced by a buf lint rule.

**Why?** Proto3 has a subtle issue with field presence detection. Consider this scenario:

```yaml
# User wants to explicitly set port to 0 (disable feature)
spec:
  port: 0
```

Without `optional`, proto3 cannot distinguish between:
- User didn't set `port` (should get default: `443`)
- User explicitly set `port` to `0` (should stay `0`)

Both look identical to the generated code because `0` is the zero value for integers. Think of it like asking someone a question: **did they not answer, or did they explicitly say "none"?** Without `optional`, you can't tell the difference.

With `optional`, the generated Go code uses pointers:
- `nil` pointer = field not set → apply default
- Non-nil pointer pointing to `0` = field explicitly set to zero → preserve user's choice

This applies to all zero values:
- `""` (empty string)
- `0` (integers)
- `0.0` (floats)
- `false` (boolean)

### The Safeguard: Automated Enforcement

A custom buf lint plugin automatically enforces the `optional` requirement. If you forget it, you'll get a clear error during `make protos`:

```
spec.proto:68:3:Field "region" has a default value but is not marked as optional. 
Scalar fields with (dev.planton.shared.options.default) must use the 'optional' keyword 
to enable proper field presence detection.

Fix: optional string region = 4 [(dev.planton.shared.options.default) = "us-west-2"];
```

This safeguard prevents a common bug where users couldn't set fields to zero values—they would always be overridden with defaults.

See [`buf/lint/planton/README.md`](../../../buf/lint/planton/README.md) for details on the lint plugin.

## Best Practices

### When to Use Defaults

✅ **Good candidates for defaults**:
- Version numbers that change infrequently (`v0.19.0`)
- Standard namespaces (`external-dns`, `cert-manager`)
- Common port numbers (`443`, `8080`)
- Feature flags with safe defaults (`enabled: false`)
- Regional preferences (`us-west-2`)
- Resource sizing with reasonable defaults (`replicas: 3`)

❌ **Fields that should NOT have defaults**:
- Security credentials (API keys, tokens, passwords)
- Unique identifiers (names, IDs)
- Required user decisions (which cluster to deploy to)
- Organization-specific values (domain names, account IDs)
- Fields where "not set" has different semantics than a specific value

### Heuristics for Good Defaults

1. **Safe by default**: Defaults should be production-safe. Don't default to debug mode or insecure settings.

2. **Rarely changed**: If 80% of users need a different value, it's not a good default.

3. **Observable**: Users should easily discover what defaults were applied (`load-manifest` command shows them).

4. **Overridable**: Defaults should be easy to override in manifests or via `--set` flags.

5. **Stable**: Avoid defaults that reference moving targets or might become outdated quickly.

### Examples

**Good default** (version with stable release):
```protobuf
optional string external_dns_version = 3 [(dev.planton.shared.options.default) = "v0.19.0"];
```

**Bad default** (security-sensitive field):
```protobuf
// DON'T DO THIS - API tokens should never have defaults
optional string api_token = 5 [(dev.planton.shared.options.default) = "default-token"];
```

**Good default** (namespace with standard convention):
```protobuf
optional string namespace = 2 [(dev.planton.shared.options.default) = "cert-manager"];
```

**Bad default** (unique identifier):
```protobuf
// DON'T DO THIS - names should be unique and user-specified
optional string name = 1 [(dev.planton.shared.options.default) = "my-resource"];
```

## How Users Experience Defaults

### Minimal Manifests

Users write minimal manifests, specifying only what's unique to their deployment:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: CertManager
metadata:
  name: cert-manager
spec:
  targetCluster:
    kubernetesProviderConfigId: k8s-cluster-01
  # namespace, version automatically defaulted
```

### Viewing Applied Defaults

Use `load-manifest` to see the effective configuration with defaults applied:

```bash
planton load-manifest cert-manager.yaml
```

Output shows defaults:
```yaml
spec:
  namespace: cert-manager        # ← Default applied
  certManagerVersion: v1.13.0    # ← Default applied
  targetCluster:
    kubernetesProviderConfigId: k8s-cluster-01
```

### Overriding Defaults

Users can override defaults in two ways:

**1. In the manifest**:
```yaml
spec:
  namespace: custom-namespace    # Override default
  certManagerVersion: v1.14.0    # Override default
```

**2. Via CLI flags**:
```bash
planton deploy cert-manager.yaml \
  --set spec.namespace=custom-namespace \
  --set spec.certManagerVersion=v1.14.0
```

## Implementation Details

This package uses protobuf reflection to:

1. **Traverse** the proto message recursively
2. **Detect** fields with the `(dev.planton.shared.options.default)` extension
3. **Check** if the field is unset (using `Has()` method, which works correctly due to `optional`)
4. **Convert** the default string to the appropriate type
5. **Set** the field value

The implementation is type-safe and handles all proto scalar types with proper validation.

## FAQ

### Why can't I set a field to its zero value without `optional`?

Proto3's implicit presence tracking cannot distinguish between "field not set" and "field set to zero value." The `optional` keyword enables explicit presence tracking using pointers in generated code, allowing the default application logic to correctly detect when a field is truly unset.

### What happens if I define a default but forget `optional`?

The build will fail during `buf lint` with a clear error message telling you exactly how to fix it. This is enforced automatically—you cannot accidentally create fields with defaults that lack proper presence tracking.

### Can I have defaults for nested messages?

No. Defaults are only supported for scalar types. However, fields *within* nested messages can have defaults:

```protobuf
message Parent {
  NestedConfig config = 1;  // No default for the message itself
}

message NestedConfig {
  optional string namespace = 1 [(dev.planton.shared.options.default) = "default"];
}
```

### Do defaults work for repeated fields or maps?

No. Lists (`repeated`) and maps cannot have defaults. It's unclear what a sensible default would be for collection types.

### How do I remove a default from a field?

Remove the `(dev.planton.shared.options.default)` extension and the `optional` keyword (if the field doesn't need presence tracking for other reasons):

```protobuf
// Before
optional string namespace = 2 [(dev.planton.shared.options.default) = "default"];

// After
string namespace = 2;  // No default
```

### What if I want different defaults for different environments?

Use the `(dev.planton.shared.options.recommended_default)` extension for suggestions, or handle environment-specific defaults in your deployment tooling rather than in proto definitions. Proto defaults should be universal.

## Related Documentation

- [Custom Buf Lint Plugin](../../../buf/lint/planton/README.md) - Enforces the `optional` requirement
- [Proto Field Options](../../../apis/project/planton/shared/options/options.proto) - Definition of the `default` extension
- [Proto Field Presence](https://protobuf.dev/programming-guides/field_presence/) - Official protobuf documentation on field presence

## Contributing

When adding new cloud resource APIs:

1. Define sensible defaults in your proto files
2. Always use `optional` for fields with defaults
3. Run `make protos` to verify lint rules pass
4. Test defaults with `load-manifest` command
5. Document non-obvious default choices in comments

