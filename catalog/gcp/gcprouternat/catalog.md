# GCP Router NAT

Deploys a Cloud Router with a NAT gateway that provides outbound internet connectivity for VMs and GKE nodes without external IP addresses (public NAT) or NAT between private networks (private NAT). The NAT can cover all subnets in a region or be scoped to specific subnets — down to individual secondary ranges — with automatic or static IP allocation, NAT64 for IPv6-only workloads, per-destination NAT rules, port and timeout tuning, and configurable translation logging. The router's own BGP surface (ASN, route advertisement, keepalive) is configurable for routers that also serve VPN or Interconnect sessions. Only the names, region, network, endpoint type, and NAT type are immutable — IP allocation, subnet scoping, port tuning, timeouts, and rules all update in place, which is what makes NAT IP rotation and fleet-wide egress tuning zero-downtime operations.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Compute Engine API enablement** -- a `google_project_service` that activates `compute.googleapis.com` on the target project
- **Cloud Router** -- a `google_compute_router` in the specified project, region, and VPC network that hosts the NAT configuration, optionally carrying full BGP configuration (ASN, advertisement mode/groups/ranges, keepalive, identifier range), a description, encrypted-Interconnect dedication, and resource-manager tags
- **Cloud NAT Gateway** -- a `google_compute_router_nat` attached to the router, configured with the selected IP allocation strategy, subnet scope, NAT64 scope, port/timeout tuning, rules, and log settings

Static NAT IPs are **referenced, never created**: each `natIps` entry points at a GcpAddress reservation, which is its own composable node with its own lifecycle. The literal IP of each entry is that address node's `address` output.

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the Cloud Router and NAT will be created. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef.
- **A VPC network** with at least one subnet in the target region. Provide the network self-link directly or reference a GcpVpcNetwork Infra Component via ValueFromRef.
- **GcpAddress reservations** (EXTERNAL, same region) when egress must come from stable, allowlistable IPs — omit `natIps` entirely for auto-allocation.

## Deploy

### Console

Open the deployment store, find **GCP Router NAT**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **All-Subnets Auto-Allocated NAT** preset in the [Presets](#presets) tab to pre-populate a NAT configuration covering all subnets with auto-allocated IPs.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpRouterNat
metadata:
  name: main-nat
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  vpcSelfLink:
    value: "projects/acme-prod-12345/global/networks/main-vpc"
  region: us-central1
  routerName: "main-router"
  natName: "main-nat"
```

```shell
planton apply -f gcp-router-nat.yaml
```

This creates a Cloud Router and NAT covering all subnets in us-central1 with auto-allocated IPs and error-only logging. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the Cloud NAT to a GCP project and VPC deployed in the same InfraPipeline:

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

The InfraPipeline resolves the dependency graph, deploys the project and VPC first, then provisions the Cloud Router and NAT with the resolved values.

## Key Configuration

These are the most important decisions when configuring a Cloud NAT. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Subnet scope** -- Leave `subnetworks` empty to cover all subnets in the region (the common case). List subnetworks (with optional per-range scoping: primary only, or named secondary ranges) when only certain subnets should have NAT egress -- useful for isolating workloads that should not reach the internet, or for NATing GKE pod ranges without the node range.

**IP allocation** -- Leave `natIps` empty for auto-allocated IPs that GCP manages automatically. Reference GcpAddress reservations when you need predictable egress addresses for partner allowlisting, compliance requirements, or API rate limit registration. Rotation is zero-downtime: add the new reservation, move the old one to `drainNatIps`, and existing connections bleed off without a cut.

**NAT64** -- IPv6-only subnets reach IPv4 destinations through the same gateway: set `sourceSubnetworkIpRangesToNat64` to `ALL_IPV6_SUBNETWORKS`, or list dual-stack subnetworks in `nat64Subnetworks`. Only one NAT per region in a network may claim the all-subnetworks mode.

**Log filter** -- `logFilter` defaults to ERRORS_ONLY, which logs port exhaustion and connection failures without high volume. Use ALL for security auditing or detailed troubleshooting. Use DISABLED for non-production environments to reduce costs.

**Router and NAT naming** -- The `routerName` and `natName` fields set the GCP resource names. Each region in a VPC needs its own Cloud Router. Multiple NAT configurations can share a router, but each must have a unique name.

**Router BGP** -- When the same router also terminates Cloud VPN or Interconnect BGP sessions, `routerBgp` controls the ASN and what routes are advertised to every peer (`DEFAULT` advertises all subnets; `CUSTOM` advertises only the listed groups/ranges).

**Teardown posture** -- `deletionPolicy` defaults to DELETE. Set PREVENT on a production egress path so a destroy fails instead of taking the fleet offline; ABANDON removes the pair from management while it keeps serving traffic.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpVpcNetwork** | `vpcSelfLink` | `status.outputs.network_self_link` |
| **GcpAddress** | `natIps[]`, `drainNatIps[]`, rule actions | `status.outputs.self_link` |
| **GcpSubnetwork** | `subnetworks[].subnetwork`, `nat64Subnetworks[]`, rule ranges | `status.outputs.subnetwork_self_link` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | Name of the Cloud NAT gateway | Inventory/automation references (private GKE clusters compose Cloud NAT on their network — no spec field links them) |
| `router_self_link` | Self-link URL of the Cloud Router | The router that VPN tunnels and Interconnect attachments compose on top of |
| `nat_ips` | Self links of the manual NAT IP reservations (empty under auto-allocation) | Partner allowlisting and compliance records — the literal IPs live on the referenced GcpAddress nodes' `address` outputs |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**All-subnets auto-allocated NAT** -- Covers all subnets in the region with auto-allocated external IPs and error-only logging. The simplest configuration for giving private GKE nodes or VMs outbound internet access. Start from the **All-Subnets Auto-Allocated NAT** preset.

**Static IP allowlisting** -- Referenced GcpAddress reservations as stable egress IPs, with subnet scoping and connection draining for rotation. Use when egress must come from known, stable IP addresses for partner allowlisting or compliance. Start from the **Static-IP Allowlisted Egress** preset.

**Private NAT between NCC spokes** -- `type: PRIVATE` with subnetwork-range rules for NAT between VPC networks joined by Network Connectivity Center. Start from the **Private NAT for Network Connectivity Center** preset.

## Works With

- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the Cloud Router and NAT are created
- [**GCP VPC Network**](/infra-catalog/gcp-vpc-network) -- provides the VPC network that the Cloud Router is attached to
- [**GCP Address**](/infra-catalog/gcp-address) -- EXTERNAL reservations referenced as stable NAT IPs
- [**GCP Subnetwork**](/infra-catalog/gcp-subnetwork) -- subnetworks scoped for NAT (and NAT64) coverage
