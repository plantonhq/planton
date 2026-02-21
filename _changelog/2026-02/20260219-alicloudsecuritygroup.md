# AlicloudSecurityGroup Component Added

**Date**: 2026-02-19
**Component**: AlicloudSecurityGroup
**Enum**: 3022
**ID Prefix**: acsg

## Summary

Added the AlicloudSecurityGroup deployment component -- a stateful virtual firewall that controls inbound and outbound traffic for VPC-based resources.

This component bundles an `alicloud_security_group` with `alicloud_security_group_rule` resources (per DD07 composite bundling), ensuring the security group is always provisioned with its intended access policy.

## What Was Created

### API Definition
- `apis/org/openmcf/provider/alicloud/alicloudsecuritygroup/v1/` -- Full proto API (spec, api, stack_input, stack_outputs)
- `AlicloudSecurityGroupSpec` with `repeated AlicloudSecurityGroupRule rules` for composite bundling
- Registered `AlicloudSecurityGroup = 3022` in `CloudResourceKind` enum under the Networking category

### IaC Modules
- **Pulumi** (Go): Creates alicloud provider, `ecs.SecurityGroup`, and `ecs.SecurityGroupRule` per rule entry with `nic_type=intranet` hardcoded and `pulumi.Parent(sg)` dependency chain
- **Terraform** (HCL): `alicloud_security_group` + `alicloud_security_group_rule` with `for_each` over rules map

### Tests
- Ginkgo/Gomega spec validation tests: 25 specs covering valid inputs (minimal, full config, rules with all fields, mixed ingress/egress, SG-to-SG references, drop policy, priority boundaries, StringValueOrRef), invalid inputs (missing required fields, wrong api_version/kind, invalid CEL values for type/ip_protocol/policy/inner_access_policy, priority out of range)

### Documentation
- README.md with configuration reference, port range format guide, output reference, and related components
- examples.md with 3 YAML examples (web tier, database tier, SG-to-SG application tier)
- catalog-page.md with full configuration reference and deployment examples
- docs/README.md with comprehensive research documentation covering normal vs enterprise SGs, rule evaluation, immutability constraints, and provider resource mapping
- Pulumi overview.md with module architecture and control flow
- 3 presets: web-tier, database-tier, bastion-host

## Spec Design Decisions

- **`security_group_name` not `name`**: Uses the provider-authentic field name. The provider deprecated `name` in v1.239.0.
- **`inner_access_policy` included**: Meaningful security control that determines intra-group traffic behavior. CEL-validated to `["Accept", "Drop"]` using provider-exact casing.
- **`nic_type` hardcoded**: Since `vpc_id` is required, all rules are VPC-based and `nic_type` must be `"intranet"`. Not exposed as a user field.
- **Single rules list with `type` field**: Follows Azure's pattern and the Alibaba Cloud provider's native model (direction as a field, not structural separation).
- **Rule defaults via `optional` + options.default**: `port_range` defaults to `"-1/-1"`, `priority` defaults to `1`, `policy` defaults to `"accept"`. Resolved in locals.go helper functions.
- **`resource_group_id` included**: Consistent with AlicloudVpc and AlicloudLogProject per DD05.
- **Fields excluded for v1**: `security_group_type` (enterprise), `ipv6_cidr_ip`, `prefix_list_id`, `source_group_owner_account`. All can be added later as non-breaking changes.

## Verification

- `go build ./...` -- PASS
- `go vet ./...` -- PASS
- `go test ./...` -- PASS (25/25 specs)
- Pulumi `go build ./...` -- PASS
- Pulumi `go vet ./...` -- PASS
- `terraform init` -- PASS (alicloud provider v1.271.0)
- `terraform validate` -- PASS
