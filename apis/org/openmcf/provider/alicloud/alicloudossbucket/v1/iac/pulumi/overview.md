# Overview

The AlicloudOssBucket Pulumi module creates a single Alibaba Cloud OSS bucket from an OpenMCF manifest. The module handles optional features (versioning, encryption, lifecycle rules, CORS, logging) by conditionally building the appropriate Pulumi resource arguments.

## Module Architecture

```
iac/pulumi/
├── main.go              Entry point: loads stack input, calls module.Resources()
└── module/
    ├── locals.go         Transforms stack input into computed values (tag merging)
    ├── main.go           Controller: creates provider, bucket, and all sub-configurations
    └── outputs.go        Defines output constant names (bucket_name, endpoints)
```

**Entry point** (`iac/pulumi/main.go`): Deserializes the Pulumi config into `AlicloudOssBucketStackInput` via `stackinput.LoadStackInput()`, then delegates to `module.Resources()`.

**Controller** (`module/main.go`): Initializes locals, creates the Alicloud provider scoped to `spec.region`, builds `oss.BucketArgs` with all configured features, provisions the bucket via `oss.NewBucket`, and exports three outputs.

**Locals** (`module/locals.go`): Builds the merged tag map. System tags (`resource`, `resource_name`, `resource_kind`) are set first. Metadata fields (`resource_id`, `organization`, `environment`) are added conditionally. User-defined `spec.tags` are merged last.

**Outputs** (`module/outputs.go`): String constants for the three export names, ensuring consistency between the Pulumi exports and `stack_outputs.proto`.

## Data Flow

```
AlicloudOssBucketStackInput
  │
  ├─ target.Metadata  ──► initializeLocals() ──► Locals.Tags (merged map)
  ├─ target.Spec.Tags ─┘
  │
  └─ target.Spec ──► Resources()
                        │
                        ├─ alicloud.NewProvider (region)
                        │
                        └─ oss.NewBucket
                              ├─ Versioning (conditional)
                              ├─ ServerSideEncryptionRule (conditional)
                              ├─ Logging (conditional)
                              ├─ CorsRules (conditional)
                              ├─ LifecycleRules (conditional)
                              │
                              ├─► Export: bucket_name
                              ├─► Export: extranet_endpoint
                              └─► Export: intranet_endpoint
```

## Design Decisions

**Single resource scope**: The bucket component creates one `oss.Bucket` with all inline configuration. Unlike the provider's newer separate-resource approach (e.g., `alicloud_oss_bucket_acl`), the inline approach keeps everything in one place for simpler state management.

**Versioning as boolean**: The proto spec exposes `versioning_enabled` (bool) rather than the provider's `Enabled`/`Suspended` enum. The `Suspended` state only applies to buckets that previously had versioning enabled, which is a runtime migration concern -- not relevant at provisioning time.

**Conditional building**: Each optional feature (versioning, encryption, logging, CORS, lifecycle) is conditionally added to `BucketArgs` only when configured. This prevents sending empty blocks to the Alibaba Cloud API.

**`optionalString` and `optionalStringFromPtr` helpers**: Proto3 `optional` fields generate Go pointer types (`*string`). The helper functions convert these to Pulumi's `StringPtrInput`, mapping empty/nil values to `nil` to avoid API errors on empty strings.

**Lifecycle rule builder**: The `buildLifecycleRules` function maps the simplified proto schema (days-based expiration, transitions, abort, noncurrent version expiration) to the Pulumi SDK's richer type hierarchy. Each lifecycle feature is conditionally included only when its corresponding field is set.

## Customization

| Goal | File to Change |
|------|---------------|
| Add spec fields to the bucket resource | `module/main.go` -- add args to `oss.BucketArgs` |
| Change tag logic or add new system tags | `module/locals.go` -- modify `initializeLocals()` |
| Add new stack outputs | `module/outputs.go` (constant) + `module/main.go` (export call) |
| Add companion resources (e.g., bucket policy) | Create a new file in `module/` and call it from `Resources()` |
