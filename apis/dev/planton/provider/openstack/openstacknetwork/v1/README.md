# OpenStack Network

Provision and manage Neutron networks in OpenStack using Planton's unified API.

## Overview

A Neutron network is the foundational networking primitive in OpenStack. It represents an isolated Layer 2 broadcast domain. Subnets, ports, routers, security groups, floating IPs, and compute instances all attach to or reference a network.

This component creates an `openstack_networking_network_v2` resource through both Pulumi and Terraform IaC modules with full feature parity.

The network name is derived from `metadata.name`.

## Prerequisites

1. **OpenStack Cloud**: Access to an OpenStack deployment with Neutron (Networking service)
2. **Credentials**: OpenStack credentials configured via the credential management system
3. **Planton CLI**: Install from [planton.dev](https://planton.dev)

## Quick Start

### Minimal Network

Create a simple private network with all defaults:

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackNetwork
metadata:
  name: dev-network
spec: {}
```

### Network with DNS Domain

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackNetwork
metadata:
  name: app-network
spec:
  description: "Application network with DNS integration"
  dns_domain: "app.internal.example.com."
  mtu: 1450
```

### Deploy

```bash
planton apply --manifest network.yaml \
  -p openstack-creds.yaml
```

## Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `description` | string | No | Human-readable description of the network |
| `admin_state_up` | bool | No | Administrative state. Default: `true` |
| `shared` | bool | No | Whether shared across tenants (admin-only). Default: `false` |
| `external` | bool | No | Whether this is an external/provider network (admin-only). Default: `false` |
| `mtu` | int32 | No | Maximum Transmission Unit in bytes. 0 = OpenStack default |
| `dns_domain` | string | No | DNS domain for the network. Must end with `.` if set |
| `port_security_enabled` | bool | No | Whether port security is enforced. Omit to use deployment default |
| `tags` | string[] | No | Tags applied to the OpenStack resource. Must be unique |
| `region` | string | No | Override the region from the provider config |

## Outputs

| Field | Description |
|-------|-------------|
| `network_id` | The unique identifier (UUID) of the network |
| `name` | The network name (from `metadata.name`) |
| `region` | Region where the network was created |

## IaC Implementations

This component is implemented with both Pulumi (Go) and Terraform (HCL) with full feature parity.

- **Terraform resource**: `openstack_networking_network_v2`
- **Pulumi resource**: `openstack.networking.Network`

## Notes

- **Admin-only fields**: `shared` and `external` typically require admin privileges. Tenant users creating private networks can leave both as `false` (the default).
- **Port security**: When omitted, the OpenStack deployment's default applies (typically `true`). Set explicitly only when you need to override the deployment default.
- **DNS integration**: Requires the `dns-integration` Neutron extension. The `dns_domain` must end with a dot (e.g., `"my-domain.example.com."`).
- **MTU**: For VXLAN overlay networks, 1450 is common. For standard Ethernet, 1500. For jumbo frames, 9000.
