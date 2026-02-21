# Pulumi Module Overview

## Module Architecture

The AliCloudKmsKey Pulumi module is organized into three files under `iac/pulumi/module/`:

| File | Responsibility |
|------|---------------|
| `main.go` | Controller -- creates the provider and KMS key resource, exports outputs |
| `locals.go` | Transformations -- tag computation, default resolution, bool-to-string conversion |
| `outputs.go` | Constants -- defines output key names exported to the stack |

The entry point binary at `iac/pulumi/main.go` loads the stack input (manifest YAML -> AliCloudKmsKeyStackInput) and delegates to `module.Resources()`.

## Control Flow

```
LoadStackInput (manifest YAML -> AliCloudKmsKeyStackInput)
    |
initializeLocals() -> Locals{Tags, AliCloudKmsKey}
    |
NewProvider("alicloud", region)
    |
kms.NewKey("name", KeyArgs{...})
    |
ctx.Export(key_id, arn)
```

## Key Design Decisions

### Bool-to-String Conversion

The Alibaba Cloud KMS provider represents `automatic_rotation` and `deletion_protection` as strings (`"Enabled"` / `"Disabled"`). The proto spec uses `bool` for a cleaner user experience. The `locals.go` file contains `automaticRotation()` and `deletionProtection()` functions that handle this conversion.

### Default Resolution

Optional fields with defaults (`key_spec`, `key_usage`, `protection_level`, `pending_window_in_days`, `automatic_rotation`, `deletion_protection`) are resolved in `locals.go` helper functions. Each function checks if the proto optional field is set; if not, it returns the documented default value.

### Tag Merging

Tags follow the standard OpenMCF pattern: base tags (`resource`, `resource_name`, `resource_kind`) are computed, then merged with optional organization/environment tags and user-specified tags. User tags take precedence.

## Output Keys

| Constant | Value | Source |
|----------|-------|--------|
| `OpKeyId` | `"key_id"` | `key.ID()` (Pulumi resource ID = KMS key ID) |
| `OpArn` | `"arn"` | `key.Arn` (computed by the provider) |
