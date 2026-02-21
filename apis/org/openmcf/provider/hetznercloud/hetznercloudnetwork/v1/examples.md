# HetznerCloudNetwork Examples

## Minimal Single-Subnet Network

The simplest working network: a single `/16` block with one cloud subnet in `eu-central`. Suitable for single-region deployments where all servers are in the same Hetzner data center.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudNetwork
metadata:
  name: simple-net
spec:
  ipRange: "10.0.0.0/16"
  subnets:
    - type: cloud
      networkZone: eu-central
      ipRange: "10.0.1.0/24"
```

---

## Multi-Zone Network

A network spanning two zones for geographic redundancy. Servers in `eu-central` and `us-east` can communicate over private IPs through the same network. Each zone gets its own subnet with non-overlapping CIDR ranges.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudNetwork
metadata:
  name: multi-zone
  org: acme-corp
  env: production
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

---

## Network with Static Routes

A network with a static route directing traffic for a remote network through a VPN gateway server within the network. The gateway server at `10.0.1.1` must have IP forwarding enabled at the OS level.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudNetwork
metadata:
  name: routed-net
  org: acme-corp
  env: production
  labels:
    topology: hub-and-spoke
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
    - destination: "192.168.100.0/24"
      gateway: "10.0.1.1"
  deleteProtection: true
```

---

## Hybrid Cloud with vSwitch

A network combining cloud servers and Hetzner Robot dedicated servers via a vSwitch subnet. The `cloud` subnet hosts regular cloud servers, while the `vswitch` subnet bridges to a Robot vSwitch (ID `12345`) for hybrid connectivity. The `exposeRoutesToVswitch` flag makes custom routes visible to the dedicated servers.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudNetwork
metadata:
  name: hybrid-net
  org: acme-corp
  env: production
  labels:
    topology: hybrid
spec:
  ipRange: "10.0.0.0/16"
  subnets:
    - type: cloud
      networkZone: eu-central
      ipRange: "10.0.1.0/24"
    - type: vswitch
      networkZone: eu-central
      ipRange: "10.0.10.0/24"
      vswitchId: 12345
  routes:
    - destination: "172.16.0.0/12"
      gateway: "10.0.1.1"
  exposeRoutesToVswitch: true
  deleteProtection: true
```

---

## InfraChart Composition with valueFrom

In an infra chart, a `HetznerCloudServer` references the network via `valueFrom` so the network ID is resolved from stack outputs. This eliminates hardcoded IDs and establishes a dependency edge in the DAG.

Network manifest:

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudNetwork
metadata:
  name: app-network
  org: acme-corp
  env: production
spec:
  ipRange: "10.0.0.0/16"
  subnets:
    - type: cloud
      networkZone: eu-central
      ipRange: "10.0.1.0/24"
  deleteProtection: true
```

Server manifest referencing the network output:

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudServer
metadata:
  name: app-01
  org: acme-corp
  env: production
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

The `valueFrom` reference ensures that:
1. The network (with its subnets) is created before the server
2. The correct numeric ID is passed to the server without manual lookup
3. Network topology changes (adding subnets, modifying routes) do not require updating the server manifest — only the network_id reference matters
