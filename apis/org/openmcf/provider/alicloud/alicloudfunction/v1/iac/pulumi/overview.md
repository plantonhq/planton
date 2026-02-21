# Pulumi Module Overview — AliCloudFunction

## Module Architecture

The module lives in `module/` and consists of three files:

```
iac/pulumi/
├── Pulumi.yaml          # Pulumi project definition
├── main.go              # Entry point — deserializes stack input, calls module
├── module/
│   ├── main.go          # Controller — creates provider and fc.V3Function
│   ├── locals.go        # Transformations — tag merging, optional-field helpers
│   └── outputs.go       # Output constants — function_id, function_name, function_arn
├── BUILD.bazel
└── Makefile
```

| File | Responsibility |
|------|---------------|
| `main.go` | Creates alicloud provider, builds `fc.V3FunctionArgs`, conditionally attaches config blocks, creates the function, exports outputs |
| `locals.go` | Initializes `Locals` struct with merged tags (standard + user); provides `optionalString`, `optionalInt`, `optionalFloat64`, `optionalBool` helpers |
| `outputs.go` | Defines output key constants: `function_id`, `function_name`, `function_arn` |

Entry point: `Resources(ctx *pulumi.Context, stackInput *AliCloudFunctionStackInput) error`

## Control Flow

```
StackInput (target manifest + provider config)
  │
  ▼
initializeLocals()
  ├── merge standard tags (resource_name, resource_kind, resource_id, organization, environment)
  └── merge user-provided spec.Tags (user tags win on conflict)
  │
  ▼
alicloud.NewProvider(region)
  │
  ▼
Build fc.V3FunctionArgs
  ├── required: FunctionName, Handler, Runtime
  ├── optional scalars: Description, Cpu, MemorySize, Timeout, DiskSize, InstanceConcurrency
  ├── optional: InternetAccess, EnvironmentVariables, Layers, Tags, ResourceGroupId
  └── optional: Role (from StringValueOrRef)
  │
  ▼
Conditionally attach config blocks (only when spec field is non-nil):
  ├── code         → codeArgs()
  ├── vpcConfig    → vpcConfigArgs()
  ├── logConfig    → logConfigArgs()
  ├── customContainerConfig → customContainerConfigArgs()
  ├── customRuntimeConfig   → customRuntimeConfigArgs()
  ├── instanceLifecycleConfig → instanceLifecycleConfigArgs()
  ├── nasConfig    → nasConfigArgs()
  └── gpuConfig    → gpuConfigArgs()
  │
  ▼
fc.NewV3Function(ctx, functionName, args, pulumi.Provider(alicloudProvider))
  │
  ▼
ctx.Export(function_id, function_name, function_arn)
```

## Key Components

### Tag Merging (`locals.go`)

Standard tags are computed from `metadata`:

| Tag | Source |
|-----|--------|
| `resource` | `"true"` (hardcoded) |
| `resource_name` | `metadata.name` |
| `resource_kind` | `"alicloud_function"` (from CloudResourceKind enum) |
| `resource_id` | `metadata.id` (only if non-empty) |
| `organization` | `metadata.org` (only if non-empty) |
| `environment` | `metadata.env` (only if non-empty) |

User tags from `spec.tags` are merged last. On key conflicts, user tags win.

### Optional Field Helpers (`locals.go`)

Four helpers convert proto optional types to Pulumi input types, returning `nil`
for unset values to avoid overriding provider defaults:

- `optionalString(s string) → pulumi.StringPtrInput`
- `optionalInt(v *int32) → pulumi.IntPtrInput`
- `optionalFloat64(v *float64) → pulumi.Float64PtrInput`
- `optionalBool(v *bool) → pulumi.BoolPtrInput`

### Config Block Builders (`main.go`)

Each optional config section has a dedicated builder function that converts the
proto message into Pulumi SDK args:

| Function | Proto Message | Pulumi Type |
|----------|--------------|-------------|
| `codeArgs()` | `AliCloudFunctionCode` | `fc.V3FunctionCodePtrInput` |
| `vpcConfigArgs()` | `AliCloudFunctionVpcConfig` | `fc.V3FunctionVpcConfigPtrInput` |
| `logConfigArgs()` | `AliCloudFunctionLogConfig` | `fc.V3FunctionLogConfigPtrInput` |
| `customContainerConfigArgs()` | `AliCloudFunctionCustomContainerConfig` | `fc.V3FunctionCustomContainerConfigPtrInput` |
| `customRuntimeConfigArgs()` | `AliCloudFunctionCustomRuntimeConfig` | `fc.V3FunctionCustomRuntimeConfigPtrInput` |
| `instanceLifecycleConfigArgs()` | `AliCloudFunctionInstanceLifecycleConfig` | `fc.V3FunctionInstanceLifecycleConfigPtrInput` |
| `nasConfigArgs()` | `AliCloudFunctionNasConfig` | `fc.V3FunctionNasConfigPtrInput` |
| `gpuConfigArgs()` | `AliCloudFunctionGpuConfig` | `fc.V3FunctionGpuConfigPtrInput` |

### Health Check Builders

Two separate health check builders handle the different Pulumi types for
container vs. custom runtime health checks:

- `containerHealthCheckArgs()` → `fc.V3FunctionCustomContainerConfigHealthCheckConfigPtrInput`
- `runtimeHealthCheckArgs()` → `fc.V3FunctionCustomRuntimeConfigHealthCheckConfigPtrInput`

## Resource Relationships

The module creates a flat resource structure (no parent chaining):

```
Provider (alicloud)
  └── fc.V3Function (function_name)
```

There is only one infrastructure resource. The provider is bound to the function
via `pulumi.Provider(alicloudProvider)`.

## Design Decisions

**Single resource, no bundling**: The component creates exactly one
`fc.V3Function`. Triggers, aliases, versions, and provisioned concurrency are
separate lifecycle concerns. Bundling them would mean changing a trigger forces a
full function plan/update cycle.

**Nil-guarded config blocks**: Each optional config section is only included when
the spec field is non-nil. This prevents the provider from receiving empty/zero
config blocks that would override server-side defaults (e.g., an empty
`vpc_config` block would clear VPC attachment).

**StringValueOrRef resolution**: Foreign-key fields (`role`, `vpcConfig.vpcId`,
`vpcConfig.securityGroupId`, `logConfig.project`) call `.GetValue()` to extract
the resolved string. The OpenMCF runtime resolves `ref`-based values before the
Pulumi module executes.

**No default resolution in Go**: Unlike the AliCloudLogProject module, this
module does not apply Go-level defaults for optional scalars (`cpu`,
`memorySize`, `timeout`, etc.). These fields use `optionalInt`/`optionalFloat64`
which return `nil` for unset values, deferring to the FC provider's own defaults.

## Customization Guide

| Goal | File to Modify | Notes |
|------|---------------|-------|
| Add a new top-level field (e.g., `tracingConfig`) | `main.go` (add to `V3FunctionArgs`) | Also update `spec.proto` and regenerate |
| Add a new output (e.g., `function_url`) | `outputs.go` (add constant), `main.go` (add `ctx.Export`) | Also update `stack_outputs.proto` |
| Change tag behavior | `locals.go`, `initializeLocals()` | Modify tag map construction |
| Support function versioning | New file `version.go` | Would need `fc.V3FunctionVersion` resource |
| Add trigger support | Separate OpenMCF component recommended | Triggers have independent lifecycles |

## Next Steps

- [`examples.md`](./examples.md) — Runnable manifest examples
- [`README.md`](./README.md) — CLI usage and credentials
- [`../hack/manifest.yaml`](../hack/manifest.yaml) — Minimal test manifest
