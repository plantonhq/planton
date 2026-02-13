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

### Provider-Authentic Values

This component uses Azure's exact API values for enum-like fields:

- **Direction**: `"Inbound"`, `"Outbound"` (not uppercase)
- **Access**: `"Allow"`, `"Deny"` (not uppercase)
- **Protocol**: `"Tcp"`, `"Udp"`, `"Icmp"`, `"*"` (not uppercase)

### Address Prefixes

Source and destination addresses support:

- CIDR blocks: `"10.0.0.0/8"`, `"192.168.1.0/24"`
- Single IPs: `"10.0.0.1"`
- Azure service tags: `"VirtualNetwork"`, `"AzureLoadBalancer"`, `"Internet"`
- Wildcard: `"*"` (any address)
- Multiple CIDRs via the plural `_prefixes` fields

## Configuration Options

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `region` | string | yes | -- | Azure region |
| `resource_group` | StringValueOrRef | yes | -- | Resource group |
| `name` | string | yes | -- | NSG name (1-80 chars) |
| `security_rules` | list | no | [] | Security rules |

### Security Rule Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `name` | string | yes | -- | Rule name (1-80 chars) |
| `description` | string | no | -- | Description (max 140 chars) |
| `priority` | int32 | yes | -- | Priority (100-4096) |
| `direction` | string | yes | -- | "Inbound" or "Outbound" |
| `access` | string | yes | -- | "Allow" or "Deny" |
| `protocol` | string | yes | -- | "Tcp", "Udp", "Icmp", or "*" |
| `source_port_range` | string | no | "*" | Source port/range |
| `destination_port_range` | string | yes | -- | Destination port/range |
| `source_address_prefix` | string | no | "*" | Source CIDR/tag |
| `destination_address_prefix` | string | no | "*" | Destination CIDR/tag |
| `source_address_prefixes` | list | no | -- | Multiple source CIDRs |
| `destination_address_prefixes` | list | no | -- | Multiple destination CIDRs |

## Outputs

| Output | Description |
|--------|-------------|
| `nsg_id` | Azure Resource Manager ID of the NSG |
| `nsg_name` | Name of the NSG |

## Infra Chart Usage

AzureNetworkSecurityGroup is a key component in the **enterprise-network-foundation**
infra chart, where per-tier NSGs enforce traffic segmentation:

```
AzureVpc (VNet)
  └── AzureSubnet (web-tier)
        └── AzureNetworkSecurityGroup (web-nsg) ── association
  └── AzureSubnet (app-tier)
        └── AzureNetworkSecurityGroup (app-nsg) ── association
  └── AzureSubnet (data-tier)
        └── AzureNetworkSecurityGroup (data-nsg) ── association
```

The NSG-to-subnet association is created by the infra chart, not by this component.
This keeps the NSG lifecycle independent of any particular subnet.
