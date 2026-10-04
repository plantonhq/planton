# AzureNetworkSecurityGroup

An Azure Network Security Group (NSG) is a stateful packet-filtering firewall that
controls inbound and outbound traffic for Azure resources. NSGs evaluate traffic
against priority-ordered security rules based on 5-tuple matching (source IP, source
port, destination IP, destination port, protocol) combined with an Allow/Deny decision.

## When to Use

Use AzureNetworkSecurityGroup when you need to:

- **Control subnet-level traffic** -- Attach an NSG to a subnet to enforce security
  policies for all resources in that subnet (VMs, AKS nodes, App Service VNet-integrated apps)
- **Implement zero-trust networking** -- Define explicit allow rules and a catch-all deny
  rule to ensure only authorized traffic flows between network tiers
- **Segment enterprise networks** -- Create per-tier NSGs (web, app, data, management) with
  rules that enforce the principle of least privilege between tiers
- **Meet compliance requirements** -- Audit and enforce network access policies with
  deterministic, version-controlled rule sets

## Key Concepts

### Priority Ordering

Rules are evaluated in priority order within each direction (Inbound or Outbound). Lower
priority numbers are evaluated first (100 is highest priority, 4096 is lowest). The first
matching rule determines the traffic decision. If no user-defined rule matches, Azure's
implicit default rules apply.

### Azure Default Rules

Every NSG automatically includes three implicit default rules per direction (priorities
65000-65500) that cannot be deleted:

- **AllowVNetInBound** (65000) -- Allow traffic between resources in the same VNet
- **AllowAzureLoadBalancerInBound** (65001) -- Allow Azure Load Balancer health probes
- **DenyAllInBound** (65500) -- Deny all other inbound traffic

### Enum Values

The `direction`, `access`, and `protocol` fields are closed enums; the IaC modules
map them to Azure's ARM values at deploy time:

- **Direction**: `INBOUND`, `OUTBOUND`
- **Access**: `ALLOW`, `DENY`
- **Protocol**: `ANY` (ARM's `*`), `TCP`, `UDP`, `ICMP`, `AH`, `ESP`

### Address Prefixes

Sources and destinations each take exactly one addressing style:

- A single prefix: a CIDR (`"10.0.0.0/8"`), an IP (`"10.0.0.1"`), an Azure service
  tag (`"VirtualNetwork"`, `"AzureLoadBalancer"`, `"Internet"`), or `"*"` (any)
- Multiple CIDRs/IPs via the plural `_prefixes` fields (service tags and `"*"` are
  singular-only)
- Application security group IDs via the `_application_security_group_ids` fields
  (identity-based addressing, up to 10 plain ARM IDs)

Leaving all three unset means any (`*`).

## Configuration Options

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `region` | string | yes | -- | Azure region |
| `resource_group` | StringValueOrRef | yes | -- | Resource group |
| `name` | string | yes | -- | NSG name (1-80 chars) |
| `security_rules` | list | no | [] | Security rules |
| `tags` | map | no | {} | Free-form tags merged over Planton-derived resource tags (user tag wins on key conflict) |

### Security Rule Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `name` | string | yes | -- | Rule name (1-80 chars) |
| `description` | string | no | -- | Description (max 140 chars) |
| `priority` | int32 | yes | -- | Priority (100-4096) |
| `direction` | enum | yes | -- | `INBOUND` or `OUTBOUND` |
| `access` | enum | yes | -- | `ALLOW` or `DENY` |
| `protocol` | enum | yes | -- | `ANY`, `TCP`, `UDP`, `ICMP`, `AH`, or `ESP` |
| `source_port_range` | string | no | any | Source port/range; never combined with `source_port_ranges` |
| `source_port_ranges` | list | no | -- | Multiple source ports/ranges; never combined with `source_port_range` |
| `destination_port_range` | string | yes* | -- | Destination port/range |
| `destination_port_ranges` | list | yes* | -- | Multiple destination ports/ranges |
| `source_address_prefix` | string | no | any | Source CIDR/tag/`*` (at most one source addressing style) |
| `source_address_prefixes` | list | no | -- | Multiple source CIDRs (at most one source addressing style) |
| `source_application_security_group_ids` | list | no | -- | Source application security group IDs, up to 10 (at most one source addressing style) |
| `destination_address_prefix` | string | no | any | Destination CIDR/tag/`*` (at most one destination addressing style) |
| `destination_address_prefixes` | list | no | -- | Multiple destination CIDRs (at most one destination addressing style) |
| `destination_application_security_group_ids` | list | no | -- | Destination application security group IDs, up to 10 (at most one destination addressing style) |

\* Exactly one of `destination_port_range` or `destination_port_ranges` must be set (use `"*"` for any port).

## Outputs

| Output | Description |
|--------|-------------|
| `network_security_group_id` | Azure Resource Manager ID of the NSG |
| `network_security_group_name` | Name of the NSG |

## Infra Chart Usage

AzureNetworkSecurityGroup is a key component in the **enterprise-network-foundation**
Infra Chart, where per-tier NSGs enforce traffic segmentation:

```
AzureVirtualNetwork (VNet)
  └── AzureSubnet (web-tier)
        └── AzureNetworkSecurityGroup (web-nsg) ── association
  └── AzureSubnet (app-tier)
        └── AzureNetworkSecurityGroup (app-nsg) ── association
  └── AzureSubnet (data-tier)
        └── AzureNetworkSecurityGroup (data-nsg) ── association
```

The NSG-to-subnet association is created by the Infra Chart, not by this component.
This keeps the NSG lifecycle independent of any particular subnet.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
