# OpenStack Network

Deploys an OpenStack Neutron network, providing an isolated Layer 2 broadcast domain that serves as the foundation for subnets, ports, routers, and instance attachments.

## What Gets Created

When you deploy an OpenStackNetwork resource, OpenMCF provisions:

- **Neutron Network** — an `openstack_networking_network_v2` resource with the configured administrative state, MTU, port security settings, and optional DNS domain integration

## Prerequisites

- **OpenStack credentials** configured via environment variables or OpenMCF provider config
- **Admin privileges** if creating shared networks (`shared: true`) or external/provider networks (`external: true`)
- **DNS integration extension** enabled in Neutron if using `dnsDomain`

## Quick Start

Create a file `network.yaml`:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackNetwork
metadata:
  name: my-network
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OpenStackNetwork.my-network
spec: {}
```

Deploy:

```shell
openmcf apply -f network.yaml
```

This creates a Neutron network named `my-network` with default settings: admin state up, port security enabled, and standard MTU.

## Configuration Reference

### Required Fields

All spec fields are optional. The network name is derived from `metadata.name`.

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | — | Human-readable description of the network, visible in the OpenStack API and Horizon. |
| `adminStateUp` | `bool` | `true` | Administrative state of the network. When `false`, the network is down and does not forward traffic. |
| `shared` | `bool` | `false` | When `true`, the network is shared across all tenants/projects. Requires admin privileges. |
| `external` | `bool` | `false` | When `true`, marks this as an external (provider) network used for floating IP allocation and router gateways. Requires admin privileges. |
| `mtu` | `int` | platform default | Maximum Transmission Unit in bytes. Common values: `1500` (standard Ethernet), `1450` (VXLAN overlay), `9000` (jumbo frames). |
| `dnsDomain` | `string` | — | DNS domain for auto-assigned port DNS names. Must end with a dot (e.g., `my-network.example.com.`). Requires the dns-integration Neutron extension. |
| `portSecurityEnabled` | `bool` | platform default | Controls port security enforcement on ports created on this network. When enabled, only traffic matching security groups and allowed address pairs is permitted. |
| `tags` | `string[]` | `[]` | Tags for filtering and organization in the OpenStack API. Must be unique. |
| `region` | `string` | provider default | Overrides the region from the provider config for this network. |

## Examples

### Basic Network

A simple network with default settings, suitable for development environments:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackNetwork
metadata:
  name: dev-network
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OpenStackNetwork.dev-network
spec:
  description: Development environment network
```

### Network with Custom MTU and DNS

A network configured for VXLAN overlay with DNS integration for automatic port name resolution:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackNetwork
metadata:
  name: overlay-network
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.OpenStackNetwork.overlay-network
spec:
  description: VXLAN overlay network with DNS integration
  mtu: 1450
  dnsDomain: overlay.internal.example.com.
  tags:
    - staging
    - overlay
```

### External Provider Network

An admin-created external network for floating IP allocation and router gateways:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackNetwork
metadata:
  name: external-net
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OpenStackNetwork.external-net
spec:
  description: External provider network for floating IPs
  external: true
  shared: true
  portSecurityEnabled: false
  mtu: 1500
  tags:
    - external
    - provider
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `network_id` | `string` | UUID of the created Neutron network |
| `name` | `string` | Name of the network, derived from `metadata.name` |
| `region` | `string` | OpenStack region where the network was created |

## Related Components

- [OpenStackSubnet](/docs/catalog/openstack/openstacksubnet) — defines IP address ranges and DHCP settings on the network
- [OpenStackNetworkPort](/docs/catalog/openstack/openstacknetworkport) — creates ports with specific IPs and security groups on the network
- [OpenStackRouter](/docs/catalog/openstack/openstackrouter) — provides routing between networks and external connectivity
- [OpenStackInstance](/docs/catalog/openstack/openstackinstance) — attaches compute instances to the network
- [OpenStackFloatingIp](/docs/catalog/openstack/openstackfloatingip) — allocates floating IPs from external networks
