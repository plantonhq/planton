# CloudflareWorker: Add worker_name Field to Spec

**Date**: December 25, 2025
**Type**: Enhancement
**Components**: API Definitions, Pulumi CLI Integration, Terraform Provider, CloudflareWorker Provider

## Summary

Added a new required `worker_name` field to the `CloudflareWorkerSpec` protobuf definition, moving the worker name from the nested `script.name` field to a top-level spec field. This change improves the API design by making the worker name more prominent and consistent with other deployment components. Updated all IaC modules (Pulumi and Terraform), tests, and documentation to use the new field.

## Problem Statement / Motivation

The CloudflareWorker component had an inconsistent API design where the worker name was buried inside the `script` configuration object. This created several issues:

### Pain Points

- **Inconsistent API**: Unlike other deployment components in Planton, the worker name was not a top-level spec field
- **Poor Discoverability**: Users had to navigate to `spec.script.name` to set the worker name, which was not intuitive
- **Semantic Confusion**: The `script` object was meant to represent the worker script bundle configuration, not the worker's identity
- **Redundancy**: The worker name was conflated with script configuration, even though they serve different purposes

## Solution / What's New

Extracted the worker name to a top-level `worker_name` field in the `CloudflareWorkerSpec`, making it a first-class required field with proper validation.

### Schema Changes

**Before**:
```protobuf
message CloudflareWorkerSpec {
  string account_id = 1 [(buf.validate.field).required = true];
  CloudflareWorkerScript script = 2 [(buf.validate.field).required = true];
  // ...
}

message CloudflareWorkerScript {
  string name = 1 [(buf.validate.field).string.min_len = 1];
  CloudflareWorkerScriptBundleR2Object bundle = 2 [(buf.validate.field).required = true];
}
```

**After**:
```protobuf
message CloudflareWorkerSpec {
  string account_id = 1 [(buf.validate.field).required = true];
  
  // The name of the Cloudflare Worker.
  // This is the worker name that will be visible in the Cloudflare dashboard.
  string worker_name = 2 [
    (buf.validate.field).required = true,
    (buf.validate.field).string.min_len = 1,
    (buf.validate.field).string.max_len = 63
  ];
  
  CloudflareWorkerScript script = 3 [(buf.validate.field).required = true];
  // ...
}

message CloudflareWorkerScript {
  CloudflareWorkerScriptBundleR2Object bundle = 1 [(buf.validate.field).required = true];
}
```

## Implementation Details

### 1. Protocol Buffer Changes

**File**: `apis/dev/planton/provider/cloudflare/cloudflareworker/v1/spec.proto`

- Added `worker_name` as field #2 in `CloudflareWorkerSpec`
- Made it required with `buf.validate.field` options:
  - `required = true`
  - `min_len = 1` (cannot be empty)
  - `max_len = 63` (Cloudflare worker name limit)
- Removed `name` field from `CloudflareWorkerScript`
- Renumbered subsequent fields in `CloudflareWorkerSpec` (script → 3, kv_bindings → 4, etc.)
- Renumbered `bundle` field in `CloudflareWorkerScript` (1 instead of 2)

### 2. Test Updates

**File**: `apis/dev/planton/provider/cloudflare/cloudflareworker/v1/spec_test.go`

- Updated all test cases to include `WorkerName` field in `CloudflareWorkerSpec`
- Removed `Name` field from all `CloudflareWorkerScript` struct literals
- Added new test case to verify `worker_name` is required:
  ```go
  ginkgo.It("should return error if worker_name is missing", func() {
    input := &CloudflareWorker{
      Spec: &CloudflareWorkerSpec{
        AccountId: "00000000000000000000000000000000",
        // WorkerName is missing - should fail validation
        Script: &CloudflareWorkerScript{
          Bundle: &CloudflareWorkerScriptBundleR2Object{
            Bucket: "test-bucket",
            Path:   "test/script.js",
          },
        },
      },
    }
    err := protovalidate.Validate(input)
    gomega.Expect(err).NotTo(gomega.BeNil())
  })
  ```

All 8 existing tests updated, 1 new test added (9 tests total, all passing).

### 3. Pulumi Module Updates

**File**: `apis/dev/planton/provider/cloudflare/cloudflareworker/v1/iac/pulumi/module/worker_script.go`

Changed from:
```go
ScriptName: pulumi.String(locals.CloudflareWorker.Spec.Script.Name),
```

To:
```go
ScriptName: pulumi.String(locals.CloudflareWorker.Spec.WorkerName),
```

**File**: `apis/dev/planton/provider/cloudflare/cloudflareworker/v1/iac/pulumi/module/route.go`

Changed from:
```go
Script: pulumi.String(locals.CloudflareWorker.Spec.Script.Name),
```

To:
```go
Script: pulumi.String(locals.CloudflareWorker.Spec.WorkerName),
```

### 4. Terraform Module Updates

**File**: `apis/dev/planton/provider/cloudflare/cloudflareworker/v1/iac/tf/locals.tf`

Changed from:
```hcl
script_name = var.spec.script.name
```

To:
```hcl
script_name = var.spec.worker_name
```

### 5. Documentation Updates

**File**: `apis/dev/planton/provider/cloudflare/cloudflareworker/v1/examples.md`

Updated all 8 example manifests to use the new field structure:

**Before**:
```yaml
spec:
  accountId: "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
  script:
    name: hello-worker
    bundle:
      bucket: my-workers-bucket
      path: builds/hello-worker-v1.0.0.js
```

**After**:
```yaml
spec:
  accountId: "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
  workerName: hello-worker
  script:
    bundle:
      bucket: my-workers-bucket
      path: builds/hello-worker-v1.0.0.js
```

Updated examples:
1. Minimal Worker (No Route)
2. Worker with Custom Domain
3. API Gateway with KV Storage
4. Webhook Handler
5. Authentication Middleware
6. Multi-Environment Deployment (staging + production)
7. Worker with Environment Variables
8. A/B Testing Worker

## Benefits

### For API Consistency
- Worker name is now at the same level as `account_id`, making the API more intuitive
- Aligns with naming patterns in other deployment components (e.g., `PostgresKubernetes`, `GcpGkeCluster`)

### For Developers
- **Clearer Intent**: The worker name is immediately visible in the spec without navigating nested objects
- **Better Validation**: Field-level validation with min/max length constraints provides early error detection
- **Easier Configuration**: Users can set the worker name without thinking about script configuration

### For Maintainability
- **Separation of Concerns**: Worker identity (`worker_name`) is now separate from script configuration (`script.bundle`)
- **Future-Proof**: Adding more script configuration options won't pollute the worker name field

## Impact

### Breaking Change

This is a **breaking change** for existing CloudflareWorker manifests. Users must update their YAML files to use the new field structure.

**Migration Required**:
```yaml
# Old format (no longer valid)
spec:
  script:
    name: my-worker
    bundle: { ... }

# New format (required)
spec:
  workerName: my-worker
  script:
    bundle: { ... }
```

### Affected Components

- **Protocol Buffers**: Field numbering changed in `CloudflareWorkerSpec` and `CloudflareWorkerScript`
- **Go Stubs**: Regenerated with `make protos`
- **TypeScript Stubs**: Regenerated for web console
- **Pulumi Module**: Two files updated (`worker_script.go`, `route.go`)
- **Terraform Module**: One file updated (`locals.tf`)
- **Tests**: Nine test cases updated/added
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

**Result**: ✅ All tests passing, zero build errors

## Design Decisions

### Why Top-Level Field?

**Considered Alternatives**:
1. **Keep `script.name`** - Rejected because it conflates worker identity with script configuration
2. **Use `metadata.name`** - Rejected because metadata is for resource naming in the control plane, not provider-specific identifiers
3. **Make it optional** - Rejected because every worker requires a name in Cloudflare

**Chosen Approach**: Top-level required field with explicit validation
- Most intuitive for users
- Consistent with other deployment components
- Clearly separates identity from configuration

### Field Number Renumbering

Protocol buffer field numbers were renumbered to maintain sequential ordering:
- `worker_name` inserted at position 2 (after `account_id`)
- `script` moved from 2 → 3
- `kv_bindings` moved from 3 → 4
- `dns` moved from 4 → 5
- `compatibility_date` moved from 5 → 6
- `usage_model` moved from 6 → 7
- `env` moved from 7 → 8

This maintains logical grouping: identity fields first, then configuration.

### Validation Rules

Chosen validation constraints:
- **Required**: Every worker must have a name
- **Min Length 1**: Cannot be empty string
- **Max Length 63**: Cloudflare's documented limit for worker names

These rules ensure the field value is always valid for the Cloudflare Workers API.

## Testing Strategy

### Test Coverage

1. **Positive Cases** (6 tests):
   - Minimal valid configuration
   - Worker with environment variables and secrets
   - Worker with route pattern and custom domain
   
2. **Negative Cases** (3 tests):
   - Missing `worker_name` (new test)
   - Missing `account_id`
   - Invalid `account_id` format
   - Invalid `compatibility_date` format

### Validation Testing

The new required field validation is tested with:
```go
ginkgo.It("should return error if worker_name is missing", func() {
  input := &CloudflareWorker{
    Spec: &CloudflareWorkerSpec{
      AccountId: "00000000000000000000000000000000",
      // WorkerName intentionally omitted
      Script: &CloudflareWorkerScript{ /* ... */ },
    },
  }
  err := protovalidate.Validate(input)
  gomega.Expect(err).NotTo(gomega.BeNil())
})
```

This ensures the `buf.validate` rules are correctly enforced.

## Files Changed

**Protocol Buffers** (1 file):
- `apis/dev/planton/provider/cloudflare/cloudflareworker/v1/spec.proto`

**Tests** (1 file):
- `apis/dev/planton/provider/cloudflare/cloudflareworker/v1/spec_test.go`

**Pulumi** (2 files):
- `apis/dev/planton/provider/cloudflare/cloudflareworker/v1/iac/pulumi/module/worker_script.go`
- `apis/dev/planton/provider/cloudflare/cloudflareworker/v1/iac/pulumi/module/route.go`

**Terraform** (1 file):
- `apis/dev/planton/provider/cloudflare/cloudflareworker/v1/iac/tf/locals.tf`

**Documentation** (1 file):
- `apis/dev/planton/provider/cloudflare/cloudflareworker/v1/examples.md`

**Total**: 6 files changed

## Related Work

This change is part of ongoing API refinement efforts to improve consistency across deployment components in Planton. Similar patterns should be considered for other providers where critical fields are nested unnecessarily.

**Future Considerations**:
- Audit other deployment components for similar API inconsistencies
- Document field-level validation best practices
- Consider adding `worker_name` to stack outputs for easier reference

---

**Status**: ✅ Production Ready
**Migration**: Required for existing CloudflareWorker users

