---
title: "Subnet"
description: "Subnet deployment documentation"
icon: "package"
order: 100
componentName: "openstacksubnet"
---

# OpenStack Subnet

Deploys an OpenStack Neutron subnet within a network, providing IP address allocation via a CIDR block with configurable gateway, DHCP, DNS, and allocation pool settings.

## What Gets Created

When you deploy an OpenStackSubnet resource, OpenMCF provisions:

- **Neutron Subnet** — an `openstack.networking.Subnet` resource with the configured CIDR, IP version, gateway, DHCP settings, DNS nameservers, allocation pools, and tags

## Prerequisites

- **OpenStack credentials** configured via environment variables or OpenMCF provider config
- **An existing Neutron network** — every subnet belongs to exactly one network, referenced by `networkId`

## Quick Start

Create a file `subnet.yaml`:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSubnet
metadata:
  name: my-subnet
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OpenStackSubnet.my-subnet
spec:
  networkId:
    value: "<network-uuid>"
  cidr: "192.168.1.0/24"
```

Deploy:

```shell
openmcf apply -f subnet.yaml
```

This creates a Neutron subnet named `my-subnet` on the specified network with a `/24` CIDR, IPv4, DHCP enabled, and an auto-assigned gateway.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `networkId` | `StringValueOrRef` | ID of the network this subnet belongs to. Can reference an OpenStackNetwork resource via `valueFrom`. | Required |
| `cidr` | `string` | IP address range in CIDR notation (e.g., `192.168.1.0/24` or `2001:db8::/64`). | Required. Must match CIDR format. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ipVersion` | `int32` | `4` | IP protocol version. Must be `4` (IPv4) or `6` (IPv6). |
| `gatewayIp` | `string` | auto-assigned | IP address of the subnet gateway. If omitted and `noGateway` is false, OpenStack assigns the first usable IP. Mutually exclusive with `noGateway`. |
| `noGateway` | `bool` | `false` | Disables the gateway on this subnet. Use for isolated subnets (e.g., storage networks). Mutually exclusive with `gatewayIp`. |
| `enableDhcp` | `bool` | `true` | Controls whether DHCP is enabled. When enabled, OpenStack's DHCP agent assigns IPs to ports on this subnet. |
| `dnsNameservers` | `string[]` | `[]` | DNS server IP addresses pushed to instances via DHCP. |
| `allocationPools` | `AllocationPool[]` | entire CIDR | Sub-ranges of the CIDR from which IPs are allocated. Each pool has a `start` and `end` IP. If omitted, the entire CIDR minus gateway and broadcast addresses is used. |
| `description` | `string` | — | Human-readable description visible in the OpenStack API and Horizon. |
| `tags` | `string[]` | `[]` | Tags for filtering and organization in the OpenStack API. Must be unique. |
| `region` | `string` | provider default | Overrides the region from the provider config for this subnet. |

**AllocationPool object:**

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `start` | `string` | First IP address in the allocation range (e.g., `192.168.1.100`). | Required |
| `end` | `string` | Last IP address in the allocation range (e.g., `192.168.1.200`). | Required |

## Examples

### Basic Subnet

A subnet with default settings on an existing network, suitable for development environments:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSubnet
metadata:
  name: dev-subnet
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OpenStackSubnet.dev-subnet
spec:
  networkId:
    value: "abc12345-def6-7890-abcd-ef1234567890"
  cidr: "192.168.1.0/24"
  description: Development subnet
```

### Subnet with DNS and Custom Gateway

A subnet referencing a managed OpenStackNetwork resource, with custom DNS servers and an explicit gateway:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSubnet
metadata:
  name: app-subnet
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.OpenStackSubnet.app-subnet
spec:
  networkId:
    valueFrom:
      name: app-network
  cidr: "10.0.0.0/16"
  gatewayIp: "10.0.0.1"
  dnsNameservers:
    - "8.8.8.8"
    - "8.8.4.4"
  tags:
    - staging
    - app-tier
```

### Full-Featured Subnet with Allocation Pools

A production subnet with allocation pools to reserve IP ranges, DNS servers, tags, and a specific region:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSubnet
metadata:
  name: prod-subnet
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OpenStackSubnet.prod-subnet
spec:
  networkId:
    valueFrom:
      name: prod-network
  cidr: "10.100.0.0/16"
  ipVersion: 4
  gatewayIp: "10.100.0.1"
  enableDhcp: true
  dnsNameservers:
    - "10.100.0.10"
    - "8.8.8.8"
  allocationPools:
    - start: "10.100.1.0"
      end: "10.100.127.255"
    - start: "10.100.200.0"
      end: "10.100.254.255"
  description: Production application subnet with reserved ranges
  tags:
    - production
    - managed
  region: RegionOne
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `subnet_id` | `string` | UUID of the created Neutron subnet |
| `name` | `string` | Name of the subnet, derived from `metadata.name` |
| `cidr` | `string` | CIDR block of the subnet |
| `gateway_ip` | `string` | Gateway IP address of the subnet (empty if `noGateway` was set) |
| `network_id` | `string` | ID of the parent network |
| `region` | `string` | OpenStack region where the subnet was created |

## Related Components

- [OpenStackNetwork](/docs/catalog/openstack/network) — the parent network that this subnet belongs to
- [OpenStackRouterInterface](/docs/catalog/openstack/router-interface) — attaches a subnet to a router for inter-network routing
- [OpenStackLoadBalancer](/docs/catalog/openstack/load-balancer) — places a load balancer VIP on a subnet
- [OpenStackLoadBalancerMember](/docs/catalog/openstack/load-balancer-member) — registers backend members on a subnet
- [OpenStackContainerClusterTemplate](/docs/catalog/openstack/container-cluster-template) — uses a subnet as the fixed network for container clusters
- [OpenStackInstance](/docs/catalog/openstack/instance) — attaches compute instances to networks via subnets
