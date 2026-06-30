# Unified Apply/Destroy Commands with Provisioner Auto-Detection

**Date**: November 18, 2025  
**Type**: Feature  
**Impact**: High - Simplifies user experience with kubectl-like interface

## Summary

Added `planton apply` and `planton destroy` commands that provide a kubectl-like experience by automatically detecting the IaC provisioner (Pulumi, Tofu, or Terraform) from manifest labels and routing to the appropriate CLI. This eliminates the need to remember provisioner-specific commands and provides a unified interface for infrastructure management.

## Motivation

Previously, users had to use different commands based on the provisioner:
- Pulumi: `planton pulumi update`
- Tofu: `planton tofu apply`
- Terraform: Manual terraform commands

This created several challenges:
- Cognitive overhead remembering which command to use for each resource
- Inconsistent command structure across provisioners
- Difficult to create universal automation scripts
- Barrier to entry for new users unfamiliar with specific provisioners

The goal was to bring the experience closer to kubectl's simplicity: `kubectl apply` and `kubectl delete` work regardless of resource type.

## What's New

### Provisioner Label

Manifests can now specify which IaC provisioner to use via a label:

```yaml
metadata:
  labels:
    planton.dev/provisioner: pulumi  # or tofu or terraform
```

The label value is case-insensitive, so `pulumi`, `Pulumi`, and `PULUMI` are all valid.

### Unified Apply Command

Single command to apply infrastructure changes regardless of provisioner:

```bash
# Using -f flag (kubectl-style)
planton apply -f manifest.yaml

# Using --manifest flag (existing style)
planton apply --manifest manifest.yaml

# With kustomize
planton apply --kustomize-dir _kustomize --overlay prod

# With field overrides
planton apply -f manifest.yaml --set spec.replicas=3
```

### Unified Destroy Command

Single command to destroy infrastructure with alias support:

```bash
# Using destroy command
planton destroy -f manifest.yaml

# Using delete alias (kubectl-style)
planton delete -f manifest.yaml

# With kustomize
planton destroy --kustomize-dir _kustomize --overlay prod
```

### Interactive Provisioner Selection

If the `planton.dev/provisioner` label is not present in the manifest, the CLI will prompt you to select a provisioner interactively:

```
Select provisioner [Pulumi]/tofu/terraform: 
```

Simply press Enter to use the default (Pulumi), or type your choice.

## Implementation Details

### New Packages Created

1. **`pkg/iac/provisionerlabels`** - Provisioner label constants
   - `ProvisionerLabelKey = "planton.dev/provisioner"`

2. **`pkg/iac/provisioner`** - Provisioner detection and extraction logic
   - `ProvisionerType` enum (Pulumi, Tofu, Terraform)
   - `ExtractFromManifest()` - Extracts provisioner from manifest labels
   - `FromString()` - Converts string to provisioner type (case-insensitive)

3. **`internal/cli/prompt`** - Interactive user prompts
   - `PromptForProvisioner()` - Prompts user to select provisioner with Pulumi as default

### New Commands

- **`cmd/planton/root/apply.go`** - Unified apply command
  - Supports `-f` shorthand for `--manifest`
  - Automatically detects provisioner from manifest
  - Routes to appropriate provisioner CLI
  - Falls back to interactive prompt if label missing

- **`cmd/planton/root/destroy.go`** - Unified destroy command
  - Has `delete` as an alias for kubectl compatibility
  - Same detection and routing logic as apply
  - Supports all provisioner-specific flags

### Modified Components

- **`cmd/planton/root.go`** - Registered new commands
- **`internal/cli/cliprint/print.go`** - Added Tofu/Terraform print helpers
  - `PrintTofuSuccess()` and `PrintTofuFailure()`
  - `PrintTerraformSuccess()` and `PrintTerraformFailure()`

## Command Behavior

### Apply Command Flow

1. Load and validate manifest (supports file, kustomize, or input-dir)
2. Apply any field overrides specified via `--set` flags
3. Detect provisioner from manifest label
4. If provisioner missing, prompt user interactively (default: Pulumi)
5. Route to appropriate provisioner:
   - **Pulumi**: Call `pulumistack.Run()` with `PulumiOperationType_update`
   - **Tofu**: Call `tofumodule.RunCommand()` with `TerraformOperationType_apply`
   - **Terraform**: Currently returns error (not yet implemented)

### Destroy Command Flow

Same as apply, but routes to destroy operations:
- **Pulumi**: `PulumiOperationType_destroy`
- **Tofu**: `TerraformOperationType_destroy`
- **Terraform**: Currently returns error (not yet implemented)

## Supported Flags

Both commands support all standard flags from their respective provisioners:

### Common Flags
- `-f, --manifest` - Path to manifest file
- `--kustomize-dir` - Kustomize directory
- `--overlay` - Kustomize overlay
- `--set` - Field overrides (key=value pairs)
- `--module-dir` - Provisioner module directory

### Pulumi-Specific Flags
- `--stack` - Stack FQDN (or extracted from manifest)
- `--yes` - Auto-approve
- `--diff` - Show detailed diffs

### Tofu/Terraform-Specific Flags
- `--auto-approve` - Skip interactive approval

### Provider Credentials
- `--aws-provider-config`
- `--azure-provider-config`
- `--cloudflare-provider-config`
- `--confluent-provider-config`
- `--gcp-provider-config`
- `--kubernetes-provider-config`
- `--atlas-provider-config`
- `--snowflake-provider-config`

## Examples

### Example 1: Pulumi Resource with Label

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesExternalDns
metadata:
  name: external-dns-prod
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/stack.name: prod.KubernetesExternalDns.external-dns
spec:
  # ... resource spec
```

```bash
# Apply
planton apply -f external-dns.yaml

# Destroy
planton destroy -f external-dns.yaml
```

### Example 2: Tofu Resource with Label

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsVpc
metadata:
  name: app-vpc
  labels:
    planton.dev/provisioner: tofu
    terraform.planton.dev/backend.type: s3
    terraform.planton.dev/backend.object: terraform-state/vpc/prod.tfstate
spec:
  # ... resource spec
```

```bash
# Apply
planton apply -f vpc.yaml --auto-approve

# Destroy (using delete alias)
planton delete -f vpc.yaml --auto-approve
```

### Example 3: No Label - Interactive Prompt

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpCloudSql
metadata:
  name: app-database
spec:
  # ... resource spec
```

```bash
planton apply -f database.yaml
# Output:
# ✓ Manifest loaded
# ✓ Manifest validated
# • Detecting provisioner...
# ℹ Provisioner not specified in manifest
# Select provisioner [Pulumi]/tofu/terraform: tofu
# ✓ Using provisioner: tofu
# ...
```

### Example 4: Kustomize with Overrides

```bash
# Apply with kustomize and overrides
planton apply \
  --kustomize-dir _kustomize \
  --overlay production \
  --set spec.replicas=5 \
  --set spec.version=v2.0.0

# Destroy using kustomize
planton destroy \
  --kustomize-dir _kustomize \
  --overlay production
```

## Migration Guide

This is a **non-breaking change**. All existing commands continue to work:

```bash
# Old way (still works)
planton pulumi update --manifest manifest.yaml
planton tofu apply --manifest manifest.yaml

# New unified way (recommended)
planton apply -f manifest.yaml
planton destroy -f manifest.yaml
```

To adopt the new commands:

1. **Optional**: Add `planton.dev/provisioner` label to your manifests
2. Use `apply` instead of `pulumi update` or `tofu apply`
3. Use `destroy` or `delete` instead of `pulumi destroy` or `tofu destroy`

If you don't add the provisioner label, the CLI will prompt you to select one.

## Benefits

1. **Simplified Mental Model**: One command for all provisioners
2. **kubectl-like Experience**: Familiar `-f` flag and `apply`/`delete` commands
3. **Reduced Errors**: No need to remember provisioner-specific commands
4. **Better Automation**: Write scripts that work regardless of provisioner
5. **Lower Barrier to Entry**: New users don't need to learn multiple command patterns
6. **Backward Compatible**: All existing commands still work

## Future Enhancements

- Add support for Terraform provisioner (currently shows helpful error)
- Auto-detection of provisioner based on resource Kind (fallback if label missing)
- Support for `--watch` flag to continuously reconcile state
- Add `planton diff` command to preview changes without applying
- Integration with CI/CD systems for automated provisioner detection

## Breaking Changes

None. This feature is completely additive and backward compatible.

## Notes

- The provisioner label is case-insensitive for user convenience
- Interactive prompt defaults to Pulumi (press Enter to select)
- The `delete` command is an alias for `destroy` to match kubectl conventions
- Invalid provisioner values result in clear error messages
- All existing flags and options are preserved and work as before

