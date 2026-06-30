---
title: "Private Network"
description: "Private Network deployment documentation"
icon: "package"
order: 100
componentName: "scalewayprivatenetwork"
---

# Scaleway Private Network

Deploys a Scaleway Private Network inside an existing VPC, with optional IPv4/IPv6 subnet configuration and default route propagation. The Private Network serves as the primary attachment point for Kapsule clusters, RDB instances, Redis clusters, Load Balancers, and other Scaleway resources that require private connectivity.

## What Gets Created

When you deploy a ScalewayPrivateNetwork resource, Planton provisions:

- **Private Network** — a `network.PrivateNetwork` resource attached to the specified VPC, with built-in DHCP and IPAM-managed addressing
- **IPv4 Subnet** — either the user-specified CIDR or an auto-allocated subnet from Scaleway's IPAM service
- **IPv6 Subnets** — created only when `ipv6Subnets` entries are provided
- **Scaleway Tags** — resource kind, name, organization, and environment labels applied as flat `key=value` tags

## Prerequisites

- **Scaleway credentials** configured via environment variables or Planton provider config
- **An existing VPC** in the target region — either a literal VPC UUID or an Planton-managed ScalewayVpc resource whose output can be referenced via `valueFrom`

## Quick Start

Create a file `private-network.yaml`:

```yaml
apiVersion: scaleway.planton.dev/v1
kind: ScalewayPrivateNetwork
metadata:
  name: my-network
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.ScalewayPrivateNetwork.my-network
spec:
  vpcId: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
  region: fr-par
```

Deploy:

```shell
planton apply -f private-network.yaml
```

This creates a Private Network in the `fr-par` region with an IPAM-auto-allocated IPv4 subnet. The allocated CIDR is available in stack outputs as `ipv4_subnet_cidr`.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `vpcId` | `StringValueOrRef` | UUID of the VPC in which to create this Private Network. Can be a literal UUID or a `valueFrom` reference to a ScalewayVpc resource's `status.outputs.vpc_id`. The Private Network's region must match the VPC's region. | Required |
| `region` | `string` | Scaleway region for the Private Network (e.g., `"fr-par"`, `"nl-ams"`, `"pl-waw"`). Cannot be changed after creation. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ipv4Subnet` | `string` | Auto-allocated by IPAM | IPv4 subnet in CIDR notation (e.g., `"192.168.0.0/24"`, `"10.0.1.0/24"`). When multiple Private Networks share a VPC, specify non-overlapping ranges to ensure correct routing. |
| `ipv6Subnets` | `string[]` | `[]` | IPv6 subnets in CIDR notation (e.g., `"fd46:78ab:30b8:177c::/64"`). Multiple entries are supported for dual-stack networking. |
| `enableDefaultRoutePropagation` | `bool` | `false` | When `true`, resources in this Private Network receive the VPC's default routes, enabling communication with resources in other Private Networks within the same VPC. |

## Examples

### Minimal Private Network with Auto-Allocated Subnet

A Private Network with no explicit subnet — Scaleway's IPAM assigns one automatically:

```yaml
apiVersion: scaleway.planton.dev/v1
kind: ScalewayPrivateNetwork
metadata:
  name: dev-network
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.ScalewayPrivateNetwork.dev-network
spec:
  vpcId: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
  region: fr-par
```

### Private Network with Explicit IPv4 Subnet and Route Propagation

A Private Network with a controlled address range and cross-network routing enabled:

```yaml
apiVersion: scaleway.planton.dev/v1
kind: ScalewayPrivateNetwork
metadata:
  name: app-network
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.ScalewayPrivateNetwork.app-network
spec:
  vpcId: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
  region: fr-par
  ipv4Subnet: "10.0.1.0/24"
  enableDefaultRoutePropagation: true
```

### Dual-Stack Network with VPC Reference

A Private Network referencing an Planton-managed VPC, with both IPv4 and IPv6 subnets:

```yaml
apiVersion: scaleway.planton.dev/v1
kind: ScalewayPrivateNetwork
metadata:
  name: dual-stack-network
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.ScalewayPrivateNetwork.dual-stack-network
spec:
  vpcId:
    valueFrom:
      kind: ScalewayVpc
      name: main-vpc
      fieldPath: status.outputs.vpc_id
  region: nl-ams
  ipv4Subnet: "172.16.0.0/22"
  ipv6Subnets:
    - "fd46:78ab:30b8:177c::/64"
  enableDefaultRoutePropagation: true
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `private_network_id` | `string` | UUID of the created Private Network. This is the primary cross-resource reference consumed by downstream components (Kapsule clusters, RDB instances, Redis clusters, Load Balancers, etc.) via `valueFrom`. |
| `ipv4_subnet_cidr` | `string` | IPv4 CIDR of the subnet associated with this Private Network. Reflects the requested `ipv4Subnet` if specified, or the CIDR auto-allocated by Scaleway's IPAM service. |

## Related Components

- [ScalewayVpc](/docs/catalog/scaleway/vpc) — the parent VPC that contains this Private Network
- [ScalewayKapsuleCluster](/docs/catalog/scaleway/kapsule-cluster) — attaches Kubernetes clusters to this Private Network for private pod-to-service communication
- [ScalewayRdbInstance](/docs/catalog/scaleway/rdb-instance) — attaches managed database instances to this Private Network for private database connectivity
- [ScalewayRedisCluster](/docs/catalog/scaleway/redis-cluster) — attaches Redis clusters to this Private Network
- [ScalewayInstanceSecurityGroup](/docs/catalog/scaleway/instance-security-group) — controls network access for compute instances within this Private Network
