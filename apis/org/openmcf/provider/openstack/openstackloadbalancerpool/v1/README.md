# OpenStack Load Balancer Pool

Provision and manage Octavia backend pools in OpenStack using OpenMCF's unified API.

## Overview

An Octavia pool groups backend members (servers) and defines the protocol and load-balancing algorithm used to distribute traffic from a listener. Each pool is the default pool for exactly one listener, and members and health monitors attach to the pool.

This component creates an `openstack_lb_pool_v2` resource through both Pulumi and Terraform IaC modules with full feature parity. The pool name is derived from `metadata.name`.

## Prerequisites

1. **OpenStack Cloud**: Access to an OpenStack deployment with Octavia
2. **Credentials**: OpenStack credentials configured via the credential management system
3. **OpenMCF CLI**: Install from [openmcf.org](https://openmcf.org)
4. **Listener**: An existing OpenStack load balancer listener (see `OpenStackLoadBalancerListener`)

## Quick Start

### Minimal Pool

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerPool
metadata:
  name: web-pool
spec:
  listener_id:
    value: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  protocol: "HTTP"
  lb_method: "ROUND_ROBIN"
```

### Deploy

```bash
openmcf apply --manifest pool.yaml -p openstack-creds.yaml
```

## Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `listener_id` | StringValueOrRef | Yes | Listener FK. ForceNew |
| `protocol` | string | Yes | Backend protocol: HTTP, HTTPS, TCP, UDP, PROXY. ForceNew |
| `lb_method` | string | Yes | Algorithm: ROUND_ROBIN, LEAST_CONNECTIONS, SOURCE_IP, SOURCE_IP_PORT |
| `persistence` | object | No | Session persistence config (type + optional cookie_name) |
| `description` | string | No | Human-readable description |
| `admin_state_up` | bool | No | Administrative state. Default: true |
| `tags` | string[] | No | Tags. Must be unique |
| `region` | string | No | Override region from provider config |

## Outputs

| Field | Description |
|-------|-------------|
| `pool_id` | UUID of the pool |
| `name` | Pool name |
| `protocol` | Backend protocol |
| `lb_method` | Load-balancing algorithm |
| `region` | Region where the pool was created |

## Foreign Key Relationships

**Inbound:** `listener_id` -> `OpenStackLoadBalancerListener.status.outputs.listener_id`

**Outbound:**
- `OpenStackLoadBalancerMember.spec.pool_id` -> `pool_id`
- `OpenStackLoadBalancerMonitor.spec.pool_id` -> `pool_id`

## IaC Implementations

- **Terraform resource**: `openstack_lb_pool_v2`
- **Pulumi resource**: `openstack.loadbalancer.Pool`

## Notes

- **listener_id vs loadbalancer_id**: Exposes listener_id only (80/20 design).
- **Session persistence**: SOURCE_IP, HTTP_COOKIE, APP_COOKIE (requires cookie_name).
- **Admin state**: Set admin_state_up false to stop the pool from receiving traffic.
- **Tags**: Must be unique within the resource.
