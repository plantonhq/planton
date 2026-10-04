# Manifest Backend Configuration Support

**Date**: September 16, 2025  
**Type**: Feature  
**PR**: [#240](https://github.com/plantonhq/planton/pull/240)

## Summary

Added support for embedding backend configuration (Pulumi stack info, Terraform/Tofu backend details) directly in manifest labels. This enables self-contained manifests that can be deployed from URLs without requiring additional CLI flags or configuration.

## Motivation

Previously, users had to specify backend configuration through CLI flags every time they deployed:
- Pulumi required `--stack org/project/stack` flag
- Terraform/Tofu required manual backend configuration

This made it difficult to:
- Share deployable manifests that "just work"
- Deploy from URLs without additional context
- Create quick-start examples and demos
- Automate deployments in CI/CD pipelines

## What's New

### Pulumi Backend Configuration via Labels

Manifests can now specify Pulumi stack information using labels:

```yaml
metadata:
  labels:
    # Option 1: Full stack FQDN (recommended)
    pulumi.planton.dev/stack.fqdn: "myorg/project/stack"
    
    # Option 2: Individual components
    pulumi.planton.dev/organization: "myorg"
    pulumi.planton.dev/project: "my-project"  
    pulumi.planton.dev/stack.name: "production"
```

### Terraform/Tofu Backend Configuration via Labels

Manifests can specify remote state backend configuration:

```yaml
metadata:
  labels:
    # Backend type: s3, gcs, azurerm, local
    terraform.planton.dev/backend.type: "s3"
    # Backend-specific configuration
    terraform.planton.dev/backend.object: "bucket-name/path/to/state.tfstate"
```

### CLI Usage

Deploy without backend flags:

```bash
# Pulumi - no --stack flag needed
planton pulumi update --manifest https://example.com/manifest.yaml

# Tofu - backend auto-configured
planton tofu apply --manifest manifest.yaml
```

### Priority and Fallback

- **Pulumi**: Manifest labels take precedence over CLI flags
- **Tofu/Terraform**: Manifest labels used if present, otherwise defaults to local backend
- CLI flags still work and can override manifest labels when explicitly provided

## Implementation Details

### New Packages Created

1. **`pkg/iac/pulumi/pulumilabels`** - Pulumi label constants
2. **`pkg/iac/pulumi/backendconfig`** - Pulumi backend extraction logic
3. **`pkg/iac/tofu/tofulabels`** - Terraform/Tofu label constants
4. **`pkg/iac/tofu/backendconfig`** - Terraform/Tofu backend extraction logic
5. **`pkg/kubernetes/kuberneteslabels`** - Kubernetes-specific labels (relocated from overridelabels)

### Modified Core Components

- **`pkg/iac/pulumi/pulumistack/Run()`** - Now extracts and uses backend config from manifests
- **`pkg/iac/tofu/tofumodule/RunCommand()`** - Now extracts and uses backend config from manifests
- **`pkg/reflection/metadatareflect`** - Added `ExtractLabels()` helper function

### Removed Legacy Code

- Deleted `pkg/overridelabels` package (functionality moved to domain-specific packages)

## Backend Type Support

### Supported Terraform/Tofu Backends

- **S3 (AWS)**: `terraform.planton.dev/backend.object: "bucket/key"`
- **GCS (Google Cloud)**: `terraform.planton.dev/backend.object: "bucket/prefix"`
- **Azure Blob**: `terraform.planton.dev/backend.object: "container/path"`
- **Local**: Default when no backend specified

## Security Considerations

- Backend configuration labels do NOT include credentials
- Credentials must still be provided via:
  - Environment variables
  - CLI credential flags
  - Cloud provider credential chains (IAM roles, service accounts)

## Migration Guide

No breaking changes. Existing workflows continue to function as before. To adopt the new feature:

1. Add appropriate labels to your manifests
2. Remove `--stack` flag from Pulumi commands
3. Remove manual backend configuration for Terraform/Tofu

## Examples

### Pulumi Example

```yaml
apiVersion: code2ai.planton.ai/v1
kind: MicroserviceKubernetes
metadata:
  labels:
    pulumi.planton.dev/stack.fqdn: "acme-corp/microservices/production"
spec:
  # ... rest of spec
```

### Terraform S3 Backend Example

```yaml
apiVersion: aws.planton.ai/v1
kind: EksCluster
metadata:
  labels:
    terraform.planton.dev/backend.type: "s3"
    terraform.planton.dev/backend.object: "terraform-state/eks/prod/terraform.tfstate"
spec:
  # ... rest of spec
```

## Benefits

1. **Self-contained manifests** - All configuration in one place
2. **URL deployments** - Deploy directly from Git repos or artifact stores
3. **Simplified CI/CD** - No need to pass backend configuration through pipeline variables
4. **Better demos** - Quick-start examples work out of the box
5. **Backward compatible** - No changes required for existing workflows

## Future Enhancements

- Support for additional Terraform backend types
- Backend configuration validation in manifest validation phase
- Web UI integration for backend configuration management
