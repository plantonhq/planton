# GcpFirestoreDatabase - Pulumi Module Overview

## Architecture

```
main.go (entrypoint)
  └── module.Resources()
       ├── initializeLocals()    → Locals struct
       └── firestoreDatabase()   → firestore.NewDatabase()
```

## Module Structure

| File | Purpose |
|---|---|
| `main.go` | Entry point, provider setup, calls `firestoreDatabase()` |
| `locals.go` | Initializes `Locals` struct from stack input |
| `firestore_database.go` | Creates the Firestore database with all configuration |
| `outputs.go` | Defines output constant names |

## Key Design Notes

### No Labels

Firestore databases do not support GCP labels. Unlike GcpBigtableInstance
(which computes and applies framework labels), this module's `Locals` struct
does not include a `GcpLabels` map.

### Conditional Field Setting

Optional fields are only set when the user provides them:
- `ConcurrencyMode` -- only set if non-empty (GCP defaults by database type)
- `PointInTimeRecoveryEnablement` -- only set if non-empty (defaults to DISABLED)
- `DeleteProtectionState` -- set via proactive default (DELETE_PROTECTION_DISABLED)
- `DatabaseEdition` -- only set if non-empty (defaults to STANDARD)
- `CmekConfig` -- only set if `kms_key_name` is provided
- `DeletionPolicy` -- always set to "DELETE" for IaC lifecycle management

### StringValueOrRef Usage

Two fields use `StringValueOrRef` for cross-resource references:
- `spec.ProjectId.GetValue()` -- resolves project ID
- `spec.KmsKeyName.GetValue()` -- resolves KMS key fully qualified name

### Deletion Policy

The module always sets `DeletionPolicy` to `"DELETE"`. Without this, the
Pulumi/Terraform default is `"ABANDON"`, which would leave the database
behind when the stack is destroyed. This is an internal IaC concern, not
exposed to users.

## Dependencies

- `github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/firestore` -- Firestore database resource
- `github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider` -- GCP provider setup
