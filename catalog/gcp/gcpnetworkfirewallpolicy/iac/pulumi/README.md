# GcpNetworkFirewallPolicy - Pulumi Module

This Pulumi module provisions one network firewall policy on one of two provider resource families — global (`compute.NetworkFirewallPolicy` + `Rule` + `Association`) when `spec.region` is empty, regional (`compute.RegionNetworkFirewallPolicy` + `Rule` + `Association`) when it is set — with one rule per `spec.rules` entry and one association per `spec.associations` entry. It is the Pulumi-side implementation of the Planton `GcpNetworkFirewallPolicy` resource kind and has feature parity with the Terraform module.

## Overview

Google models the two scopes as separate resources with identical rule and association shapes, so the spec is one shape and `Resources` picks the family once (the `GcpHealthCheck` grain), mirroring the Terraform module's count guards. Each rule is rendered once into provider-neutral inputs (`rule_inputs.go`) and mapped onto the family's own structs, so the send posture lives in one place. Rules are keyed by priority (the Pulumi resource name carries it): a renumbered rule is replaced, every other rule is left alone. The module enables the Compute Engine API first.

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --stack org/project/stack
planton pulumi destroy --manifest ../../e2e/manifest.yaml --stack org/project/stack
```

Credentials are provided via stack input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Module Layout

| File | Purpose |
|------|---------|
| `main.go` | Provider setup, Compute API enablement, the global/regional switch |
| `locals.go` | Policy name default, `IsRegional`, resolved project, association names |
| `rule_inputs.go` | A rule and its match rendered once into provider-neutral inputs; the send posture |
| `global_policy.go` | The global family: policy, rules, associations, outputs |
| `regional_policy.go` | The regional family: the same plus `region` |
| `outputs.go` | Output key constants (must match `outputs.proto`) |

## Outputs

| Name | Description |
|------|-------------|
| `policy_name` | The policy's name in GCP |
| `policy_id` | The server-assigned numeric ID |
| `self_link` | Self-link URL of the policy (global or regional form) |
| `region` | The region, or empty for a global policy |
| `rule_tuple_count` | Google's complexity measure for the rule set |
| `association_names` | Association names in declaration order |

## Send Posture

- `project`, `policy_type`, `description`, `deletion_policy` sent only when set; `deletion_policy` fanned to the policy, every rule, and every association.
- Every rule list sent only when non-empty (a true nil, never an empty array); `target_type` and the two network-context enums sent only when set (Optional+Computed); `security_profile_group` and `tls_inspect` travel together.
- `disabled` and `enable_logging` sent as the spec states them (`false` is the API's own default).
- Secure tags render as `{ name = tagValues/{id} }` blocks; the block's `state` is read-only.

## Notes

- **A rule's priority is its identity.** Changing a priority destroys that rule and creates a new one; change a rule's content in place instead.
- **A network carries one global and, per region, one regional policy association at a time.**
- **`policy_type` follows the network profile.** RDMA and ULL types exist only on the regional family (spec CEL).
- **Free.** Firewall policies carry no meter; rule logging bills as Cloud Logging ingestion.
