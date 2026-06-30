---
title: "Router Interface"
description: "Router Interface deployment documentation"
icon: "package"
order: 100
componentName: "openstackrouterinterface"
---

# OpenStack Router Interface

Deploys an OpenStack Neutron router interface, attaching a router to a subnet by creating a port on the subnet and binding it to the router. This is the join between Layer 2 (subnet) and Layer 3 (router) — without it, a subnet has no route to other subnets or to external networks.

## What Gets Created

When you deploy an OpenStackRouterInterface resource, Planton provisions:

- **Neutron Router Interface** — an `openstack_networking_router_interface_v2` resource that creates a port on the specified subnet and attaches it to the specified router, establishing L3 connectivity for the subnet

## Prerequisites

- **OpenStack credentials** configured via environment variables or Planton provider config
- **An existing router** — either an OpenStackRouter resource or a pre-existing router UUID
- **An existing subnet** — either an OpenStackSubnet resource or a pre-existing subnet UUID

## Quick Start

Create a file `router-interface.yaml`:

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackRouterInterface
metadata:
  name: my-router-interface
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OpenStackRouterInterface.my-router-interface
spec:
  routerId:
    value: <router-uuid>
  subnetId:
    value: <subnet-uuid>
```

Deploy:

```shell
planton apply -f router-interface.yaml
```

This attaches the specified router to the specified subnet, enabling L3 routing for instances on that subnet.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `routerId` | `StringValueOrRef` | ID of the router to attach the subnet to. Can reference an OpenStackRouter resource via `valueFrom`. | required |
| `subnetId` | `StringValueOrRef` | ID of the subnet to connect to the router. Can reference an OpenStackSubnet resource via `valueFrom`. | required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `region` | `string` | provider default | Overrides the region from the provider config for this router interface. |

## Examples

### Basic with Literal UUIDs

Attach a pre-existing router to a pre-existing subnet using their UUIDs directly:

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackRouterInterface
metadata:
  name: web-subnet-attachment
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OpenStackRouterInterface.web-subnet-attachment
spec:
  routerId:
    value: 3a1f2b4c-5d6e-7f8a-9b0c-1d2e3f4a5b6c
  subnetId:
    value: 7c8d9e0f-1a2b-3c4d-5e6f-7a8b9c0d1e2f
```

### Foreign Key References

Wire the router interface to OpenStackRouter and OpenStackSubnet resources in the same deployment using `valueFrom`, so the IDs are resolved automatically:

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackRouterInterface
metadata:
  name: app-subnet-to-edge
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: staging.OpenStackRouterInterface.app-subnet-to-edge
spec:
  routerId:
    valueFrom:
      kind: OpenStackRouter
      name: edge-router
      fieldPath: status.outputs.router_id
  subnetId:
    valueFrom:
      kind: OpenStackSubnet
      name: app-subnet
      fieldPath: status.outputs.subnet_id
```

### Multi-Region with Explicit Region

Attach a router to a subnet in a specific OpenStack region, overriding the provider default:

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackRouterInterface
metadata:
  name: region2-db-subnet
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.OpenStackRouterInterface.region2-db-subnet
spec:
  routerId:
    valueFrom:
      kind: OpenStackRouter
      name: region2-router
      fieldPath: status.outputs.router_id
  subnetId:
    valueFrom:
      kind: OpenStackSubnet
      name: db-subnet
      fieldPath: status.outputs.subnet_id
  region: RegionTwo
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `port_id` | `string` | UUID of the port auto-created by the router interface attachment. This is also the Terraform resource ID. |
| `router_id` | `string` | UUID of the router this interface is attached to |
| `subnet_id` | `string` | UUID of the subnet connected to the router |
| `region` | `string` | OpenStack region where the router interface was created |

## Related Components

- [OpenStackRouter](/docs/catalog/openstack/router) — the L3 router that this interface attaches to
- [OpenStackSubnet](/docs/catalog/openstack/subnet) — the subnet that this interface connects to the router
- [OpenStackNetwork](/docs/catalog/openstack/network) — the Layer 2 network that subnets belong to
- [OpenStackFloatingIp](/docs/catalog/openstack/floating-ip) — allocates floating IPs from the external network reachable via the router
- [OpenStackInstance](/docs/catalog/openstack/instance) — compute instances that use the subnet's default gateway provided by the router interface
