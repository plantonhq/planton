# GCP Hierarchical Firewall Policy

Creates a hierarchical firewall policy: an ordered set of firewall rules that lives on the organization or a folder and is enforced on every VPC network beneath the node it is associated with -- in every project, present and future -- before any network-level policy or legacy VPC firewall rule gets a say. It is the tool a platform team uses to state the rules nobody in the organization may undo ("no SSH from the internet", "deny known-malicious sources", "allow the health-check ranges") and to delegate everything else downward with `goto_next`. Rules key on the modern matching surface the legacy VPC firewall never had: secure tags, address groups, domain names, threat-intelligence lists, geography, and Cloud NGFW security profile groups with TLS inspection.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Firewall policy** -- the `compute_firewall_policy` container on the organization or folder, named by Google with a server-assigned numeric ID
- **Rules** -- one `compute_firewall_policy_rule` per `rules` entry, keyed by priority
- **Associations** -- one `compute_firewall_policy_association` per `associations` entry, attaching the policy to the organization or a folder

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module whose identity holds `roles/compute.orgFirewallPolicyAdmin` on the organization (or on the folder the policy lives under and every folder it is associated with). Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.

### GCP Organization

- **A Google Cloud Organization** -- hierarchical policies do not exist in a project-only account. Have the numeric organization ID (`gcloud organizations list`), or declare the folder with `GcpFolder` and reference its `folder_id` output.
- **Secure tags** for tag-keyed rules are `GcpTagValue` resources under a `GcpTagKey` with `purpose: GCE_FIREWALL`; reference the value's `name` output.

## Deploy

### Console

Open the deployment store, find **GCP Hierarchical Firewall Policy**, and click **Deploy**. The creation wizard walks you through the parent, the rules, and where the policy is enforced. Start from the **Org Deny SSH From Internet** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpHierarchicalFirewallPolicy
metadata:
  name: org-baseline
  org: acme-corp
  env: prod
spec:
  parent:
    organizationId: "123456789012"
  description: Organization-wide baseline nobody may undo
  rules:
    - priority: 1000
      action: deny
      direction: INGRESS
      description: No SSH from the internet, anywhere in the organization
      enableLogging: true
      match:
        srcIpRanges: ["0.0.0.0/0"]
        layer4Configs:
          - ipProtocol: tcp
            ports: ["22"]
    - priority: 2000
      action: goto_next
      direction: INGRESS
      description: Everything else is decided lower down
      match:
        layer4Configs:
          - ipProtocol: all
  associations:
    - name: org-baseline-org
      target:
        organizationId: "123456789012"
  deletionPolicy: PREVENT
```

```shell
planton apply -f org-baseline.yaml
```

This creates a policy under the organization with two rules -- a logged deny of SSH from anywhere, then a delegation of everything else -- and enforces it on the whole organization, guarded against accidental destroy. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a landing zone, the policy references its `GcpFolder` via ValueFromRef and its rules reference the `GcpTagValue` and `GcpVpcNetwork` blocks the chart declares:

```yaml
spec:
  parent:
    folderId:
      valueFrom:
        kind: GcpFolder
        name: production
        fieldPath: status.outputs.folder_id
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
    - target:
        folderId:
          valueFrom:
            kind: GcpFolder
            name: production
            fieldPath: status.outputs.folder_id
```

The InfraPipeline deploys the folder and the tag values, then the policy, and destroys the policy before them -- the order Google requires.

## Key Configuration

These are the most important decisions when configuring a hierarchical policy. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Parent versus associations** -- `parent` is where the policy lives (IAM, quota); `associations` is where it is enforced. Usually the same node, but one policy can be enforced on several folders. A node carries one hierarchical policy at a time.

**Priority is identity** -- Google keys rules by priority; renumbering recreates the rule. Leave gaps (1000, 2000, ...) so a rule can be slotted in later. Priorities 2147483646 and 2147483647 are Google's implied `goto_next` rules.

**`goto_next` is the delegation lever** -- an organization policy that ends every decision with `allow` or `deny` leaves the projects nothing to decide. The everyday shape is a few non-negotiable denies (and allows for Google's health-check and IAP ranges), then `goto_next` for the rest.

**Direction decides which match fields matter** -- on `INGRESS`, the `src*` fields describe where traffic comes from; on `EGRESS`, the `dest*` fields describe where it goes. Secure tags and networks exist only on the VPC side, so `srcSecureTags` and `srcNetworks` are meaningful on ingress rules only.

**Targets narrow the rule** -- `targetResources` (networks), `targetSecureTags`, or `targetServiceAccounts` restrict which VMs the rule applies to; empty means every VM beneath the association. Tags and service accounts cannot be combined on one rule.

**Deletion policy** -- `DELETE` (default) detaches, removes, and deletes; `PREVENT` fails the destroy, the guard for an organization's baseline; `ABANDON` leaves everything live and enforcing.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpFolder** | `parent.folderId`, `associations[].target.folderId` | `status.outputs.folder_id` |
| **GcpVpcNetwork** | `rules[].targetResources[]`, `rules[].match.srcNetworks[]` | `status.outputs.network_self_link` |
| **GcpTagValue** | `rules[].targetSecureTags[]`, `rules[].match.srcSecureTags[]` | `status.outputs.name` |
| **GcpServiceAccount** | `rules[].targetServiceAccounts[]` | `status.outputs.email` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `policy_id` | The server-assigned numeric ID | `gcloud compute firewall-policies describe`; firewall log correlation |
| `short_name` | The short name | Tooling |
| `self_link` | Self-link URL | Tooling |
| `parent` | `organizations/{id}` or `folders/{id}` | Audit |
| `rule_tuple_count` | Google's complexity measure | Quota watch before adding wide rules |
| `association_names` | Association names in declaration order | A detach targets one |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Organization baseline** -- a logged deny of SSH from the internet and `goto_next` for the rest, enforced on the organization. Start from the **Org Deny SSH From Internet** preset.

**Folder allow-list with delegation** -- allow intra-VPC traffic and Google's health-check ranges, delegate everything else, enforced on one folder by reference. Start from the **Folder Allow Internal Goto Next** preset.

**Threat intelligence and geo-blocking** -- deny known-malicious sources and Tor exit nodes, and traffic from listed countries, at the top of the organization. Start from the **Threat Intel And Geo Block** preset.

## Works With

- [**GCP Network Firewall Policy**](/cloud-catalog/gcp-network-firewall-policy) -- the project-level policy attached to VPC networks, evaluated after this one
- [**GCP Folder**](/cloud-catalog/gcp-folder) -- the folder a policy lives under or is enforced on
- [**GCP Tag Value**](/cloud-catalog/gcp-tag-value) -- the secure tags rules key on
- [**GCP VPC Network**](/cloud-catalog/gcp-vpc-network) -- the networks `targetResources` and `srcNetworks` reference
- [**GCP Firewall Rule**](/cloud-catalog/gcp-firewall-rule) -- legacy per-network rules, evaluated last
