---
title: "Network Security  Group"
description: "Network Security  Group deployment documentation"
icon: "package"
order: 100
componentName: "azurenetworksecuritygroup"
---

# AzureNetworkSecurityGroup -- Research & Design Documentation

## Overview

Azure Network Security Groups (NSGs) are stateful, Layer 3-4 packet-filtering firewalls
that control traffic flow for Azure resources. They are the primary network access control
mechanism in Azure, analogous to:

- **AWS Security Groups** -- stateful, but AWS SGs are allow-only (no deny rules)
- **GCP Firewall Rules** -- similar priority-based evaluation, but GCP uses VPC-level rules
- **Traditional firewalls** -- NSGs are per-subnet or per-NIC, not per-network-boundary

NSGs are foundational to Azure networking. Every enterprise Azure deployment uses NSGs
to implement network segmentation, enforce the principle of least privilege, and meet
compliance requirements.

## Azure NSG Architecture

### Rule Evaluation Model

Azure evaluates NSG rules using a priority-based, first-match model:

1. **Direction separation** -- Inbound and outbound rules are evaluated independently
2. **Priority ordering** -- Rules are evaluated from lowest priority number (highest priority)
   to highest priority number (lowest priority)
3. **First match wins** -- The first rule whose 5-tuple matches the traffic determines the
   access decision (Allow or Deny)
4. **Default rules** -- If no user rule matches, Azure's implicit default rules apply

### 5-Tuple Matching

Each rule matches traffic based on:

| Field | Description | Example |
|-------|-------------|---------|
| Source IP | Source address/CIDR/tag | `10.0.1.0/24`, `VirtualNetwork`, `*` |
| Source Port | Source port/range | `*`, `1024-65535` |
| Destination IP | Destination address/CIDR/tag | `10.0.2.0/24`, `Internet` |
| Destination Port | Destination port/range | `443`, `80`, `22` |
| Protocol | Transport protocol | `Tcp`, `Udp`, `Icmp`, `*` |

### Implicit Default Rules

Every NSG has six immutable default rules (three inbound, three outbound):

**Inbound defaults:**

| Priority | Name | Action | Description |
|----------|------|--------|-------------|
| 65000 | AllowVnetInBound | Allow | VNet-to-VNet traffic |
| 65001 | AllowAzureLoadBalancerInBound | Allow | Load Balancer health probes |
| 65500 | DenyAllInBound | Deny | All other inbound traffic |

**Outbound defaults:**

| Priority | Name | Action | Description |
|----------|------|--------|-------------|
| 65000 | AllowVnetOutBound | Allow | VNet-to-VNet traffic |
| 65001 | AllowInternetOutBound | Allow | All outbound internet traffic |
| 65500 | DenyAllOutBound | Deny | All other outbound traffic |

### Stateful Behavior

NSGs are stateful: if inbound traffic is allowed, the return (response) traffic is
automatically permitted regardless of outbound rules, and vice versa. This means you
only need to define rules in one direction for established connections.

### Association Model

NSGs can be associated with:

- **Subnets** -- Rules apply to all resources in the subnet
- **Network Interfaces (NICs)** -- Rules apply to a specific VM/resource

When both subnet-level and NIC-level NSGs exist, traffic is evaluated against both:
inbound traffic hits the subnet NSG first, then the NIC NSG. Outbound traffic hits
the NIC NSG first, then the subnet NSG.

## Deployment Landscape

### Azure Terraform Provider Resources

The AzureRM Terraform provider offers two approaches for NSG rules:

1. **Inline rules** -- `security_rule` blocks within `azurerm_network_security_group`
2. **Separate rules** -- `azurerm_network_security_rule` as standalone resources

**Important:** These two approaches conflict. Using both inline and separate rules for
the same NSG causes Terraform state conflicts. OpenMCF uses separate rules exclusively
to provide per-rule lifecycle management and better error reporting.

### Azure Pulumi Provider Resources

The Pulumi Azure Classic SDK (v6) mirrors the Terraform resources:

- `network.NetworkSecurityGroup` -- NSG resource
- `network.NetworkSecurityRule` -- Individual rule resource

### NSG vs Application Security Groups (ASGs)

Azure also offers Application Security Groups (ASGs) for grouping VMs by application
role. ASGs can be used as source or destination in NSG rules instead of IP addresses.
OpenMCF omits ASG support in this 80/20 design because:

1. ASGs add complexity with marginal benefit for most deployments
2. CIDR-based rules cover the vast majority of enterprise use cases
3. ASGs can be added as a future enhancement without breaking changes

### NSG Flow Logs

Azure supports NSG Flow Logs for traffic auditing (via Network Watcher). This is
a separate resource (`azurerm_network_watcher_flow_log`) and is not bundled into
the NSG component. Enabling flow logs is an operational concern that can be handled
by a separate OpenMCF component or infra chart configuration.

## Design Rationale

### Why Strings with CEL Validation (Not Proto Enums)

The `direction`, `access`, and `protocol` fields use string types with CEL `in`
validation instead of protobuf enums. This design choice was established across
R02-R05 (AzureApplicationInsights, AzurePublicIp, AzureSubnet) and is grounded in:

1. **Provider authenticity** -- Azure API values are mixed-case strings (`"Tcp"`, `"Inbound"`,
   `"Allow"`). Proto enums use UPPER_CASE, requiring a mapping layer in IaC modules
2. **Zero mapping** -- String values pass through directly to Azure provider calls
3. **Consistency** -- All Azure resources in OpenMCF use this pattern
4. **Extensibility** -- Adding new protocol values (e.g., `"Ah"`, `"Esp"`) only requires
   updating the CEL expression, not adding proto enum values and regenerating code

### Why Separate Rules (Not Inline)

Security rules are created as separate `NetworkSecurityRule` resources rather than
inline on the NSG. This follows the AzureUserAssignedIdentity pattern (separate
role assignments) and provides:

1. **Per-rule error messages** -- If rule creation fails, the error identifies the specific rule
2. **Per-rule state management** -- Each rule has its own entry in Pulumi/Terraform state
3. **Conflict avoidance** -- Terraform's inline vs separate rule conflict is eliminated
4. **Explicit naming** -- Each rule gets a descriptive Pulumi resource name

### Why No Subnet Association

The NSG-to-subnet association is deliberately excluded from this component:

1. **Independent lifecycle** -- An NSG may be created, reviewed, and approved before
   being associated with a subnet
2. **Reuse** -- The same NSG could potentially be associated with multiple subnets
   (same rules for peer subnets)
3. **Infra chart responsibility** -- The enterprise-network-foundation chart creates
   both NSGs and subnets, then wires the associations as a separate step
4. **Separation of concerns** -- NSG defines "what traffic to allow/deny";
   association defines "where to enforce it"

### Why Description Field Was Added

The T02 plan did not include a `description` field on security rules. It was added
because:

1. Azure supports it (max 140 characters)
2. It serves a real operational need -- when reviewing NSGs in Azure Portal or via CLI,
   descriptions explain the intent behind each rule
3. It has zero cost (optional field, no IaC module complexity)
4. It follows the "production-quality infrastructure" philosophy of the platform

### 80/20 Omissions

The following Azure NSG features are deliberately omitted:

| Feature | Reason for Omission |
|---------|---------------------|
| Application Security Groups (ASGs) | Adds complexity, CIDR-based rules cover 80% of use cases |
| Plural port ranges (`source_port_ranges`) | Single port range covers most rules; multiple rules handle edge cases |
| `Ah` and `Esp` protocols | IPsec protocols are edge cases for enterprise NSGs |
| NSG Flow Logs | Separate operational concern, different lifecycle |
| Augmented security rules | Preview feature, not GA stable |

All omissions can be added later without breaking changes.

## Enterprise Deployment Patterns

### Three-Tier Architecture

The most common pattern -- web, app, and data tiers each with their own NSG:

```
Internet
  │
  ▼
[Web NSG] ── Allow 80, 443 from Internet
  │          Deny all other inbound
  ▼
[App NSG] ── Allow 8080 from web subnet
  │          Allow health probes from LB
  │          Deny all other inbound
  ▼
[Data NSG] ── Allow 5432 from app subnet
              Allow 6380 from app subnet
              Allow 22 from mgmt subnet
              Deny all other inbound
```

### Hub-and-Spoke Network

In hub-and-spoke topologies, NSGs control traffic between spokes via the hub:

- **Hub NSG** -- Controls traffic to/from shared services (DNS, NVA, bastion)
- **Spoke NSGs** -- Per-workload rules specific to each spoke's requirements
- **Gateway NSG** -- Controls traffic entering/leaving the Azure environment

### AKS Cluster NSG

AKS clusters require specific NSG rules for control plane communication:

- Allow TCP 443 (API server) from authorized IP ranges
- Allow TCP 9000 (tunnelfront) from AzureCloud service tag
- Allow UDP 1194 (tunnel) from AzureCloud service tag

## References

- [Azure NSG Documentation](https://learn.microsoft.com/en-us/azure/virtual-network/network-security-groups-overview)
- [NSG Default Rules](https://learn.microsoft.com/en-us/azure/virtual-network/network-security-groups-overview#default-security-rules)
- [Service Tags](https://learn.microsoft.com/en-us/azure/virtual-network/service-tags-overview)
- [Terraform azurerm_network_security_group](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/network_security_group)
- [Terraform azurerm_network_security_rule](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/network_security_rule)
