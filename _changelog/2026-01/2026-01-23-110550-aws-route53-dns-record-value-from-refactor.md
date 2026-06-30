# AWS Route53 DNS Record: StringValueOrRef Refactoring

**Date**: 2026-01-23
**Component**: `aws-route53-dns-record`
**Type**: Refactor

## Summary

Refactored `AwsRoute53DnsRecord` to use `StringValueOrRef` for zone and alias target fields, enabling seamless resource wiring via `value_from` references. Renamed `hosted_zone_id` to `zone_id` for consistency.

## Problem Statement

The initial `AwsRoute53DnsRecord` implementation required users to manually look up and hardcode:

1. **Route53 Zone IDs**: Users had to copy zone IDs from AWS console or other resources
2. **ALB DNS Names**: Required copying the full load balancer DNS name
3. **ALB Hosted Zone IDs**: AWS service-specific zone IDs that vary by region

This created friction in GitOps workflows where resources should reference each other declaratively, not through copy-pasted values.

## Solution

Converted three key fields to `StringValueOrRef` type with sensible defaults:

| Field | Default Kind | Default Field Path |
|-------|--------------|-------------------|
| `spec.zone_id` | `AwsRoute53Zone` | `status.outputs.zone_id` |
| `alias_target.dns_name` | `AwsAlb` | `status.outputs.load_balancer_dns_name` |
| `alias_target.zone_id` | `AwsAlb` | `status.outputs.load_balancer_hosted_zone_id` |

## Implementation Details

### Schema Changes (spec.proto)

**Before:**
```protobuf
message AwsRoute53DnsRecordSpec {
  string hosted_zone_id = 1;
  // ...
}

message AwsRoute53AliasTarget {
  string dns_name = 1;
  string hosted_zone_id = 2;
}
```

**After:**
```protobuf
message AwsRoute53DnsRecordSpec {
  StringValueOrRef zone_id = 1 [
    (default_kind) = AwsRoute53Zone,
    (default_kind_field_path) = "status.outputs.zone_id"
  ];
  // ...
}

message AwsRoute53AliasTarget {
  StringValueOrRef dns_name = 1 [
    (default_kind) = AwsAlb,
    (default_kind_field_path) = "status.outputs.load_balancer_dns_name"
  ];
  StringValueOrRef zone_id = 2 [
    (default_kind) = AwsAlb,
    (default_kind_field_path) = "status.outputs.load_balancer_hosted_zone_id"
  ];
}
```

### Updated CEL Validations

Adjusted mutual exclusivity checks to handle `StringValueOrRef` nested structure:

```protobuf
// Check alias_target presence via nested value/value_from
expression: "size(this.values) == 0 || !has(this.alias_target) || 
  (!has(this.alias_target.dns_name) && !has(this.alias_target.dns_name.value) && 
   !has(this.alias_target.dns_name.value_from))"
```

### Pulumi Module

Updated value extraction to use `GetValue()` method:

```go
zoneId := ""
if spec.ZoneId != nil {
    zoneId = spec.ZoneId.GetValue()
}
```

### Terraform Module

Updated variable structure to accept resolved `StringValueOrRef` values:

```hcl
variable "spec" {
  type = object({
    zone_id = object({
      value = optional(string)
    })
    // ...
  })
}
```

## Usage Examples

**Literal value (backward compatible pattern):**
```yaml
spec:
  zone_id:
    value: Z1234567890ABC
```

**Reference to AwsRoute53Zone:**
```yaml
spec:
  zone_id:
    value_from:
      name: my-zone
```

**ALB alias with value_from (primary use case):**
```yaml
spec:
  zone_id:
    value_from:
      name: my-zone
  alias_target:
    dns_name:
      value_from:
        name: my-alb
    zone_id:
      value_from:
        name: my-alb
    evaluate_target_health: true
```

## Files Changed

```
apis/dev/planton/provider/aws/awsroute53dnsrecord/v1/
├── spec.proto              # StringValueOrRef fields, updated validations
├── spec.pb.go              # Generated Go code
├── stack_outputs.proto     # Renamed hosted_zone_id → zone_id
├── stack_outputs.pb.go     # Generated
├── spec_test.go            # Updated test helpers and cases
├── README.md               # Updated documentation
├── examples.md             # Comprehensive value_from examples
├── iac/
│   ├── pulumi/module/
│   │   ├── main.go         # GetValue() extraction
│   │   └── outputs.go      # Renamed constant
│   ├── tf/
│   │   ├── variables.tf    # Nested object structure
│   │   ├── locals.tf       # Value extraction
│   │   ├── main.tf         # Updated references
│   │   └── outputs.tf      # Renamed output
│   └── hack/manifest.yaml  # Updated test manifest
```

## Benefits

1. **Declarative Resource Wiring**: Reference zones and ALBs by name instead of IDs
2. **Reduced Human Error**: No more copy-paste mistakes with zone IDs
3. **GitOps-Friendly**: Resources can be defined together and wired automatically
4. **Consistent Naming**: `zone_id` aligns with `AwsRoute53Zone` output field name
5. **Smart Defaults**: ALB is the default kind for alias targets (most common use case)

## Breaking Changes

- Field renamed: `hosted_zone_id` → `zone_id` in both spec and alias target
- Field types changed from `string` to `StringValueOrRef`
- Users must update manifests to use `value:` wrapper for literal values

## Migration Guide

**Before:**
```yaml
spec:
  hosted_zone_id: Z1234567890ABC
  alias_target:
    dns_name: my-alb.us-east-1.elb.amazonaws.com
    hosted_zone_id: Z35SXDOTRQ7X7K
```

**After:**
```yaml
spec:
  zone_id:
    value: Z1234567890ABC
  alias_target:
    dns_name:
      value: my-alb.us-east-1.elb.amazonaws.com
    zone_id:
      value: Z35SXDOTRQ7X7K
```

Or better, use references:
```yaml
spec:
  zone_id:
    value_from:
      name: my-zone
  alias_target:
    dns_name:
      value_from:
        name: my-alb
    zone_id:
      value_from:
        name: my-alb
```
