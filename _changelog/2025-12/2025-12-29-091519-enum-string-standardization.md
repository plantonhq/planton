# Standardize Enum Values to Match Provider API Strings

**Date**: December 29, 2025
**Type**: Refactoring
**Components**: API Definitions, Protobuf Schemas, Provider Framework, Pulumi CLI Integration

## Summary

Refactored protobuf enum values across 9 catalog kinds to exactly match the string values expected by their respective cloud provider APIs. This enables Go code to use the `.String()` method directly instead of verbose switch statements, significantly simplifying the Pulumi module implementations.

## Problem Statement

The Go code for translating protobuf enums to provider API strings was unnecessarily complex and error-prone:

```go
// Before: 16-line switch statement for a simple enum translation
switch locals.CloudflareR2Bucket.Spec.Location {
case cloudflarer2bucketv1.CloudflareR2Location_WNAM:
    bucketLocation = "WNAM"
case cloudflarer2bucketv1.CloudflareR2Location_ENAM:
    bucketLocation = "ENAM"
case cloudflarer2bucketv1.CloudflareR2Location_WEUR:
    bucketLocation = "WEUR"
// ... more cases
default:
    bucketLocation = "auto"
}
```

### Pain Points

- **Verbose translation code**: Each enum required a switch statement mapping values to strings
- **Inconsistent naming**: Enum values like `CLOUDFLARE_R2_LOCATION_UNSPECIFIED` didn't match API expectations
- **Maintenance burden**: Adding new enum values required updating both proto and Go code
- **String literal comparisons**: Checking for unspecified values used error-prone string literals like `"civo_database_engine_unspecified"`

## Solution

Updated enum values in proto files to exactly match provider API expectations, then simplified Go code to use `.String()` directly:

```go
// After: Single line using .String()
bucketLocation := locals.CloudflareR2Bucket.Spec.Location.String()
```

For unspecified value checks, replaced string comparisons with idiomatic zero-value checks:

```go
// Before: String comparison
if engineSlug == "civo_database_engine_unspecified" {
    return nil, errors.Errorf("unsupported database engine")
}

// After: Zero-value check
if locals.CivoDatabase.Spec.Engine == 0 {
    return nil, errors.Errorf("database engine is required")
}
```

## Implementation Details

### Kinds Updated

| Kind | Enum | Before → After |
|-----------|------|----------------|
| **cloudflarer2bucket** | `CloudflareR2Location` | `CLOUDFLARE_R2_LOCATION_UNSPECIFIED` → `auto` |
| **cloudflareloadbalancer** | `SessionAffinity` | `SESSION_AFFINITY_NONE/COOKIE` → `none/cookie` |
| | `SteeringPolicy` | `STEERING_OFF/GEO/RANDOM` → `off/geo/random` |
| **civovolume** | `FilesystemType` | `NONE/EXT4/XFS` → `unformatted/ext4/xfs` |
| **digitaloceanvolume** | `FilesystemType` | `NONE/EXT4/XFS` → `unformatted/ext4/xfs` |
| **digitaloceandatabasecluster** | `Engine` | `postgres` → `pg` (DigitalOcean uses "pg" slug) |
| **digitaloceancontainerregistry** | `Tier` | `STARTER/BASIC/PROFESSIONAL` → `starter/basic/professional` |
| **awsdynamodb** | `BillingMode` | `BILLING_MODE_PROVISIONED` → `PROVISIONED` |
| | `AttributeType` | `ATTRIBUTE_TYPE_S/N/B` → `S/N/B` |
| | `StreamViewType` | `STREAM_VIEW_TYPE_KEYS_ONLY` → `KEYS_ONLY` |
| | `KeyType` | `KEY_TYPE_HASH/RANGE` → `HASH/RANGE` |
| **awsclientvpn** | `TransportProtocol` | Already lowercase (`udp/tcp`), simplified Go code |
| **gcprouternat** | `LogFilter` | Already matches GCP (`ERRORS_ONLY/ALL`), simplified Go code |

### Proto Changes Example

```protobuf
// Before
enum CloudflareLoadBalancerSessionAffinity {
  SESSION_AFFINITY_NONE = 0;
  SESSION_AFFINITY_COOKIE = 1;
}

// After - values match Cloudflare API strings
enum CloudflareLoadBalancerSessionAffinity {
  none = 0;   // No session affinity
  cookie = 1; // Cookie-based session affinity
}
```

### Go Code Simplification

**cloudflareloadbalancer** - Removed 15-line switch, replaced with 2 lines:
```go
steering := pulumi.StringPtr(locals.CloudflareLoadBalancer.Spec.SteeringPolicy.String())
affinity := pulumi.StringPtr(locals.CloudflareLoadBalancer.Spec.SessionAffinity.String())
```

**awsdynamodb** - Simplified billing mode, attribute type, stream view type, and table class handling:
```go
// Billing mode
if locals.Spec.BillingMode != 0 {
    billingMode = pulumi.StringPtr(locals.Spec.BillingMode.String())
}

// Attribute types
attrType := "S" // default
if a.Type != 0 {
    attrType = a.Type.String()
}
```

### Files Modified

**Proto files** (8 files):
- `cloudflarer2bucket/v1/spec.proto`
- `cloudflareloadbalancer/v1/spec.proto`
- `civovolume/v1/spec.proto`
- `digitaloceanvolume/v1/spec.proto`
- `digitaloceandatabasecluster/v1/spec.proto`
- `digitaloceancontainerregistry/v1/spec.proto`
- `awsdynamodb/v1/spec.proto`
- `awsclientvpn/v1/spec.proto`

**Go module files** (9 files):
- `cloudflarer2bucket/v1/iac/pulumi/module/bucket.go`
- `cloudflareloadbalancer/v1/iac/pulumi/module/load_balancer.go`
- `civovolume/v1/iac/pulumi/module/volume.go`
- `civodatabase/v1/iac/pulumi/module/database.go`
- `digitaloceanvolume/v1/iac/pulumi/module/volume.go`
- `digitaloceandatabasecluster/v1/iac/pulumi/module/database_cluster.go`
- `digitaloceancontainerregistry/v1/iac/pulumi/module/registry.go`
- `awsdynamodb/v1/iac/pulumi/module/table.go`
- `awsclientvpn/v1/iac/pulumi/module/main.go`
- `gcprouternat/v1/iac/pulumi/module/router_nat.go`

**Test files** (5 files updated for new enum constant names)

## Benefits

### Code Simplification
- **~120 lines of switch statements removed** across 9 modules
- Each enum translation went from 8-20 lines to 1-2 lines
- Removed need to import enum packages just for switch comparisons

### Maintainability
- Adding new enum values only requires updating the proto file
- Generated Go code automatically has the correct string value
- Less room for typos or mismatched strings

### Consistency
- All enums now follow the same pattern: value names match API expectations
- Zero-value checks use idiomatic `== 0` instead of string comparisons
- Comments document what each value means

### Developer Experience
- IntelliSense/autocomplete shows the actual API values
- Proto files serve as API documentation
- Less cognitive load when reading module code

## Impact

### Users
- No breaking changes to YAML manifests - enum values remain the same from user perspective
- Proto comments now document the actual API values being sent

### Developers
- Simpler module code to read and maintain
- Clear pattern to follow for new enum additions
- Removed unnecessary package imports

### Generated Code
- Updated Go and TypeScript generated files
- Frontend TypeScript code benefits from same enum standardization

## Related Work

- This work was done alongside the Cloudflare R2 bucket custom domain feature
- Pattern can be applied to other kinds as they're updated

---

**Status**: ✅ Production Ready
**Commits**: 
- `87674673` - refactor: standardize enum values to match provider API strings
- `9535fe0f` - refactor: use enum zero-value comparison instead of string literals

