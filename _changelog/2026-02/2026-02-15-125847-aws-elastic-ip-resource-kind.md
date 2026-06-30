# AWS Elastic IP Resource Kind (R10)

**Date**: February 15, 2026
**Type**: Feature
**Components**: API Definitions, Pulumi Module, Terraform Module, Documentation

## Summary

Added AwsElasticIp as the twelfth new AWS resource kind in the cloud provider expansion project. This is a fundamentally simple component — a static public IPv4 address allocation — whose primary value lies in its outputs (`allocation_id`, `public_ip`) consumed by NLB, NAT Gateway, and EC2. Also backfilled the NLB's `allocation_id` field with the now-registerable `default_kind` annotation.

## Problem Statement / Motivation

The AwsNetworkLoadBalancer (R09) was shipped with a gap: its `allocation_id` field in subnet mappings had no `default_kind` annotation because AwsElasticIp didn't exist yet. This meant NLB users couldn't use the ergonomic `valueFrom` pattern to reference Elastic IPs — they had to hardcode allocation IDs. Additionally, the Elastic IP itself is a foundational networking primitive needed for static IPs on NLBs, NAT Gateways, and bastion hosts.

### Pain Points

- NLB subnet mappings required hardcoded allocation IDs (no `valueFrom` support)
- No way to declaratively manage static public IPs in Planton
- Missing dependency in the infra chart DAG for IP allocation

## Solution / What's New

### AwsElasticIp Component

A nearly-zero-config component that allocates a VPC Elastic IP:

- **3 optional spec fields**: `public_ipv4_pool`, `address`, `network_border_group`
- **1 CEL validation**: `address` requires `public_ipv4_pool` (BYOIP constraint)
- **4 outputs**: `allocation_id`, `public_ip`, `arn`, `public_dns`
- **Domain hardcoded to "vpc"** in IaC (EC2-Classic retired)

### NLB Backfill

Updated the NLB's `allocation_id` field with:
```protobuf
(dev.planton.shared.foreignkey.v1.default_kind) = AwsElasticIp
(dev.planton.shared.foreignkey.v1.default_kind_field_path) = "status.outputs.allocation_id"
```

## Implementation Details

### Proto API (4 files)

- `spec.proto` — 3 optional string fields, 1 message-level CEL validation
- `stack_outputs.proto` — 4 output fields (allocation_id, public_ip, arn, public_dns)
- `api.proto` — KRM wiring with const apiVersion/kind
- `stack_input.proto` — AwsElasticIp + AwsProviderConfig

### Validation Tests (10 tests, all passing)

- 5 happy path: minimal, network_border_group, BYOIP pool, BYOIP + address, all fields
- 5 failure: address without pool (CEL), wrong apiVersion, wrong kind, missing metadata, missing spec

### Pulumi Module (4 files)

- `main.go` — provider setup, orchestration, exports
- `locals.go` — tag initialization with AwsElasticIp enum
- `outputs.go` — 4 output key constants
- `eip.go` — single `ec2.NewEip` with conditional BYOIP/zone fields

### Terraform Module (5 files)

- `main.tf` — single `aws_eip` resource with null coalescing for optional fields
- `locals.tf` — tag merging, empty-string-to-null conversion
- Feature parity with Pulumi module

### Documentation

- `README.md` — spec reference, examples, deliberate omissions
- `examples.md` — 7 examples from minimal to Wavelength zone
- `docs/README.md` — architecture: addressing model, BYOIP, immutability, cost, security
- Catalog page — audited, zero Critical issues
- 2 presets: standard-eip, byoip-pool

### Enum Registration

`AwsElasticIp = 281` in `cloud_resource_kind.proto`, id_prefix: `awseip`

## Benefits

- **Closes the NLB gap**: `valueFrom` references from NLB to EIP now work seamlessly
- **Infra chart ready**: EIP outputs (`allocation_id`) compose into the dependency DAG
- **Zero-config for 95% of users**: Empty spec allocates a standard VPC EIP
- **BYOIP support**: 3 optional fields cover the BYOIP edge case without cluttering the UX

## Impact

- **New resource kind**: AwsElasticIp (enum 281) — the 37th cloud resource kind in Planton
- **NLB improvement**: Existing AwsNetworkLoadBalancer gains `default_kind` for `allocation_id`
- **Files created**: ~35 files in `apis/dev/planton/provider/aws/awselasticip/v1/`
- **Files modified**: NLB spec.proto, cloud_resource_kind.proto, site catalog index

## Related Work

- AwsNetworkLoadBalancer (R09) — primary consumer of EIP allocation IDs
- 20260215.02.sp.aws-resource-expansion — parent project tracking ~32 new AWS resource kinds

---

**Status**: Production Ready
