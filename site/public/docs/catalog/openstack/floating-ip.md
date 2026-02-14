---
title: "Floating IP"
description: "Floating IP deployment documentation"
icon: "package"
order: 100
componentName: "openstackfloatingip"
---

# OpenStack Floating IP

Deploys an OpenStack Neutron floating IP, allocating an external (public) IP address from a provider network. The floating IP can optionally be associated with a port for immediate external connectivity, or allocated standalone for later association via the separate OpenStackFloatingIpAssociate component.

## What Gets Created

When you deploy an OpenStackFloatingIp resource, OpenMCF provisions:

- **Neutron Floating IP** — an `openstack_networking_floatingip_v2` resource allocated from the specified external network, with optional built-in port association, specific address reservation, and tags

## Prerequisites

- **OpenStack credentials** configured via environment variables or OpenMCF provider config
- **An existing external (provider) network** — provided as a literal UUID or via `valueFrom` reference to an OpenStackNetwork resource
- **An existing port** if using built-in association via `portId`

## Quick Start

Create a file `floating-ip.yaml`:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackFloatingIp
metadata:
  name: my-fip
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OpenStackFloatingIp.my-fip
spec:
  floatingNetworkId:
    value: "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
```

Deploy:

```shell
openmcf apply -f floating-ip.yaml
```

This allocates a floating IP from the specified external network. The IP is not associated with any port (allocation-only mode). Use the `address` output for DNS configuration, firewall rules, or later association.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `floatingNetworkId` | `StringValueOrRef` | ID of the external (provider) network from which the floating IP is allocated. This is the pool of public IP addresses. Can reference an OpenStackNetwork resource via `valueFrom`. | required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `portId` | `StringValueOrRef` | — | ID of an existing port to associate with this floating IP. When set, the floating IP is immediately bound to the port, providing external connectivity to whatever is attached to that port. Omit for allocation-only mode. Can reference an OpenStackNetworkPort resource via `valueFrom`. |
| `fixedIp` | `string` | — | Fixed IP address on the port to associate the floating IP with. Only relevant when `portId` is set and the port has multiple IP addresses. If the port has a single IP, this can be omitted. Validation: can only be set when `portId` is configured. |
| `subnetId` | `string` | — | UUID of a subnet within the external network from which to allocate the floating IP. References an admin-managed subnet on the provider network. If omitted, OpenStack allocates from any available subnet. |
| `address` | `string` | — | Requests a specific floating IP address from the pool. If omitted, OpenStack allocates any available address. ForceNew: changing this value destroys and recreates the resource. Use for DNS pre-configuration, firewall whitelisting, or IP reservation. |
| `description` | `string` | — | Human-readable description of the floating IP, visible in the OpenStack API and Horizon. |
| `tags` | `string[]` | `[]` | Tags for filtering and organization in the OpenStack API and Horizon dashboard. Must be unique. |
| `region` | `string` | provider default | Overrides the region from the provider config for this floating IP. |

## Examples

### Allocation-Only Floating IP

A floating IP allocated from an external network with no port association. Suitable for reserving a public IP before the target port or instance exists:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackFloatingIp
metadata:
  name: web-fip
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OpenStackFloatingIp.web-fip
spec:
  floatingNetworkId:
    value: "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
  description: Web server floating IP (allocation only)
```

### Floating IP with Port Association

A floating IP allocated and immediately associated with a port, using `valueFrom` references to other OpenMCF resources for both the external network and the target port:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackFloatingIp
metadata:
  name: app-fip
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.OpenStackFloatingIp.app-fip
spec:
  floatingNetworkId:
    valueFrom:
      kind: OpenStackNetwork
      name: external-net
      fieldPath: status.outputs.network_id
  portId:
    valueFrom:
      kind: OpenStackNetworkPort
      name: app-port
      fieldPath: status.outputs.port_id
  description: Application server floating IP
  tags:
    - staging
    - app-tier
```

### Reserved IP Address

A floating IP with a specific address requested from the pool, useful for DNS pre-configuration or firewall whitelisting where the IP must be known before allocation:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackFloatingIp
metadata:
  name: lb-fip
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OpenStackFloatingIp.lb-fip
spec:
  floatingNetworkId:
    valueFrom:
      kind: OpenStackNetwork
      name: external-net
      fieldPath: status.outputs.network_id
  address: "203.0.113.42"
  portId:
    valueFrom:
      kind: OpenStackNetworkPort
      name: lb-port
      fieldPath: status.outputs.port_id
  description: Load balancer floating IP with reserved address
  region: RegionOne
  tags:
    - production
    - load-balancer
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `floating_ip_id` | `string` | UUID of the floating IP resource in OpenStack. |
| `address` | `string` | The allocated floating IP address (e.g., `203.0.113.42`). Primary output for DNS configuration, firewall rules, and FK target for OpenStackFloatingIpAssociate. |
| `floating_network_id` | `string` | UUID of the external network the floating IP was allocated from. |
| `port_id` | `string` | UUID of the port this floating IP is associated with. Empty if allocation-only mode. |
| `fixed_ip` | `string` | Fixed IP address on the port that the floating IP is mapped to. Empty if no port association exists. |
| `region` | `string` | OpenStack region where the floating IP was allocated. |

## Related Components

- [OpenStackNetwork](/docs/catalog/openstack/openstacknetwork) — the external (provider) network from which floating IPs are allocated
- [OpenStackNetworkPort](/docs/catalog/openstack/openstacknetworkport) — the port that a floating IP can be associated with for external connectivity
- [OpenStackFloatingIpAssociate](/docs/catalog/openstack/openstackfloatingipassociate) — associates an existing floating IP with a port as a separate DAG node in InfraCharts
- [OpenStackRouter](/docs/catalog/openstack/openstackrouter) — provides routing between tenant and external networks
- [OpenStackInstance](/docs/catalog/openstack/openstackinstance) — compute instances that gain external connectivity through floating IPs
