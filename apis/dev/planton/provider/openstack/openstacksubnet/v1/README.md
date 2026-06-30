# OpenStack Subnet

Provision and manage Neutron subnets in OpenStack using Planton's unified API.

## Overview

A Neutron subnet provides IP address allocation within a network. It defines a CIDR block, gateway, DHCP settings, and DNS configuration. Every OpenStack workload that needs IP connectivity requires at least one subnet attached to a network.

This component creates an `openstack_networking_subnet_v2` resource through both Pulumi and Terraform IaC modules with full feature parity.

The subnet name is derived from `metadata.name`.

## Prerequisites

1. **OpenStack Cloud**: Access to an OpenStack deployment with Neutron (Networking service)
2. **Credentials**: OpenStack credentials configured via the credential management system
3. **Planton CLI**: Install from [planton.dev](https://planton.dev)
4. **Network**: An existing OpenStack network (see `OpenStackNetwork`)

## Quick Start

### Minimal Subnet

Create a subnet on an existing network with a literal network ID:

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackSubnet
metadata:
  name: dev-subnet
spec:
  network_id:
    value: "e0a1f622-9aab-4a48-8c8c-3b0c7e2a9b1d"
  cidr: "192.168.1.0/24"
```

### Subnet with Foreign Key Reference

Reference a network managed by Planton using `value_from`:

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackSubnet
metadata:
  name: dev-subnet
spec:
  network_id:
    value_from:
      name: dev-network
  cidr: "10.0.0.0/16"
  dns_nameservers:
    - "8.8.8.8"
    - "8.8.4.4"
```

### Deploy

```bash
planton apply --manifest subnet.yaml \
  -p openstack-creds.yaml
```

## Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `network_id` | StringValueOrRef | Yes | ID of the parent network (FK to OpenStackNetwork) |
| `cidr` | string | Yes | IP range in CIDR notation (e.g., `192.168.1.0/24`) |
| `ip_version` | int32 | No | IP protocol version: `4` or `6`. Default: `4` |
| `gateway_ip` | string | No | Specific gateway IP. Mutually exclusive with `no_gateway` |
| `no_gateway` | bool | No | Disable gateway entirely. Mutually exclusive with `gateway_ip` |
| `enable_dhcp` | bool | No | Enable DHCP on the subnet. Default: `true` |
| `dns_nameservers` | string[] | No | DNS server IPs pushed to instances via DHCP |
| `allocation_pools` | AllocationPool[] | No | IP allocation sub-ranges within the CIDR |
| `description` | string | No | Human-readable description of the subnet |
| `tags` | string[] | No | Tags applied to the OpenStack resource. Must be unique |
| `region` | string | No | Override the region from the provider config |

### AllocationPool

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `start` | string | Yes | First IP address in the range |
| `end` | string | Yes | Last IP address in the range |

## Outputs

| Field | Description |
|-------|-------------|
| `subnet_id` | The unique identifier (UUID) of the subnet |
| `name` | The subnet name (from `metadata.name`) |
| `cidr` | The CIDR block |
| `gateway_ip` | The gateway IP (computed or specified; empty if `no_gateway`) |
| `network_id` | The parent network ID |
| `region` | Region where the subnet was created |

## Foreign Key Relationships

**Inbound (this component references):**

- `network_id` -> `OpenStackNetwork.status.outputs.network_id`

**Outbound (referenced by downstream):**

- `OpenStackRouterInterface.spec.subnet_id` -> `subnet_id`
- `OpenStackLoadBalancer.spec.vip_subnet_id` -> `subnet_id`
- `OpenStackLoadBalancerMember.spec.subnet_id` -> `subnet_id`
- `OpenStackContainerClusterTemplate.spec.fixed_subnet` -> `subnet_id`

## IaC Implementations

This component is implemented with both Pulumi (Go) and Terraform (HCL) with full feature parity.

- **Terraform resource**: `openstack_networking_subnet_v2`
- **Pulumi resource**: `openstack.networking.Subnet`

## Notes

- **Gateway behavior**: By default, OpenStack assigns the first usable IP in the CIDR as the gateway. Use `gateway_ip` to override, or `no_gateway: true` for isolated subnets.
- **DHCP**: Enabled by default. Disable only for subnets where static IP assignment is managed externally.
- **Allocation pools**: If omitted, the entire CIDR (minus gateway and network/broadcast) is allocatable. Use pools to reserve ranges for static assignments or network appliances.
- **DNS nameservers**: Pushed to instances via DHCP option 6. Common values: `8.8.8.8` (Google), `1.1.1.1` (Cloudflare), or internal DNS servers.
- **IPv6**: Set `ip_version: 6` and use an IPv6 CIDR (e.g., `2001:db8::/64`). IPv6-specific options (`ipv6_address_mode`, `ipv6_ra_mode`) can be added later.
