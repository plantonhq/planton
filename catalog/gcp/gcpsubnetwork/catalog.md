# GCP Subnetwork

Deploys a subnetwork in a custom-mode GCP VPC — the regional address space workloads actually live in. One kind carries the whole subnet surface: the primary IPv4 CIDR (expandable in place, never shrinkable), secondary ranges for GKE alias IPs, Private Google Access, dual-stack and IPv6-only addressing, VPC Flow Logs, and the special-purpose roles other networking features require — proxy-only subnets for Envoy-based load balancers and Private Service Connect subnets.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Subnetwork** -- a Compute Engine subnetwork in the specified project, region, and VPC network, carrying the primary CIDR range, purpose, stack type, and Private Google Access setting
- **Secondary IP Ranges** -- created only when `secondaryIpRanges` entries are provided; named secondary CIDR blocks used by GKE for Pod and Service IP allocation
- **VPC Flow Logs configuration** -- created only when `logConfig` is set; an empty block enables logging with GCP's defaults (5-second aggregation, 50% sampling, all metadata)
- **Compute Engine API enablement** -- `compute.googleapis.com` is enabled in the target project; tearing down the subnet never disables the API

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the subnetwork will be created. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef.
- **A VPC network** in custom subnet mode. Provide the network self-link directly or reference a GcpVpcNetwork Infra Component via ValueFromRef. The module enables the Compute Engine API itself, so the connection's principal needs permission to enable services on a fresh project.
- **An internal IPv6 range on the VPC** (only for INTERNAL `ipv6AccessType`) -- ULA-addressed subnets draw their prefix from the VPC's internal IPv6 range.

## Deploy

### Console

Open the deployment store, find **GCP Subnetwork**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **GKE-Ready Subnet** preset in the [Presets](#presets) tab to pre-populate a subnet with secondary IP ranges configured for GKE Pod and Service allocation.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpSubnetwork
metadata:
  name: gke-subnet
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  vpcSelfLink:
    value: "projects/acme-prod-12345/global/networks/main-vpc"
  region: us-central1
  ipCidrRange: "10.0.0.0/20"
  subnetworkName: "gke-us-central1"
  privateIpGoogleAccess: true
```

```shell
planton apply -f gcp-subnetwork.yaml
```

This creates a subnet with a `/20` primary range and Private Google Access enabled. No secondary ranges are configured -- add them for GKE deployments. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the subnetwork to a GCP project and VPC deployed in the same InfraPipeline:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
  vpcSelfLink:
    valueFrom:
      kind: GcpVpcNetwork
      name: main-vpc
      fieldPath: status.outputs.network_self_link
```

The InfraPipeline resolves the dependency graph, deploys the project and VPC first, then provisions the subnetwork with the resolved values.

## Key Configuration

These are the most important decisions when configuring a subnetwork. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Primary CIDR range** -- The `ipCidrRange` must not overlap any other range in the VPC. The range is mutable in ONE direction: expanding it (e.g. `/20` → `/18`) updates in place, but shrinking forces destroy and recreate — an outage for everything addressed in the subnet. Size for growth up front: `/20` (4,096 IPs) is a common team-sized default; GCP reserves 4 addresses per primary range. Enterprises with centrally allocated IP plans can source the CIDR from a Network Connectivity internal range via `reservedInternalRange` instead.

**Purpose** -- PRIVATE (the default) is a regular workload subnet. REGIONAL_MANAGED_PROXY reserves Envoy address space that regional internal and external Application Load Balancers require — every VPC region hosting one needs exactly one ACTIVE proxy-only subnet, and forgetting it is the error load-balancer deployments trip over. PRIVATE_SERVICE_CONNECT backs published PSC services. Choose at creation: purpose is immutable in practice.

**Secondary IP ranges for GKE** -- Add entries to `secondaryIpRanges` with names like `pods` and `services` to enable VPC-native GKE clusters. A `/14` range for pods supports large clusters, while a `/20` range for services covers most Kubernetes deployments. The range names are referenced by GcpGkeCluster's `clusterSecondaryRangeName` and `servicesSecondaryRangeName` fields. The `sendSecondaryIpRangeIfEmpty` latch defaults to false, so a partial manifest with an empty list leaves existing ranges untouched instead of wiping a live cluster's pod range.

**Private Google Access** -- Set `privateIpGoogleAccess` to true so VMs and GKE nodes without external IPs can reach Google APIs (Artifact Registry, Cloud Logging, Cloud Storage) through internal routing. Effectively mandatory for private-only subnets; required for private GKE clusters. Mutable.

**VPC Flow Logs** -- Setting `logConfig` (even empty) turns on flow logging at GCP's defaults: 50% sampling with all metadata. Full sampling on a busy subnet carries real Cloud Logging cost — raise `flowSampling` to 1.0 only for security forensics, and narrow with `filterExpr` (e.g. `connection.dest_port == 443`) for targeted capture.

**IP stack** -- `stackType` IPV4_ONLY is the default; IPV4_IPV6 and IPV6_ONLY require `ipv6AccessType` (EXTERNAL internet-routable GUAs, or INTERNAL ULAs that need the VPC's internal IPv6 range enabled). Moving between IPV4_ONLY and IPV4_IPV6 updates in place; moving to or from IPV6_ONLY recreates, and the access type itself is immutable.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpVpcNetwork** | `vpcSelfLink` | `status.outputs.network_self_link` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `subnetwork_self_link` | Self-link URI of the created subnetwork | GcpGkeCluster subnet reference, Compute Engine instance placement |
| `subnetwork_name` | Name of the subnetwork in GCP | Consumers that address subnets by name (e.g. Cloud Run Direct VPC egress) |
| `region` | Region where the subnetwork resides | Cloud Router and NAT region alignment |
| `ip_cidr_range` | Primary IPv4 CIDR of the subnet | Firewall rule source/destination ranges, IP planning |
| `gateway_address` | IPv4 address of the subnet's default gateway | Appliance and custom-route configuration |
| `secondary_ranges` | List of secondary range names and CIDRs | GcpGkeCluster Pod and Service secondary range name references |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**GKE-ready subnet** -- A `/20` primary range with two secondary ranges: a `/14` for pods (262,144 IPs) and a `/20` for services (4,096 IPs), plus Private Google Access enabled. The range names `pods` and `services` match the defaults expected by GcpGkeCluster presets. Start from the **GKE-Ready Subnet** preset.

**General-purpose subnet** -- A `/24` primary range for Compute Engine VMs or Cloud Run with VPC Access. No secondary ranges, since alias IPs are not needed outside GKE. Private Google Access is still enabled. Start from the **General-Purpose Subnet** preset.

**Proxy-only subnet** -- A REGIONAL_MANAGED_PROXY subnet with the ACTIVE role: the address space GCP's Envoy-based regional load balancers allocate proxies from, and the prerequisite every region hosting a regional ALB needs exactly one of. Start from the **Proxy-Only Subnet (Regional Managed Proxy)** preset.

## Works With

- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the subnetwork is created
- [**GCP VPC Network**](/infra-catalog/gcp-vpc-network) -- provides the VPC network that contains this subnetwork
- [**GCP GKE Cluster**](/infra-catalog/gcp-gke-cluster) -- consumes the subnet and its secondary ranges for VPC-native pod and service IPs