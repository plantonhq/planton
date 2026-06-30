# OpenStack Load Balancer Member

Provision and manage Octavia pool members in OpenStack using Planton's unified API.

## Overview

An Octavia pool member represents a backend server that receives traffic from the pool's load-balancing algorithm. Each member defines an IP address and port where the backend server listens, along with an optional weight for weighted load distribution.

This component creates an `openstack_lb_member_v2` resource through both Pulumi and Terraform IaC modules with full feature parity. The member name is derived from `metadata.name`.

## Prerequisites

1. **OpenStack Cloud**: Access to an OpenStack deployment with Octavia
2. **Credentials**: OpenStack credentials configured via the credential management system
3. **Planton CLI**: Install from [planton.dev](https://planton.dev)
4. **Pool**: An existing OpenStack load balancer pool (see `OpenStackLoadBalancerPool`)

## Quick Start

### Minimal Member

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackLoadBalancerMember
metadata:
  name: web-backend-1
spec:
  pool_id:
    value: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  address: "10.0.0.10"
  protocol_port: 8080
```

### Deploy

```bash
planton apply --manifest member.yaml -p openstack-creds.yaml
```

## Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `pool_id` | StringValueOrRef | Yes | Pool FK. ForceNew |
| `address` | string | Yes | Backend server IP address. ForceNew |
| `protocol_port` | int32 | Yes | Backend port (1-65535). ForceNew |
| `subnet_id` | StringValueOrRef | No | Subnet for cross-subnet routing. ForceNew |
| `weight` | int32 | No | Member weight (0-256). Default: 1 |
| `admin_state_up` | bool | No | Administrative state. Default: true |
| `tags` | string[] | No | Tags. Must be unique |
| `region` | string | No | Override region from provider config |

## Outputs

| Field | Description |
|-------|-------------|
| `member_id` | UUID of the member |
| `name` | Member name |
| `address` | Backend server IP address |
| `protocol_port` | Backend server port |
| `weight` | Member weight |
| `region` | Region where the member was created |

## Foreign Key Relationships

**Inbound:**
- `pool_id` -> `OpenStackLoadBalancerPool.status.outputs.pool_id`
- `subnet_id` -> `OpenStackSubnet.status.outputs.subnet_id` (optional)

## IaC Implementations

- **Terraform resource**: `openstack_lb_member_v2`
- **Pulumi resource**: `openstack.loadbalancer.Member`

## Notes

- **Weight**: 0 = drain (no traffic), 1-256 = weighted distribution. Octavia defaults to 1 when unset.
- **Subnet**: Required for cross-subnet routing. Octavia uses it for L3 routing when the member is on a different subnet than the VIP.
- **Admin state**: Set admin_state_up false to remove the member from rotation without deleting it.
- **Tags**: Must be unique within the resource.
