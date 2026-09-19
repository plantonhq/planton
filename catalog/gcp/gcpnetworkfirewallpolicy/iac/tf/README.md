# GcpNetworkFirewallPolicy - Terraform Module

This Terraform module provisions one network firewall policy on one of two provider resource families — global (`google_compute_network_firewall_policy` + `_rule` + `_association`) when `spec.region` is empty, regional (`google_compute_region_network_firewall_policy` + `_rule` + `_association`) when it is set — with one rule per `spec.rules` entry and one association per `spec.associations` entry. It is the Terraform-side implementation of the Planton `GcpNetworkFirewallPolicy` resource kind and has feature parity with the Pulumi module.

## Overview

Google models the two scopes as separate resources with identical rule and association shapes, so the spec is one shape and a single `is_regional` local count-gates every resource (the `GcpHealthCheck` grain). Rules are keyed by priority: a renumbered rule is replaced, every other rule is left alone — the provider's own semantics. Every catalog-producible target is a reference flattened to its handle: the project, VPC networks, tag values, service accounts, forwarding rules. The module enables the Compute Engine API first and runs on the plain `google` provider — every modeled field is GA on the pinned 8.x line.

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
cd catalog/gcp/gcpnetworkfirewallpolicy/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpNetworkFirewallPolicy spec | — |

The `spec` object includes: `project_id` (empty = the provider's default project), `policy_name` (empty defaults to `metadata.name`), `region` (empty = global), `description`, `policy_type` (sent only when set; RDMA and ULL types regional-only), `rules` (each with `priority`, `action`, `direction`, `match`, `rule_name`, `target_type`, `target_forwarding_rules`, targets, logging, security profile group), `associations` (each with `name` — empty defaults to `<policy_name>-<n>` — and `network`), and `deletion_policy`.

`variables.tf` is generated from the proto contract (`planton tofu generate-variables GcpNetworkFirewallPolicy`) and formatted with `tofu fmt`; regenerate it when the spec changes rather than editing by hand.

## Outputs

| Name | Description |
|------|-------------|
| `policy_name` | The policy's name in GCP |
| `policy_id` | The server-assigned numeric ID |
| `self_link` | Self-link URL of the policy (global or regional form) |
| `region` | The region, or empty for a global policy |
| `rule_tuple_count` | Google's complexity measure for the rule set |
| `association_names` | Association names in declaration order |

Each output selects whichever family was created with `one(concat(...))`.

## Resources Created

- `google_project_service` — enables `compute.googleapis.com`; never disabled on destroy
- `google_compute_network_firewall_policy` / `google_compute_region_network_firewall_policy` — exactly one, by `spec.region`; `name`, `policy_type`, `project`, `region` immutable; `policy_type`, `project`, `description`, `deletion_policy` sent only when set
- `..._rule` — `for_each` keyed by priority on the active family; every list sent only when non-empty; `target_type` and the two network-context enums sent only when set (Optional+Computed); `security_profile_group` and `tls_inspect` travel together; `deletion_policy` fanned
- `..._association` — `for_each` keyed by name on the active family; `attachment_target` is the network; `deletion_policy` fanned

## Notes

- **A rule's priority is its identity.** Changing a priority destroys that rule and creates a new one; change a rule's content in place instead. Leave gaps (1000, 2000, ...) so rules can be slotted in later.
- **A network carries one global and, per region, one regional policy association at a time.** Associating a second fails until the first is detached.
- **`policy_type` follows the network profile.** A policy attaches only to a network whose profile carries its type; `VPC_POLICY` is the ordinary VPC and Google's default when unset.
- **Free.** Firewall policies carry no meter; rule logging bills as Cloud Logging ingestion.
