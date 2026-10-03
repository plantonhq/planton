# GCP VPC Network

Deploys a Google Cloud VPC network — the global routing domain everything else in the GCP catalog plugs into. A GCP VPC spans every region: subnetworks, firewall rules, Cloud NAT, serverless connectors, and GKE clusters all anchor to this one resource. In custom subnet mode (the production posture) the network itself carries no IP ranges; the IP plan lives on GcpSubnetwork resources. Private services access (Cloud SQL / AlloyDB / Memorystore private IP) is composed separately: a GcpGlobalAddress VPC_PEERING range plus a GcpServiceNetworkingConnection.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Compute Engine API enablement** on the target project (never disabled on destroy)
- **VPC Network** -- a `google_compute_network` in auto or custom subnet mode, with the configured routing mode, MTU, ULA internal IPv6 allocation, firewall-policy evaluation order, BGP best-path selection, and default-route posture

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the network will be created. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef. In a Shared VPC design this is the HOST project.
- **IAM permissions** -- the connection's principal needs network administration on the target project (`compute.networks.*`) plus `serviceusage.services.enable` for the module's Compute API enablement step; `compute.routes.list`/`delete` are additionally required only when `deleteDefaultRoutesOnCreate` is true.

## Deploy

### Console

Open the deployment store, find **GCP VPC Network**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Custom Mode VPC with Regional Routing** preset in the [Presets](#presets) tab for the production-standard network.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVpcNetwork
metadata:
  name: prod-vpc
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  networkName: prod-vpc
  autoCreateSubnetworks: false
  description: Production VPC for application workloads
```

```shell
planton apply -f vpc.yaml
```

This creates an empty custom-mode network — the deliberate starting point. Subnetworks, firewall rules, and NAT are then authored as their own resources against its self link. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the network to a GCP project deployed in the same InfraPipeline:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
  networkName: prod-vpc
  autoCreateSubnetworks: false
```

The InfraPipeline resolves the dependency graph, deploys the project first, then provisions the network. This node is the composition backbone: GcpSubnetwork, GcpFirewallRule, and GcpRouterNat nodes hang off its `network_self_link` output.

## Key Configuration

These are the most important decisions when configuring a VPC network. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Subnet mode** -- `autoCreateSubnetworks: false` (custom mode, the default here) starts the network empty so every subnet is a deliberate resource with a chosen CIDR. Auto mode creates one subnet per region from `10.128.0.0/9` — convenient for sandboxes, a collision hazard for anything that will ever peer. The door is one-way: auto converts to custom, never back.

**Routing mode** -- `REGIONAL` (GCP's default when unset) keeps Cloud Router advertisements inside each router's region; `GLOBAL` lets one VPN or Interconnect serve every region. Mutable at any time.

**MTU** -- 1460 by default; up to 8896 for jumbo frames inside the VPC. Internet-bound and cross-VPC paths still clamp lower, and every VM OS must match.

**Default routes** -- `deleteDefaultRoutesOnCreate: true` births the network with NO internet path — the fail-closed posture for regulated egress. Nothing (including Cloud NAT) reaches the internet until routes are authored deliberately. Create-time only.

**Description is not a free edit** -- `description` is immutable on this resource: changing it destroys and recreates the network, and GCP refuses to delete a network that still holds subnets, peerings, or attached resources. Write it once, deliberately.

**Deletion policy** -- `PREVENT` protects the network every subnet, route, and peering depends on — destroy fails instead of cascading. `ABANDON` removes it from management while it keeps serving in GCP. The default `DELETE` already refuses while dependent resources remain.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `network_self_link` | Full self-link URL of the network | GcpSubnetwork, GcpFirewallRule, GcpRouterNat, GKE clusters — the catalog's most-consumed FK anchor |
| `network_name` | Name of the network | GcpCloudRun VPC access, GcpServerlessVpcConnector network placement |
| `network_id` | GCP resource path (projects/PROJECT/global/networks/NAME) | GcpCloudSql private-network configuration |
| `gateway_ipv4` | IPv4 of the default internet gateway (when present) | Route debugging, custom route composition |
| `internal_ipv6_range` | The ULA /48 assigned when internal IPv6 is enabled | IPv6 subnet planning |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Custom mode with regional routing** -- The production standard: an empty custom-mode network with regional routing, ready for deliberate GcpSubnetwork authoring. Start from the **Custom Mode VPC with Regional Routing** preset.

**Custom mode with global routing** -- The multi-region/hybrid variant: global routing lets a single Cloud VPN or Interconnect attachment serve subnets in every region. Start from the **Custom Mode VPC with Global Routing** preset.

## Works With

- [**GCP Subnetwork**](/infra-catalog/gcp-subnetwork) -- the regional address spaces authored into this network
- [**GCP Firewall Rule**](/infra-catalog/gcp-firewall-rule) -- allows and denies traffic on this network
- [**GCP Router NAT**](/infra-catalog/gcp-router-nat) -- managed egress for instances without external IPs
- [**GCP Serverless VPC Connector**](/infra-catalog/gcp-serverless-vpc-connector) -- bridges Cloud Run / Cloud Functions into this network
- [**GCP Address**](/infra-catalog/gcp-address) -- anchors peering and interconnect ranges to this network
- [**GCP Global Address**](/infra-catalog/gcp-global-address) -- reserves the private-services-access range for managed services
- [**GCP Service Networking Connection**](/infra-catalog/gcp-service-networking-connection) -- peers the reserved range to Google's service producers for Cloud SQL / AlloyDB / Memorystore private IP
- [**GCP GKE Cluster**](/infra-catalog/gcp-gke-cluster) -- VPC-native clusters consume this network and its subnets
- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the network is created
