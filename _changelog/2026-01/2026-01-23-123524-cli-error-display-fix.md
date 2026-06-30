# Fix HCL Execution Error Display in CLI

**Date**: January 23, 2026
**Type**: Bug Fix
**Components**: CLI Error Handling, User Experience

## Summary

Fixed a critical bug where Terraform/OpenTofu execution errors were being silently swallowed by the CLI. Previously, when IaC execution failed, users only saw a generic "Terraform execution failed" message without the actual error details, making debugging impossible.

## Problem Statement

When running `planton preview/apply` with a Terraform module that failed, the terminal output showed:

```
🤝 Handing off to Terraform...
   Output below is from Terraform

✖ Terraform execution failed
   Check the above output from Terraform CLI to understand the root cause
```

But there was **no output above** - the actual error was discarded, leaving users with no way to diagnose the failure.

### Root Cause

In the error handling code across multiple command handlers, errors returned from `tofumodule.RunCommand()` were being caught but never displayed:

```go
if err != nil {
    cliprint.PrintTerraformFailure()  // Generic message only
    os.Exit(1)
}
```

The error variable was discarded, and only a generic failure message was printed.

## Solution

### 1. Created New Error Display Function

Added `printHclExecutionError()` in `internal/cli/iacrunner/run_tofu.go` that uses the `ui` package for beautiful error display:

```go
func printHclExecutionError(binary provisioner.HclBinary, err error) {
    title := fmt.Sprintf("%s Execution Failed", binary.DisplayName())
    ui.ErrorWithoutExit(title, err.Error(),
        "Check the module configuration for syntax errors",
        "Ensure all required provider credentials are configured")
}
```

### 2. Updated Error Handling in All Command Handlers

Updated the error handling pattern in all affected files to display the error before the generic failure message:

```go
if err != nil {
    ui.ErrorWithoutExit("Terraform Execution Failed", err.Error(),
        "Check the module configuration for syntax errors",
        "Ensure all required provider credentials are configured")
    cliprint.PrintTerraformFailure()
    os.Exit(1)
}
```

## Expected Output After Fix

Users will now see beautiful, actionable error output:

```
✗  Terraform Execution Failed

   failed to initialize terraform module: failed to execute terraform command
   /usr/local/bin/terraform init --var-file .terraform/terraform.tfvars: exit status 1

   Hint: Check the module configuration for syntax errors
   Hint: Ensure all required provider credentials are configured

✖ Terraform execution failed
   Check the above output from Terraform CLI to understand the root cause
```

## Files Changed

| File | Change |
|------|--------|
| `internal/cli/iacrunner/run_tofu.go` | Added `printHclExecutionError()`, updated `runHcl()` error handling |
| `cmd/planton/root/tofu/apply.go` | Added ui import, display error before failure message |
| `cmd/planton/root/tofu/destroy.go` | Added ui import, display error before failure message |
| `cmd/planton/root/tofu/plan.go` | Added ui import, display error before failure message |
| `cmd/planton/root/tofu/refresh.go` | Added ui import, display error before failure message |
| `cmd/planton/root/tofu/init.go` | Added ui import, display error before failure message |
| `cmd/planton/root/terraform/apply.go` | Added ui import, display error before failure message |
| `cmd/planton/root/terraform/destroy.go` | Added ui import, display error before failure message |
| `cmd/planton/root/terraform/plan.go` | Added ui import, display error before failure message |
| `cmd/planton/root/terraform/refresh.go` | Added ui import, display error before failure message |
| `cmd/planton/root/terraform/init.go` | Added ui import, display error before failure message |

## Benefits

### For Users

- **Actionable errors**: Users can now see exactly what went wrong
- **Beautiful formatting**: Errors use the `ui` package for consistent, styled output
- **Helpful hints**: Each error includes troubleshooting suggestions
- **Debugging enabled**: The full error chain is displayed, making root cause analysis possible

### For Developers

- **Consistent pattern**: All command handlers now follow the same error display pattern
- **Extensible**: Easy to add more specific error handling for different failure modes

## Impact

- **All CLI commands affected**: Both unified commands (`apply`, `preview`, etc.) and explicit subcommands (`tofu apply`, `terraform apply`, etc.) are fixed
- **No breaking changes**: Output format is additive - the generic failure message is still displayed after the detailed error

---

**Status**: ✅ Complete
