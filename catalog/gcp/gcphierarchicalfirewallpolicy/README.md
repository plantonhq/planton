# GCP Hierarchical Firewall Policy

Creates a hierarchical firewall policy — an ordered set of firewall rules that lives on the organization or a folder and is enforced on every VPC network beneath the node it is associated with, in every project, present and future, before any network-level policy or legacy VPC rule gets a say. It is how a platform team states the rules nobody in the organization may undo ("no SSH from the internet", "deny known-malicious sources") and delegates everything else downward with `goto_next`. Rules key on the modern matching surface: secure tags, address groups, domain names, threat-intelligence lists, geography, and Cloud NGFW security profile groups.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Firewall policy** -- the `compute_firewall_policy` container on the organization or folder, named by Google with a numeric ID
- **Rules** -- one `compute_firewall_policy_rule` per `rules` entry, keyed by priority
- **Associations** -- one `compute_firewall_policy_association` per `associations` entry, attaching the policy to the organization or a folder

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module whose identity holds `roles/compute.orgFirewallPolicyAdmin` on the organization (or on the folder the policy lives under and every folder it is associated with).
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Organization

- **A Google Cloud Organization** -- hierarchical policies do not exist in a project-only account. Have the numeric organization ID (`gcloud organizations list`) or declare the folder with `GcpFolder` and reference its `folder_id` output.
- **Secure tags** for tag-keyed rules are `GcpTagValue` resources whose tag key has `purpose: GCE_FIREWALL`; reference their `name` output.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpHierarchicalFirewallPolicy
metadata:
  name: org-baseline
spec:
  parent:
    organizationId: "123456789012"
  rules:
    - priority: 1000
      action: deny
      direction: INGRESS
      enableLogging: true
      match:
        srcIpRanges: ["0.0.0.0/0"]
        layer4Configs:
          - ipProtocol: tcp
            ports: ["22"]
    - priority: 2000
      action: goto_next
      direction: INGRESS
      match:
        layer4Configs:
          - ipProtocol: all
  associations:
    - target:
        organizationId: "123456789012"
```

```shell
planton apply -f org-baseline.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `parent` | `message` | Exactly one of `organizationId` (numeric) or `folderId` (`GcpFolder` reference or numeric literal). Immutable. |
| `rules[].priority` | `int32` | 0 (first) to 2147483645 (last); unique; the rule's identity -- changing it recreates the rule. |
| `rules[].action` | `string` | `allow`, `deny`, `goto_next`, or `apply_security_profile_group`. |
| `rules[].direction` | `string` | `INGRESS` or `EGRESS`. |
| `rules[].match.layer4Configs` | `list` | At least one `{ipProtocol, ports}`; `ports` only for tcp/udp. |
| `associations[].target` | `message` | Exactly one of `organizationId` or `folderId` (`GcpFolder` reference). Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `shortName` | `string` | `metadata.name` | RFC 1035, unique in the organization. Immutable. |
| `description` | `string` | — | Mutable. |
| `rules[].description` | `string` | — | Mutable. |
| `rules[].disabled` | `bool` | `false` | Keep the rule, stop enforcing it. |
| `rules[].enableLogging` | `bool` | `false` | Log decided connections; not on `goto_next`. |
| `rules[].match.srcIpRanges` / `destIpRanges` | `list` | — | CIDRs, max 5000 each. |
| `rules[].match.srcAddressGroups` / `destAddressGroups` | `list` | — | Address-group names, max 10 each. |
| `rules[].match.srcFqdns` / `destFqdns` | `list` | — | Domain names, max 100 each. |
| `rules[].match.srcRegionCodes` / `destRegionCodes` | `list` | — | ISO 3166-1 alpha-2, max 5000 each. |
| `rules[].match.srcThreatIntelligences` / `destThreatIntelligences` | `list` | — | Google Threat Intelligence list names. |
| `rules[].match.srcSecureTags` | `list` | — | `GcpTagValue` references (ingress only), max 256. |
| `rules[].match.srcNetworks` | `list` | — | `GcpVpcNetwork` references (ingress only). |
| `rules[].match.srcNetworkContext` / `destNetworkContext` | `string` | unset | `INTERNET`, `INTRA_VPC`, `NON_INTERNET`, `VPC_NETWORKS`, `UNSPECIFIED`. |
| `rules[].targetResources` | `list` | all networks | `GcpVpcNetwork` references the rule applies to. |
| `rules[].targetSecureTags` | `list` | — | `GcpTagValue` references, max 256; not with `targetServiceAccounts`. |
| `rules[].targetServiceAccounts` | `list` | — | `GcpServiceAccount` references; not with `targetSecureTags`. |
| `rules[].securityProfileGroup` | `string` | — | Required for and only for `apply_security_profile_group`. |
| `rules[].tlsInspect` | `bool` | `false` | Only with `apply_security_profile_group`. |
| `associations[].name` | `string` | `<shortName>-<n>` | RFC 1035, unique in the policy. Immutable. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON` for the policy, every rule, every association. |

### Validation Rules

- Exactly one arm on `parent` and on every association `target`.
- Rule priorities unique; association names unique.
- `securityProfileGroup` present exactly when `action` is `apply_security_profile_group`; `tlsInspect` only then; no `enableLogging` on `goto_next`.
- `targetSecureTags` and `targetServiceAccounts` are mutually exclusive.
- `ports` only with `tcp`, `udp`, `6`, or `17`; `ipProtocol` is a well-known name, `all`, or a number 0-255.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `policy_id` | `string` | The server-assigned numeric ID (Google's `name`) |
| `short_name` | `string` | The short name |
| `self_link` | `string` | Self-link URL |
| `parent` | `string` | `organizations/{id}` or `folders/{id}` |
| `rule_tuple_count` | `int64` | Google's complexity measure for the rule set |
| `association_names` | `list` | Association names in declaration order |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **A rule's priority is its identity.** Renumbering recreates the rule; change content in place instead. Leave gaps between priorities.
- **A node carries one hierarchical policy at a time.** Associating a second policy with the same folder fails until the first is detached.
- **`goto_next` is the delegation lever.** Without it, a matching `allow` or `deny` at the organization ends the decision for every project beneath.
- **Free.** Firewall policies have no meter; rule logging bills as Cloud Logging ingestion.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- [GcpNetworkFirewallPolicy](/docs/catalog/gcp/gcpnetworkfirewallpolicy) — the project-level policy attached to VPC networks
- [GcpFirewallRule](/docs/catalog/gcp/gcpfirewallrule) — legacy per-network VPC firewall rules, evaluated after every policy
- [GcpFolder](/docs/catalog/gcp/gcpfolder) — the folder a policy lives under or is enforced on
- [GcpTagValue](/docs/catalog/gcp/gcptagvalue) — the secure tags rules key on
- [GcpVpcNetwork](/docs/catalog/gcp/gcpvpcnetwork) — the networks `targetResources` and `srcNetworks` reference

## Additional Resources

- [Hierarchical firewall policies](https://cloud.google.com/firewall/docs/firewall-policies)
- [Firewall policy rule components](https://cloud.google.com/firewall/docs/firewall-policies-rule-details)
- [Policy and rule evaluation order](https://cloud.google.com/firewall/docs/firewall-policies-overview#rule-evaluation)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
