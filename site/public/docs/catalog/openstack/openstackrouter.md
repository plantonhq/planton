---
title: "Router"
description: "Router deployment documentation"
icon: "package"
order: 100
componentName: "openstackrouter"
---

# OpenStackRouter Research Documentation

## OpenStack Neutron Router

### What It Is

A Neutron router is a virtual L3 networking device that forwards packets between subnets and, when connected to an external network, provides north-south connectivity (SNAT/DNAT) to the outside world. Routers are the essential bridge between isolated tenant networks and the internet.

### Terraform Resource

- **Type**: `openstack_networking_router_v2`
- **Provider**: `terraform-provider-openstack/openstack` v3.x
- **API**: Neutron L3 API (`/v2.0/routers`)

### Pulumi Resource

- **Type**: `openstack.networking.Router`
- **SDK**: `pulumi-openstack` v5.x

### Architecture

```
External/Provider Network (admin-managed)
         │
         │ external_gateway (external_network_id)
         │
    ┌────▼────┐
    │  Router  │  ◄── SNAT, DNAT, DVR
    └────┬────┘
         │
    ┌────▼────────────────┐
    │  Router Interfaces  │  ◄── openstack_networking_router_interface_v2
    ├─────────┬───────────┤
    │         │           │
  Subnet A  Subnet B  Subnet C   (tenant subnets)
```

### Key Concepts

**External Gateway**: Setting `external_network_id` connects the router to a provider network. This is how tenant traffic reaches the internet. The cloud admin creates the external network; tenants reference it by UUID.

**SNAT (Source NAT)**: When enabled (default), all outbound traffic from tenant subnets is translated to the router's external IP. This allows instances without floating IPs to reach the internet.

**DVR (Distributed Virtual Router)**: In standard mode, all L3 traffic flows through a centralized network node. DVR distributes routing to each compute node, eliminating the bottleneck. DVR is a create-time decision and cannot be changed.

**External Fixed IPs**: When a router gets an external gateway, OpenStack allocates one or more IPs on the external network. You can request specific IPs or let OpenStack auto-allocate.

### Spec Field Rationale (80/20 Analysis)

The Terraform `openstack_networking_router_v2` resource has 16 schema fields. We include 8:

| Included | Why |
|----------|-----|
| `external_network_id` | Core field -- defines external connectivity |
| `admin_state_up` | Control router operational state |
| `enable_snat` | Critical for network architecture decisions |
| `distributed` | DVR vs centralized is a major architectural choice |
| `external_fixed_ips` | Control external IP allocation |
| `description` | Standard field, pattern consistency |
| `tags` | Standard field, pattern consistency |
| `region` | Standard field, pattern consistency |

| Excluded | Why |
|----------|-----|
| `external_qos_policy_id` | Niche QoS feature |
| `flavor_id` | Deployment-specific router flavors (rare) |
| `external_subnet_ids` | Conflicts with external_fixed_ip, niche use case |
| `tenant_id` | Admin-only (cross-tenant router creation) |
| `value_specs` | Custom provider extensions, non-portable |
| `availability_zone_hints` | Niche, AZ-aware routing |
| `vendor_options` | Provider-specific workarounds |
| `name` | Derived from metadata.name (standard pattern) |

### Foreign Key Design

`external_network_id` is an **optional** `StringValueOrRef`:

- **Optional** because routers can exist without external gateways (internal routing only)
- **StringValueOrRef** because in InfraCharts, the external network may be a resource in the same chart (wired via `value_from`) or a pre-existing UUID (literal `value`)
- **default_kind = OpenStackNetwork**: Canvas UI auto-suggests OpenStackNetwork resources
- **default_kind_field_path = "status.outputs.network_id"**: Auto-wires to the network's UUID output

This is the first OpenStack component with an **optional** FK (Subnet's `network_id` is required).

### CEL Validations

Two message-level CEL expressions enforce the TF provider's `RequiredWith` constraints:

1. **enable_snat requires external_network_id**: Setting SNAT without a gateway is meaningless
2. **external_fixed_ips requires external_network_id**: Requesting external IPs without a gateway is meaningless

### Output Design

- `router_id`: Primary FK output, referenced by `OpenStackRouterInterface`
- `external_gateway_ip`: Convenience output -- extracts the first IP from the external_fixed_ips list. Saves users from parsing the raw list. Empty when no external gateway is configured.

### Downstream Dependencies

The `router_id` output is consumed by:
- `OpenStackRouterInterface` (attaches subnets to the router)

### Common Patterns

**Developer Environment**: Router connects a developer's isolated network to the external network for internet access. Combined with SecurityGroup, FloatingIp, and Instance.

**Kubernetes Environment**: Router provides external connectivity for Magnum cluster nodes. Combined with ContainerClusterTemplate and ContainerCluster.

**Project Landing Zone**: Router establishes baseline network connectivity for a new tenant project. Combined with Project, Network, Subnet, and SecurityGroup.
