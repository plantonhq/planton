# CivoDnsRecord: Convert zone_id to StringValueOrRef

**Date**: January 23, 2026
**Type**: Refactoring
**Components**: CivoDnsRecord, API Definitions, Pulumi Module

## Summary

Updated the `zone_id` field in `CivoDnsRecordSpec` from a plain `string` type to `StringValueOrRef`, enabling users to either provide a literal zone ID or reference the output from a `CivoDnsZone` resource. This aligns CivoDnsRecord with the established pattern used by AwsRoute53DnsRecord and GcpDnsRecord.

## Problem Statement / Motivation

The CivoDnsRecord component was initially forged with `zone_id` as a plain string field. This was inconsistent with other DNS record components in the project which use `StringValueOrRef` for zone references.

### Pain Points

- Users couldn't wire CivoDnsRecord to CivoDnsZone outputs using `value_from`
- Inconsistent API patterns across DNS record components (AWS, GCP, Azure all use `StringValueOrRef`)
- Manual copy-paste of zone IDs instead of declarative resource references

## Solution / What's New

Updated the `zone_id` field to use the `StringValueOrRef` type with proper default kind hints:

```protobuf
dev.planton.shared.foreignkey.v1.StringValueOrRef zone_id = 1 [
  (buf.validate.field).required = true,
  (dev.planton.shared.foreignkey.v1.default_kind) = CivoDnsZone,
  (dev.planton.shared.foreignkey.v1.default_kind_field_path) = "status.outputs.zone_id"
];
```

### Usage Patterns

Users can now specify zone_id in two ways:

```yaml
# Literal value
spec:
  zone_id: "zone-abc123"

# OR reference from CivoDnsZone
spec:
  zone_id:
    value_from:
      name: my-dns-zone
```

## Implementation Details

### Files Changed

1. **spec.proto** - Updated field type and added foreign key annotations
2. **locals.go** - Added `ZoneId` field extraction using `GetValue()`
3. **dns_record.go** - Updated to use `locals.ZoneId` for Pulumi resource
4. **spec_test.go** - Updated all tests to use `StringValueOrRef` with helper function

### Pulumi Module Update

```go
// locals.go
type Locals struct {
    // ... other fields
    ZoneId string  // Extracted from StringValueOrRef
}

func initializeLocals(...) *Locals {
    // Extract zone ID from StringValueOrRef
    locals.ZoneId = target.Spec.ZoneId.GetValue()
    // ...
}
```

### Test Helper Function

Added a helper to simplify test construction:

```go
func strVal(s string) *foreignkeyv1.StringValueOrRef {
    return &foreignkeyv1.StringValueOrRef{
        LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: s},
    }
}
```

## Benefits

- **Consistent API patterns** across all DNS record components
- **Declarative resource wiring** via `value_from` references
- **Reduced manual errors** by eliminating copy-paste of zone IDs
- **Better developer experience** with IDE autocomplete for referenced resources

## Impact

- **API Change**: The `zone_id` field type changed from `string` to `StringValueOrRef`
- **Backward Compatible**: Users can still provide literal strings via the `value` field
- **All tests pass**: 21 test cases covering valid and invalid inputs

## Related Work

- AwsRoute53DnsRecord uses same pattern for `zone_id`
- GcpDnsRecord uses same pattern for `managed_zone` and `project_id`
- Part of broader DNS record component standardization effort

---

**Status**: ✅ Production Ready
