# GCP Address

Deploys a regional Compute Engine address reservation — a static IP (or reserved CIDR range) at regional scope. Two primary use cases: external static IPs for Cloud NAT, regional load balancers, and internet-facing VMs; and internal reservations for GCE endpoints, internal LB VIPs, VPC-peering ranges, IPsec interconnect, and DNS resolver endpoints. Supports ValueFromRef wiring to GCP projects, VPC networks, and subnetworks. For global-scope reservations (HTTP(S) LB frontends, private-services-access ranges, Private Service Connect), use GcpGlobalAddress instead.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Compute Engine API enablement** on the target project (never disabled on destroy)
- **Regional Address Reservation** -- a `google_compute_address` pinning a public IP (EXTERNAL) or a private IP / CIDR range (INTERNAL) in the specified region, with the purpose-appropriate network or subnetwork anchor

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the reservation will be created. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef.
- **IAM permissions** -- `roles/compute.networkAdmin` on the target project.
- **For internal purposes** -- an existing VPC network (peering/interconnect ranges) or a subnetwork in the reservation's region (GCE endpoints, DNS resolvers), referenced directly or via ValueFromRef.

## Deploy

### Console

Open the deployment store, find **GCP Address**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **External NAT IP** preset in the [Presets](#presets) tab for the most common reservation.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpAddress
metadata:
  name: nat-ip
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  addressName: nat-external-ip
  region: us-central1
  addressType: EXTERNAL
  networkTier: PREMIUM
  description: Static IP for Cloud NAT in us-central1
```

```shell
planton apply -f address.yaml
```

This reserves a public IPv4 address GCP assigns — the reservation pins it permanently, so the NAT gateway that consumes it keeps a stable, allow-listable egress address. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire an internal endpoint reservation to its subnetwork — and the consuming VM to the reserved address:

```yaml
spec:
  addressName: data-vm-ip
  region: us-central1
  addressType: INTERNAL
  purpose: GCE_ENDPOINT
  subnetwork:
    valueFrom:
      kind: GcpSubnetwork
      name: app-subnet
      fieldPath: status.outputs.subnetwork_self_link
```

The consuming GcpComputeInstance then references this reservation from `networkInterfaces[].networkIp` via `status.outputs.address` — the VM keeps the same private IP across rebuilds.

## Key Configuration

These are the most important decisions when configuring an address reservation. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Region** -- A regional address serves only resources in its region: the NAT gateway, internal LB, or VM NIC that consumes it must live there. Immutable.

**Address type** -- `EXTERNAL` (the GCP default when unset) reserves a public IP; `INTERNAL` reserves a private IP or range inside your VPC. The fork gates everything else: the network tier is external-only, the purpose is internal-only.

**Purpose (internal only)** -- Decides what anchors the reservation: `VPC_PEERING`/`IPSEC_INTERCONNECT` ranges hang off the VPC network (with a `prefixLength` sizing the block), while `GCE_ENDPOINT`/`DNS_RESOLVER` addresses come from a subnetwork's range. `SHARED_LOADBALANCER_VIP` backs internal LB frontends. `PRIVATE_SERVICE_CONNECT` is global-only — use GcpGlobalAddress.

**Everything is ForceNew** -- Any change destroys and recreates the reservation, and a replaced EXTERNAL reservation returns a DIFFERENT public IP. Treat the spec as create-time.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpVpcNetwork** (optional) | `network` (VPC_PEERING / IPSEC_INTERCONNECT) | `status.outputs.network_self_link` |
| **GcpSubnetwork** (optional) | `subnetwork` (GCE_ENDPOINT / DNS_RESOLVER) | `status.outputs.subnetwork_self_link` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `address` | The reserved IP (or the start of the reserved range) | GcpComputeInstance `networkIp`/`natIp`, GcpDnsRecord targets, GcpVertexAiNotebook |
| `self_link` | Self-link URL of the reservation | GcpRouterNat `natIps`/`drainNatIps` |
| `name` | Name of the reservation in GCP | Monitoring, gcloud runbooks |
| `region` | The plain region name | Scope-compatibility checks in compositions |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**External NAT IP** -- A Premium-tier public IPv4 for Cloud NAT, so all egress traffic leaves through one stable, allow-listable address. Start from the **External NAT IP** preset.

**Internal LB VIP** -- An internal reservation with the `SHARED_LOADBALANCER_VIP` purpose — the stable frontend address for an internal load balancer. Start from the **Internal Load Balancer VIP** preset.

**Internal GCE Endpoint** -- A subnetwork-anchored private IP a VM references from its NIC, keeping clients and DNS stable across VM rebuilds. Start from the **Internal GCE Endpoint** preset.

## Works With

- [**GCP Compute Instance**](/infra-catalog/gcp-compute-instance) -- consumes reserved internal (networkIp) and external (natIp) addresses on its NICs
- [**GCP Router NAT**](/infra-catalog/gcp-router-nat) -- consumes external reservations (by self link) as its NAT IPs
- [**GCP DNS Record**](/infra-catalog/gcp-dns-record) -- targets the reserved address for internal load balancer records
- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the reservation is created
- [**GCP VPC Network**](/infra-catalog/gcp-vpc-network) -- anchors peering and interconnect ranges
- [**GCP Subnetwork**](/infra-catalog/gcp-subnetwork) -- anchors endpoint and resolver addresses
- [**GCP Global Address**](/infra-catalog/gcp-global-address) -- the global-scope sibling for HTTP(S) LB frontends and PSA ranges
