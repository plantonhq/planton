# AliCloudVpc Component Added

**Date**: 2026-02-19
**Component**: AliCloudVpc
**Enum**: 3020
**ID Prefix**: acvpc

## Summary

Added the AliCloudVpc deployment component -- the networking foundation for virtually every other Alibaba Cloud resource in the catalog.

This component manages an Alibaba Cloud Virtual Private Cloud (VPC), creating the isolated virtual network along with its automatically provisioned virtual router and system route table.

## What Was Created

### API Definition
- `apis/org/openmcf/provider/alicloud/alicloudvpc/v1/` -- Full proto API (spec, api, stack_input, stack_outputs)
- Registered `AliCloudVpc = 3020` in `CloudResourceKind` enum under the Networking category

### IaC Modules
- **Pulumi** (Go): Creates alicloud provider and a single `vpc.Network` resource with all spec fields mapped
- **Terraform** (HCL): Single `alicloud_vpc` resource with matching variables, outputs, and tag merging

### Tests
- Ginkgo/Gomega spec validation tests: 11 specs covering valid inputs (minimal, full config, IPv6, various CIDR ranges), missing required fields (region, vpc_name, cidr_block), wrong api_version/kind, missing metadata, and vpc_name max length

### Documentation
- README.md with configuration reference, CIDR guidance, output reference, and related components
- examples.md with 3 YAML examples (minimal, production with tags, IPv6-enabled)

## Spec Design Decisions

- **`vpc_name` not `name`**: Uses the provider-authentic field name (`vpc_name`), consistent with `project_name` and `role_name` on other Alibaba Cloud components. TF deprecated `name` in v1.119.0.
- **`cidr_block` required**: TF marks it Optional+Computed, but we require it because auto-assigned CIDRs are unpredictable and break downstream VSwitch planning.
- **Tags included**: Consistent with all existing Alibaba Cloud components.
- **No bundled sub-resources**: VPC stands alone per DD07 -- it's a foundation resource referenced by many downstream components.
- **Fields excluded for v1**: `classic_link_enabled`, `dns_hostname_status`, IPAM fields, `ipv6_isp`, `user_cidrs`, system route table customization, `dry_run`, `is_default`, `force_delete`. All can be added later as non-breaking changes.

## Verification

- `go build ./...` -- PASS
- `go vet ./...` -- PASS
- `go test ./...` -- PASS (11/11 specs)
- `terraform init` -- PASS (alicloud provider v1.271.0)
- `terraform validate` -- PASS
