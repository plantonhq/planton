# KubernetesExternalDns: StringValueOrRef Migration for AKS and Cloudflare

**Date**: November 24, 2025
**Type**: Enhancement
**Components**: API Definitions, Pulumi Integration, Terraform Integration, Kubernetes Provider

## Summary

Migrated `dns_zone_id` fields in `KubernetesExternalDnsAksConfig` and `KubernetesExternalDnsCloudflareConfig` from plain `string` to `StringValueOrRef`, enabling support for both literal DNS zone ID values and foreign key references. This change brings AKS and Cloudflare configurations into alignment with the existing GKE and EKS patterns, providing consistent cross-provider foreign key support while maintaining backward compatibility through the `.value` accessor.

## Problem Statement / Motivation

The KubernetesExternalDns component had an inconsistency in how DNS zone IDs were handled across different cloud providers. GKE and EKS configurations already supported `StringValueOrRef` for their zone ID fields, allowing users to either provide literal zone IDs or reference them from other resources. However, AKS and Cloudflare configurations were still using plain `string` types, limiting flexibility and creating an inconsistent API surface.

### Pain Points

- **Inconsistent API**: Different providers had different configuration patterns for the same conceptual field
- **Limited flexibility**: AKS and Cloudflare users couldn't reference DNS zones from other Planton resources
- **Foreign key limitations**: No support for dynamic zone ID resolution at runtime
- **Pattern divergence**: New providers had to choose between the old pattern (string) and new pattern (StringValueOrRef)

## Solution / What's New

Updated the protobuf schema to use `StringValueOrRef` for `dns_zone_id` in both AKS and Cloudflare configurations, then propagated this change through the entire stack:

### Architecture Changes

```
Proto Schema (spec.proto)
    ↓
Generated Go Stubs (spec.pb.go)
    ↓
Pulumi Module (iac/pulumi/module/main.go)
    ↓
Terraform Module (iac/tf/locals.tf, main.tf)
    ↓
Examples & Documentation
    ↓
Test Suite (spec_test.go)
```

### Key Changes

**Proto Schema** (`spec.proto`):
```protobuf
// AKS Config - BEFORE
message KubernetesExternalDnsAksConfig {
  string dns_zone_id = 1 [(buf.validate.field).required = true];
}

// AKS Config - AFTER
message KubernetesExternalDnsAksConfig {
  dev.planton.shared.foreignkey.v1.StringValueOrRef dns_zone_id = 1 
    [(buf.validate.field).required = true];
}

// Cloudflare Config - BEFORE
message KubernetesExternalDnsCloudflareConfig {
  string api_token = 1 [(buf.validate.field).required = true];
  string dns_zone_id = 2 [(buf.validate.field).required = true];
}

// Cloudflare Config - AFTER
message KubernetesExternalDnsCloudflareConfig {
  string api_token = 1 [(buf.validate.field).required = true];
  dev.planton.shared.foreignkey.v1.StringValueOrRef dns_zone_id = 2 
    [(buf.validate.field).required = true];
}
```

**YAML Manifest** (user-facing):
```yaml
# BEFORE
spec:
  aks:
    dns_zone_id: /subscriptions/.../dnszones/example.com

# AFTER
spec:
  aks:
    dns_zone_id:
      value: /subscriptions/.../dnszones/example.com
      # OR reference another resource:
      # ref: azure-dns-zone-prod
```

## Implementation Details

### 1. Pulumi Module Updates

**File**: `apis/dev/planton/provider/kubernetes/kubernetesexternaldns/v1/iac/pulumi/module/main.go`

Updated AKS configuration to use `.GetValue()` accessor and switched from `domainFilters` to `zoneIdFilters` for consistency:

```go
// AKS - BEFORE
case spec.GetAks() != nil:
    aks := spec.GetAks()
    values["provider"] = pulumi.String("azure")
    if aks.DnsZoneId != "" {
        values["domainFilters"] = pulumi.StringArray{pulumi.String(aks.DnsZoneId)}
    }

// AKS - AFTER
case spec.GetAks() != nil:
    aks := spec.GetAks()
    values["provider"] = pulumi.String("azure")
    values["zoneIdFilters"] = pulumi.StringArray{
        pulumi.String(aks.DnsZoneId.GetValue()),
    }
```

Updated Cloudflare configuration to use `.GetValue()` accessor:

```go
// Cloudflare - BEFORE
extraArgs := pulumi.StringArray{
    pulumi.String("--cloudflare-dns-records-per-page=5000"),
    pulumi.String(fmt.Sprintf("--zone-id-filter=%s", cf.DnsZoneId)),
}

// Cloudflare - AFTER
extraArgs := pulumi.StringArray{
    pulumi.String("--cloudflare-dns-records-per-page=5000"),
    pulumi.String(fmt.Sprintf("--zone-id-filter=%s", cf.DnsZoneId.GetValue())),
}
```

**Consistency improvement**: Changed AKS from `domainFilters` to `zoneIdFilters`, matching the pattern used by GKE and EKS. This ensures uniform zone scoping behavior across all providers.

### 2. Terraform Module Updates

**File**: `apis/dev/planton/provider/kubernetes/kubernetesexternaldns/v1/iac/tf/locals.tf`

```hcl
# AKS - BEFORE
aks_dns_zone_id = local.is_aks ? try(var.spec.aks.dns_zone_id, "") : ""

# AKS - AFTER
aks_dns_zone_id = local.is_aks ? try(var.spec.aks.dns_zone_id.value, "") : ""

# Cloudflare - BEFORE
cf_dns_zone_id = local.is_cloudflare ? try(var.spec.cloudflare.dns_zone_id, "") : ""

# Cloudflare - AFTER
cf_dns_zone_id = local.is_cloudflare ? try(var.spec.cloudflare.dns_zone_id.value, "") : ""
```

**File**: `apis/dev/planton/provider/kubernetes/kubernetesexternaldns/v1/iac/tf/main.tf`

```hcl
# AKS - BEFORE
dynamic "set" {
  for_each = local.is_aks && local.aks_dns_zone_id != "" ? [1] : []
  content {
    name  = "domainFilters[0]"
    value = local.aks_dns_zone_id
  }
}

# AKS - AFTER
dynamic "set" {
  for_each = local.is_aks && local.aks_dns_zone_id != "" ? [1] : []
  content {
    name  = "zoneIdFilters[0]"
    value = local.aks_dns_zone_id
  }
}
```

### 3. Test Suite Updates

**File**: `apis/dev/planton/provider/kubernetes/kubernetesexternaldns/v1/spec_test.go`

Updated all AKS and Cloudflare test cases to use the new `StringValueOrRef` structure:

```go
// AKS - BEFORE
Aks: &KubernetesExternalDnsAksConfig{
    DnsZoneId: "my-azure-dns-zone-id",
}

// AKS - AFTER
Aks: &KubernetesExternalDnsAksConfig{
    DnsZoneId: &foreignkeyv1.StringValueOrRef{
        LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
            Value: "my-azure-dns-zone-id",
        },
    },
}

// Cloudflare - BEFORE
Cloudflare: &KubernetesExternalDnsCloudflareConfig{
    ApiToken:  "my-cloudflare-api-token",
    DnsZoneId: "1234567890abcdef1234567890abcdef",
}

// Cloudflare - AFTER
Cloudflare: &KubernetesExternalDnsCloudflareConfig{
    ApiToken: "my-cloudflare-api-token",
    DnsZoneId: &foreignkeyv1.StringValueOrRef{
        LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
            Value: "1234567890abcdef1234567890abcdef",
        },
    },
}
```

**Test fix**: Corrected the "minimal AKS config" test case to properly validate that `dns_zone_id` is required (since it's marked with `buf.validate.field.required = true` in the proto).

### 4. Documentation Updates

Updated examples across multiple files:

**Files Updated**:
- `examples.md` (root-level YAML examples)
- `iac/pulumi/examples.md` (Pulumi Go examples)
- `iac/tf/examples.md` (Terraform examples)

All examples now show the correct `StringValueOrRef` structure:

```yaml
# YAML Example
spec:
  aks:
    dns_zone_id:
      value: /subscriptions/.../Microsoft.Network/dnszones/example.com
      
  cloudflare:
    api_token: your-token
    dns_zone_id:
      value: 1234567890abcdef1234567890abcdef
```

## Benefits

### For API Users

1. **Consistent API surface**: All providers now use the same pattern for DNS zone IDs
2. **Foreign key support**: Can reference DNS zones from other Planton resources
3. **Flexibility**: Choose between literal values and references based on use case
4. **Better composition**: Resources can reference each other, reducing duplication

### For Developers

1. **Pattern consistency**: New providers can follow established `StringValueOrRef` pattern
2. **Reduced cognitive load**: Same approach works across all providers
3. **Type safety**: Proto validation ensures required fields are provided
4. **Better maintainability**: Single pattern to understand and maintain

### Examples of New Capabilities

**Literal Value** (backward compatible):
```yaml
spec:
  cloudflare:
    api_token: secret-token
    dns_zone_id:
      value: abc123def456
```

**Foreign Key Reference** (new capability):
```yaml
spec:
  cloudflare:
    api_token: secret-token
    dns_zone_id:
      ref: cloudflare-dns-zone-prod
```

The foreign key reference will be resolved at runtime to fetch the actual zone ID from the referenced resource.

## Impact

### Breaking Change: Migration Required

**Users must update their manifests** to wrap `dns_zone_id` values in a `value` field for AKS and Cloudflare configurations.

**Migration Path**:

1. **Identify affected manifests**: Any KubernetesExternalDns resources using AKS or Cloudflare
2. **Update YAML structure**: Wrap zone IDs with `value:` key
3. **Test validation**: Ensure manifests pass proto validation
4. **Deploy**: Apply updated manifests

**Example Migration**:

```yaml
# OLD (will fail validation)
spec:
  cloudflare:
    api_token: xxx
    dns_zone_id: abc123

# NEW (required)
spec:
  cloudflare:
    api_token: xxx
    dns_zone_id:
      value: abc123
```

### Components Affected

- **Proto Schema**: `spec.proto` changed for AKS and Cloudflare configs
- **Go Stubs**: Auto-generated from proto changes
- **Pulumi Module**: Updated accessor calls and Helm values
- **Terraform Module**: Updated variable access patterns
- **Test Suite**: All test cases updated and passing
- **Documentation**: All examples updated across 3 files

### Verification

All changes validated through:

1. **Proto compilation**: `make protos` completed successfully
2. **Component tests**: All 13 tests passing (13/13 ✓)
3. **Build validation**: `make build` completed successfully
4. **Code review**: All accessor patterns verified for consistency

## Related Work

- **GKE and EKS configurations**: Already used `StringValueOrRef` for zone IDs
- **Foreign key framework**: Part of broader foreign key support across Planton
- **Provider consistency**: Aligns with ongoing effort to standardize provider APIs
- **Previous changelog**: `2025-10-17-external-dns-cloudflare-support.md` added initial Cloudflare support

## Code Metrics

- **Files Modified**: 7 implementation files + examples
- **Test Cases Updated**: 6 test scenarios (AKS + Cloudflare)
- **Proto Fields Changed**: 2 fields across 2 config messages
- **Examples Updated**: 8 examples across 3 documentation files
- **Lines Changed**: ~50 lines (focused, surgical changes)

## Design Decisions

### Why StringValueOrRef for Zone IDs?

DNS zone IDs are commonly managed as separate resources in Planton. Using `StringValueOrRef` allows users to:

1. Reference centrally-managed DNS zones
2. Avoid hardcoding zone IDs across multiple manifests
3. Enable GitOps workflows where references are resolved at apply time
4. Support multi-environment deployments with zone references

### Why Change AKS from domainFilters to zoneIdFilters?

The ExternalDNS Helm chart supports multiple filtering mechanisms:

- `domainFilters`: Filters by domain name patterns (e.g., `example.com`)
- `zoneIdFilters`: Filters by exact zone IDs (more precise)

Using `zoneIdFilters` provides:

1. **Precision**: Exact zone ID matching prevents accidental cross-zone modifications
2. **Consistency**: All providers (GKE, EKS, AKS) now use zone ID filtering
3. **Safety**: Zone ID is more restrictive than domain pattern matching
4. **Alignment**: Matches the intent of the `dns_zone_id` field name

### Backward Compatibility Strategy

Since this is a **breaking change** to the proto schema:

- ✅ **Clear migration path**: Documented in this changelog
- ✅ **Validation errors**: Proto validation will catch old format immediately
- ✅ **Consistent pattern**: Matches existing GKE/EKS patterns users already know
- ✅ **Minimal impact**: Only affects AKS and Cloudflare users (subset of total users)

## Testing Strategy

### Test Coverage

**Unit Tests** (`spec_test.go`):
- ✅ Valid AKS config with StringValueOrRef
- ✅ Missing dns_zone_id validation error
- ✅ Valid Cloudflare config with StringValueOrRef
- ✅ Cloudflare with proxy enabled
- ✅ Missing api_token validation error
- ✅ Missing dns_zone_id validation error

**Integration Validation**:
- ✅ Proto regeneration (`make protos`)
- ✅ Go compilation (`make build`)
- ✅ Pulumi module syntax validation
- ✅ Terraform locals syntax validation

### Test Results

```bash
$ go test ./apis/dev/planton/provider/kubernetes/kubernetesexternaldns/v1/ -v

Running Suite: KubernetesExternalDns Suite
Random Seed: 1763988160

Will run 13 of 13 specs
•••••••••••••

Ran 13 of 13 Specs in 0.010 seconds
SUCCESS! -- 13 Passed | 0 Failed | 0 Pending | 0 Skipped
```

```bash
$ make build

Computing main repo mapping: 
Loading: 0 packages loaded
INFO: Build completed successfully
```

## Future Enhancements

### Short Term

- Add migration script to automatically update existing manifests
- Create validation warnings for deprecated patterns
- Update CLI to provide helpful error messages for old format

### Long Term

- Consider applying StringValueOrRef to other resource ID fields across providers
- Enhance foreign key resolution to support cross-namespace references
- Add runtime validation for foreign key reference cycles

## Troubleshooting

### Common Issues

**Error: validation error - dns_zone_id: value is required**

**Cause**: Manifest still uses old string format instead of StringValueOrRef

**Solution**: Wrap zone ID with `value:` key:
```yaml
dns_zone_id:
  value: your-zone-id-here
```

**Error: cannot use "zone-id" as *StringValueOrRef**

**Cause**: Code still accessing field as plain string

**Solution**: Use `.GetValue()` accessor:
```go
// Wrong
zoneId := config.DnsZoneId

// Correct
zoneId := config.DnsZoneId.GetValue()
```

**Terraform Error: object required**

**Cause**: Terraform variable still expects string

**Solution**: Update Terraform to access nested value:
```hcl
# Wrong
dns_zone_id = var.spec.aks.dns_zone_id

# Correct
dns_zone_id = var.spec.aks.dns_zone_id.value
```

---

**Status**: ✅ Production Ready
**Timeline**: Completed November 24, 2025

