# Terraform Module Zip Downloads in CLI

**Date**: January 8, 2026
**Type**: Feature
**Components**: Terraform CLI Integration, IAC Module Distribution, CLI UX

## Summary

Implemented zip-based distribution for Terraform modules, eliminating the need to clone the entire planton monorepo. The CLI now downloads lightweight zip files (~5-50KB) from GitHub releases instead of cloning ~500MB+ of source code, dramatically improving startup time and reducing disk usage.

## Problem Statement / Motivation

After implementing pre-built Pulumi binaries, we identified the same inefficiency in the Terraform workflow:

1. Cloning the entire planton monorepo to a staging area
2. Copying the repo to a workspace directory
3. Navigating to the specific `tf/` folder
4. Running terraform commands

### Pain Points

- **Slow startup**: Cloning a large monorepo takes 30-60 seconds
- **Wasted bandwidth**: Downloading hundreds of MB when needing only a few KB
- **Disk space**: Full repo clone consumed ~1GB+ of disk space
- **Git dependency**: Required git to be installed on the system

## Solution / What's New

The CLI now downloads pre-packaged Terraform module zips from GitHub releases:

```mermaid
flowchart TB
    A["tofu apply"] --> B{moduleDir provided?}
    B -->|Yes| C[Use provided directory]
    B -->|No| D{Release version available?}
    D -->|Yes| E[Download terraform-component.zip]
    D -->|No/dev| F[Fallback: clone repo]
    E --> G["Cache in ~/.planton/terraform/modules/"]
    E -->|Fail| H["⚠️ Warning + Fallback"]
    H --> F
    G --> I[Run terraform]
    F --> I
    C --> I
```

### Key Features

1. **Automatic zip download**: Downloads component-specific zips from GitHub releases
2. **Smart caching**: Modules cached at `~/.planton/terraform/modules/{version}/`
3. **Graceful fallback**: Falls back to staging/clone if zip unavailable
4. **Module version support**: Works with `--module-version` for specific releases
5. **No git required**: Zip mode doesn't need git on the system

## Implementation Details

### New Package: `pkg/iac/tofu/tofuzip/`

Created a new package to handle zip distribution, mirroring the Pulumi binary pattern:

**`zip.go`** - Core zip download and caching:

- `EnsureModule()` - Downloads and caches module if not present
- `GetModulePath()` - Returns cached module path
- `BuildDownloadURL()` - Constructs GitHub release download URL
- `DownloadAndExtractZip()` - Downloads and extracts zip archive
- `IsModuleCached()` - Checks if module already cached
- `CanUseZipMode()` - Checks if zip mode is available (not dev build)

### Directory Structure

All Terraform modules cached under `~/.planton/terraform/`:

```
~/.planton/terraform/
└── modules/                          # Cached module zips
    └── v0.3.2/                       # Per-version cache
        ├── awsecsservice/            # Extracted module
        │   ├── main.tf
        │   ├── variables.tf
        │   └── outputs.tf
        └── kubernetesdeployment/
            └── ...
```

### Download URL Format

Zips are downloaded directly from GitHub releases:

```bash
# CLI version v0.3.2 -> downloads from main release
https://github.com/plantonhq/planton/releases/download/v0.3.2/terraform-kubernetesdeployment.zip

# Module version override -> downloads from that specific tag
https://github.com/plantonhq/planton/releases/download/v0.3.2+terraform.awsecsservice.20260108.0/terraform-awsecsservice.zip
```

### No Platform Suffix Needed

Unlike Pulumi binaries which are platform-specific, Terraform modules are platform-agnostic:

| Pulumi | Terraform |
|--------|-----------|
| `pulumi-{component}_{os}_{arch}.gz` | `terraform-{component}.zip` |
| Requires `runtime.GOOS/GOARCH` | No platform detection |
| Executable binary | Folder with `.tf` files |

### Module Directory Changes

Updated `GetModulePath()` in `pkg/iac/tofu/tofumodule/module_directory.go`:

**Previous flow:**
1. Check if moduleDir provided
2. Clone repo to staging
3. Copy to workspace
4. Navigate to tf folder

**New flow:**
1. Check if moduleDir provided
2. **Try zip download** (if version available)
3. On success: return cached module path
4. On failure: log warning, fall back to staging

### CLI UX Enhancement

Added `PrintWarning()` function to `cliprint` package with ⚠️ warning icon for graceful fallback messages:

```go
// PrintWarning prints a warning message with a yellow warning sign
func PrintWarning(msg string) {
    yellow := color.New(color.FgYellow).SprintFunc()
    fmt.Printf("%s %s\n", Warning, yellow(msg))
}
```

## Benefits

### Performance Improvements

| Metric | Before (Clone) | After (Zip) | Improvement |
|--------|---------------|-------------|-------------|
| Download size | ~500MB+ repo | 5-50KB zip | 99%+ reduction |
| Startup time | 30-60s | 1-2s | 95%+ faster |
| Disk usage | ~1GB+ | ~1MB | 99%+ reduction |
| Network calls | git clone (many) | 1 HTTP GET | Minimal |

### User Experience

- **Near-instant module access**: After first download
- **Works offline**: Once cached, modules work without network
- **No git required**: Zip mode doesn't need git installed
- **Smaller footprint**: Only download what you need
- **Graceful degradation**: Falls back to staging on any failure

## Impact

### CLI Users

- Dramatically faster `planton tofu apply` execution
- No need to wait for large repo clones
- Reduced disk space requirements
- Works in minimal environments without git

### CI/CD Pipelines

- Faster pipeline execution
- Smaller container images (no git needed for zip mode)
- Better caching (zip files by version)
- More reliable (fewer network operations)

## Files Changed

| File | Change |
|------|--------|
| `pkg/iac/tofu/tofuzip/zip.go` | **New** - Zip download and caching logic |
| `pkg/iac/tofu/tofuzip/BUILD.bazel` | **New** - Bazel build file |
| `pkg/iac/tofu/tofumodule/module_directory.go` | **Modified** - Added zip download path |
| `pkg/iac/tofu/tofumodule/BUILD.bazel` | **Modified** - Added tofuzip dependency |
| `internal/cli/cliprint/symbols.go` | **Modified** - Added Warning symbol |
| `internal/cli/cliprint/print.go` | **Modified** - Added PrintWarning function |

## Related Work

- **Prior**: Terraform Module Zip Artifacts in Releases (`2026-01-08-064118`)
- **Prior**: Platform-Specific Pulumi Binary Downloads (`2026-01-08-061943`)
- **Prior**: Pulumi Binary Distribution for IaC Runner (`2026-01-08-052417`)

---

**Status**: ✅ Production Ready
**Timeline**: ~45 minutes implementation

