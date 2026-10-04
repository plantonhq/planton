# Tofu Backend Config Package

This package extracts Terraform/OpenTofu backend configuration from Planton
resource manifests using standardized `metadata.annotations`.

## Overview

The `backendconfig` package reads, parses, and validates Terraform/OpenTofu
backend configuration from manifest annotations. It ensures backend
configurations are complete and valid before they are used to initialize
Terraform state management.

Backend configuration lives in annotations — never labels — because
`metadata.labels` are derived into cloud-provider tags by planton IaC modules;
a platform key there would leak internal configuration onto the user's real
provider resources.

## Provisioner-Aware Annotation Keys

The annotation key prefix must match the provisioner in use; there is no
cross-prefix fallback.

| Annotation (prefix `terraform.` or `tofu.`) | Required | Purpose |
|---------------------------------------------|----------|---------|
| `<prefix>planton.dev/backend.type` | Yes | Backend type: `s3`, `gcs`, `azurerm`, `remote`, `local` |
| `<prefix>planton.dev/backend.bucket` | For remote backends | Bucket / container name |
| `<prefix>planton.dev/backend.key` | For remote backends | State file path within the bucket |
| `<prefix>planton.dev/backend.region` | S3 only (or `auto`) | AWS region, or `auto` for S3-compatible stores |
| `<prefix>planton.dev/backend.endpoint` | Only when `region: auto` | Custom S3-compatible endpoint (R2, MinIO) |

Key constants are built by `pkg/iac/tofu/tofuannotationkeys`.

## Core Types

### TofuBackendConfig

```go
type TofuBackendConfig struct {
    BackendType     string // "s3", "gcs", "azurerm", "remote", "local"
    BackendBucket   string // bucket or container name for remote backends
    BackendKey      string // state file path within the bucket
    BackendRegion   string // region for S3 backends ("auto" for S3-compatible)
    BackendEndpoint string // custom S3-compatible endpoint (R2, MinIO, etc.)
    S3Compatible    bool   // true when endpoint set or region == "auto"
}
```

## Key Functions

### ExtractFromManifest

```go
func ExtractFromManifest(manifest proto.Message, provisionerType string) (*TofuBackendConfig, error)
```

Extracts backend configuration from a manifest's `metadata.annotations`.

**Parameters:**
- `manifest` - The proto message containing the manifest
- `provisionerType` - The provisioner type (`"terraform"` or `"tofu"`), which
  selects the annotation prefix

**Behavior:**
1. Reads annotations only (labels are never consulted)
2. Returns `nil, nil` when none of the backend annotations are present
   (allows fallback to CLI flags or defaults)
3. Extracts whatever annotations exist; completeness is checked separately by
   `Validate()` (see `validate.go`), which reports every missing field with its
   exact annotation name

**Example Usage:**

```go
import (
    "github.com/plantonhq/planton/pkg/iac/tofu/backendconfig"
)

// Extract backend config for OpenTofu
config, err := backendconfig.ExtractFromManifest(manifest, "tofu")
if err != nil {
    return err
}

if config == nil {
    // No backend config in manifest - use CLI flags or defaults
    config = getDefaultBackendConfig()
}
```

## Validation

`Validate()` enforces per-backend completeness:

- `s3` requires `bucket`, `key`, and `region`; when `region` is `auto`, an
  `endpoint` is also required
- `gcs` and `azurerm` require `bucket` and `key`
- `local` requires no further fields

Validation errors carry the exact annotation name for each missing field, so
the user can fix the manifest without consulting docs.

## Backend-Specific Formats

### S3 Backend
```yaml
metadata:
  annotations:
    tofu.planton.dev/backend.type: s3
    tofu.planton.dev/backend.bucket: my-terraform-bucket
    tofu.planton.dev/backend.key: vpc/production/terraform.tfstate
    tofu.planton.dev/backend.region: us-west-2
```

### S3-Compatible Backend (Cloudflare R2, MinIO)
```yaml
metadata:
  annotations:
    tofu.planton.dev/backend.type: s3
    tofu.planton.dev/backend.bucket: my-r2-state
    tofu.planton.dev/backend.key: prod/terraform.tfstate
    tofu.planton.dev/backend.region: auto
    tofu.planton.dev/backend.endpoint: https://account-id.r2.cloudflarestorage.com
```

When `region: auto` is detected, the generated backend block includes the
compatibility skip flags that S3-compatible stores require.

### GCS Backend
```yaml
metadata:
  annotations:
    tofu.planton.dev/backend.type: gcs
    tofu.planton.dev/backend.bucket: my-terraform-bucket
    tofu.planton.dev/backend.key: kubernetes/staging/cluster
```

### Azure Storage Backend
```yaml
metadata:
  annotations:
    tofu.planton.dev/backend.type: azurerm
    tofu.planton.dev/backend.bucket: tfstate
    tofu.planton.dev/backend.key: rds/production
```

### Local Backend
```yaml
metadata:
  annotations:
    tofu.planton.dev/backend.type: local
```

## Testing

Run tests:
```bash
go test ./pkg/iac/tofu/backendconfig -v
```

Coverage includes: valid configurations for every backend type, nil return when
no backend annotations are present, platform keys under `labels` being ignored,
partial configurations reported field-by-field, and S3-compatible detection.

## Design Principles

1. **Fail-Safe**: Returns nil instead of error when annotations are absent
2. **Strict Validation**: Prevents partial or invalid configurations, with
   per-field actionable errors
3. **Backend Agnostic**: Doesn't handle backend-specific authentication
4. **Annotations Only**: Labels are user cloud-tag territory and are never read

## Related Packages

- `pkg/iac/tofu/tofuannotationkeys`: Defines the annotation key constants
- `pkg/reflection/metadatareflect`: Provides annotation extraction functionality
- `pkg/iac/tofu/generators`: Consumes backend configuration to generate `backend.tf`
