# OpenStackRouter

An OpenStack Neutron router that provides L3 routing between subnets and optionally connects to an external network for internet access.

## Overview

A router is the backbone of OpenStack networking. It connects tenant subnets to each other (east-west traffic) and to external/provider networks (north-south traffic) via SNAT and DNAT.

## Key Features

- **External Gateway**: Connect to a provider network for internet access via `external_network_id`
- **SNAT Control**: Enable or disable Source NAT on the external gateway
- **DVR Support**: Distributed Virtual Router mode eliminates the centralized L3 agent bottleneck
- **Fixed IP Control**: Request specific external IPs for predictable addressing
- **Foreign Key**: `external_network_id` uses `StringValueOrRef` -- works with literal UUIDs and InfraChart value_from references

## Minimal Example

An internal-only router (no external connectivity):

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackRouter
metadata:
  name: internal-router
spec: {}
```

## Router with External Gateway

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackRouter
metadata:
  name: edge-router
spec:
  external_network_id:
    value: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  enable_snat: true
```

## InfraChart Usage (value_from)

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackRouter
metadata:
  name: dev-router
spec:
  external_network_id:
    value_from:
      name: public-network
```

The `value_from` reference is resolved by the InfraChart DAG engine, wiring this router to the OpenStackNetwork named `public-network`.

## Outputs

| Output | Description |
|--------|-------------|
| `router_id` | UUID of the created router (FK for RouterInterface) |
| `name` | Router name |
| `external_network_id` | External network UUID (empty if internal-only) |
| `external_gateway_ip` | Primary external IP address (empty if internal-only) |
| `region` | OpenStack region |

## Terraform Resource

`openstack_networking_router_v2`

## Pulumi Resource

`openstack.networking.Router`
