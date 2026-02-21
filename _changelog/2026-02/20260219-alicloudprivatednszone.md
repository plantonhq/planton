# AlicloudPrivateDnsZone Component Added

**Date**: 2026-02-19
**Component**: AlicloudPrivateDnsZone
**Enum**: 3042
**ID Prefix**: acpz

## Summary

Added the AlicloudPrivateDnsZone deployment component -- a composite resource that bundles a Private Zone (pvtz_zone), VPC attachment(s) (pvtz_zone_attachment), and DNS records (pvtz_zone_record) into a single deployable unit for VPC-internal DNS resolution.

Private zones resolve domain names only within attached VPCs. They are invisible to the public internet, making them ideal for service discovery, database endpoint management, and split-horizon DNS.

## What Was Created

### API Definition
- `apis/org/openmcf/provider/alicloud/alicloudprivatednszone/v1/` -- Full proto API (spec, api, stack_input, stack_outputs)
- Registered `AlicloudPrivateDnsZone = 3042` in `CloudResourceKind` enum under the DNS category
- Nested messages: `AlicloudPrivateDnsZoneVpcAttachment` (with StringValueOrRef for vpc_id), `AlicloudPrivateDnsZoneRecord` (with CEL validation for record types)

### IaC Modules
- **Pulumi** (Go): Creates pvtz.Zone, pvtz.ZoneAttachment (with cross-region VPC support), and pvtz.ZoneRecord for each spec.records entry. Records are parented to the zone for clean dependency tracking.
- **Terraform** (HCL): alicloud_pvtz_zone + alicloud_pvtz_zone_attachment (dynamic blocks for VPC list) + alicloud_pvtz_zone_record (for_each). Tag merging in locals.tf.

### Tests
- Ginkgo/Gomega spec validation tests: 18 specs covering valid inputs (minimal, full config, multi-VPC, cross-region, MX with priority, all record types, records with remarks) and invalid inputs (missing region, missing zone_name, empty vpc_attachments, missing vpc_id, invalid record type, empty rr, empty value, wrong api_version, wrong kind, missing metadata, missing spec, zone_name too long)

### Documentation
- README.md with configuration reference tables and related components
- examples.md with 3 YAML examples (minimal, service discovery with records, multi-VPC database zone)
- catalog-page.md with detailed user-facing documentation
- docs/README.md with comprehensive research documentation
- Pulumi and Terraform READMEs, overview.md, examples

### Presets
- 01-internal-service-discovery -- single VPC, A records for service endpoints
- 02-multi-vpc-database-zone -- multi-VPC (cross-region), database endpoint discovery with governance

## Spec Design Decisions

- **Multiple VPC attachments** (divergence from T02): T02 specified a single `vpc_id`, but the provider supports multiple VPC attachments including cross-region. Private zones are commonly shared across VPCs (app + management + monitoring). `repeated vpc_attachments` is the correct design.
- **Record types restricted**: Only A, CNAME, MX, PTR, SRV, TXT are supported by Private Zone (unlike public Alidns which also supports AAAA, NS, CAA). CEL validation enforces this.
- **Omitted fields**: proxy_pattern, sync_status, user_info (cross-account sync), record status (ENABLE/DISABLE). These are niche features that add complexity without broad utility for v1.
- **Tags on zone**: Consistent with all other Alibaba Cloud components.
- **resource_group_id**: Per DD05, plain string (not StringValueOrRef).

## Verification

- `go build ./...` -- PASS
- `go vet ./...` -- PASS
- `go test ./...` -- PASS (18/18 specs)
- `terraform init` -- PASS (alicloud provider v1.271.0)
- `terraform validate` -- PASS
