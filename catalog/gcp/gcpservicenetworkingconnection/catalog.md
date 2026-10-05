# GCP Service Networking Connection

Establishes private services access — the VPC peering between your network and a Google managed-services producer that lets Cloud SQL, AlloyDB, Memorystore (PRIVATE_SERVICE_ACCESS mode), and Filestore hand out private IPs from ranges you reserved inside your VPC. The connection composes two first-class resources: the VPC network being peered and one or more INTERNAL `VPC_PEERING` address ranges (GcpGlobalAddress) the producer carves service subnets from. One connection exists per (network, service) pair — capacity grows by appending ranges to this resource, never by adding a second connection.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Private services access connection** — a `google_service_networking_connection` peering your network with the producer (default `servicenetworking.googleapis.com`), backed by the reserved ranges; the peering appears on the VPC named after the service (e.g. `servicenetworking-googleapis-com`)
- **Service Networking API enablement** — `servicenetworking.googleapis.com` enabled in the target project (the producer-side control plane; never disabled on destroy)
- **Compute Engine API enablement** — `compute.googleapis.com` enabled in the target project (the network and reserved ranges live in Compute; never disabled on destroy)

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** — an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A VPC network** for the `network` field — the consumer VPC to peer (a GcpVpcNetwork reference, a network name, or a full self-link).
- **At least one reserved peering range** for `reservedPeeringRanges` — a GcpGlobalAddress with `addressType: INTERNAL` and purpose `VPC_PEERING`, referenced by NAME (not self-link or CIDR). A /16 is the common default — producers cannot use space that is too fragmented, and a too-small range surfaces later as instance-creation failures, not at connection time.

## Deploy

### Console

Open the deployment store, find **GCP Service Networking Connection**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the network, the service producer, and the reserved range list. Start from the **Private Services Access for Managed Services** preset in the [Presets](#presets) tab to pre-populate the canonical shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpServiceNetworkingConnection
metadata:
  name: prod-vpc-psa
  org: acme-corp
  env: prod
spec:
  network:
    value: prod-vpc
  reservedPeeringRanges:
    - value: prod-vpc-psa-range
```

```shell
planton apply -f psa-connection.yaml
```

This peers `prod-vpc` with `servicenetworking.googleapis.com` (the default producer behind Cloud SQL, AlloyDB, Memorystore, and Filestore private IP), drawing service subnets from the named reserved range. An Infra Job tracks the provisioning in real time.

### InfraChart

When the VPC and reserved range are deployed in the same chart, wire the connection to them with ValueFromRef:

```yaml
spec:
  network:
    valueFrom:
      kind: GcpVpcNetwork
      name: prod-vpc
      fieldPath: status.outputs.network_self_link
  reservedPeeringRanges:
    - valueFrom:
        kind: GcpGlobalAddress
        name: prod-vpc-psa-range
        fieldPath: status.outputs.name
```

The InfraPipeline deploys the network and the address range first, then establishes the peering — the ordering (network → range → connection → private-IP instances) that downstream AlloyDB and Cloud SQL resources require.

## Key Configuration

These are the most important decisions when configuring a service networking connection. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**One connection per pair — widen, never duplicate** — GCP rejects a second connection for the same (network, service) pair. When the producer runs out of address space, append ranges to THIS resource's `reservedPeeringRanges`: growth is additive and safe (already-provisioned producer subnets are untouched), while a second connection resource fails its create.

**Immutable network and service** — `network` and `service` are create-time; changing either destroys and recreates the connection, severing private connectivity for every producer resource on it. The connection has no cloud-side name of its own — GCP addresses it by the pair, and `metadata.name` names only the Planton resource.

**Range sizing** — reserve generously up front (a /16 is the standard default). A too-small or fragmented range does not fail at connection time; it fails LATER, as instance-creation errors when the producer cannot allocate — the worst time to be appending slivers.

**Adopting a pre-existing connection** — when a connection for the pair already exists outside management (a gcloud or console flow), the create fails with "Cannot modify allocated ranges". `updateOnCreationFail: true` converts that failure into an in-place update of the existing connection's ranges — a deliberate adoption lever. Leave it false otherwise: silently adopting someone else's peering is worse than failing loudly.

**Teardown ordering** — GCP refuses to delete the connection while any producer still holds subnets in the reserved ranges: destroy the private-IP instances first, then the connection, then the address ranges. The producer releases its hold asynchronously — live-measured in 2026-08, a deleted private-IP Cloud SQL instance kept the connection rejecting deletion for over 40 minutes. `deletionPolicy: PREVENT` is the posture under a production database fleet; `ABANDON` is the escape hatch for stuck teardowns, at the price of an unmanaged peering.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |
| **GcpVpcNetwork** | `network` | `status.outputs.network_self_link` |
| **GcpGlobalAddress** (per entry) | `reservedPeeringRanges` | `status.outputs.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `peering` | Name of the VPC peering GCP created on the network (e.g. `servicenetworking-googleapis-com`) | Auditing the network's peerings list, peering-level route settings |
| `network` | Self-link of the peered network as the connection resolved it | Confirming which network the producer is attached to without chasing the reference chain |

Private-IP service instances do not reference these outputs — they consume the connection by its existence on the network, which is why deploy order (connection before instances) matters more than wiring.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Private services access** — the canonical single-range setup: one VPC, one /16 `VPC_PEERING` range, the default producer. The prerequisite step for every private-IP database on the network. Start from the **Private Services Access for Managed Services** preset.

**Multi-range growth** — two reserved ranges from the start, for networks expecting heavy private-IP growth — sizing once instead of appending under incident pressure. Start from the **Growing Producer Capacity with Multiple Ranges** preset.

## Works With

- [**GCP Global Address**](/infra-catalog/gcp-global-address) — reserves the `VPC_PEERING` address space the producer carves subnets from
- [**GCP VPC Network**](/infra-catalog/gcp-vpc-network) — the consumer network being peered
- [**GCP AlloyDB Cluster**](/infra-catalog/gcp-alloydb-cluster) — private IP requires this connection on the network first
- [**GCP Cloud SQL**](/infra-catalog/gcp-cloud-sql) — private IP requires this connection on the network first
- [**GCP Project**](/infra-catalog/gcp-project) — provides the project where the required APIs are enabled
