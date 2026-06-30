# ScalewayInstanceSecurityGroup Resource Kind (R04)

**Date**: February 13, 2026
**Type**: Feature
**Components**: API Definitions, Protobuf Schemas, Pulumi IaC Module, Terraform IaC Module, Resource Management

## Summary

Implemented ScalewayInstanceSecurityGroup as the fourth Scaleway resource kind in Planton. This is a standalone, non-composite, zonal firewall resource that wraps `scaleway_instance_security_group`. It introduces Scaleway's default-policy-based firewall model (allowlist vs denylist) and provides the `security_group_id` output that downstream `ScalewayInstance` resources will reference.

## Problem Statement / Motivation

The Scaleway cloud provider expansion requires firewall capabilities before Instance resources can be implemented. In Scaleway's architecture, security groups are assigned at the Instance level (not at the network level), so they must exist before Instances can be created.

### Pain Points

- No firewall resource existed for Scaleway in Planton
- ScalewayInstance (R06) depends on `security_group_id` -- this is a prerequisite
- The kapsule-environment infra chart needs security group support for worker node firewalling

## Solution / What's New

A complete ScalewayInstanceSecurityGroup resource kind with:

- **Proto schemas**: spec with `ScalewaySecurityGroupInboundRule`/`OutboundRule` messages, api, stack_input, stack_outputs
- **Pulumi Go module**: Maps proto rules to `instance.SecurityGroupInboundRuleArgs`/`OutboundRuleArgs`, creates `instance.SecurityGroup`
- **Terraform HCL module**: Uses `dynamic` blocks for inline inbound/outbound rules
- **Documentation**: README.md with configuration reference, security best practices, and infra chart integration guide; examples.md with 8 real-world patterns
- **Validation tests**: Ginkgo/Gomega protovalidate tests covering valid and invalid inputs

### Key Design Decisions

1. **Not a composite resource** -- Wraps exactly one Terraform resource. Rules are inline on the security group, not separate resources.
2. **No StringValueOrRef inputs** -- Standalone resource with zero upstream dependencies. Referenced downstream by ScalewayInstance.
3. **Scaleway-native terminology** -- Actions are "accept"/"drop" (not "allow"/"deny"). Protocols are uppercase ("TCP", "UDP", "ICMP", "ANY").
4. **Unified `port_range` field** -- Single string field accepting both "80" and "22-23", following the CivoFirewall precedent. Avoids the API's confusing `port` vs `port_range` split.
5. **Default policy model exposed** -- `inbound_default_policy` and `outbound_default_policy` let users choose between allowlist ("drop" default) and denylist ("accept" default) models.
6. **Top-level rule messages** -- `ScalewaySecurityGroupInboundRule` and `ScalewaySecurityGroupOutboundRule` are defined at the package level (not nested inside spec) for cleaner code generation.

## Implementation Details

### Files Created (18 files)

**Proto schemas (4)**:
- `apis/dev/planton/provider/scaleway/scalewayinstancesecuritygroup/v1/spec.proto`
- `apis/dev/planton/provider/scaleway/scalewayinstancesecuritygroup/v1/api.proto`
- `apis/dev/planton/provider/scaleway/scalewayinstancesecuritygroup/v1/stack_input.proto`
- `apis/dev/planton/provider/scaleway/scalewayinstancesecuritygroup/v1/stack_outputs.proto`

**Pulumi Go module (7)**:
- `apis/.../iac/pulumi/main.go` -- Entry point
- `apis/.../iac/pulumi/module/main.go` -- Resources() orchestrator
- `apis/.../iac/pulumi/module/locals.go` -- Locals struct + tag generation
- `apis/.../iac/pulumi/module/security_group.go` -- SecurityGroup provisioner
- `apis/.../iac/pulumi/module/outputs.go` -- Output constants
- `apis/.../iac/pulumi/Pulumi.yaml` -- Runtime config
- `apis/.../iac/pulumi/Makefile` -- Dev targets

**Terraform HCL module (5)**:
- `apis/.../iac/tf/main.tf` -- Security group with dynamic rule blocks
- `apis/.../iac/tf/variables.tf` -- Input variables
- `apis/.../iac/tf/locals.tf` -- Local values and tags
- `apis/.../iac/tf/outputs.tf` -- Output definitions
- `apis/.../iac/tf/provider.tf` -- Provider configuration (zone-based)

**Documentation + Tests (2)**:
- `apis/.../README.md` -- Comprehensive documentation
- `apis/.../examples.md` -- 8 real-world usage patterns
- `apis/.../spec_test.go` -- Protovalidate tests

### Verification Results

- `make protos` -- Zero warnings
- `make generate-cloud-resource-kind-map` -- ScalewayInstanceSecurityGroup registered (15 remaining skip)
- `go build` -- Clean
- `go vet` -- Clean
- `go test` -- All tests pass
- `terraform validate` -- "Success! The configuration is valid."

## Benefits

- Enables the Instance resource (R06) to reference security groups via StringValueOrRef
- Supports both allowlist and denylist firewall models via default policies
- Preserves Scaleway-native semantics (actions, protocols, stateful mode, SMTP security)
- Ready for infra chart composition in kapsule-environment and instance-based charts

## Impact

- **Resource count**: 4 of 19 Scaleway resource kinds now implemented
- **Dependency chain**: Unblocks ScalewayInstance (R06) which depends on security_group_id
- **Infra chart readiness**: Layer 0/1 standalone resource, composable via valueFrom references

## Related Work

- R01: ScalewayVpc (foundation, regional)
- R02: ScalewayPrivateNetwork (first StringValueOrRef, regional)
- R03: ScalewayPublicGateway (first composite, zonal)
- R04: ScalewayInstanceSecurityGroup (this change -- standalone, zonal)
- Next: R05 ScalewayLoadBalancer (complex composite)

---

**Status**: Production Ready
**Timeline**: Single session implementation
