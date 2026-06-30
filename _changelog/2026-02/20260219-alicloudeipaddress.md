# AliCloudEipAddress Component Added

**Date**: 2026-02-19
**Component**: AliCloudEipAddress
**Enum**: 3023
**ID Prefix**: aceip

## Summary

Added the AliCloudEipAddress deployment component -- a standalone Elastic IP Address that can be associated with NAT gateways, ALB/NLB load balancers, VPN gateways, and ECS instances.

This component allocates a static, public IPv4 address that persists independently of the resource lifecycle, allowing it to be released from one resource and re-associated with another without changing the address.

## What Was Created

### API Definition
- `apis/dev/planton/provider/alicloud/alicloudeipaddress/v1/` -- Full proto API (spec, api, stack_input, stack_outputs)
- Registered `AliCloudEipAddress = 3023` in `CloudResourceKind` enum under the Networking category

### IaC Modules
- **Pulumi** (Go): Creates alicloud provider and a single `ecs.EipAddress` resource with bandwidth int-to-string conversion and default resolution for optional fields
- **Terraform** (HCL): Single `alicloud_eip_address` resource with matching variables, outputs, and tag merging

### Tests
- Ginkgo/Gomega spec validation tests: 15 specs covering valid inputs (minimal, full config, PayByTraffic, China ISPs, L2/special ISPs, bandwidth boundaries), missing required fields (region), wrong api_version/kind, missing metadata, invalid internet_charge_type, invalid ISP, bandwidth out of range, and address_name max length

### Documentation
- README.md with configuration reference, ISP values table, bandwidth/charging explanation, and related components
- examples.md with 3 YAML examples (minimal, NAT gateway, high-bandwidth production)
- catalog-page.md with full configuration reference and examples
- Pulumi overview.md documenting module architecture, control flow, and implementation details
- Pulumi/TF README.md and examples.md

### Presets
- 01-standard: 5 Mbps, PayByTraffic, BGP (development/staging)
- 02-high-bandwidth: 100 Mbps, PayByBandwidth, BGP_PRO (production)

## Spec Design Decisions

- **`address_name` not `name`**: Uses the provider-authentic field name. TF deprecated `name` in v1.126.0.
- **`bandwidth` as int32 (not string)**: The provider uses string, but bandwidth in Mbps is inherently numeric. Using int32 for better YAML UX with int-to-string conversion in IaC code.
- **`bandwidth` optional, default 5**: The provider defaults to "5". Made optional to match, since 5 Mbps is a sensible default for most development use cases.
- **All 10 ISP values included**: T02 spec listed 5, but the provider supports 10. Including all to avoid artificially limiting users in finance cloud or international regions.
- **`description` and `tags` added**: Not in T02 spec but present in every other Alibaba Cloud component. Added for consistency.
- **Fields excluded for v1**: `payment_type`/`period`/`pricing_cycle` (Subscription EIPs cannot be deleted via API -- significant footgun), `deletion_protection` (cross-cutting decision), `high_definition_monitor_log_status`/`log_project`/`log_store` (niche), `mode` (set by association), `zone`/`ip_address`/`netmode`/`security_protection_types`/`activity_id`/`auto_pay`/`allocation_id`/`public_ip_address_pool_id` (niche or internal).

## Verification

- `go build ./...` -- PASS
- `go vet ./...` -- PASS
- `go test ./...` -- PASS (15/15 specs)
- `terraform init` -- PASS (alicloud provider v1.271.0)
- `terraform validate` -- PASS
