# OpenStackRouterInterface

Attaches an OpenStack Neutron router to a subnet, enabling L3 routing for that subnet's traffic.

## Overview

A router interface is a "join" resource -- it connects a router (L3) to a subnet (L2) by creating a port on the subnet and attaching it to the router. Without this, instances on a subnet can communicate with each other but have no route to other subnets or to the external network.

In the `openstack/developer-environment` InfraChart, the router interface is the piece that connects the developer's isolated subnet to the router that provides internet access.

## Key Features

- **Two Required Foreign Keys**: `router_id` references an OpenStackRouter, `subnet_id` references an OpenStackSubnet
- **StringValueOrRef**: Both FKs support literal UUIDs and InfraChart `value_from` references for DAG wiring
- **Pure Join Resource**: Only 3 spec fields -- the simplest OpenStack component
- **ForceNew**: All fields are immutable after creation. Any change recreates the interface.

## Minimal Example

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackRouterInterface
metadata:
  name: dev-router-subnet
spec:
  router_id:
    value: "router-uuid-here"
  subnet_id:
    value: "subnet-uuid-here"
```

## InfraChart Usage (value_from)

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackRouterInterface
metadata:
  name: dev-router-subnet
spec:
  router_id:
    value_from:
      name: dev-router
  subnet_id:
    value_from:
      name: dev-subnet
```

Both `value_from` references are resolved by the InfraChart DAG engine, wiring this interface to the named OpenStackRouter and OpenStackSubnet resources.

## Outputs

| Output | Description |
|--------|-------------|
| `port_id` | UUID of the auto-created port (also the TF resource ID) |
| `router_id` | UUID of the router |
| `subnet_id` | UUID of the subnet |
| `region` | OpenStack region |

## Terraform Resource

`openstack_networking_router_interface_v2`

## Pulumi Resource

`openstack.networking.RouterInterface`
