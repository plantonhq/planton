# OpenStack Load Balancer Listener

Provision and manage Octavia listeners in OpenStack using OpenMCF's unified API.

## Overview

An Octavia listener defines a protocol and port combination on a load balancer that accepts incoming traffic. Each listener is associated with exactly one load balancer and forwards traffic to a backend pool. For TLS termination, the listener references a Barbican secret container holding the certificate.

This component creates an `openstack_lb_listener_v2` resource through both Pulumi and Terraform IaC modules with full feature parity.

The listener name is derived from `metadata.name`.

## Prerequisites

1. **OpenStack Cloud**: Access to an OpenStack deployment with Octavia (Load Balancing service)
2. **Credentials**: OpenStack credentials configured via the credential management system
3. **OpenMCF CLI**: Install from [openmcf.org](https://openmcf.org)
4. **Load Balancer**: An existing OpenStack load balancer (see `OpenStackLoadBalancer`)

## Quick Start

### Minimal HTTP Listener

Create a listener on an existing load balancer with a literal ID:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerListener
metadata:
  name: http-listener
spec:
  loadbalancer_id:
    value: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  protocol: "HTTP"
  protocol_port: 80
```

### Listener with Foreign Key Reference

Reference a load balancer managed by OpenMCF using `value_from`:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerListener
metadata:
  name: http-listener
spec:
  loadbalancer_id:
    value_from:
      name: app-lb
  protocol: "HTTP"
  protocol_port: 80
  description: "HTTP listener for web application"
```

### Deploy

```bash
openmcf apply --manifest listener.yaml \
  -p openstack-creds.yaml
```

## Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `loadbalancer_id` | StringValueOrRef | Yes | Load balancer to attach to (FK to OpenStackLoadBalancer) |
| `protocol` | string | Yes | Protocol: HTTP, HTTPS, TCP, UDP, TERMINATED_HTTPS. ForceNew |
| `protocol_port` | int32 | Yes | Port number (1-65535). ForceNew |
| `description` | string | No | Human-readable description |
| `connection_limit` | int32 | No | Max connections. -1 = unlimited. Default: Octavia default |
| `default_tls_container_ref` | string | No | Barbican TLS secret URI. Required for TERMINATED_HTTPS |
| `insert_headers` | map<string,string> | No | HTTP headers to insert (e.g., X-Forwarded-For) |
| `allowed_cidrs` | string[] | No | CIDRs allowed to access this listener |
| `admin_state_up` | bool | No | Administrative state. Default: `true` |
| `tags` | string[] | No | Tags applied to the OpenStack resource. Must be unique |
| `region` | string | No | Override the region from the provider config |

## Outputs

| Field | Description |
|-------|-------------|
| `listener_id` | The unique identifier (UUID) of the listener |
| `name` | The listener name (from `metadata.name`) |
| `protocol` | The protocol the listener accepts |
| `protocol_port` | The port on which the listener accepts traffic |
| `region` | Region where the listener was created |

## Foreign Key Relationships

**Inbound (this component references):**

- `loadbalancer_id` -> `OpenStackLoadBalancer.status.outputs.loadbalancer_id`

**Outbound (referenced by downstream):**

- `OpenStackLoadBalancerPool.spec.listener_id` -> `listener_id`

## IaC Implementations

This component is implemented with both Pulumi (Go) and Terraform (HCL) with full feature parity.

- **Terraform resource**: `openstack_lb_listener_v2`
- **Pulumi resource**: `openstack.loadbalancer.Listener`

## Notes

- **Protocol selection**: Choose HTTP for unencrypted web traffic, HTTPS for pass-through TLS, TCP for raw TCP, UDP for DNS/gaming, and TERMINATED_HTTPS for TLS termination at the load balancer.
- **TLS termination**: When using TERMINATED_HTTPS, provide a `default_tls_container_ref` pointing to a Barbican secret container with the certificate, private key, and optional intermediates.
- **Insert headers**: Only applicable to HTTP and TERMINATED_HTTPS protocols. Common headers: `X-Forwarded-For`, `X-Forwarded-Proto`, `X-Forwarded-Port`.
- **Allowed CIDRs**: When set, acts as a whitelist -- only traffic from listed CIDRs reaches the listener. When empty, all traffic is allowed.
- **Connection limit**: Set to -1 for unlimited connections. The default (unset) uses the Octavia default for the load balancer's flavor.
- **Admin state**: Set `admin_state_up: false` to temporarily stop accepting traffic without deleting the listener.
- **Tags**: Must be unique within the resource. Used for filtering and organization in OpenStack.
