# GcpFirewallRule Deployment Component

**Date**: February 15, 2026
**Type**: Feature
**Components**: GCP Provider, API Definitions, Pulumi Module, Terraform Module

## Summary

Added the `GcpFirewallRule` deployment component to Planton, enabling declarative provisioning of GCP compute firewall rules. This is the first resource in the GCP expansion effort, bringing the total GCP resource count from 19 to 20. The component features a clean `action` + `rules` abstraction over Terraform's raw `allow`/`deny` blocks, cross-field CEL validation for GCP's tag/service-account mutual exclusion constraint, and full infra-chart composability via `StringValueOrRef` on project and network fields.

## Problem Statement / Motivation

GCP VPC firewall rules are a foundational networking primitive required by virtually every infra chart that provisions GCP resources. Without a `GcpFirewallRule` component, infra charts for GKE environments, Cloud Run backends, Spanner applications, and serverless API backends cannot express network security rules declaratively.

### Pain Points

- No way to provision GCP firewall rules through Planton
- Infra charts needing network security had to rely on external processes
- Missing a Layer 1 networking building block that sits between VPC and higher-level services

## Solution / What's New

A complete deployment component following the ideal state checklist:

### Proto API (4 files)

- `spec.proto` with 16 fields covering all 80/20 firewall configuration
- `stack_outputs.proto` with `firewall_self_link`, `firewall_name`, `creation_timestamp`
- `api.proto` with KRM envelope (apiVersion, kind, metadata, spec, status)
- `stack_input.proto` wiring target + GcpProviderConfig

### Key Design Decisions

1. **`action` + `rules` abstraction**: Instead of exposing Terraform's `allow`/`deny` blocks directly, the spec uses a single `action` field ("ALLOW"/"DENY") paired with a `rules` repeated field of `GcpFirewallProtocolPort`. The Pulumi module maps this to the correct `FirewallAllowArray` or `FirewallDenyArray`. This is cleaner from a user perspective while maintaining full fidelity with the GCP API.

2. **Cross-field CEL validation**: Two message-level CEL rules enforce GCP constraints at the schema level:
   - INGRESS rules must specify at least one source field
   - Tag-based and service-account-based targeting are mutually exclusive

3. **`log_config` as structured message**: Instead of a boolean, logging configuration uses a `GcpFirewallLogConfig` message with a required `metadata` field, matching the modern GCP API (the deprecated `enable_logging` boolean is not exposed).

4. **`StringValueOrRef` on `project_id` and `network`**: Enables infra-chart composability -- firewall rules can reference a GcpProject's project_id and a GcpVpc's network_self_link via `valueFrom`.

## Implementation Details

### Pulumi Module (`iac/pulumi/module/`)

```
module/
  main.go        # Resources() entry point
  locals.go      # Locals struct with labels
  outputs.go     # Output key constants
  firewall.go    # compute.NewFirewall with action->allow/deny mapping
```

The `firewall.go` file contains the core mapping logic:

```go
switch spec.Action {
case "ALLOW":
    args.Allows = mapToAllowRules(spec.Rules)
case "DENY":
    args.Denies = mapToDenyRules(spec.Rules)
}
```

GCP compute firewall rules do not support labels, so while `GcpLabels` are computed in `locals.go` for consistency, they are not applied to the resource.

### Terraform Module (`iac/tf/`)

Uses `dynamic` blocks to conditionally create `allow` or `deny` blocks based on the `action` variable, mirroring the Pulumi logic.

### Validation Tests

17 Ginkgo/Gomega tests covering:
- 6 positive cases (valid INGRESS/EGRESS, tags, service accounts, logging, disabled)
- 11 negative cases (missing fields, invalid direction/action, INGRESS without source, tag/SA mixing, priority bounds, service account limits, invalid log_config)

## Benefits

- GCP users can now provision firewall rules declaratively through Planton
- Infra charts can compose firewall rules with VPCs and other networking resources
- Schema-level validation catches misconfigurations before deployment
- Clean abstraction reduces cognitive load vs raw Terraform/Pulumi

## Impact

- **New resource kind**: GcpFirewallRule (enum 620, id_prefix `gcpfwr`)
- **Files created**: ~40 files (protos, stubs, Go modules, TF modules, docs, presets)
- **Test coverage**: 17 validation tests, all passing
- **Build status**: Go build clean, Terraform validate clean

## Related Work

- Part of the GCP resource expansion project (20260215.01.sp.gcp-resource-expansion)
- First of 21 new GCP resources planned
- Next in queue: GcpGlobalAddress (R02)

---

**Status**: Production Ready
**Timeline**: Single session
