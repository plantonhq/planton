# Fix CLI Backend Config Passthrough to RunCommand

**Date**: January 23, 2026
**Type**: Bug Fix
**Components**: CLI Backend Configuration, Unified Commands

## Summary

Fixed an architectural bug where CLI backend flags (`--backend-type`, `--backend-bucket`, etc.) were being processed and displayed correctly by the unified commands but were not being passed to `tofumodule.RunCommand`, causing failures when manifest labels had incomplete backend configuration.

## Problem Statement

When running unified commands like `planton preview` with CLI backend flags:

```bash
planton preview -i manifest.yaml --backend-type=s3 --local-module
```

The CLI would:
1. Correctly merge CLI flags with manifest labels
2. Display the correct backend config (showing `Type: s3`)
3. Then fail with: `"unsupported backend type from manifest labels:"` (empty type)

### Root Cause

There was an architectural disconnect in how backend configuration flowed through the code:

1. `run_tofu.go` called `buildAndValidateBackendConfig()` which correctly merged CLI flags with manifest labels
2. The merged config was displayed via `ui.BackendConfigSummary()` 
3. But `tofumodule.RunCommand()` did NOT receive this config
4. Instead, `RunCommand` independently extracted from manifest labels
5. The manifest had `terraform.planton.dev/backend.key` but NO `backend.type`
6. This created a non-nil config with empty `BackendType`, triggering the error

## Solution

Modified `RunCommand` to accept an optional `backendConfig` parameter:

```go
func RunCommand(
    binaryName string,
    // ... other params ...
    providerConfig *stackinputproviderconfig.ProviderConfig,
    backendConfig *backendconfig.TofuBackendConfig,  // NEW
) error {
    // If backendConfig is provided, use it directly
    // Otherwise, fall back to extracting from manifest (legacy path)
}
```

### Behavior

- **Unified commands** (`apply`, `preview`, `destroy`, etc.): Pass the merged CLI+manifest config to `RunCommand`
- **Direct subcommands** (`tofu apply`, `terraform apply`, etc.): Pass `nil`, falling back to manifest label extraction (backward compatible)

## Files Changed

| File | Change |
|------|--------|
| `pkg/iac/tofu/tofumodule/run_command.go` | Added `backendConfig` parameter, use if non-nil |
| `internal/cli/iacrunner/run_tofu.go` | Pass `backendCfg` to RunCommand |
| `cmd/planton/root/tofu/apply.go` | Pass `nil` for backendConfig |
| `cmd/planton/root/tofu/destroy.go` | Pass `nil` for backendConfig |
| `cmd/planton/root/tofu/plan.go` | Pass `nil` for backendConfig |
| `cmd/planton/root/tofu/refresh.go` | Pass `nil` for backendConfig |
| `cmd/planton/root/terraform/apply.go` | Pass `nil` for backendConfig |
| `cmd/planton/root/terraform/destroy.go` | Pass `nil` for backendConfig |
| `cmd/planton/root/terraform/plan.go` | Pass `nil` for backendConfig |
| `cmd/planton/root/terraform/refresh.go` | Pass `nil` for backendConfig |

## Expected Behavior After Fix

1. User runs: `planton preview -i manifest.yaml --backend-type=s3 --local-module`
2. `run_tofu.go` builds backend config merging CLI flags with manifest labels
3. Config with `BackendType: "s3"` is passed to `RunCommand`
4. `RunCommand` uses the provided config instead of extracting from manifest
5. Terraform init/apply runs successfully with S3 backend

## Related Changes

This fix builds on the previous error display fix (2026-01-23-123524-cli-error-display-fix.md) which made the error visible in the first place.

---

**Status**: Complete
