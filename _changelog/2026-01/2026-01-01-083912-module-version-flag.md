# Add `--module-version` Flag for Version-Specific Module Checkout

**Date:** 2026-01-01  
**Type:** Feature Enhancement  
**Scope:** CLI, Pulumi, OpenTofu  

---

## Summary

Added a new `--module-version` flag that allows users to specify a particular version (tag, branch, or commit SHA) of the IaC modules to use in the workspace copy, without affecting the staging area.

## Problem

After implementing the staging-based module cache, users needed a way to:
- Use a specific module version different from what's in staging
- Test with older/newer versions without affecting the staging area
- Debug issues by matching exact module versions used in previous deployments
- Work with commit SHAs in addition to tags

## Solution

### New `--module-version` Flag

Added `--module-version` to all infrastructure commands:
- `planton apply --module-version v0.2.273`
- `planton destroy --module-version main`
- `planton plan --module-version abc1234`
- `planton refresh --module-version <version>`
- `planton init --module-version <version>`

The version argument supports:
- **Tags**: `v0.2.273`, `v1.0.0`
- **Branches**: `main`, `develop`
- **Commit SHAs**: `abc1234`, full 40-char SHA

### Behavior

```
┌─────────────────────────────────────────────────────────────────┐
│                                                                 │
│  User runs: planton apply --module-version v0.2.273    │
│                                                                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  1. Ensure staging exists (clone if not)                        │
│                                                                 │
│  2. Copy staging → workspace                                    │
│     ~/.planton/staging/planton                 │
│           ↓ (cp -a)                                             │
│     ~/.planton/pulumi/<stack>/planton          │
│                                                                 │
│  3. In workspace copy: git checkout v0.2.273                   │
│     (staging unchanged - still on its current version)          │
│                                                                 │
│  4. Execute Pulumi/Tofu operation                               │
│                                                                 │
│  5. Cleanup workspace copy (unless --no-cleanup)                │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Key Points

1. **Staging is unaffected**: The `--module-version` checkout happens in the workspace copy, not in staging
2. **Requires network for new versions**: If the version isn't in the cloned staging repo, `git fetch --all` is run in the workspace
3. **Supports any git ref**: Tags, branches, and commit SHAs all work

## Implementation Details

### Files Changed

| File | Change |
|------|--------|
| `internal/cli/flag/flag.go` | Added `ModuleVersion` flag constant |
| `internal/cli/staging/staging.go` | Added `CheckoutVersionInWorkspace()` function |
| `pkg/iac/pulumi/pulumimodule/module_directory.go` | Updated `GetPath()` to accept `moduleVersion` |
| `pkg/iac/tofu/tofumodule/module_directory.go` | Updated `GetModulePath()` to accept `moduleVersion` |
| `pkg/iac/pulumi/pulumistack/*.go` | Updated all stack operations to pass `moduleVersion` |
| `pkg/iac/tofu/tofumodule/run_command.go` | Updated to pass `moduleVersion` |
| `cmd/planton/root/*.go` | Added `--module-version` flag to unified commands |
| `cmd/planton/root/pulumi.go` | Added `--module-version` to Pulumi parent command |
| `cmd/planton/root/pulumi/*.go` | Updated handlers to pass `moduleVersion` |
| `cmd/planton/root/tofu/*.go` | Added `--module-version` flag and handlers |

### New Function

```go
// CheckoutVersionInWorkspace checks out a specific version in a workspace copy
func CheckoutVersionInWorkspace(workspacePath, version string) error {
    if version == "" {
        return nil
    }
    
    // Fetch all to ensure version is available
    fetchCmd := exec.Command("git", "-C", workspacePath, "fetch", "--all", "--tags")
    if err := fetchCmd.Run(); err != nil {
        return errors.Wrap(err, "failed to fetch in workspace copy")
    }
    
    // Checkout the specified version
    checkoutCmd := exec.Command("git", "-C", workspacePath, "checkout", version)
    if err := checkoutCmd.Run(); err != nil {
        return errors.Wrapf(err, "failed to checkout version %s", version)
    }
    
    return nil
}
```

## Usage Examples

```bash
# Use a specific release tag
planton apply -f manifest.yaml --module-version v0.2.273

# Use latest development branch
planton plan -f manifest.yaml --module-version main

# Debug with exact commit SHA from previous deployment
planton destroy -f manifest.yaml --module-version a1b2c3d4

# Combine with --no-cleanup to inspect the workspace
planton apply -f manifest.yaml --module-version v0.2.270 --no-cleanup
```

## Checkout Command Update

The `planton checkout` command was already designed to support any git ref (tag, branch, or commit SHA). This changelog confirms that behavior and extends the same flexibility to `--module-version`.

## Impact

- **Flexibility**: Users can now pin specific module versions without touching staging
- **Debugging**: Easier to reproduce issues by using exact commit SHAs
- **Testing**: Can test new module versions before updating staging
- **CI/CD**: Pipelines can use specific versions while local development uses latest

---

*This enhancement builds on the staging-based module cache feature to provide fine-grained version control for IaC modules.*

