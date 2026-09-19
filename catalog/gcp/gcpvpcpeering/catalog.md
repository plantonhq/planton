# GCP VPC Peering

Manages one side of a Google Cloud VPC Network Peering: the peering entry on your network that points at another network -- in the same project, another project, or another organization -- and the routes that side exchanges. Peered networks reach each other over internal IPs with no gateway, no bandwidth cap beyond the VMs' own, and no trip through the internet. Set `peerNetwork` to create the peering; leave it empty to manage the route exchange of a peering Google created for you (the private-services-access peering behind Cloud SQL private IP and Memorystore).

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions exactly one of:

- **Network peering** (`peerNetwork` set) -- the `compute_network_peering` entry on your network with its route-exchange flags
- **Peering routes config** (`peerNetwork` empty) -- the `compute_network_peering_routes_config` on an existing peering named `peeringName`

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with `roles/compute.networkAdmin` on the network's project (and on the peer's project when creating a peering to it). Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.

### GCP Networks

- **Your network must exist** -- declare it with `GcpVpcNetwork` and reference its `network_self_link` output, or pass the self link as a literal.
- **Subnet ranges must not overlap** between the two networks; Google rejects the peering otherwise.
- **Peering is not transitive**: A-B and B-C do not connect A to C. A hub-and-spoke needs a peering per spoke.
- **A peering is two half-entries.** It goes ACTIVE only when both sides exist and point at each other.

## Deploy

### Console

Open the deployment store, find **GCP VPC Peering**, and click **Deploy**. The creation wizard walks you through the two networks and the route exchange. Start from the **Peer Two VPCs Side A** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVpcPeering
metadata:
  name: hub-to-spoke
  org: acme-corp
  env: prod
spec:
  network:
    value: projects/acme-net/global/networks/hub
  peerNetwork:
    value: projects/acme-app/global/networks/spoke
  exportCustomRoutes: true
```

```shell
planton apply -f vpc-peering.yaml
```

This creates the hub's side of a hub-spoke peering and shares the hub's VPN/Interconnect routes with the spoke. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, both sides reference their `GcpVpcNetwork`s via ValueFromRef, and the two sides are two resources:

```yaml
# side A, on the hub
spec:
  network:
    valueFrom:
      kind: GcpVpcNetwork
      name: hub
      fieldPath: status.outputs.network_self_link
  peerNetwork:
    valueFrom:
      kind: GcpVpcNetwork
      name: spoke
      fieldPath: status.outputs.network_self_link
  exportCustomRoutes: true
---
# side B, on the spoke
spec:
  network:
    valueFrom:
      kind: GcpVpcNetwork
      name: spoke
      fieldPath: status.outputs.network_self_link
  peerNetwork:
    valueFrom:
      kind: GcpVpcNetwork
      name: hub
      fieldPath: status.outputs.network_self_link
  importCustomRoutes: true
```

The InfraPipeline deploys both networks, then both peering sides; the provider serializes the two sides so declaring them together is safe, and each side's `state` output reads `ACTIVE` once the pair is up.

## Key Configuration

These are the most important decisions when configuring a peering. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Network and peer network** -- your side and the other side. Both immutable. Leave `peerNetwork` empty to switch to the routes-config form on an existing peering.

**Custom routes** -- `exportCustomRoutes` on one side and `importCustomRoutes` on the other is what lets an on-premises network reached over VPN or Interconnect see the peer (or, on the `servicenetworking-googleapis-com` peering, what lets on-premises reach Cloud SQL). Both default false; both change in place.

**Public-IP subnet routes** -- whether privately-used public IP ranges cross the peering. Export defaults true, import false; both immutable in the create form.

**Update strategy** -- `CONSENSUS` makes a changed setting take effect only once both sides agree; use it when the two sides are owned by different teams and a one-sided flip must never change what traffic flows.

**Deletion policy** -- `DELETE` (default) removes this side's entry and the pair goes INACTIVE; `PREVENT` fails the destroy, the guard for a peering production traffic depends on; `ABANDON` leaves the entry live. The routes-config form has no destroy effect in GCP.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpVpcNetwork** | `network` | `status.outputs.network_self_link` |
| **GcpVpcNetwork** | `peerNetwork` | `status.outputs.network_self_link` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `peering_name` | The peering entry's name on this side | Tooling; `gcloud compute networks peerings list` |
| `network` | This side's network | Tooling |
| `state` | `ACTIVE` or `INACTIVE` (create form) | Readiness checks |
| `state_details` | GCP's explanation of the state (create form) | Diagnosing an INACTIVE pair |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Peer two VPCs** -- the pair of sides for a hub-spoke peering. Start from the **Peer Two VPCs Side A** and **Peer Two VPCs Side B** presets.

**Service networking routes** -- export custom routes on the Google-managed private-services peering so on-premises reaches Cloud SQL. Start from the **Service Networking Routes** preset.

## Works With

- [**GCP VPC Network**](/cloud-catalog/gcp-vpc-network) -- the networks on both sides
- [**GCP Service Networking Connection**](/cloud-catalog/gcp-service-networking-connection) -- the Google-managed peering the routes-config form tunes
- [**GCP HA VPN Gateway**](/cloud-catalog/gcp-ha-vpn-gateway) -- the dynamic routes `exportCustomRoutes` shares with the peer
