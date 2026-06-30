# GCP Project Delete Protection Field

**Date**: December 20, 2025  
**Type**: Enhancement  
**Components**: GCP Provider, API Definitions, Pulumi Module, Terraform Module

## Summary

Added a `delete_protection` boolean field to the GcpProject spec that configures GCP's native project deletion protection. When enabled, the project cannot be deleted until the flag is explicitly set to false. This is passed as an argument to the GCP API (not a Pulumi/Terraform lifecycle option), providing an additional layer of protection against accidental deletion of critical projects.

## Problem Statement / Motivation

Production GCP projects containing critical infrastructure can be accidentally deleted through automation errors, misconfigurations, or human mistakes. While GCP has a 30-day soft-delete recovery window, this doesn't protect against authorized-but-erroneous `terraform destroy` or `pulumi destroy` commands.

### Pain Points

- No declarative way to enable GCP's native deletion protection through Planton
- Production projects vulnerable to accidental deletion via IaC operations
- Users had to manually enable deletion protection through GCP Console or gcloud CLI
- Inconsistency between what's declared in manifests and actual GCP project protection state

## Solution / What's New

Added a new optional boolean field `delete_protection` to the `GcpProjectSpec` message that maps directly to GCP's native `deletion_policy` resource attribute.

### Field Definition

```protobuf
// If true, enables deletion protection on the GCP project.
// When enabled, the project cannot be deleted until this flag is set to false.
// This is a GCP-native feature (not a Pulumi/Terraform lifecycle option)
// that provides an additional layer of protection against accidental deletion.
// Defaults to false.
optional bool delete_protection = 10 [(dev.planton.shared.options.default) = "false"];
```

### Mapping Logic

| `delete_protection` | GCP `deletion_policy` |
|---------------------|----------------------|
| `true`              | `"PREVENT"`          |
| `false` (default)   | `"DELETE"`           |

## Implementation Details

### Proto Schema Update

**File**: `apis/dev/planton/provider/gcp/gcpproject/v1/spec.proto`

Added field number 10 to `GcpProjectSpec` with default value of `false` and appropriate documentation explaining that this is a GCP-native feature.

### Pulumi Module Update

**File**: `apis/dev/planton/provider/gcp/gcpproject/v1/iac/pulumi/module/project.go`

```go
// Determine deletion policy based on delete_protection flag
// When delete_protection is true, set to "PREVENT" to block project deletion
deletionPolicy := "DELETE"
if locals.GcpProject.Spec.GetDeleteProtection() {
    deletionPolicy = "PREVENT"
}

projectArgs := &organizations.ProjectArgs{
    // ... other fields ...
    DeletionPolicy: pulumi.String(deletionPolicy),
}
```

### Terraform Module Updates

**File**: `apis/dev/planton/provider/gcp/gcpproject/v1/iac/tf/variables.tf`

```hcl
variable "spec" {
  type = object({
    # ... existing fields ...
    delete_protection = optional(bool)  # If true, enables GCP-native deletion protection
  })
}
```

**File**: `apis/dev/planton/provider/gcp/gcpproject/v1/iac/tf/locals.tf`

```hcl
# Deletion policy configuration
# When delete_protection is true, set to "PREVENT" to block project deletion
deletion_policy = var.spec.delete_protection == true ? "PREVENT" : "DELETE"
```

**File**: `apis/dev/planton/provider/gcp/gcpproject/v1/iac/tf/main.tf`

```hcl
resource "google_project" "this" {
  # ... other fields ...
  deletion_policy = local.deletion_policy
}
```

## Usage Examples

### YAML Manifest

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpProject
metadata:
  name: prod-payment-processing
spec:
  parentType: folder
  parentId: "345678901234"
  billingAccountId: "ABCDEF-123456-ABCDEF"
  labels:
    env: prod
    criticality: high
  disableDefaultNetwork: true
  deleteProtection: true  # CRITICAL: Prevent accidental deletion
  enabledApis:
    - compute.googleapis.com
    - storage.googleapis.com
```

### Terraform HCL

```hcl
module "prod_project" {
  source = "../../tf"
  
  spec = {
    project_id              = "prod-payment-proc"
    parent_type             = "folder"
    parent_id               = "345678901234"
    billing_account_id      = "ABCDEF-123456-ABCDEF"
    disable_default_network = true
    delete_protection       = true  # Prevent accidental deletion
  }
}
```

## Benefits

- **Production Safety**: Critical projects are protected from accidental deletion
- **Declarative Control**: Protection state is managed as code, not manual configuration
- **IaC Consistency**: Both Pulumi and Terraform modules support the same field
- **Default Safe**: Defaults to `false` for backward compatibility with existing manifests
- **GCP-Native**: Uses GCP's actual deletion protection, not just IaC lifecycle hooks

## Impact

### Users
- Can now declaratively enable deletion protection on production GCP projects
- Protection persists even if Pulumi/Terraform state is corrupted or lost
- Clear documentation and examples for best practices

### Developer Experience
- Consistent field naming (`delete_protection`) across the API
- Clear mapping to GCP's underlying `deletion_policy` attribute
- Updated examples in all three locations (YAML, Terraform, Pulumi)

## Files Changed

| File | Change |
|------|--------|
| `apis/.../gcpproject/v1/spec.proto` | Added `delete_protection` field |
| `apis/.../gcpproject/v1/iac/pulumi/module/project.go` | Added DeletionPolicy logic |
| `apis/.../gcpproject/v1/iac/tf/variables.tf` | Added variable |
| `apis/.../gcpproject/v1/iac/tf/locals.tf` | Added local mapping |
| `apis/.../gcpproject/v1/iac/tf/main.tf` | Added deletion_policy argument |
| `apis/.../gcpproject/v1/examples.md` | Updated with delete protection examples |
| `apis/.../gcpproject/v1/iac/tf/examples.md` | Updated production example |
| `apis/.../gcpproject/v1/iac/pulumi/examples.md` | Added new protection example |

## Validation

- ✅ Proto stubs regenerated with `make protos`
- ✅ Component tests passed: `go test ./apis/dev/planton/provider/gcp/gcpproject/v1/`
- ✅ Full build completed: `make build`
- ✅ All tests passed: `make test`

## Related Work

- **GcpProjectLien** (mentioned in docs/README.md): Separate resource for project liens - complementary protection mechanism
- **Research Document**: `v1/docs/README.md` discusses deletion safety and the 30-day recovery window

---

**Status**: ✅ Production Ready  
**Timeline**: Single session implementation
