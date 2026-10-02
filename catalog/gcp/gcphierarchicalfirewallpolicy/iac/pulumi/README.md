# GcpHierarchicalFirewallPolicy - Pulumi Module

This Pulumi module provisions one hierarchical firewall policy (`compute.FirewallPolicy`), one rule per `spec.rules` entry (`compute.FirewallPolicyRule`), and one association per `spec.associations` entry (`compute.FirewallPolicyAssociation`). It is the Pulumi-side implementation of the Planton `GcpHierarchicalFirewallPolicy` resource kind and has feature parity with the Terraform module.

## Overview

The policy lives on the organization or a folder (`spec.parent`, exactly one arm, rendered to Google's `organizations/{id}` or `folders/{id}` string in `locals.go`) and is enforced wherever `spec.associations` attach it. Rules are keyed by priority (the Pulumi resource name carries it): a renumbered rule is replaced, every other rule is left alone — the provider's own semantics. Every catalog-producible target is a reference resolved to its handle before the module runs: folders, VPC networks, tag values, service accounts. No project is involved: hierarchical policies are organization resources.

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --stack org/project/stack
planton pulumi destroy --manifest ../../e2e/manifest.yaml --stack org/project/stack
```

Credentials are provided via stack input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Module Layout

| File | Purpose |
|------|---------|
| `main.go` | Provider setup, then the policy, its rules, its associations |
| `locals.go` | Short name default, parent rendering, association names and targets |
| `firewall_policy.go` | The policy container and every output |
| `rules.go` | One `FirewallPolicyRule` per entry; the match block; secure-tag and reference-list rendering |
| `associations.go` | One `FirewallPolicyAssociation` per entry |
| `outputs.go` | Output key constants (must match `outputs.proto`) |

## Outputs

| Name | Description |
|------|-------------|
| `policy_id` | The server-assigned numeric ID (Google's `name` for the policy) |
| `short_name` | The policy's short name |
| `self_link` | Self-link URL of the policy |
| `parent` | `organizations/{id}` or `folders/{id}` |
| `rule_tuple_count` | Google's complexity measure for the rule set |
| `association_names` | Association names in declaration order |

## Send Posture

- `description` and `deletion_policy` sent only when set; `deletion_policy` fanned to the policy, every rule, and every association.
- Every rule list (targets, IP ranges, FQDNs, region codes, threat lists, address groups, networks) sent only when non-empty; the two network-context enums sent only when set (Optional+Computed); `security_profile_group` and `tls_inspect` travel together.
- `disabled` and `enable_logging` sent as the spec states them (`false` is the API's own default).
- Secure tags render as `{ name = tagValues/{id} }` blocks; the block's `state` is read-only.

## Notes

- **A rule's priority is its identity.** Changing a priority destroys that rule and creates a new one; change a rule's content in place instead.
- **A node carries one hierarchical policy association at a time.**
- **Google's implied rules** at priorities 2147483646 and 2147483647 (`goto_next`) are always present; the spec caps user priorities below them.
- **Free.** Firewall policies carry no meter; rule logging bills as Cloud Logging ingestion.
