# AliCloudNatGateway

**Date**: 2026-02-19
**Type**: New Component
**Resource Kind**: AliCloudNatGateway (enum 3024, id_prefix: acnat)

## Summary

Added AliCloudNatGateway component that provisions an Alibaba Cloud Enhanced NAT Gateway with bundled EIP association and SNAT entries. This is a composite component (per DD07) that creates a fully functional NAT setup as a single deployable unit.

## What's Included

- **Proto API**: spec.proto with 13 fields, stack_outputs.proto with 4 outputs, api.proto, stack_input.proto
- **Validations**: CEL validations for nat_type, payment_type, internet_charge_type, specification; length constraints on nat_gateway_name
- **Tests**: spec_test.go with 7 valid-input and 10 invalid-input test cases
- **Pulumi Module**: main.go, locals.go, outputs.go, snat_entries.go -- uses ecs.GetEipAddresses data source for EIP IP resolution
- **Terraform Module**: main.tf, snat_entries.tf, variables.tf, outputs.tf, locals.tf, provider.tf -- uses data.alicloud_eip_addresses for EIP IP resolution
- **Documentation**: catalog-page.md, examples.md, README.md, docs/README.md, Pulumi overview.md
- **Presets**: 3 presets (single-vswitch, multi-az-production, cidr-based-snat)
- **Registration**: Enum 3024 in cloud_resource_kind.proto

## Design Decisions

- EIP IP address resolved via data source lookup (not a separate user-facing field)
- Single EIP support for v1 (80/20 rule; multi-EIP can be added in v2)
- Enhanced NAT type as default; Normal exposed as optional for legacy compatibility
- PayByLcu billing as default (modern capacity-unit billing)
- Both source_vswitch_id (StringValueOrRef) and source_cidr (string) supported for SNAT entries

## Dependencies

- AliCloudVpc (vpc_id)
- AliCloudVswitch (vswitch_id, source_vswitch_id in SNAT entries)
- AliCloudEipAddress (eip_id)
