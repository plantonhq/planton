# GcpHierarchicalFirewallPolicy - Terraform Module

This Terraform module provisions one hierarchical firewall policy (`google_compute_firewall_policy`), one rule per `spec.rules` entry (`google_compute_firewall_policy_rule`), and one association per `spec.associations` entry (`google_compute_firewall_policy_association`). It is the Terraform-side implementation of the Planton `GcpHierarchicalFirewallPolicy` resource kind and has feature parity with the Pulumi module.

## Overview

The policy lives on the organization or a folder (`spec.parent`, exactly one arm, rendered to Google's `organizations/{id}` or `folders/{id}` string) and is enforced wherever `spec.associations` attach it. Rules are keyed by priority: a renumbered rule is replaced, every other rule is left alone — the provider's own semantics. Every catalog-producible target is a reference flattened to its handle: folders, VPC networks, tag values, service accounts. The module runs on the plain `google` provider — every modeled field is GA on the pinned 8.x line. No project is involved: hierarchical policies are organization resources.

## Usage with Planton CLI

```shell
planton tofu init --manifest ../../e2e/manifest.yaml
planton tofu plan --manifest ../../e2e/manifest.yaml
planton tofu apply --manifest ../../e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest ../../e2e/manifest.yaml --auto-approve
```

Credentials are provided via stack input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Direct Terraform Usage

```bash
cd catalog/gcp/gcphierarchicalfirewallpolicy/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpHierarchicalFirewallPolicy spec | — |

The `spec` object includes: `parent` (`organization_id` or `folder_id`), `short_name` (empty defaults to `metadata.name`), `description`, `rules` (each with `priority`, `action`, `direction`, `match`, targets, logging, security profile group), `associations` (each with `name` — empty defaults to `<short_name>-<n>` — and a `target` with `organization_id` or `folder_id`), and `deletion_policy`.

`variables.tf` is generated from the proto contract (`planton tofu generate-variables GcpHierarchicalFirewallPolicy`) and formatted with `tofu fmt`; regenerate it when the spec changes rather than editing by hand.

## Outputs

| Name | Description |
|------|-------------|
| `policy_id` | The server-assigned numeric ID (Google's `name` for the policy) |
| `short_name` | The policy's short name |
| `self_link` | Self-link URL of the policy |
| `parent` | `organizations/{id}` or `folders/{id}` |
| `rule_tuple_count` | Google's complexity measure for the rule set |
| `association_names` | Association names in declaration order |

## Resources Created

- `google_compute_firewall_policy` — `parent` and `short_name` immutable; `description` and `deletion_policy` sent only when set
- `google_compute_firewall_policy_rule` — `for_each` keyed by priority; every list sent only when non-empty; the two network-context enums sent only when set (Optional+Computed); `security_profile_group` and `tls_inspect` travel together; `deletion_policy` fanned
- `google_compute_firewall_policy_association` — `for_each` keyed by name; `attachment_target` rendered from the target message; `deletion_policy` fanned

## Notes

- **A rule's priority is its identity.** Changing a priority destroys that rule and creates a new one; change a rule's content in place instead. Leave gaps (1000, 2000, ...) so rules can be slotted in later.
- **A node carries one hierarchical policy association at a time.** Associating a second policy with the same folder fails until the first is detached.
- **Google's implied rules** at priorities 2147483646 and 2147483647 (`goto_next`) are always present; the spec caps user priorities below them.
- **Free.** Firewall policies carry no meter; rule logging bills as Cloud Logging ingestion.
