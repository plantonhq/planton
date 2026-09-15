# DigitalOcean VPC

Deploys a DigitalOcean Virtual Private Cloud with configurable CIDR range and region placement, providing an isolated private network for Droplets, Kubernetes clusters, databases, and load balancers. The one decision that outlives everything else is the IP range: it is immutable, and replacing it means evacuating every resource inside the network first -- so choose between an explicit, planned range and letting DigitalOcean assign one before the first member lands.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **DigitalOcean VPC** -- a regional private network with the configured IP range (or a DigitalOcean-assigned range if omitted) and optional description

## Before You Deploy

### Planton Setup

- **DigitalOcean Provider Connection** -- an active connection in the Connect module with a DigitalOcean API token. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### DigitalOcean Account

- **A target region** -- DigitalOcean VPCs are regional and cannot span regions. Choose the region where your resources will be deployed (e.g., `nyc1`, `sfo3`, `fra1`).
- **IP address planning** (only for explicit ranges) -- if corporate ranges, VPNs, or VPC peering will ever touch this network, plan non-overlapping CIDR blocks before deployment. DigitalOcean accepts prefix lengths from /16 through /24, and the range is immutable after creation.

## Deploy

### Console

Open the deployment store, find **DigitalOcean VPC**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Standard VPC** preset in the [Presets](#presets) tab to pre-populate a working configuration with an explicit /16 CIDR block.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanVpc
metadata:
  name: app-network
  org: acme-corp
  env: prod
spec:
  region: nyc1
```

```shell
planton apply -f do-vpc.yaml
```

This creates a VPC in the NYC1 region with a DigitalOcean-assigned, non-conflicting IP range (reported through the `ip_range` output). A Stack Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring a VPC. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**IP range** -- `ipRangeCidr` accepts CIDR blocks with prefix lengths /16 through /24 (e.g., `"10.10.0.0/16"`). When omitted, DigitalOcean assigns a non-conflicting range and reports it through the `ip_range` output -- fine when nothing will ever peer with this network or VPN into it. The moment corporate ranges enter the picture, choose the range yourself and record the allocation, because auto-assigned 10.x blocks will eventually collide with somebody's office network. The range is immutable: changing it later replaces the VPC, which means evacuating every member first.

**Sizing inside /16-/24** -- A /24 (256 addresses) sounds roomy for a dozen Droplets, but managed resources quietly consume addresses too: every DOKS node, load balancer, and database cluster member takes one, and a Kubernetes cluster autoscaling to twenty nodes eats twenty. A /20 (4,096) is a comfortable default for an environment; reserve /24s for genuinely small, fixed-size networks.

**Region** -- The `region` field is create-time only, and every resource placed in this VPC must be in the same region. Cross-region private connectivity does not come from this kind -- it comes from VPC peering or from routing over the public network with TLS -- so a workload that spans regions needs one VPC per region with non-overlapping ranges, designed up front so future peering stays possible.

**Regional defaults** -- Whether a VPC is the region's DEFAULT is computed by DigitalOcean and cannot be set here. Treat the default VPC as the untyped landing zone and this kind's VPCs as the deliberate ones: wire every resource's `vpc` reference explicitly instead of relying on regional defaults. One trap: in a region with no VPC yet, the first VPC created becomes the default and cannot be deleted, so a destroy of it fails -- seed the region's default (or create any resource there without a `vpc`) before declaring your first deliberate VPC in a new region.

**Teardown order** -- DigitalOcean refuses to delete a VPC that still contains resources. Tear environments down in dependency order -- workloads, then load balancers and databases, then the VPC last.

## Outputs and Dependencies

### What This Component Consumes

This component has no foreign key dependencies.

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `vpc_id` | Unique VPC identifier (UUID) on DigitalOcean | Droplet VPC placement, Kubernetes cluster networking, database cluster VPC attachment, load balancer VPC placement |
| `ip_range` | The VPC's CIDR range as DigitalOcean reports it (covers the auto-assigned case) | Firewall rules, peering and VPN planning |
| `urn` | The VPC's uniform resource name (`do:vpc:<uuid>`) | DigitalOcean project assignment, audit |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard VPC** -- explicit /16 CIDR block providing 65,536 IPs, suitable for production environments with multiple Droplets, Kubernetes clusters, and databases. Adjust the second octet to avoid overlap when running multiple VPCs. Start from the **Standard VPC** preset.

**One VPC per environment per region** -- separate VPCs for prod and staging in each region, with explicitly planned, non-overlapping ranges. The isolation is free, blast radius shrinks, and future peering between any pair stays possible because no two ranges collide.

## Works With

- [**DigitalOcean Droplet**](/cloud-catalog/digital-ocean-droplet) -- placed in this network via its `vpc` reference
- [**DigitalOcean Kubernetes Cluster**](/cloud-catalog/digital-ocean-kubernetes-cluster) -- cluster networking consumes `vpc_id`
- [**DigitalOcean Database Cluster**](/cloud-catalog/digital-ocean-database-cluster) -- private database attachment inside the VPC
- [**DigitalOcean Load Balancer**](/cloud-catalog/digital-ocean-load-balancer) -- balancer placement via its `vpc` reference
- [**DigitalOcean App Platform App**](/cloud-catalog/digital-ocean-app) -- App Platform egress placement (`spec.vpc`)
