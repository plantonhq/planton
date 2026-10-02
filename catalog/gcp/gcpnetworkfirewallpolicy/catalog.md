# GCP Network Firewall Policy

Creates a network firewall policy: an ordered set of firewall rules owned by a project and associated with one or more of its VPC networks. It is Google's replacement for legacy per-network firewall rules -- one policy holds the whole rule set, can be attached to several networks, and carries the matching surface the legacy rules lack: secure tags, address groups, domain names, threat-intelligence lists, geography, and layer-7 inspection through a Cloud NGFW security profile group. One kind covers both scopes: leave `region` empty for a GLOBAL policy that governs all of a network's traffic (the right default), or set it for a REGIONAL policy that governs one region's traffic -- the scope an internal managed load balancer's rules must live at.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Firewall policy** -- the `compute_network_firewall_policy` (global) or `compute_region_network_firewall_policy` (regional), chosen by `region`
- **Rules** -- one `..._rule` per `rules` entry, keyed by priority
- **Associations** -- one `..._association` per `associations` entry, attaching the policy to a VPC network in the project
- **API enablement** -- `compute.googleapis.com`, never disabled on destroy

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with `roles/compute.securityAdmin` on the project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.

### GCP Network

- **The VPC must exist** in the same project -- declare it with `GcpVpcNetwork` and reference its `network_self_link` output, or pass the self link as a literal.
- **Secure tags** for tag-keyed rules are `GcpTagValue` resources under a `GcpTagKey` with `purpose: GCE_FIREWALL` bound to the network; reference the value's `name` output.

## Deploy

### Console

Open the deployment store, find **GCP Network Firewall Policy**, and click **Deploy**. The creation wizard walks you through the scope, the rules, and the networks to attach. Start from the **Global Baseline Policy** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpNetworkFirewallPolicy
metadata:
  name: baseline
  org: acme-corp
  env: prod
spec:
  description: Baseline for the main VPC -- IAP SSH and health checks in, everything else denied
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
    - priority: 1100
      action: allow
      direction: INGRESS
      ruleName: allow-health-checks
      match:
        srcIpRanges: ["35.191.0.0/16", "130.211.0.0/22"]
        layer4Configs:
          - ipProtocol: tcp
    - priority: 65000
      action: deny
      direction: INGRESS
      ruleName: deny-all-ingress
      enableLogging: true
      match:
        srcIpRanges: ["0.0.0.0/0"]
        layer4Configs:
          - ipProtocol: all
  associations:
    - name: baseline-main
      network:
        value: projects/acme-net/global/networks/main
```

```shell
planton apply -f baseline.yaml
```

This creates a global policy in the provider's project with three rules -- allow Identity-Aware Proxy SSH, allow Google's health-check probers, deny and log everything else -- attached to the `main` network. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, the policy references its `GcpVpcNetwork` and the `GcpTagValue` blocks the chart declares via ValueFromRef:

```yaml
spec:
  rules:
    - priority: 1000
      action: allow
      direction: INGRESS
      match:
        srcSecureTags:
          - valueFrom:
              kind: GcpTagValue
              name: tier-frontend
              fieldPath: status.outputs.name
        layer4Configs:
          - ipProtocol: tcp
            ports: ["8080"]
      targetSecureTags:
        - valueFrom:
            kind: GcpTagValue
            name: tier-backend
            fieldPath: status.outputs.name
  associations:
    - network:
        valueFrom:
          kind: GcpVpcNetwork
          name: main
          fieldPath: status.outputs.network_self_link
```

The InfraPipeline deploys the network and the tag values, then the policy, and destroys the policy before them -- the order Google requires.

## Key Configuration

These are the most important decisions when configuring a network firewall policy. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Global or regional** -- empty `region` governs every region of the attached networks and is the everyday choice; a region governs only that region's traffic, and is required for `INTERNAL_MANAGED_LB` rules and for the RDMA and ULL policy types. Immutable.

**Priority is identity** -- Google keys rules by priority; renumbering recreates the rule. `ruleName` is a free label. Leave gaps (1000, 2000, ...) so a rule can be slotted in later.

**Where it sits in the order** -- hierarchical policies decide first; this policy's `allow` or `deny` only reaches traffic they passed with `goto_next`. Global network policy, then regional, then legacy VPC rules.

**Direction decides which match fields matter** -- on `INGRESS`, the `src*` fields describe where traffic comes from; on `EGRESS`, the `dest*` fields describe where it goes. `srcSecureTags` and `srcNetworks` are meaningful on ingress rules only.

**Targets narrow the rule** -- `targetSecureTags` or `targetServiceAccounts` restrict which VMs a rule applies to; `targetType: INTERNAL_MANAGED_LB` with `targetForwardingRules` applies it to internal Application Load Balancers instead of VMs.

**Deletion policy** -- `DELETE` (default) detaches, removes, and deletes; `PREVENT` fails the destroy; `ABANDON` leaves everything live and enforcing.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpVpcNetwork** | `associations[].network`, `rules[].match.srcNetworks[]` | `status.outputs.network_self_link` |
| **GcpTagValue** | `rules[].targetSecureTags[]`, `rules[].match.srcSecureTags[]` | `status.outputs.name` |
| **GcpServiceAccount** | `rules[].targetServiceAccounts[]` | `status.outputs.email` |
| **GcpGlobalForwardingRule** | `rules[].targetForwardingRules[]` | `status.outputs.self_link` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `policy_name` | The policy's name | `gcloud compute network-firewall-policies describe` |
| `policy_id` | The server-assigned numeric ID | Firewall log correlation |
| `self_link` | Self-link URL (global or regional form) | Tooling |
| `region` | The region, or empty for global | A consumer confirming scope |
| `rule_tuple_count` | Google's complexity measure | Quota watch before adding wide rules |
| `association_names` | Association names in declaration order | A detach targets one |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Global baseline** -- allow IAP SSH and Google health checks, deny and log everything else, attached to one network. Start from the **Global Baseline Policy** preset.

**Regional policy for an internal load balancer** -- `region` set and a rule with `targetType: INTERNAL_MANAGED_LB` naming the forwarding rule. Start from the **Regional ILB Policy** preset.

**Secure-tag micro-segmentation** -- rules keyed on `srcSecureTags` and `targetSecureTags` from `GcpTagValue` references, so tiers talk only to the tiers they should. Start from the **Secure Tag Microsegmentation** preset.

## Works With

- [**GCP Hierarchical Firewall Policy**](/cloud-catalog/gcp-hierarchical-firewall-policy) -- organization- and folder-level policies evaluated before this one
- [**GCP VPC Network**](/cloud-catalog/gcp-vpc-network) -- the networks the policy attaches to
- [**GCP Tag Value**](/cloud-catalog/gcp-tag-value) -- the secure tags rules key on
- [**GCP Global Forwarding Rule**](/cloud-catalog/gcp-global-forwarding-rule) -- the load balancers `targetForwardingRules` reference
- [**GCP Firewall Rule**](/cloud-catalog/gcp-firewall-rule) -- legacy per-network rules, evaluated last
