# GCP Components: Remove metadata.name Dependency, Add Explicit Resource Name Fields

**Date**: November 19, 2025
**Type**: Breaking Change, Refactoring
**Components**: GCP Provider, API Definitions, Pulumi Integration, Terraform Integration

## Summary

Refactored 6 GCP deployment components to eliminate dependency on `metadata.name` for actual GCP resource naming. Each component now has explicit name fields in `spec.proto` (e.g., `network_name`, `bucket_name`, `cluster_name`) that directly control the GCP resource names, while `metadata.name` remains for Planton organizational labels. This provides better clarity, portability, and alignment with GCP's explicit naming requirements.

## Problem Statement

Previously, GCP components used `metadata.name` for dual purposes:

1. **Planton organizational identifier** - Resource tracking, labels, references
2. **Actual GCP resource name** - The name of the VPC network, GCS bucket, GKE cluster, etc.

This coupling created several issues:

### Pain Points

- **Ambiguity**: Unclear whether `metadata.name` was for Planton tracking or GCP resource naming
- **Portability concerns**: Kubernetes-style metadata pattern didn't clearly map to cloud provider resources
- **Implicit behavior**: Users had to understand that metadata.name became the GCP resource name
- **Limited flexibility**: Couldn't have different tracking names vs actual GCP names
- **Inconsistency with provider APIs**: GCP APIs require explicit resource names, not metadata-derived ones
- **Documentation complexity**: Examples had to explain the metadata.name → resource name mapping

### Example of the Old Pattern

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpVpc
metadata:
  name: my-vpc  # Used for BOTH Planton tracking AND GCP network name
spec:
  project_id:
    value: my-project-123
  auto_create_subnetworks: false
```

Behind the scenes:
```go
// Pulumi code
Name: pulumi.String(locals.GcpVpc.Metadata.Name),  // Implicit!
```

```hcl
# Terraform code
name = var.metadata.name  # Not obvious this becomes GCP resource name
```

## Solution / What's New

We systematically added explicit name fields to the `spec` section of 6 GCP components, making GCP resource names first-class configuration:

### Components Refactored

1. **GcpVpc** → Added `network_name`
2. **GcpSubnetwork** → Added `subnetwork_name`
3. **GcpRouterNat** → Added `router_name` AND `nat_name` (creates 2 resources)
4. **GcpGkeCluster** → Added `cluster_name`
5. **GcpGkeNodePool** → Added `node_pool_name`
6. **GcpGcsBucket** → Added `bucket_name`

### New Pattern

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpVpc
metadata:
  name: prod-vpc-resource  # Planton organizational identifier
spec:
  project_id:
    value: my-project-123
  network_name: prod-network  # EXPLICIT GCP VPC name
  auto_create_subnetworks: false
```

Now in code:
```go
// Pulumi - explicit and clear
Name: pulumi.String(locals.GcpVpc.Spec.NetworkName),
```

```hcl
# Terraform - obvious what becomes GCP resource name
name = var.spec.network_name
```

### Key Architectural Decision

**Separation of Concerns**:
- **GCP Resource Names** (spec fields): Control actual cloud provider resource names
- **Planton Labels** (metadata.name): Organizational tracking via `gcplabelkeys.ResourceName`

This means labels continue to use `metadata.name`, while GCP resources use the new spec fields.

## Implementation Details

### Phase 1: Proto Schema Updates

Added required name fields to all 6 component spec files with comprehensive validation:

**Example: GcpVpc**
```protobuf
// spec.proto
message GcpVpcSpec {
  // ... existing fields ...
  
  // Name of the VPC network to create in GCP.
  // Must be 1-63 characters, lowercase letters, numbers, or hyphens.
  // Must start with a lowercase letter and end with a lowercase letter or number.
  // Example: "my-vpc-network", "prod-network"
  string network_name = 4 [
    (buf.validate.field).required = true,
    (buf.validate.field).string.pattern = "^[a-z]([a-z0-9-]{0,61}[a-z0-9])?$"
  ];
}
```

**GCP Naming Validation Patterns**:
- VPC/Subnet/Router/NAT: `^[a-z]([a-z0-9-]{0,61}[a-z0-9])?$` (1-63 chars)
- GKE Cluster/NodePool: `^[a-z]([a-z0-9-]{0,38}[a-z0-9])?$` (1-40 chars)
- GCS Bucket: `^[a-z0-9]([a-z0-9-._]*[a-z0-9])?$` with 3-63 char length validation

### Phase 2: Pulumi Code Updates

**File**: `apis/dev/planton/provider/gcp/<component>/v1/iac/pulumi/module/*.go`

Updated resource creation to use spec name fields:

```go
// Before
Name: pulumi.String(locals.GcpVpc.Metadata.Name),

// After
Name: pulumi.String(locals.GcpVpc.Spec.NetworkName),
```

**Labels kept using metadata.name**:
```go
// locals.go - labels continue to use metadata.name
locals.GcpLabels = map[string]string{
    gcplabelkeys.Resource:     strconv.FormatBool(true),
    gcplabelkeys.ResourceName: locals.GcpVpc.Metadata.Name,  // ← Unchanged
    gcplabelkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_GcpVpc.String()),
}
```

### Phase 3: Terraform Code Updates

**File**: `apis/dev/planton/provider/gcp/<component>/v1/iac/tf/variables.tf`

Added name fields with validation:

```hcl
variable "spec" {
  type = object({
    # ... existing fields ...
    network_name = string
  })
  
  validation {
    condition     = can(regex("^[a-z]([a-z0-9-]{0,61}[a-z0-9])?$", var.spec.network_name))
    error_message = "Network name must be 1-63 characters, lowercase letters, numbers, or hyphens..."
  }
}
```

**File**: `main.tf`

```hcl
# Before
resource "google_compute_network" "vpc" {
  name = var.metadata.name
  # ...
}

# After
resource "google_compute_network" "vpc" {
  name = var.spec.network_name  # ← Explicit and clear
  # ...
}
```

**File**: `locals.tf`

Labels continue using `metadata.name`:
```hcl
locals {
  gcp_labels = {
    resource     = var.metadata.name  # ← For labels/tracking
    resource-id  = var.metadata.id
    # ...
  }
}
```

### Phase 4: Test Enhancements

**File**: `spec_test.go`

Added comprehensive validation tests for each component:

```go
// Positive test - valid name
Spec: &GcpVpcSpec{
    ProjectId: &foreignkeyv1.StringValueOrRef{
        LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
    },
    NetworkName: "test-vpc-network",  // ← New required field
},

// Negative test - missing name
Spec: &GcpVpcSpec{
    ProjectId: &foreignkeyv1.StringValueOrRef{
        LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
    },
    // NetworkName missing - should fail validation
},

// Negative test - invalid format
NetworkName: "INVALID-NAME",  // Uppercase not allowed
```

Added 3 test cases per component:
- ✅ Valid name (happy path)
- ❌ Missing name (required field validation)
- ❌ Invalid format (pattern validation)

### Phase 5: Documentation Updates

**File**: `examples.md`

Updated all YAML examples to include the new name fields:

```yaml
# Before
spec:
  project_id:
    value: my-dev-project-123
  auto_create_subnetworks: false

# After
spec:
  project_id:
    value: my-dev-project-123
  network_name: dev-network  # ← NEW: Explicit GCP VPC name
  auto_create_subnetworks: false
```

**File**: `hack/manifest.yaml`

Updated test manifests with realistic name values for local development.

## Special Case: GcpRouterNat

This component required **two name fields** because it creates two GCP resources:

```protobuf
message GcpRouterNatSpec {
  // ... existing fields ...
  
  // Name of the Cloud Router to create in GCP
  string router_name = 6 [
    (buf.validate.field).required = true,
    (buf.validate.field).string.pattern = "^[a-z]([a-z0-9-]{0,61}[a-z0-9])?$"
  ];
  
  // Name of the NAT configuration on the Cloud Router
  string nat_name = 7 [
    (buf.validate.field).required = true,
    (buf.validate.field).string.pattern = "^[a-z]([a-z0-9-]{0,61}[a-z0-9])?$"
  ];
}
```

Usage in Pulumi:
```go
// Create Router
compute.NewRouter(ctx, "router", &compute.RouterArgs{
    Name: pulumi.String(locals.GcpRouterNat.Spec.RouterName),
    // ...
})

// Create NAT
compute.NewRouterNat(ctx, "router-nat", &compute.RouterNatArgs{
    Name: pulumi.String(locals.GcpRouterNat.Spec.NatName),
    // ...
})
```

## Benefits

### 1. Clarity and Explicitness

**Before**: Implicit mapping
```yaml
metadata:
  name: my-vpc  # Is this for Planton or GCP? Both? Unclear.
```

**After**: Crystal clear
```yaml
metadata:
  name: my-vpc-resource  # Planton identifier
spec:
  network_name: my-vpc   # GCP VPC name - explicit!
```

### 2. Better Documentation

Examples are now self-documenting:
- No need to explain "metadata.name becomes the GCP resource name"
- Each name field has inline comments explaining its purpose
- Validation patterns show allowed formats upfront

### 3. Portability and Flexibility

Users can now:
- Use different names for Planton tracking vs GCP resources
- Apply consistent Planton naming while using GCP-specific resource names
- Follow organizational naming conventions for both layers independently

### 4. Alignment with Provider APIs

GCP APIs explicitly require resource names in their request payloads. Our API now mirrors this pattern:

```go
// GCP API pattern
&compute.NetworkArgs{
    Name: "explicit-network-name",  // ← Required by GCP
}

// Our API now matches this pattern
Spec: &GcpVpcSpec{
    NetworkName: "explicit-network-name",  // ← Explicit in our API too
}
```

### 5. Type Safety and Validation

All name fields include:
- Required field validation
- Pattern validation (GCP naming rules)
- Min/max length validation (where applicable)
- Compile-time type checking

### 6. Consistent Pattern Across Components

All GCP components now follow the same pattern:
- Spec field for resource name
- Metadata.name for Planton tracking
- Labels use metadata.name
- GCP resources use spec name

## Impact

### Breaking Change for Users

**Migration Required**: All existing manifests using these 6 components must add the new name field(s) to their spec.

**Before**:
```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpVpc
metadata:
  name: prod-network
spec:
  project_id:
    value: prod-project
```

**After**:
```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpVpc
metadata:
  name: prod-network
spec:
  project_id:
    value: prod-project
  network_name: prod-network  # ← REQUIRED NEW FIELD
```

### Components Affected

Users must update manifests for:
- ✅ GcpVpc (add `network_name`)
- ✅ GcpSubnetwork (add `subnetwork_name`)
- ✅ GcpRouterNat (add `router_name` and `nat_name`)
- ✅ GcpGkeCluster (add `cluster_name`)
- ✅ GcpGkeNodePool (add `node_pool_name`)
- ✅ GcpGcsBucket (add `bucket_name`)

### Files Modified

**Total: ~54 files across 6 components**

Per component:
- 1 `spec.proto` (schema definition)
- 1 `spec.pb.go` (regenerated stub)
- 3-4 Pulumi module files (`locals.go`, `*.go`)
- 3-4 Terraform files (`variables.tf`, `locals.tf`, `main.tf`)
- 1 `spec_test.go` (validation tests)
- 1 `examples.md` (all examples updated)
- 1 `hack/manifest.yaml` (test manifest)

### Code Quality

- ✅ All proto validations tested
- ✅ All component tests passed (6/6)
- ✅ Full build successful
- ✅ Full test suite passed
- ✅ Feature parity maintained between Pulumi and Terraform

## Implementation Timeline

**Execution**: Single systematic refactoring pass
**Phases**:
1. Proto schema updates (6 files)
2. Proto stub regeneration (`make protos`)
3. Pulumi implementation updates (12 files)
4. Terraform implementation updates (18 files)
5. Test enhancements (6 files)
6. Documentation updates (12+ files)
7. Verification (component tests + full build)

**Result**: All changes validated and tested in one pass.

## Technical Details

### Naming Validation Patterns

Each component enforces GCP-specific naming rules:

**Standard Pattern** (VPC, Subnet, Router, NAT):
- 1-63 characters
- Lowercase letters, numbers, hyphens only
- Must start with lowercase letter
- Must end with lowercase letter or number
- Pattern: `^[a-z]([a-z0-9-]{0,61}[a-z0-9])?$`

**GKE Pattern** (Cluster, NodePool):
- 1-40 characters (GKE restriction)
- Same format rules
- Pattern: `^[a-z]([a-z0-9-]{0,38}[a-z0-9])?$`

**GCS Bucket Pattern**:
- 3-63 characters (globally unique requirement)
- Lowercase letters, numbers, hyphens, or dots
- Cannot be formatted as IP address
- Pattern: `^[a-z0-9]([a-z0-9-._]*[a-z0-9])?$`

### Label Assignment Strategy

**Critical distinction**: Labels use `metadata.name`, resources use spec name fields.

**Pulumi**:
```go
// In locals.go
locals.GcpLabels = map[string]string{
    gcplabelkeys.ResourceName: locals.GcpVpc.Metadata.Name,  // ← For tracking
}

// In vpc.go
networkArgs := &compute.NetworkArgs{
    Name: pulumi.String(locals.GcpVpc.Spec.NetworkName),  // ← For GCP
}
```

**Terraform**:
```hcl
# In locals.tf
locals {
  gcp_labels = {
    resource = var.metadata.name  # ← For labels
  }
}

# In main.tf
resource "google_compute_network" "vpc" {
  name = var.spec.network_name  # ← For GCP
}
```

This ensures:
- GCP resources get explicit names from spec
- Planton labels maintain organizational identifiers
- No conflict between tracking and naming

### Terraform Variable Validation

Added comprehensive validation blocks:

```hcl
variable "spec" {
  type = object({
    network_name = string
  })
  
  validation {
    condition     = can(regex("^[a-z]([a-z0-9-]{0,61}[a-z0-9])?$", var.spec.network_name))
    error_message = "Network name must be 1-63 characters, lowercase letters, numbers, or hyphens, starting with a letter and ending with a letter or number."
  }
}
```

This provides:
- Pre-deployment validation
- Clear error messages for invalid names
- Consistent validation across Pulumi and Terraform

### Test Coverage

Enhanced `spec_test.go` for all components with:

**Required field tests**:
```go
// Test missing name field
Spec: &GcpVpcSpec{
    ProjectId: &foreignkeyv1.StringValueOrRef{...},
    // NetworkName missing
},
// Expect validation error
```

**Pattern validation tests**:
```go
// Test invalid format
NetworkName: "INVALID-NAME",  // Uppercase not allowed
// Expect validation error
```

**Happy path tests**:
```go
// Test valid name
NetworkName: "valid-network-name",
// Expect no error
```

## Before/After Comparison

### GcpVpc Example

**Before**:
```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpVpc
metadata:
  name: prod-network  # Dual purpose - confusing
spec:
  project_id:
    value: prod-project
  auto_create_subnetworks: false
```

**After**:
```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpVpc
metadata:
  name: prod-network  # Planton tracking identifier
spec:
  project_id:
    value: prod-project
  network_name: prod-network  # Explicit GCP VPC name
  auto_create_subnetworks: false
```

### GcpRouterNat Example (Two Resources)

**After**:
```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpRouterNat
metadata:
  name: prod-nat-gateway
spec:
  vpc_self_link:
    ref:
      kind: GcpVpc
      name: prod-vpc
  router_name: prod-router-uscentral1  # Cloud Router name
  nat_name: prod-nat-uscentral1        # NAT config name
  region: us-central1
```

Now it's explicit that this resource creates:
1. A Cloud Router named `prod-router-uscentral1`
2. A NAT configuration named `prod-nat-uscentral1`

## Migration Guide

### For Existing Manifests

**Step 1**: Identify which GCP components you're using

**Step 2**: Add the appropriate name field(s) to each manifest's spec:

| Component | Field(s) to Add |
|-----------|----------------|
| GcpVpc | `network_name` |
| GcpSubnetwork | `subnetwork_name` |
| GcpRouterNat | `router_name`, `nat_name` |
| GcpGkeCluster | `cluster_name` |
| GcpGkeNodePool | `node_pool_name` |
| GcpGcsBucket | `bucket_name` |

**Step 3**: Set the name field value (typically same as metadata.name for consistency):

```yaml
metadata:
  name: my-resource
spec:
  # ... existing fields ...
  <name_field>: my-resource  # Use same name or customize
```

**Step 4**: Validate your manifest:
- Ensure name follows GCP naming rules (lowercase, hyphens, starts with letter)
- Check length constraints (1-63 chars for most, 1-40 for GKE)
- Verify no uppercase letters or special characters

### Validation Errors You Might See

If you forget the name field:
```
validation error:
- spec.network_name: value is required [required]
```

If you use invalid format:
```
validation error:
- spec.network_name: value must match pattern ^[a-z]([a-z0-9-]{0,61}[a-z0-9])?$ [string.pattern]
```

## Related Work

### Architecture Alignment

This refactoring aligns with Planton's design principles from `architecture/deployment-component.md`:

> **80/20 Scoping**: Fields reflect research findings (not every possible provider option)

Explicit name fields are essential (part of the 20%) and should be in the spec.

### Future Enhancements

This pattern can be extended to other cloud providers:
- AWS components (VPC, subnets, clusters)
- Azure components (VNets, subnets, AKS)
- Other GCP components not yet refactored

### Consistency with Kubernetes Resources

This follows the pattern already established in Kubernetes workload components, where resource names are explicit in the spec rather than derived from metadata.

## Lessons Learned

### What Worked Well

1. **Systematic approach**: Updating all 6 components in one pass ensured consistency
2. **Test-driven**: Adding validation tests before implementation caught issues early
3. **Separation of concerns**: Keeping labels using metadata.name while resources use spec fields provided clear boundaries
4. **Pattern validation**: GCP naming rules enforced at proto level prevent deployment failures

### Challenges Addressed

1. **Label vs Resource naming**: Initially unclear whether labels should use new fields - decision: keep using metadata.name
2. **GcpRouterNat complexity**: Required two name fields for two resources - solved with clear naming (router_name, nat_name)
3. **Test coverage**: Needed to add missing test cases for new required fields across all components

---

**Status**: ✅ Complete, Production Ready
**Validation**: All tests passed, build successful
**Breaking Change**: Yes - requires manifest updates
**Rollout**: Immediate (single PR)


