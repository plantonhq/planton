# Hetzner Cloud Network

Creates a private network in Hetzner Cloud with subnets and optional static routes. The network provides isolated IPv4 connectivity between cloud resources using RFC 1918 addresses. Subnets are declared inline and can span multiple network zones for cross-region private connectivity. At least one subnet is required because servers and load balancers attach to subnets, not directly to the network.

## What Gets Created

- **Network** — an `hcloud_network` resource with a top-level CIDR block, standard labels computed from resource metadata, optional delete protection, and optional vSwitch route exposure.
- **Subnets** (1 per entry in `subnets`) — `hcloud_network_subnet` resources, each assigned to a network zone with a CIDR range carved from the network's address space. Three types are supported: `cloud` (standard), `server` (Robot dedicated servers), and `vswitch` (Robot vSwitch bridge).
- **Routes** (1 per entry in `routes`, optional) — `hcloud_network_route` resources defining static routing within the network. Created only when `routes` is non-empty.

## Prerequisites

- **Hetzner Cloud API token** configured via environment variable (`HCLOUD_TOKEN`) or OpenMCF provider config
- **A Hetzner Robot vSwitch** if using `vswitch`-type subnets (requires the vSwitch numeric ID)

## Quick Start

Create a file `network.yaml`:

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudNetwork
metadata:
  name: my-network
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.HetznerCloudNetwork.my-network
spec:
  ipRange: "10.0.0.0/16"
  subnets:
    - type: cloud
      networkZone: eu-central
      ipRange: "10.0.1.0/24"
```

Deploy:

```shell
openmcf apply -f network.yaml
```

This creates a private network with a single cloud subnet in the `eu-central` zone. Servers can be attached to this network to communicate over private IPs.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `ipRange` | `string` | CIDR block for the network. Must be one of the RFC 1918 private ranges (`10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`). All subnet ranges must fall within this block. | `min_len: 1` |
| `subnets` | `repeated Subnet` | Subnets within the network. At least one is required. | `min_items: 1` |
| `subnets[].type` | `enum` (`cloud`, `server`, `vswitch`) | Subnet type. `cloud` for standard cloud servers, `server` for Robot dedicated servers, `vswitch` for Robot vSwitch bridge. | Required, defined values only |
| `subnets[].networkZone` | `string` | Hetzner Cloud network zone. Known zones: `eu-central`, `us-east`, `us-west`, `ap-southeast`. | `min_len: 1` |
| `subnets[].ipRange` | `string` | CIDR range for the subnet. Must be a subset of the network's `ipRange` and must not overlap with other subnets or route destinations. | `min_len: 1` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `routes` | `repeated Route` | empty | Static routes within the network. Each route specifies a destination CIDR and a gateway IP within one of the network's subnets. |
| `routes[].destination` | `string` | — | Destination CIDR for the route. Must not overlap with subnet ranges. |
| `routes[].gateway` | `string` | — | Gateway IP address within one of the network's subnets. Cannot be the first IP of the network range or `172.31.1.1`. |
| `deleteProtection` | `bool` | `false` | Prevent accidental deletion of the network via the Hetzner Cloud API. |
| `exposeRoutesToVswitch` | `bool` | `false` | Expose the network's routes to vSwitch connections. Only takes effect when a vSwitch subnet is active. |
| `subnets[].vswitchId` | `int64` | — | Hetzner Robot vSwitch ID. Required when subnet type is `vswitch`, ignored otherwise. |

### Cross-Field Validation (CEL Rules)

| Constraint | Message |
|-----------|---------|
| `vswitchId` required for vswitch type | "vswitch_id is required when subnet type is vswitch" |

## Examples

### Minimal Single-Subnet Network

A single cloud subnet in `eu-central` — the simplest working network configuration.

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudNetwork
metadata:
  name: simple-net
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.HetznerCloudNetwork.simple-net
spec:
  ipRange: "10.0.0.0/16"
  subnets:
    - type: cloud
      networkZone: eu-central
      ipRange: "10.0.1.0/24"
```

### Multi-Zone Production Network

Two subnets in different zones for geographic redundancy, with delete protection enabled.

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudNetwork
metadata:
  name: multi-zone
  org: acme-corp
  env: production
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: infrastructure
    pulumi.openmcf.org/stack.name: production.HetznerCloudNetwork.multi-zone
spec:
  ipRange: "10.0.0.0/16"
  subnets:
    - type: cloud
      networkZone: eu-central
      ipRange: "10.0.1.0/24"
    - type: cloud
      networkZone: us-east
      ipRange: "10.0.2.0/24"
  deleteProtection: true
```

### Network with Static Routes

A network with custom routes directing traffic for a remote network through a VPN gateway server.

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudNetwork
metadata:
  name: routed-net
  org: acme-corp
  env: production
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: infrastructure
    pulumi.openmcf.org/stack.name: production.HetznerCloudNetwork.routed-net
spec:
  ipRange: "10.0.0.0/16"
  subnets:
    - type: cloud
      networkZone: eu-central
      ipRange: "10.0.1.0/24"
    - type: cloud
      networkZone: eu-central
      ipRange: "10.0.2.0/24"
  routes:
    - destination: "172.16.0.0/12"
      gateway: "10.0.1.1"
  deleteProtection: true
```

### Server Composition via valueFrom

A network referenced by a HetznerCloudServer using `valueFrom`. The server receives the network's numeric ID from the network's stack outputs, establishing a dependency edge in the deployment DAG.

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudNetwork
metadata:
  name: app-network
  org: acme-corp
  env: production
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: infrastructure
    pulumi.openmcf.org/stack.name: production.HetznerCloudNetwork.app-network
spec:
  ipRange: "10.0.0.0/16"
  subnets:
    - type: cloud
      networkZone: eu-central
      ipRange: "10.0.1.0/24"
  deleteProtection: true
```

The server references this network:

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudServer
metadata:
  name: app-01
  org: acme-corp
  env: production
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: infrastructure
    pulumi.openmcf.org/stack.name: production.HetznerCloudServer.app-01
spec:
  serverType: cx22
  image: ubuntu-24.04
  location: fsn1
  networkId:
    valueFrom:
      kind: HetznerCloudNetwork
      name: app-network
      fieldPath: status.outputs.network_id
```

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `network_id` | `string` | Hetzner Cloud numeric ID of the created network. Referenced by HetznerCloudServer and HetznerCloudLoadBalancer via `StringValueOrRef`. |

## Related Components

- [HetznerCloudServer](/docs/catalog/hetznercloud/hetznercloudserver) — Attaches to the network for private connectivity between servers
- [HetznerCloudLoadBalancer](/docs/catalog/hetznercloud/hetznercloudloadbalancer) — Attaches to the network to reach backend servers over private IPs
- [HetznerCloudFirewall](/docs/catalog/hetznercloud/hetznercloudfirewall) — Controls inbound/outbound traffic for servers on the network
- [HetznerCloudSshKey](/docs/catalog/hetznercloud/hetznercloudsshkey) — Foundation resource commonly deployed alongside networks for server provisioning
