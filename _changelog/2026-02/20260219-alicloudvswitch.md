# AlicloudVswitch Component Added

**Date**: 2026-02-19
**Component**: AlicloudVswitch
**Enum**: 3021
**ID Prefix**: acvsw

## Summary

Added the AlicloudVswitch deployment component -- the subnet-equivalent resource in Alibaba Cloud networking. This is the first Alibaba Cloud component that uses `StringValueOrRef` for a cross-resource dependency (`vpc_id` referencing `AlicloudVpc`), establishing the foreign-key pattern for all downstream Alibaba Cloud networking components.

## What Was Created

### API Definition
- `apis/org/openmcf/provider/alicloud/alicloudvswitch/v1/` -- Full proto API (spec, api, stack_input, stack_outputs)
- Registered `AlicloudVswitch = 3021` in `CloudResourceKind` enum under the Networking category

### IaC Modules
- **Pulumi** (Go): Creates alicloud provider and a single `vpc.NewSwitch` resource; resolves `vpc_id` from `StringValueOrRef` via `GetValue()`
- **Terraform** (HCL): Single `alicloud_vswitch` resource with matching variables, outputs, and tag merging; `vpc_id` received as a pre-resolved string

### Tests
- Ginkgo/Gomega spec validation tests: 17 specs covering valid inputs (minimal, full config, IPv6 bounds, cross-resource reference via `value_from`), missing required fields (region, vpc_id, zone_id, cidr_block, vswitch_name), wrong api_version/kind, missing metadata, vswitch_name max length, and IPv6 CIDR mask bounds

### Documentation
- README.md with configuration reference, CIDR planning guidance, immutable fields callout, and related components
- examples.md with 4 YAML examples (minimal, production with tags, cross-resource reference, IPv6-enabled)
- docs/README.md with comprehensive research documentation
- Presets: 3 ready-to-deploy templates (dev single-zone, production app tier, IPv6-enabled)

## Spec Design Decisions

- **`region` included (deviation from T02)**: T02 spec said "no region field" but every IaC module needs it for provider configuration. Consistent with all other components.
- **`vpc_id` as StringValueOrRef**: First use in Alibaba Cloud provider. Uses `default_kind = AlicloudVpc` and `default_kind_field_path = "status.outputs.vpc_id"`.
- **`vswitch_name` not `name`**: Uses provider-authentic field name. TF deprecated `name` in v1.119.0.
- **`ipv6_cidr_block_mask` validation**: Range 0-255 with `IGNORE_IF_ZERO_VALUE` so zero (proto default) is not flagged.
- **`is_default` excluded**: Creating default VSwitches is an edge case not needed for managed infrastructure.
- **Tags included**: Consistent with AlicloudVpc.

## Verification

- `go build ./...` -- PASS
- `go vet ./...` -- PASS
- `go test ./...` -- PASS (17/17 specs)
- `terraform init` -- PASS (alicloud provider v1.271.0)
- `terraform validate` -- PASS
