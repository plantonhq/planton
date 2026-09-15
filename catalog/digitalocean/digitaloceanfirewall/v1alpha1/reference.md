# DigitalOceanFirewall

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `digital-ocean.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

DigitalOceanFirewallSpec models the full `digitalocean_firewall` surface: a
named rule set (inbound and/or outbound) applied to Droplets directly by ID
or indirectly by Droplet tag. Every argument the provider accepts is
representable here; rule sources and destinations can name other Planton
resources (Droplets, load balancers, Kubernetes clusters) as references
instead of hand-copied IDs, so firewalls compose in infra charts.

## Example

```yaml
# Example DigitalOceanFirewall manifests. Deploy with:
#   planton apply -f manifest.yaml
#
# Document 1 -- a web-tier firewall: HTTPS from everywhere, SSH from a
# management network only, ping allowed, all outbound open. No targeting
# (attach Droplets later by tag or by reference).
#
# Document 2 -- a database-tier firewall targeting Droplets by literal ID
# and by tag: inbound only from the app tier's tag and a private CIDR,
# outbound restricted to DNS and package mirrors over HTTPS.
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanFirewall
metadata:
  name: example-dofw-web
spec:
  firewallName: web-tier-firewall
  inboundRules:
    - protocol: tcp
      portRange: "443"
      sourceAddresses:
        - 0.0.0.0/0
        - ::/0
    - protocol: tcp
      portRange: "22"
      sourceAddresses:
        - 203.0.113.0/24
    - protocol: icmp
      sourceAddresses:
        - 0.0.0.0/0
  outboundRules:
    - protocol: tcp
      portRange: all
      destinationAddresses:
        - 0.0.0.0/0
        - ::/0
    - protocol: udp
      portRange: all
      destinationAddresses:
        - 0.0.0.0/0
        - ::/0
---
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanFirewall
metadata:
  name: example-dofw-db
spec:
  firewallName: db-tier-firewall
  dropletIds:
    - value: "123456789"
  tags:
    - db-tier
  inboundRules:
    - protocol: tcp
      portRange: "5432"
      sourceTags:
        - app-tier
    - protocol: tcp
      portRange: "5432"
      sourceAddresses:
        - 10.10.0.0/16
  outboundRules:
    - protocol: udp
      portRange: "53"
      destinationAddresses:
        - 0.0.0.0/0
    - protocol: tcp
      portRange: "443"
      destinationAddresses:
        - 0.0.0.0/0
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.firewallName` | `string` | yes |  |  |
| `spec.inboundRules` | `[]DigitalOceanFirewallInboundRule` |  |  |  |
| `spec.inboundRules[].protocol` | `string` | yes |  |  |
| `spec.inboundRules[].portRange` | `string` |  |  |  |
| `spec.inboundRules[].sourceAddresses` | `[]string` |  |  |  |
| `spec.inboundRules[].sourceTags` | `[]string` |  |  |  |
| `spec.inboundRules[].sourceDropletIds` | `[]string \| valueFrom` |  |  | DigitalOceanDroplet (`status.outputs.droplet_id`) |
| `spec.inboundRules[].sourceKubernetesIds` | `[]string \| valueFrom` |  |  | DigitalOceanKubernetesCluster (`status.outputs.cluster_id`) |
| `spec.inboundRules[].sourceLoadBalancerUids` | `[]string \| valueFrom` |  |  | DigitalOceanLoadBalancer (`status.outputs.load_balancer_id`) |
| `spec.outboundRules` | `[]DigitalOceanFirewallOutboundRule` |  |  |  |
| `spec.outboundRules[].protocol` | `string` | yes |  |  |
| `spec.outboundRules[].portRange` | `string` |  |  |  |
| `spec.outboundRules[].destinationAddresses` | `[]string` |  |  |  |
| `spec.outboundRules[].destinationTags` | `[]string` |  |  |  |
| `spec.outboundRules[].destinationDropletIds` | `[]string \| valueFrom` |  |  | DigitalOceanDroplet (`status.outputs.droplet_id`) |
| `spec.outboundRules[].destinationKubernetesIds` | `[]string \| valueFrom` |  |  | DigitalOceanKubernetesCluster (`status.outputs.cluster_id`) |
| `spec.outboundRules[].destinationLoadBalancerUids` | `[]string \| valueFrom` |  |  | DigitalOceanLoadBalancer (`status.outputs.load_balancer_id`) |
| `spec.tags` | `[]string` |  |  |  |
| `spec.dropletIds` | `[]string \| valueFrom` |  |  | DigitalOceanDroplet (`status.outputs.droplet_id`) |

## Field Details

### spec.firewallName

`string` · required

Name of the firewall. Must be unique per account. The API accepts up to
255 characters of letters, numbers, colons, dashes, and underscores.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"255"}}

### spec.inboundRules

`[]DigitalOceanFirewallInboundRule`

Inbound rules: traffic allowed *to* the protected Droplets. Traffic not
matched by any inbound rule is dropped (DigitalOcean firewalls are
default-deny for configured directions).

- rule: port_range is required when protocol is tcp or udp

### spec.inboundRules[].protocol

`string` · required

The traffic protocol this rule matches.

- rule: {"required":true,"string":{"in":["tcp","udp","icmp"]}}

### spec.inboundRules[].portRange

`string`

Ports to allow: a single port ("80"), a range ("8000-9000"), or "all".
Required for tcp/udp; omit for icmp (the provider drops any port_range
set on an icmp rule when it reads state back). Note the read-back
normalization: the API reports "all ports" as port 0, which the provider
writes back as the literal string "all" — so "1-65535" reads back as
"all" after apply. Prefer writing "all" to avoid a permanent diff.

### spec.inboundRules[].sourceAddresses

`[]string`

IPv4/IPv6 addresses or CIDR ranges traffic is allowed from
(e.g. "192.0.2.0/24", "0.0.0.0/0", "::/0").

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.inboundRules[].sourceTags

`[]string`

Droplet tag names; traffic from any Droplet carrying one of these tags
is allowed. Tag values are case-insensitive for set membership.

- rule: {"repeated":{"items":{"string":{"pattern":"^[a-zA-Z0-9:\\-_]{1,255}$"}}}}

### spec.inboundRules[].sourceDropletIds

`[]string | valueFrom`

Droplets traffic is allowed from, as literal numeric Droplet IDs or
references to DigitalOceanDroplet resources.

- references: DigitalOceanDroplet (`status.outputs.droplet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanDroplet, name: <that resource's name>, fieldPath: status.outputs.droplet_id}} -- a bare string does not parse

### spec.inboundRules[].sourceKubernetesIds

`[]string | valueFrom`

Kubernetes cluster IDs traffic is allowed from, as literal UUIDs or
references to DigitalOceanKubernetesCluster resources.

Containment-exempt: a firewall is an account-level rule set applied to
Droplets; a cluster named here is a TRUSTED SOURCE of traffic, never the
place the firewall lives. On a diagram the firewall stands beside the
Droplets it protects, with a line to the cluster whose traffic it admits.

- references: DigitalOceanKubernetesCluster (`status.outputs.cluster_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanKubernetesCluster, name: <that resource's name>, fieldPath: status.outputs.cluster_id}} -- a bare string does not parse

### spec.inboundRules[].sourceLoadBalancerUids

`[]string | valueFrom`

Load balancer UIDs traffic is allowed from, as literal UUIDs or
references to DigitalOceanLoadBalancer resources.

- references: DigitalOceanLoadBalancer (`status.outputs.load_balancer_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanLoadBalancer, name: <that resource's name>, fieldPath: status.outputs.load_balancer_id}} -- a bare string does not parse

### spec.outboundRules

`[]DigitalOceanFirewallOutboundRule`

Outbound rules: traffic allowed *from* the protected Droplets to the
configured destinations.

- rule: port_range is required when protocol is tcp or udp

### spec.outboundRules[].protocol

`string` · required

The traffic protocol this rule matches.

- rule: {"required":true,"string":{"in":["tcp","udp","icmp"]}}

### spec.outboundRules[].portRange

`string`

Ports to allow: a single port ("80"), a range ("8000-9000"), or "all".
Required for tcp/udp; omit for icmp. The same read-back normalization as
inbound rules applies: the API reports "all ports" as "all".

### spec.outboundRules[].destinationAddresses

`[]string`

IPv4/IPv6 addresses or CIDR ranges traffic is allowed to
(e.g. "0.0.0.0/0", "::/0").

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.outboundRules[].destinationTags

`[]string`

Droplet tag names; traffic to any Droplet carrying one of these tags is
allowed. Tag values are case-insensitive for set membership.

- rule: {"repeated":{"items":{"string":{"pattern":"^[a-zA-Z0-9:\\-_]{1,255}$"}}}}

### spec.outboundRules[].destinationDropletIds

`[]string | valueFrom`

Droplets traffic is allowed to, as literal numeric Droplet IDs or
references to DigitalOceanDroplet resources.

- references: DigitalOceanDroplet (`status.outputs.droplet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanDroplet, name: <that resource's name>, fieldPath: status.outputs.droplet_id}} -- a bare string does not parse

### spec.outboundRules[].destinationKubernetesIds

`[]string | valueFrom`

Kubernetes cluster IDs traffic is allowed to, as literal UUIDs or
references to DigitalOceanKubernetesCluster resources.

Containment-exempt: a cluster named here is a permitted DESTINATION of
the protected Droplets' traffic, never the place the firewall lives (see
the inbound rule's source_kubernetes_ids).

- references: DigitalOceanKubernetesCluster (`status.outputs.cluster_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanKubernetesCluster, name: <that resource's name>, fieldPath: status.outputs.cluster_id}} -- a bare string does not parse

### spec.outboundRules[].destinationLoadBalancerUids

`[]string | valueFrom`

Load balancer UIDs traffic is allowed to, as literal UUIDs or references
to DigitalOceanLoadBalancer resources.

- references: DigitalOceanLoadBalancer (`status.outputs.load_balancer_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanLoadBalancer, name: <that resource's name>, fieldPath: status.outputs.load_balancer_id}} -- a bare string does not parse

### spec.tags

`[]string`

(Optional) Droplet tag names this firewall applies to: any Droplet
carrying one of these tags is protected, and membership follows the tag
automatically as Droplets come and go. DigitalOcean creates tags
implicitly when first referenced. The API documents a maximum of 5 tags
per firewall (enforced server-side, not by the provider). Tag values are
case-insensitive for set membership.

- rule: {"repeated":{"items":{"string":{"pattern":"^[a-zA-Z0-9:\\-_]{1,255}$"}}}}

### spec.dropletIds

`[]string | valueFrom`

(Optional) Droplets this firewall applies to, as literal numeric Droplet
IDs or references to DigitalOceanDroplet resources. The API documents a
maximum of 10 Droplets per firewall (enforced server-side). For dynamic
membership prefer tags, which track Droplets automatically.

- references: DigitalOceanDroplet (`status.outputs.droplet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanDroplet, name: <that resource's name>, fieldPath: status.outputs.droplet_id}} -- a bare string does not parse

## Validation Rules

- `spec.at_least_one_rule`: at least one inbound_rule or outbound_rule must be specified

## Outputs

Reference an output from another manifest as `valueFrom: {kind: DigitalOceanFirewall, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.firewall_id` | `string` | The unique ID of the firewall (a UUID assigned by DigitalOcean). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.inboundRules[].sourceDropletIds` | DigitalOceanDroplet | `status.outputs.droplet_id` |
| `spec.inboundRules[].sourceKubernetesIds` | DigitalOceanKubernetesCluster | `status.outputs.cluster_id` |
| `spec.inboundRules[].sourceLoadBalancerUids` | DigitalOceanLoadBalancer | `status.outputs.load_balancer_id` |
| `spec.outboundRules[].destinationDropletIds` | DigitalOceanDroplet | `status.outputs.droplet_id` |
| `spec.outboundRules[].destinationKubernetesIds` | DigitalOceanKubernetesCluster | `status.outputs.cluster_id` |
| `spec.outboundRules[].destinationLoadBalancerUids` | DigitalOceanLoadBalancer | `status.outputs.load_balancer_id` |
| `spec.dropletIds` | DigitalOceanDroplet | `status.outputs.droplet_id` |

## See Also

- [Overview](../README.md)
