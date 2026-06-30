# OpenStack Load Balancer

Provision and manage Octavia load balancers in OpenStack using Planton's unified API.

## Overview

An Octavia load balancer provides a Virtual IP (VIP) endpoint on a subnet. Listeners, pools, members, and health monitors are attached to it to distribute traffic across backend servers. The VIP can be associated with a floating IP for external access.

This component creates an `openstack_lb_loadbalancer_v2` resource through both Pulumi and Terraform IaC modules with full feature parity.

The load balancer name is derived from `metadata.name`.

## Prerequisites

1. **OpenStack Cloud**: Access to an OpenStack deployment with Octavia (Load Balancing service)
2. **Credentials**: OpenStack credentials configured via the credential management system
3. **Planton CLI**: Install from [planton.dev](https://planton.dev)
4. **Subnet**: An existing OpenStack subnet (see `OpenStackSubnet`)

## Quick Start

### Minimal Load Balancer

Create a load balancer on an existing subnet with a literal subnet ID:

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackLoadBalancer
metadata:
  name: dev-lb
spec:
  vip_subnet_id:
    value: "e0a1f622-9aab-4a48-8c8c-3b0c7e2a9b1d"
```

### Load Balancer with Foreign Key Reference

Reference a subnet managed by Planton using `value_from`:

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackLoadBalancer
metadata:
  name: dev-lb
spec:
  vip_subnet_id:
    value_from:
      name: dev-subnet
  description: "Development load balancer"
```

### Deploy

```bash
planton apply --manifest loadbalancer.yaml \
  -p openstack-creds.yaml
```

## Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `vip_subnet_id` | StringValueOrRef | Yes | Subnet for VIP allocation (FK to OpenStackSubnet) |
| `vip_address` | string | No | Specific VIP address. Auto-allocated if omitted. ForceNew |
| `description` | string | No | Human-readable description |
| `admin_state_up` | bool | No | Administrative state. Default: `true` |
| `flavor_id` | string | No | Octavia flavor ID for resource limits. ForceNew |
| `tags` | string[] | No | Tags applied to the OpenStack resource. Must be unique |
| `region` | string | No | Override the region from the provider config |

## Outputs

| Field | Description |
|-------|-------------|
| `loadbalancer_id` | The unique identifier (UUID) of the load balancer |
| `name` | The load balancer name (from `metadata.name`) |
| `vip_address` | The Virtual IP address (computed or specified) |
| `vip_port_id` | The Neutron port ID of the VIP |
| `vip_subnet_id` | The subnet where the VIP was allocated |
| `region` | Region where the load balancer was created |

## Foreign Key Relationships

**Inbound (this component references):**

- `vip_subnet_id` -> `OpenStackSubnet.status.outputs.subnet_id`

**Outbound (referenced by downstream):**

- `OpenStackLoadBalancerListener.spec.loadbalancer_id` -> `loadbalancer_id`

## IaC Implementations

This component is implemented with both Pulumi (Go) and Terraform (HCL) with full feature parity.

- **Terraform resource**: `openstack_lb_loadbalancer_v2`
- **Pulumi resource**: `openstack.loadbalancer.LoadBalancer`

## Notes

- **VIP allocation**: By default, Octavia auto-allocates an available IP from the specified subnet. Use `vip_address` to request a specific IP (must be within the subnet's CIDR).
- **Floating IP**: To make the load balancer externally accessible, associate a floating IP with the `vip_port_id` output using `OpenStackFloatingIpAssociate`.
- **Flavors**: Octavia flavors control resource limits (bandwidth, connections). If your cloud provides multiple flavors, specify one via `flavor_id`.
- **Admin state**: Set `admin_state_up: false` to create a load balancer in disabled state (useful for maintenance windows).
- **Tags**: Must be unique within the resource. Used for filtering and organization in OpenStack.
