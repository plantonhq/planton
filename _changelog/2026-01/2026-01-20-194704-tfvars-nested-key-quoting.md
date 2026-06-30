# TFVars Nested Map Key Quoting

**Date**: January 20, 2026
**Type**: Bug Fix
**Components**: Terraform/Tofu Integration, TFVars Generation

## Summary

Fixed the tfvars generator to properly quote nested map keys containing special characters (periods, slashes), which are common in Kubernetes-style labels. This resolves HCL parsing errors during `tofu refresh`, `plan`, and `apply` operations.

## Problem Statement / Motivation

When deploying OpenFGA resources with Kubernetes-style labels, the `tofu` commands failed with:

```
error: Variables not allowed
Variables may not be used here.
```

This error appeared 7 times during `refresh`, `plan`, and `apply` operations.

### Root Cause

The tfvars generator in `pkg/iac/tofu/tfvars/tfvars.go` was outputting map keys without quotes:

```hcl
labels = {
    infra-hub.planton.ai/infra-project.name = "planton-gcp-dev-fga-stack"
    planton.dev/provisioner = "terraform"
}
```

In HCL, unquoted keys containing periods are interpreted as **variable references** (like `var.something.nested`). The parser tried to evaluate `infra-hub.planton.ai/infra-project.name` as a nested variable access, which is not allowed in tfvars context.

### Why This Surfaced Now

- Kubernetes-style labels use periods (`.`) and slashes (`/`) as separators
- Previous deployments may have used simpler label keys without special characters
- The OpenFGA deployment was the first to expose this latent bug with complex label keys

## Solution / What's New

Added conditional quoting for map keys based on nesting level:

- **Top-level variable names** (indentLevel 0): NOT quoted (per HCL specification)
- **Nested map keys** (indentLevel > 0): Quoted to safely handle special characters

### Before

```hcl
metadata = {
  labels = {
    infra-hub.planton.ai/infra-project.name = "value"
  }
}
```

### After

```hcl
metadata = {
  "labels" = {
    "infra-hub.planton.ai/infra-project.name" = "value"
  }
}
```

## Implementation Details

### New Helper Function

```go
// formatKey formats a map key for HCL output.
// Top-level keys (indentLevel 0) must NOT be quoted in tfvars files.
// Nested map keys (indentLevel > 0) are quoted to safely handle special
// characters like periods and slashes that are common in Kubernetes-style labels.
func formatKey(key string, indentLevel int) string {
    if indentLevel == 0 {
        return key
    }
    return fmt.Sprintf("%q", key)
}
```

### Updated writeHCL Function

The `writeHCL` function now uses `formatKey()` for all map key outputs:

```go
formattedKey := formatKey(snakeKey, indentLevel)
buf.WriteString(fmt.Sprintf("%s%s = ", indent, formattedKey))
```

## Files Changed

| File | Change |
|------|--------|
| `pkg/iac/tofu/tfvars/tfvars.go` | Added `formatKey()` helper, updated all key formatting calls |
| `pkg/iac/tofu/tfvars/tfvars_test.go` | Updated expected output to include quoted nested keys |

## Benefits

### For Users

- OpenFGA and other deployments with Kubernetes-style labels now work correctly
- No changes required to manifests or label conventions

### For Developers

- Clear separation between top-level and nested key handling
- Well-documented behavior in code comments
- Test coverage for the new format

## Testing

- Unit tests pass with updated expected output format
- End-to-end OpenFGA deployment verified successfully
- HCL parser correctly accepts the generated tfvars

## Related Work

- Builds on the IaC-agnostic provider config fix (2026-01-20-190320)
- Part of OpenFGA deployment component rollout

---

**Status**: ✅ Production Ready
**Verified**: End-to-end OpenFGA deployment successful
