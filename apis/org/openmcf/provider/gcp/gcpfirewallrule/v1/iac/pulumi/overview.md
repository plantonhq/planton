# GCP Firewall Rule Pulumi Module - Architecture Overview

## Purpose

This document explains the internal architecture of the Pulumi module for deploying GCP compute firewall rules. It is intended for contributors, maintainers, and advanced users who want to understand how the module works under the hood.

## Design Philosophy

1. **Action Abstraction**: The user specifies `action: ALLOW` or `action: DENY` alongside a list of `rules` (protocol/port entries). The module maps this to either Pulumi's `Allows` or `Denies` argument — never both on the same resource.
2. **No Labels on Firewall**: GCP compute firewall rules do not support the `labels` field. Labels are computed in `locals.go` for consistency with other OpenMCF GCP components but are not applied to the resource.
3. **No API Enablement**: Unlike the GcpVpc module, this module does **not** create a `google_project_service` resource for `compute.googleapis.com`. The Compute Engine API is assumed to be already enabled (typically by a GcpProject or GcpVpc component).
4. **Single Resource**: The module creates exactly one `google_compute_firewall` resource. No supporting resources are needed.

## Module Structure

```
iac/pulumi/
├── main.go               # Entry point: loads stack input and invokes module
├── Pulumi.yaml            # Project metadata (name, runtime, description)
├── Makefile               # Build automation
├── debug.sh               # Local debugging helper
└── module/
    ├── main.go            # Module coordinator: initializes locals, gets provider, calls firewall()
    ├── firewall.go        # Firewall resource creation + action-to-allow/deny mapping
    ├── locals.go          # Compute labels and extract config from stack input
    └── outputs.go         # Output key constants
```

### Separation of Concerns

- **`main.go`** (entry point): Minimal — loads `GcpFirewallRuleStackInput` from Pulumi config and delegates to `module.Resources()`.
- **`module/main.go`**: Orchestrator — calls `initializeLocals()`, obtains a GCP provider, then calls `firewall()`.
- **`module/locals.go`**: Pure data transformation — extracts the `GcpFirewallRule` target and computes GCP labels (even though they are not applied to the firewall resource).
- **`module/firewall.go`**: Infrastructure logic — builds `compute.FirewallArgs`, maps action to allow/deny blocks, creates the resource, exports outputs.
- **`module/outputs.go`**: Constants for output keys (`firewall_self_link`, `firewall_name`, `creation_timestamp`).

## Data Flow

```
User YAML manifest
    ↓
GcpFirewallRuleStackInput (protobuf)
    ↓
main.go → module.Resources()
    ↓
initializeLocals() → Locals struct
    ↓
pulumigoogleprovider.Get() → GCP Provider
    ↓
firewall() → compute.NewFirewall()
    ↓
ctx.Export() → Stack Outputs
```

## How Action Maps to Allow/Deny Blocks

The core abstraction is in `firewall.go`:

```go
switch spec.Action {
case "ALLOW":
    args.Allows = mapToAllowRules(spec.Rules)
case "DENY":
    args.Denies = mapToDenyRules(spec.Rules)
}
```

Each `GcpFirewallProtocolPort` in the spec's `rules` list is converted to either a `FirewallAllowArgs` or `FirewallDenyArgs`:

| Spec Field | `action: ALLOW` | `action: DENY` |
|------------|-----------------|----------------|
| `rules[].protocol` | `Allows[].Protocol` | `Denies[].Protocol` |
| `rules[].ports` | `Allows[].Ports` | `Denies[].Ports` |

GCP requires that a firewall rule has **either** `allow` blocks **or** `deny` blocks, never both. The `action` field ensures exactly one path is taken.

### Helper Functions

- **`mapToAllowRules()`** — converts `[]*GcpFirewallProtocolPort` to `compute.FirewallAllowArray`
- **`mapToDenyRules()`** — converts `[]*GcpFirewallProtocolPort` to `compute.FirewallDenyArray`
- **`toPulumiStringArray()`** — converts `[]string` to `pulumi.StringArray` (used for ports, CIDR ranges, tags, and service accounts)

## Output Mapping

Three outputs are exported after the firewall rule is created:

| Output Constant | Pulumi Attribute | Description |
|----------------|------------------|-------------|
| `OpFirewallSelfLink` | `createdFirewall.SelfLink` | Full self-link URI (e.g., `projects/…/global/firewalls/…`) |
| `OpFirewallName` | `createdFirewall.Name` | Rule name as it exists in GCP |
| `OpCreationTimestamp` | `createdFirewall.CreationTimestamp` | RFC 3339 creation timestamp |

These outputs populate `status.outputs` on the `GcpFirewallRule` resource and can be referenced by other OpenMCF components.

## Design Decisions

### Why No Labels on the Firewall Resource

GCP compute firewall rules (`google_compute_firewall`) do not support the `labels` argument. Unlike VPCs, subnets, and instances, firewall rules are a label-less resource in the GCP API. The `locals.go` file still computes labels for consistency with the label strategy used across all GCP components, but they are not passed to the resource.

### Why No API Enablement

The GcpVpc module creates a `google_project_service` to enable `compute.googleapis.com`. Firewall rules always exist within a VPC, so by the time you create a firewall rule, the Compute API is already active. Duplicating the API enablement would add an unnecessary dependency and slow down deployments.

### Why action + rules Instead of Separate allow/deny Fields

The protobuf spec uses a single `action` string (`ALLOW` or `DENY`) plus a shared `rules` list, rather than separate `allow` and `deny` repeated fields. This mirrors the GCP console UX where you choose an action once and then define matching rules. It prevents the user from accidentally specifying both allow and deny blocks, which GCP rejects.

### Why priority Has a Default

GCP assigns priority `1000` when unset, and the OpenMCF spec mirrors this with `(org.openmcf.shared.options.default) = "1000"`. Making the default explicit ensures YAML manifests that omit `priority` behave predictably and makes the default visible in documentation and validation.

### Why source_ranges Is Validated for INGRESS

The spec-level CEL validation rule `ingress_requires_source` ensures INGRESS rules always specify at least one of `source_ranges`, `source_tags`, or `source_service_accounts`. Without this, GCP would create a rule matching **no** source, which is almost always a mistake.

## Conditional Field Mapping

The `firewall()` function only sets optional Pulumi arguments when the corresponding spec field is non-empty:

```go
if len(spec.SourceRanges) > 0 {
    args.SourceRanges = toPulumiStringArray(spec.SourceRanges)
}
```

This pattern is applied to: `sourceRanges`, `destinationRanges`, `sourceTags`, `targetTags`, `sourceServiceAccounts`, `targetServiceAccounts`, `description`, and `logConfig`. Unset fields are omitted from the Pulumi resource args, letting GCP apply its own defaults.

## Error Handling

Errors are wrapped with context using `github.com/pkg/errors`:

```go
return errors.Wrap(err, "failed to create firewall rule")
```

The error chain flows as:
1. `firewall()` returns a wrapped error
2. `module.Resources()` wraps again: "failed to create firewall rule"
3. Pulumi CLI displays the full chain with stack trace

## Extensibility

### Adding New Fields

1. Add the field to `spec.proto`
2. Regenerate Go stubs: `make protos`
3. Map the new field in `firewall.go` within the `compute.FirewallArgs` construction
4. Test: `pulumi preview`

### Adding New Outputs

1. Add the field to `stack_outputs.proto`
2. Add a constant in `outputs.go`
3. Add a `ctx.Export()` call in `firewall.go`

## Related

- [Component README](../../README.md) — full API reference
- [Pulumi README](README.md) — deployment instructions
- [Examples](../../examples.md) — comprehensive usage examples
