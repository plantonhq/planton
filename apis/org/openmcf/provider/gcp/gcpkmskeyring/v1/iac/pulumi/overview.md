# GcpKmsKeyRing — Pulumi Architecture Overview

## Execution Flow

```
main.go
  └── module.Resources(ctx, stackInput)
        ├── initializeLocals(ctx, stackInput)     → Locals struct
        ├── pulumigoogleprovider.Get(ctx, config)  → GCP provider
        └── keyRing(ctx, locals, gcpProvider)       → kms.KeyRing resource
              ├── ctx.Export("key_ring_id", ...)
              └── ctx.Export("key_ring_name", ...)
```

## Resource Mapping

| OpenMCF Spec Field | Pulumi Argument | Notes |
|--------------------|-----------------|-------|
| `project_id` | `Project` | Via `StringValueOrRef.GetValue()` |
| `key_ring_name` | `Name` | Direct mapping |
| `location` | `Location` | Direct mapping |

## Key Design Decisions

1. **No labels on resource**: KMS key rings do not support GCP labels. Labels are computed in `locals.go` for pattern consistency but not passed to the resource.
2. **ID output via `.ID()`**: The `key_ring_id` output uses Pulumi's `.ID()` which returns the fully qualified resource path. This is the exact format that `kms.CryptoKey` expects for its `KeyRing` argument.
3. **Single resource file**: Only `key_ring.go` needed — the resource is simple enough that splitting further would be over-engineering.
