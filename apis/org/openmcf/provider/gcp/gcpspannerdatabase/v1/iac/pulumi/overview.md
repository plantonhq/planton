# GcpSpannerDatabase - Pulumi Module Overview

## Architecture

```
main.go (entrypoint)
  └── module.Resources()
       ├── initializeLocals()  → Locals struct
       └── spannerDatabase()   → spanner.NewDatabase()
```

## Module Structure

| File | Purpose |
|---|---|
| `main.go` | Entry point, provider setup, calls `spannerDatabase()` |
| `locals.go` | Initializes `Locals` struct from stack input |
| `spanner_database.go` | Creates the Spanner database with all configuration |
| `outputs.go` | Defines output constant names |

## Key Design Notes

### No Labels

Spanner databases do not support GCP labels. Unlike GcpSpannerInstance (which computes and applies framework labels), this module's `Locals` struct does not include a `GcpLabels` map.

### Conditional Field Setting

Optional fields are only set when the user provides them:
- `DatabaseDialect` -- only set if non-empty (GCP defaults to GOOGLE_STANDARD_SQL)
- `VersionRetentionPeriod` -- only set if non-empty (GCP defaults to "1h")
- `EncryptionConfig` -- only set if `kms_key_name` is provided
- `Ddls` -- only set if the list is non-empty
- `DefaultTimeZone` -- only set if non-empty
- `EnableDropProtection` -- only set if true

### StringValueOrRef Usage

Three fields use `StringValueOrRef` for cross-resource references:
- `spec.ProjectId.GetValue()` -- resolves project ID
- `spec.Instance.GetValue()` -- resolves Spanner instance name
- `spec.KmsKeyName.GetValue()` -- resolves KMS key fully qualified name

### Output Construction

The `database_id` output is constructed from known inputs rather than extracted from the resource, matching the fully qualified path format: `projects/{project}/instances/{instance}/databases/{name}`.

## Dependencies

- `github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/spanner` -- Spanner database resource
- `github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider` -- GCP provider setup
