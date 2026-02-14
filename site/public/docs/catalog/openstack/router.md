---
title: "Router"
description: "Router deployment documentation"
icon: "package"
order: 100
componentName: "openstackrouter"
---

# OpenStack Router

Deploys an OpenStack Neutron router, providing L3 routing between tenant subnets and, optionally, external network connectivity via SNAT/DNAT. Routers are the backbone of OpenStack networking — they connect isolated subnets to each other and to the outside world.

## What Gets Created

When you deploy an OpenStackRouter resource, OpenMCF provisions:

- **Neutron Router** — an `openstack_networking_router_v2` resource with the configured external gateway, SNAT settings, DVR mode, external fixed IPs, and tags

## Prerequisites

- **OpenStack credentials** configured via environment variables or OpenMCF provider config
- **External network UUID** if connecting the router to an external (provider) network for internet access — this network is typically created by a cloud administrator
- **Admin privileges** if setting `distributed` mode on deployments that restrict DVR to admin users

## Quick Start

Create a file `router.yaml`:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackRouter
metadata:
  name: my-router
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OpenStackRouter.my-router
spec: {}
```

Deploy:

```shell
openmcf apply -f router.yaml
```

This creates a Neutron router named `my-router` with default settings: admin state up and no external gateway (internal routing only).

## Configuration Reference

### Required Fields

All spec fields are optional. The router name is derived from `metadata.name`.

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `externalNetworkId` | `StringValueOrRef` | — | ID of the external (provider) network used as the router's gateway. When set, the router gains external connectivity and can perform SNAT. Can reference an OpenStackNetwork resource via `valueFrom`. |
| `adminStateUp` | `bool` | `true` | Administrative state of the router. When `false`, the router is disabled and does not forward traffic. |
| `enableSnat` | `bool` | platform default | Controls whether Source NAT is enabled on the router's external gateway. Only valid when `externalNetworkId` is configured. |
| `distributed` | `bool` | platform default | Controls whether the router uses Distributed Virtual Router (DVR) mode. DVR distributes routing to each compute node, eliminating the centralized L3 agent bottleneck. Create-time setting only — cannot be changed after creation. |
| `externalFixedIps` | `ExternalFixedIp[]` | `[]` | Fixed IP addresses to allocate on the external network for the router's gateway. Only valid when `externalNetworkId` is configured. If omitted, OpenStack auto-allocates. |
| `externalFixedIps[].subnetId` | `string` | — | UUID of a subnet on the external network from which to allocate the IP. |
| `externalFixedIps[].ipAddress` | `string` | — | Specific IP address to allocate on the external network. Must be within the range of the specified subnet. |
| `description` | `string` | — | Human-readable description of the router, visible in the OpenStack API and Horizon. |
| `tags` | `string[]` | `[]` | Tags for filtering and organization in the OpenStack API. Must be unique. |
| `region` | `string` | provider default | Overrides the region from the provider config for this router. |

**Validation rules:**

- `enableSnat` can only be set when `externalNetworkId` is configured.
- `externalFixedIps` can only be specified when `externalNetworkId` is configured.

## Examples

### Internal Router

A router without an external gateway, providing routing between tenant subnets only:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackRouter
metadata:
  name: internal-router
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OpenStackRouter.internal-router
spec:
  description: Internal routing between dev subnets
  tags:
    - dev
    - internal
```

### Router with External Gateway

A router connected to an external network for internet access, referencing the network by UUID:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackRouter
metadata:
  name: gateway-router
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.OpenStackRouter.gateway-router
spec:
  externalNetworkId:
    value: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  enableSnat: true
  description: Staging router with external connectivity
  tags:
    - staging
    - gateway
```

### DVR Router with Foreign Key Reference

A distributed router that references an OpenStackNetwork resource for its external gateway using `valueFrom`, with a specific external IP allocation:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackRouter
metadata:
  name: prod-router
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OpenStackRouter.prod-router
spec:
  externalNetworkId:
    valueFrom:
      kind: OpenStackNetwork
      name: external-net
      fieldPath: status.outputs.network_id
  enableSnat: true
  distributed: true
  externalFixedIps:
    - subnetId: "f1e2d3c4-b5a6-7890-abcd-ef1234567890"
      ipAddress: "203.0.113.10"
  description: Production DVR router with dedicated external IP
  tags:
    - prod
    - dvr
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `routerId` | `string` | UUID of the created Neutron router. Primary output used as a foreign key by downstream components. |
| `name` | `string` | Name of the router, derived from `metadata.name`. |
| `externalNetworkId` | `string` | ID of the external network used as the router's gateway. Empty if no external gateway is configured. |
| `externalGatewayIp` | `string` | Primary external IP address allocated to the router's gateway. Empty if no external gateway is configured. |
| `region` | `string` | OpenStack region where the router was created. |

## Related Components

- [OpenStackNetwork](/docs/catalog/openstack/network) — provides the Layer 2 network that the router connects to as an external gateway
- [OpenStackSubnet](/docs/catalog/openstack/subnet) — defines IP address ranges on networks; subnets are attached to routers via router interfaces
- [OpenStackRouterInterface](/docs/catalog/openstack/router-interface) — attaches a subnet to this router, enabling routing for that subnet's traffic
- [OpenStackFloatingIp](/docs/catalog/openstack/floating-ip) — allocates floating IPs from the external network for 1:1 NAT to instances
- [OpenStackSecurityGroup](/docs/catalog/openstack/security-group) — controls traffic filtering for ports on networks connected to this router
- [OpenStackInstance](/docs/catalog/openstack/instance) — compute instances whose traffic is routed by this router
