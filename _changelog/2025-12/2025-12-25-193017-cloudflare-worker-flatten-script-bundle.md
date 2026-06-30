# CloudflareWorker: Flatten script_bundle Field to Spec Level

**Date**: December 25, 2025
**Type**: Refactoring
**Components**: API Definitions, Pulumi CLI Integration, Terraform Provider, CloudflareWorker Provider

## Summary

Flattened the `script.bundle` nested field to a top-level `script_bundle` field in `CloudflareWorkerSpec`, removing the now-unnecessary `CloudflareWorkerScript` wrapper message. This simplifies the API by eliminating unnecessary nesting when the wrapper message contained only a single field.

## Problem Statement / Motivation

Following the previous refactor that moved `worker_name` from `script.name` to a top-level field, the `CloudflareWorkerScript` message was left with only a single field: `bundle`. This created unnecessary indirection:

### Pain Points

- **Unnecessary Nesting**: The `CloudflareWorkerScript` message served no purpose beyond wrapping a single `bundle` field
- **Extra Indentation**: YAML manifests required an extra level of nesting (`script.bundle.bucket` instead of `scriptBundle.bucket`)
- **Semantic Confusion**: The "script" wrapper implied additional script-related configuration, but only contained bundle location
- **Inconsistent Naming**: The internal `bundle` field duplicated the concept already present in the message name `CloudflareWorkerScriptBundleR2Object`

## Solution / What's New

Removed the intermediate `CloudflareWorkerScript` message and promoted the bundle configuration directly to `CloudflareWorkerSpec` as `script_bundle`.

### Schema Changes

**Before**:
```protobuf
message CloudflareWorkerSpec {
  string account_id = 1 [...];
  string worker_name = 2 [...];
  CloudflareWorkerScript script = 3 [(buf.validate.field).required = true];
  // ...
}

message CloudflareWorkerScript {
  CloudflareWorkerScriptBundleR2Object bundle = 1 [(buf.validate.field).required = true];
}

message CloudflareWorkerScriptBundleR2Object {
  string bucket = 1 [...];
  string path = 2 [...];
}
```

**After**:
```protobuf
message CloudflareWorkerSpec {
  string account_id = 1 [...];
  string worker_name = 2 [...];
  CloudflareWorkerScriptBundle script_bundle = 3 [(buf.validate.field).required = true];
  // ...
}

message CloudflareWorkerScriptBundle {
  string bucket = 1 [...];
  string path = 2 [...];
}
```

### YAML Manifest Changes

**Before**:
```yaml
spec:
  accountId: "..."
  workerName: hello-worker
  script:
    bundle:
      bucket: my-workers-bucket
      path: builds/hello-worker-v1.0.0.js
```

**After**:
```yaml
spec:
  accountId: "..."
  workerName: hello-worker
  scriptBundle:
    bucket: my-workers-bucket
    path: builds/hello-worker-v1.0.0.js
```

## Implementation Details

### 1. Protocol Buffer Changes

**File**: `apis/dev/planton/provider/cloudflare/cloudflareworker/v1/spec.proto`

- Replaced `CloudflareWorkerScript script = 3` with `CloudflareWorkerScriptBundle script_bundle = 3`
- Removed `CloudflareWorkerScript` message entirely
- Renamed `CloudflareWorkerScriptBundleR2Object` to `CloudflareWorkerScriptBundle` (simpler name)
- Added documentation comments to the new message

### 2. Pulumi Module Updates

**File**: `apis/dev/planton/provider/cloudflare/cloudflareworker/v1/iac/pulumi/module/worker_script.go`

Changed from:
```go
script := locals.CloudflareWorker.Spec.Script
bundle := script.Bundle
```

To:
```go
bundle := locals.CloudflareWorker.Spec.ScriptBundle
```

### 3. Terraform Module Updates

**File**: `apis/dev/planton/provider/cloudflare/cloudflareworker/v1/iac/tf/locals.tf`

Changed from:
```hcl
r2_bucket = var.spec.script.bundle.bucket
r2_path   = var.spec.script.bundle.path
```

To:
```hcl
r2_bucket = var.spec.script_bundle.bucket
r2_path   = var.spec.script_bundle.path
```

**File**: `apis/dev/planton/provider/cloudflare/cloudflareworker/v1/iac/tf/variables.tf`

Changed from:
```hcl
script = object({
  bundle = object({
    bucket = string
    path   = string
  })
})
```

To:
```hcl
script_bundle = object({
  bucket = string
  path   = string
})
```

### 4. Test Updates

**File**: `apis/dev/planton/provider/cloudflare/cloudflareworker/v1/spec_test.go`

Updated all 9 test cases to use the new field structure:

Changed from:
```go
Script: &CloudflareWorkerScript{
  Bundle: &CloudflareWorkerScriptBundleR2Object{
    Bucket: "test-bucket",
    Path:   "test/script.js",
  },
},
```

To:
```go
ScriptBundle: &CloudflareWorkerScriptBundle{
  Bucket: "test-bucket",
  Path:   "test/script.js",
},
```

### 5. Documentation Updates

**File**: `apis/dev/planton/provider/cloudflare/cloudflareworker/v1/examples.md`

Updated all 8 example manifests plus the bundle creation section to use the new field structure.

## Benefits

### For API Simplicity
- Removes one level of unnecessary nesting
- Simpler YAML manifests (one less indentation level)
- Message names now align with their purpose

### For Developers
- **Cleaner Access Path**: `spec.scriptBundle.bucket` vs `spec.script.bundle.bucket`
- **Less Cognitive Load**: No need to understand why there's a wrapper message
- **Better IntelliSense**: Fewer types to navigate in IDE autocomplete

### For Maintainability
- **Fewer Messages**: Removed `CloudflareWorkerScript` and `CloudflareWorkerScriptBundleR2Object`
- **Single Source of Truth**: Bundle configuration lives in one place
- **Aligned Naming**: `CloudflareWorkerScriptBundle` clearly describes its purpose

## Impact

### Breaking Change

This is a **breaking change** for existing CloudflareWorker manifests. Users must update their YAML files to use the new field structure.

**Migration Required**:
```yaml
# Old format (no longer valid)
spec:
  script:
    bundle:
      bucket: my-bucket
      path: scripts/worker.js

# New format (required)
spec:
  scriptBundle:
    bucket: my-bucket
    path: scripts/worker.js
```

### Affected Components

- **Protocol Buffers**: Message removed, field renamed
- **Go Stubs**: Regenerated with `make protos`
- **TypeScript Stubs**: Regenerated for web console
- **Pulumi Module**: One file updated (`worker_script.go`)
- **Terraform Module**: Two files updated (`locals.tf`, `variables.tf`)
- **Tests**: Nine test cases updated
- **Documentation**: Eight examples updated

### Validation

All changes verified with:
```bash
# Regenerate protocol buffer stubs
make protos

# Run component-specific tests
go test ./apis/dev/planton/provider/cloudflare/cloudflareworker/v1/

# Full project build
make build

# Full test suite
make test
```

**Result**: ✅ All 9 tests passing, zero build errors

## Design Decisions

### Why Flatten the Structure?

**Considered Alternatives**:
1. **Keep `script.bundle`** - Rejected because the wrapper message serves no purpose with only one field
2. **Rename to `bundle`** - Rejected because `script_bundle` more clearly indicates this is for script content
3. **Keep wrapper for future extensibility** - Rejected because YAGNI (You Aren't Gonna Need It); we can always add new fields to `CloudflareWorkerScriptBundle` if needed

**Chosen Approach**: Flatten to `script_bundle` at spec level
- Reduces nesting by one level
- Maintains clear semantics with `script_bundle` name
- Follows same pattern as `worker_name` (top-level identity/config)

### Message Renaming

Renamed `CloudflareWorkerScriptBundleR2Object` to `CloudflareWorkerScriptBundle`:
- Simpler, more memorable name
- "R2Object" suffix was implementation detail that leaked into API
- New name focuses on what it represents, not where it's stored

## Files Changed

**Protocol Buffers** (1 file):
- `apis/dev/planton/provider/cloudflare/cloudflareworker/v1/spec.proto`

**Tests** (1 file):
- `apis/dev/planton/provider/cloudflare/cloudflareworker/v1/spec_test.go`

**Pulumi** (1 file):
- `apis/dev/planton/provider/cloudflare/cloudflareworker/v1/iac/pulumi/module/worker_script.go`

**Terraform** (2 files):
- `apis/dev/planton/provider/cloudflare/cloudflareworker/v1/iac/tf/locals.tf`
- `apis/dev/planton/provider/cloudflare/cloudflareworker/v1/iac/tf/variables.tf`

**Documentation** (1 file):
- `apis/dev/planton/provider/cloudflare/cloudflareworker/v1/examples.md`

**Total**: 6 files changed

## Related Work

This change completes the CloudflareWorker API cleanup started with the `worker_name` field extraction in the previous changelog (`2025-12-25-150748-cloudflare-worker-name-field-refactor.md`).

**Combined API Improvements**:
1. ✅ `worker_name` extracted to top-level (previous change)
2. ✅ `script.bundle` flattened to `script_bundle` (this change)

The CloudflareWorkerSpec now has a clean, flat structure with all essential fields at the top level.

---

**Status**: ✅ Production Ready
**Migration**: Required for existing CloudflareWorker users

