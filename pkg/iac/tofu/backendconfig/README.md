# Tofu Backend Config Package

This package provides functionality to extract Terraform/OpenTofu backend configuration from Planton resource manifests using standardized labels.

## Overview

The `backendconfig` package implements the logic to read, parse, and validate Terraform/OpenTofu backend configuration from manifest labels. It ensures backend configurations are complete and valid before they're used to initialize Terraform state management.

## Provisioner-Aware Labels

The package supports provisioner-specific labels:

| Provisioner | Backend Type Label | Backend Object Label |
|-------------|-------------------|---------------------|
| Terraform | `terraform.planton.dev/backend.type` | `terraform.planton.dev/backend.object` |
| OpenTofu | `tofu.planton.dev/backend.type` | `tofu.planton.dev/backend.object` |

For backward compatibility, OpenTofu also accepts the legacy `terraform.*` labels if `tofu.*` labels are not present.

## Core Types

### TofuBackendConfig

```go
type TofuBackendConfig struct {
    BackendType   string  // Backend type: "s3", "gcs", "azurerm", "local"
    BackendObject string  // Backend-specific object path
}
```

## Key Functions

### ExtractFromManifest

```go
func ExtractFromManifest(manifest proto.Message, provisionerType string) (*TofuBackendConfig, error)
```

Extracts Terraform/OpenTofu backend configuration from a manifest's metadata labels.

**Parameters:**
- `manifest` - The proto message containing the manifest
- `provisionerType` - The provisioner type ("terraform" or "tofu")

**Behavior:**
1. Checks for provisioner-specific labels first (e.g., `tofu.planton.dev/*`)
2. Falls back to legacy `terraform.*` labels if provisioner-specific labels are not found
3. Returns `nil, nil` if no backend labels are present (allows fallback to CLI/defaults)
4. Returns an error if labels are partially specified
5. Validates backend type against supported backends
6. Ensures non-empty values for both labels

**Example Usage:**

```go
import (
    "github.com/plantonhq/planton/pkg/iac/tofu/backendconfig"
)

// Extract backend config for OpenTofu
config, err := backendconfig.ExtractFromManifest(manifest, "tofu")
if err != nil {
    // Handle error - invalid backend configuration
    return err
}

if config == nil {
    // No backend config in manifest - use CLI flags or defaults
    config = getDefaultBackendConfig()
}

// Use the extracted configuration
fmt.Printf("Backend: %s://%s\n", config.BackendType, config.BackendObject)
```

## Validation Rules

### Supported Backend Types

The package validates that the backend type is one of:
- `s3` - Amazon S3
- `gcs` - Google Cloud Storage
- `azurerm` - Azure Storage
- `local` - Local filesystem

### Label Consistency

1. **All or Nothing**: Both labels must be specified together or neither
2. **Non-Empty**: Both label values must be non-empty strings
3. **Valid Type**: Backend type must be from the supported list

### Error Messages

```go
// No labels in manifest
"no labels found in manifest"

// Partial configuration
"both terraform.planton.dev/backend.type and terraform.planton.dev/backend.object must be specified together"

// Empty values
"Terraform backend labels cannot be empty"

// Invalid backend type
"unsupported backend type: <type>"
```

## Backend-Specific Formats

### S3 Backend
```yaml
backend.type: "s3"
backend.object: "my-terraform-bucket/vpc/production/main"
```

Translates to Terraform configuration:
```hcl
backend "s3" {
  bucket = "my-terraform-bucket"
  key    = "vpc/production/main"
  # region, dynamodb_table, etc. from environment
}
```

### GCS Backend
```yaml
backend.type: "gcs"
backend.object: "my-terraform-bucket/kubernetes/staging/cluster"
```

Translates to:
```hcl
backend "gcs" {
  bucket = "my-terraform-bucket"
  prefix = "kubernetes/staging/cluster"
  # credentials from environment
}
```

### Azure Storage Backend
```yaml
backend.type: "azurerm"
backend.object: "tfstate/rds/production"
```

Translates to:
```hcl
backend "azurerm" {
  container_name = "tfstate"
  key           = "rds/production"
  # storage_account_name, etc. from environment
}
```

## Testing

The package includes comprehensive test coverage:

```go
// Valid backends
- S3, GCS, Azure, Local configurations
- Nil return when no backend labels present

// Error cases
- Missing one of the two required labels
- Empty label values
- Unsupported backend types
- No labels in manifest
```

Run tests:
```bash
go test ./pkg/iac/tofu/backendconfig -v
```

## Integration Pattern

Here's how the CLI might integrate this package:

```go
func initializeTerraformBackend(manifest proto.Message, cliBackendOverride string) error {
    // Try to extract from manifest
    manifestConfig, err := backendconfig.ExtractFromManifest(manifest)
    if err != nil {
        return fmt.Errorf("invalid backend config in manifest: %w", err)
    }
    
    // Determine final backend configuration
    var backendType, backendObject string
    
    if cliBackendOverride != "" {
        // CLI override takes precedence
        backendType, backendObject = parseCliBackend(cliBackendOverride)
    } else if manifestConfig != nil {
        // Use manifest configuration
        backendType = manifestConfig.BackendType
        backendObject = manifestConfig.BackendObject
    } else {
        // Use defaults
        backendType = "local"
        backendObject = ".terraform/terraform.tfstate"
    }
    
    // Initialize Terraform with backend
    return initTerraform(backendType, backendObject)
}
```

## Design Principles

1. **Fail-Safe**: Returns nil instead of error when labels are absent
2. **Strict Validation**: Prevents partial or invalid configurations
3. **Backend Agnostic**: Doesn't handle backend-specific authentication
4. **Clear Errors**: Provides actionable error messages

## Advanced Usage

### Dynamic Backend Selection

```go
// Select backend based on environment
func selectBackend(manifest proto.Message, env string) (*TofuBackendConfig, error) {
    config, err := backendconfig.ExtractFromManifest(manifest)
    if err != nil {
        return nil, err
    }
    
    if config == nil {
        // Generate default based on environment
        switch env {
        case "production":
            return &TofuBackendConfig{
                BackendType:   "s3",
                BackendObject: "prod-tfstate/default",
            }, nil
        default:
            return &TofuBackendConfig{
                BackendType:   "local",
                BackendObject: fmt.Sprintf(".terraform/%s.tfstate", env),
            }, nil
        }
    }
    
    return config, nil
}
```

### Backend Migration Helper

```go
// Check if backend configuration has changed
func hasBackendChanged(oldManifest, newManifest proto.Message) (bool, error) {
    oldConfig, err := backendconfig.ExtractFromManifest(oldManifest)
    if err != nil {
        return false, err
    }
    
    newConfig, err := backendconfig.ExtractFromManifest(newManifest)
    if err != nil {
        return false, err
    }
    
    // Handle nil cases
    if oldConfig == nil && newConfig == nil {
        return false, nil
    }
    if oldConfig == nil || newConfig == nil {
        return true, nil
    }
    
    // Compare configurations
    return oldConfig.BackendType != newConfig.BackendType ||
           oldConfig.BackendObject != newConfig.BackendObject, nil
}
```

## Related Packages

- `pkg/iac/tofu/tofulabels`: Defines the label constants
- `pkg/reflection/metadatareflect`: Provides label extraction functionality
- `pkg/iac/tofu/runner`: Consumes backend configuration for Terraform execution
