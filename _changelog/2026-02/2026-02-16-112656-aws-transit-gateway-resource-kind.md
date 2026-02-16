# AWS Transit Gateway Resource Kind

**Date**: February 16, 2026
**Type**: Feature
**Components**: API Definitions, Pulumi CLI Integration, AWS Provider, Protobuf Schemas

## Summary

Added AwsTransitGateway (R25) as a new cloud resource kind in OpenMCF, providing a complete deployment component for AWS Transit Gateway with bundled VPC attachments. This is the final Phase 2 component in the AWS resource expansion project, bringing the total to 29 AWS resource kinds.

## Problem Statement / Motivation

Organizations with multiple VPCs need a centralized networking hub to replace complex VPC peering meshes. AWS Transit Gateway is the standard solution, but configuring it manually or with raw Terraform involves multiple interdependent resources (Transit Gateway, VPC attachments, route tables, propagations) that are tedious to manage.

### Pain Points

- VPC peering scales as N*(N-1)/2 connections -- impractical beyond 5-10 VPCs
- Transit Gateway configuration requires coordinating multiple AWS resources with correct enable/disable toggles
- Appliance mode for firewall inspection VPCs is easy to misconfigure, leading to asymmetric routing
- No existing OpenMCF abstraction for multi-VPC hub networking

## Solution / What's New

A complete AwsTransitGateway deployment component that bundles the Transit Gateway with inline VPC attachments, following the same pattern as AwsNetworkLoadBalancer (bundled listeners) and AwsSnsTopic (bundled subscriptions).

### Key Design Decisions

- **Bundled VPC attachments**: VPC attachments are `repeated` fields in the spec rather than separate components, because a TGW without attachments is useless and the 80% use case is a single team managing both
- **Boolean feature toggles**: AWS uses "enable"/"disable" strings; the proto uses clean bools with `recommended_default` annotations, and the IaC modules convert internally
- **80/20 scoping**: Excludes custom route tables, static routes, cross-region peering, and multicast domains from v1 -- these are specialized features that can be added as separate components later
- **Rich outputs**: Exports route table IDs so future components can add static routes without modifying the TGW itself

## Implementation Details

### Proto API (4 files)

- `spec.proto` -- 11 spec fields + nested `AwsTransitGatewayVpcAttachment` message with 8 fields
- `stack_outputs.proto` -- 6 outputs including `vpc_attachment_ids` map
- `api.proto` -- KRM envelope with `aws.openmcf.org/v1` API version
- `stack_input.proto` -- Standard AWS stack input with provider config

### Validations

- CEL: ASN range (16-bit: 64512-65534, 32-bit: 4200000000-4294967294)
- CEL: CIDR blocks max 5 (AWS hard limit)
- CEL: Attachment name format (lowercase alphanumeric + hyphens)
- buf.validate: Required fields (vpc_attachments, name, vpc_id, subnet_ids)
- 24 spec tests covering happy paths and failure cases

### Pulumi Module (5 files)

- `main.go` -- Entry point orchestrating TGW + attachments + outputs
- `locals.go` -- Tag initialization and `enableDisable()` bool-to-string converter
- `outputs.go` -- Output key constants
- `transit_gateway.go` -- Creates `ec2transitgateway.TransitGateway` with all feature toggles
- `vpc_attachment.go` -- Iterates `spec.vpc_attachments`, creates one `ec2transitgateway.VpcAttachment` per entry with `DependsOn`

### Terraform Module (5 files)

- `main.tf` -- Transit Gateway + VPC attachments via `for_each`
- `variables.tf` -- Typed input variables matching the proto spec
- `locals.tf` -- Tag merging and enable/disable lookup map
- `outputs.tf` -- All 6 outputs matching stack_outputs.proto
- `provider.tf` -- AWS provider with credential passthrough

### Registration

- Enum `AwsTransitGateway = 282` with `id_prefix: "awstgw"` in `cloud_resource_kind.proto`

## Benefits

- **Single resource** replaces 3+ manually coordinated Terraform resources
- **Full-mesh by default** -- attachments auto-associate and auto-propagate for immediate connectivity
- **Firewall-ready** -- appliance mode support for centralized inspection patterns
- **Future-proof** -- exported route table IDs enable adding static routes without touching the TGW

## Impact

- **AWS resource coverage**: 29 resource kinds (28 previously completed + Transit Gateway)
- **Phase 2 completion**: All 10 Phase 2 (Important Services) components are now done
- **Infra charts**: Enables multi-VPC hub patterns in future infra chart compositions

## Related Work

- Part of `20260215.02.sp.aws-resource-expansion` (parent: `20260212.01.openmcf-cloud-provider-expansion`)
- Phase 3 (7 specialized components) and existing component fixes (7) remain
- Reference patterns: AwsNetworkLoadBalancer (bundled listeners), AwsElasticIp (simple networking)

---

**Status**: Production Ready
