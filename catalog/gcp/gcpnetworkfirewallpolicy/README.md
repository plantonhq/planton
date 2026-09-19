# GCP Network Firewall Policy

Creates a network firewall policy — an ordered set of firewall rules owned by a project and associated with one or more of its VPC networks, either globally (all regions) or for one region. It is Google's replacement for legacy per-network firewall rules: one policy holds the whole rule set, attaches to several networks, and carries the matching surface legacy rules lack — secure tags, address groups, domain names, threat-intelligence lists, geography, and Cloud NGFW security profile groups. One kind covers both scopes: leave `region` empty for a global policy, set it for a regional one.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Firewall policy** -- the `compute_network_firewall_policy` (global) or `compute_region_network_firewall_policy` (regional), chosen by `region`
- **Rules** -- one `..._rule` per `rules` entry, keyed by priority
- **Associations** -- one `..._association` per `associations` entry, attaching the policy to a VPC network
- **API enablement** -- `compute.googleapis.com`, never disabled on destroy

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with `roles/compute.securityAdmin` on the project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Network

- **The VPC must exist** in the same project -- declare it with `GcpVpcNetwork` and reference its `network_self_link` output, or pass the self link as a literal.
- **Secure tags** for tag-keyed rules are `GcpTagValue` resources under a `GcpTagKey` with `purpose: GCE_FIREWALL`; reference the value's `name` output.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpNetworkFirewallPolicy
metadata:
  name: baseline
spec:
  rules:
    - priority: 1000
      action: allow
      direction: INGRESS
      ruleName: allow-iap-ssh
      match:
        srcIpRanges: ["35.235.240.0/20"]
        layer4Configs:
          - ipProtocol: tcp
            ports: ["22"]
    - priority: 65000
      action: deny
      direction: INGRESS
      enableLogging: true
      match:
        srcIpRanges: ["0.0.0.0/0"]
        layer4Configs:
          - ipProtocol: all
  associations:
    - network:
        value: projects/acme-net/global/networks/main
```

```shell
planton apply -f baseline.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `rules[].priority` | `int32` | 0 (first) to 2147483647 (last); unique; the rule's identity -- changing it recreates the rule. |
| `rules[].action` | `string` | `allow`, `deny`, `goto_next`, or `apply_security_profile_group`. |
| `rules[].direction` | `string` | `INGRESS` or `EGRESS`. |
| `rules[].match.layer4Configs` | `list` | At least one `{ipProtocol, ports}`; `ports` only for tcp/udp. |
| `associations[].network` | `StringValueOrRef` | The VPC, as a `GcpVpcNetwork` reference or its self link. Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The project, as a `GcpProject` reference or a literal. Immutable. |
| `policyName` | `string` | `metadata.name` | RFC 1035. Immutable. |
| `region` | `string` | global | A region for a regional policy. Immutable. |
| `description` | `string` | — | Mutable. |
| `policyType` | `string` | `VPC_POLICY` | `VPC_POLICY`; regional only: `RDMA_ROCE_POLICY`, `RDMA_FALCON_POLICY`, `ULL_POLICY`. Sent only when set. Immutable. |
| `rules[].ruleName` | `string` | — | A mutable label; not the rule's identity. |
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
| `rules[].targetType` | `string` | `INSTANCES` | `INSTANCES` or `INTERNAL_MANAGED_LB`. Sent only when set. |
| `rules[].targetForwardingRules` | `list` | — | Forwarding-rule references; required for and only for `INTERNAL_MANAGED_LB`. |
| `rules[].targetSecureTags` | `list` | — | `GcpTagValue` references, max 256. |
| `rules[].targetServiceAccounts` | `list` | — | `GcpServiceAccount` references. |
| `rules[].securityProfileGroup` | `string` | — | Required for and only for `apply_security_profile_group`. |
| `rules[].tlsInspect` | `bool` | `false` | Only with `apply_security_profile_group`. |
| `associations[].name` | `string` | `<policyName>-<n>` | RFC 1035, unique in the policy. Immutable. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON` for the policy, every rule, every association. |

### Validation Rules

- Rule priorities unique; association names unique.
- `policyType` RDMA and ULL values only with `region` set.
- `securityProfileGroup` present exactly when `action` is `apply_security_profile_group`; `tlsInspect` only then; no `enableLogging` on `goto_next`.
- `targetForwardingRules` non-empty exactly when `targetType` is `INTERNAL_MANAGED_LB`.
- `ports` only with `tcp`, `udp`, `6`, or `17`; `ipProtocol` is a well-known name, `all`, or a number 0-255.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `policy_name` | `string` | The policy's name in GCP |
| `policy_id` | `string` | The server-assigned numeric ID |
| `self_link` | `string` | Self-link URL (global or regional form) |
| `region` | `string` | The region, or empty for global |
| `rule_tuple_count` | `int64` | Google's complexity measure for the rule set |
| `association_names` | `list` | Association names in declaration order |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **A rule's priority is its identity.** Renumbering recreates the rule; change content in place instead. Leave gaps between priorities.
- **A network carries one global and, per region, one regional policy at a time.**
- **Hierarchical policies win first.** An organization- or folder-level deny cannot be undone here; a `goto_next` there is what lets this policy decide.
- **Free.** Firewall policies have no meter; rule logging bills as Cloud Logging ingestion.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- [GcpHierarchicalFirewallPolicy](/docs/catalog/gcp/gcphierarchicalfirewallpolicy) — organization- and folder-level policies evaluated before this one
- [GcpFirewallRule](/docs/catalog/gcp/gcpfirewallrule) — legacy per-network VPC firewall rules, evaluated after
- [GcpVpcNetwork](/docs/catalog/gcp/gcpvpcnetwork) — the networks the policy attaches to
- [GcpTagValue](/docs/catalog/gcp/gcptagvalue) — the secure tags rules key on
- [GcpGlobalForwardingRule](/docs/catalog/gcp/gcpglobalforwardingrule) — the load balancers `targetForwardingRules` reference

## Additional Resources

- [Network firewall policies](https://cloud.google.com/firewall/docs/network-firewall-policies)
- [Regional network firewall policies](https://cloud.google.com/firewall/docs/regional-firewall-policies)
- [Firewall policy rule components](https://cloud.google.com/firewall/docs/firewall-policies-rule-details)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
