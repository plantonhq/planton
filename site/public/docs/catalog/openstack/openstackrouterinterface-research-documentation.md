---
title: "OpenStackRouterInterface Research Documentation"
description: "OpenStackRouterInterface Research Documentation deployment documentation"
icon: "package"
order: 100
componentName: "openstackrouterinterface"
---

# OpenStackRouterInterface Research Documentation

## OpenStack Neutron Router Interface

### What It Is

A router interface is a binding between a Neutron router and a subnet. Under the hood, it creates a port on the subnet with the subnet's gateway IP and attaches that port to the router. This port becomes the default gateway for all instances on the subnet.

A router interface is not a standalone resource in the traditional sense -- it's a relationship. The Neutron API uses `PUT /v2.0/routers/{router_id}/add_router_interface` with either a `subnet_id` or `port_id`. The Terraform provider wraps this into a resource lifecycle.

### Terraform Resource

- **Type**: `openstack_networking_router_interface_v2`
- **Provider**: `terraform-provider-openstack/openstack` v3.x
- **API**: Neutron L3 API (`/v2.0/routers/{router_id}/add_router_interface`)

### Pulumi Resource

- **Type**: `openstack.networking.RouterInterface`
- **SDK**: `pulumi-openstack` v5.x

### Architecture

```
┌──────────────────────────────────────────────────┐
│                    Router                        │
│  (L3 device -- routes between subnets and        │
│   to external networks via SNAT/DNAT)            │
└──────────┬───────────────┬───────────────────────┘
           │               │
    ┌──────▼──────┐ ┌──────▼──────┐
    │   Router    │ │   Router    │    ◄── Router Interfaces
    │  Interface  │ │  Interface  │        (this component)
    │  (port A)   │ │  (port B)   │
    └──────┬──────┘ └──────┬──────┘
           │               │
    ┌──────▼──────┐ ┌──────▼──────┐
    │  Subnet A   │ │  Subnet B   │    ◄── Subnets
    │ 10.0.1.0/24 │ │ 10.0.2.0/24 │
    └─────────────┘ └─────────────┘
```

Each router interface creates a port on the subnet. The port's IP becomes the default gateway for instances on that subnet.

### Spec Field Rationale (80/20 Analysis)

The Terraform `openstack_networking_router_interface_v2` resource has 5 schema fields. We include 3:

| Included | Why |
|----------|-----|
| `router_id` | Core required field -- which router to attach to |
| `subnet_id` | Core attachment mode -- identifies the subnet to connect |
| `region` | Standard field for multi-region deployments |

| Excluded | Why |
|----------|-----|
| `port_id` | Alternative attachment mode for advanced use cases (specific IP/MAC control). 95%+ of usage is subnet-based. Can be added later. |
| `force_destroy` | Operational escape hatch for messy deletions. Proper dependency ordering handles this. |

This component has the highest inclusion ratio in the project (3 of 5 fields = 60%), but the 2 excluded fields are genuinely niche.

### Foreign Key Design

Both `router_id` and `subnet_id` are **required** `StringValueOrRef` fields:

- **Required** because a router interface without both a router and a subnet is meaningless
- **StringValueOrRef** because in InfraCharts, both the router and subnet may be resources in the same chart (wired via `value_from`) or pre-existing UUIDs (literal `value`)
- **default_kind annotations**: `OpenStackRouter` and `OpenStackSubnet` respectively, enabling Canvas UI auto-suggestion

This is the first OpenStack component with **two required FKs** in one spec, establishing the pattern for `OpenStackVolumeAttach` (Instance + Volume), `OpenStackFloatingIpAssociate` (FloatingIp + Port), and other join resources.

### Resource Identity

Unlike most OpenStack resources, a router interface has no user-visible "name" attribute. The resource is identified by the port it creates:

- **Terraform**: `d.SetId(r.PortID)` -- the resource ID is the port UUID
- **Pulumi**: The resource is named from `metadata.name`, but the OpenStack-side identifier is the port UUID
- **OpenMCF**: `metadata.name` provides the KRM identity for DAG wiring and human identification

### ForceNew Behavior

All fields on `openstack_networking_router_interface_v2` are `ForceNew`:
- `router_id`: ForceNew
- `subnet_id`: ForceNew
- `port_id`: ForceNew
- `region`: ForceNew

This means any change to the spec recreates the router interface. This is expected -- you can't re-attach a router to a different subnet in-place. OpenStack's API supports `add_router_interface` and `remove_router_interface` but not "modify."

### Output Design

- `port_id`: The auto-created port's UUID. This is the Terraform resource ID and is useful for debugging. The port is owned by the router interface -- deleting the interface also deletes the port.
- `router_id` and `subnet_id`: Echoed back for convenience and for use in downstream debugging/references.
- `region`: Standard output for multi-region awareness.

### Downstream Dependencies

No downstream components currently reference router interface outputs. The router interface is a terminal node in the dependency graph -- it consumes `router_id` and `subnet_id` but doesn't produce outputs consumed by other components.

### Common Patterns

**Developer Environment**: Router interface connects the developer's subnet to the edge router for internet access. The InfraChart wires: Network -> Subnet -> RouterInterface <- Router <- ExternalNetwork.

**Multiple Subnets**: A single router can have multiple interfaces, each connecting a different subnet. In InfraCharts, this is expressed as multiple OpenStackRouterInterface resources with the same `router_id` `value_from` but different `subnet_id` `value_from` references.

**Kubernetes Environment**: Router interface connects the Magnum cluster's network to the external network, allowing nodes to pull images and communicate with the Kubernetes API.
