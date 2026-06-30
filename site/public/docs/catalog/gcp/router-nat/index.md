---
title: "Router NAT"
description: "Router NAT deployment documentation"
icon: "package"
order: 100
componentName: "gcprouternat"
---

# GCP Router NAT

Deploys a GCP Cloud Router with a Cloud NAT gateway to provide outbound internet connectivity for private instances in a VPC. The component supports automatic or manual (static) IP allocation, per-subnet scoping, and configurable NAT translation logging.

## What Gets Created

When you deploy a GcpRouterNat resource, Planton provisions:

- **Cloud Router** — a regional `google_compute_router` attached to the specified VPC network
- **Cloud NAT Gateway** — a `google_compute_router_nat` on the router, configured with the chosen IP allocation strategy, subnet coverage, and log settings
- **Static External IP Addresses** — one `google_compute_address` per entry in `natIpNames`, created only when manual IP allocation is specified; omitted when using auto-allocation

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **A GCP project** where the Cloud Router and NAT will be created
- **An existing VPC network** (self-link or name) in the target project
- **A target region** where private instances need outbound internet access
- **Subnet self-links** if you want to restrict NAT to specific subnets (optional — omit to cover all subnets)
- **Static IP names** if you need deterministic egress IPs for external allowlisting (optional — omit for auto-allocation)

## Quick Start

Create a file `router-nat.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpRouterNat
metadata:
  name: my-nat
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.GcpRouterNat.my-nat
spec:
  projectId: my-gcp-project-123
  vpcSelfLink: https://www.googleapis.com/compute/v1/projects/my-gcp-project-123/global/networks/my-vpc
  region: us-central1
  routerName: my-router
  natName: my-nat
```

Deploy:

```shell
planton apply -f router-nat.yaml
```

This creates a Cloud Router and NAT gateway covering all subnets in `us-central1` with auto-allocated IPs and `ERRORS_ONLY` logging enabled by default.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `projectId` | `StringValueOrRef` | GCP project ID where the Cloud Router and NAT will be created. | Required. Can reference a GcpProject resource via `valueFrom`. |
| `vpcSelfLink` | `StringValueOrRef` | Self-link or name of the target VPC network. | Required. Can reference a GcpVpc resource via `valueFrom`. |
| `region` | `string` | GCP region for the Cloud Router and NAT. | Required |
| `routerName` | `string` | Name of the Cloud Router to create. | Required. Must match `^[a-z]([a-z0-9-]{0,61}[a-z0-9])?$` |
| `natName` | `string` | Name of the NAT configuration on the Cloud Router. | Required. Must match `^[a-z]([a-z0-9-]{0,61}[a-z0-9])?$` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `subnetworkSelfLinks` | `StringValueOrRef[]` | `[]` (all subnets) | Specific subnets to enable NAT on. When empty, NAT covers all subnets in the region. |
| `natIpNames` | `StringValueOrRef[]` | `[]` (auto-allocate) | Names for static external IP addresses to use for NAT. When empty, GCP auto-allocates IPs. When specified, one regional static address is created per entry. |
| `logFilter` | `GcpRouterNatLogFilter` | `ERRORS_ONLY` | Log filter for NAT translation logging. Values: `DISABLED` (no logging), `ERRORS_ONLY` (log translation errors), `ALL` (log all translations). |

## Examples

### All-Subnets NAT with Auto-Allocated IPs

The most common configuration — provides outbound internet access for every subnet in the region with automatic IP management:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpRouterNat
metadata:
  name: uscentral1-nat
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.GcpRouterNat.uscentral1-nat
spec:
  projectId: my-gcp-project-123
  vpcSelfLink: https://www.googleapis.com/compute/v1/projects/my-gcp-project-123/global/networks/my-vpc
  region: us-central1
  routerName: dev-uscentral1-router
  natName: dev-uscentral1-nat
  logFilter: ERRORS_ONLY
```

### NAT with Static IPs for Allowlisting

Use manual IP allocation when external partners need to allowlist your egress IPs:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpRouterNat
metadata:
  name: prod-nat
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpRouterNat.prod-nat
spec:
  projectId: my-gcp-project-123
  vpcSelfLink: https://www.googleapis.com/compute/v1/projects/my-gcp-project-123/global/networks/prod-vpc
  region: us-central1
  routerName: prod-uscentral1-router
  natName: prod-uscentral1-nat
  natIpNames:
    - prod-nat-ip-0
    - prod-nat-ip-1
  logFilter: ERRORS_ONLY
```

### Subnet-Scoped NAT with Full Logging

Restrict NAT to specific subnets and enable full translation logging for security auditing:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpRouterNat
metadata:
  name: audit-nat
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpRouterNat.audit-nat
spec:
  projectId: my-gcp-project-123
  vpcSelfLink: https://www.googleapis.com/compute/v1/projects/my-gcp-project-123/global/networks/prod-vpc
  region: europe-west1
  routerName: prod-euwest1-router
  natName: prod-euwest1-nat
  subnetworkSelfLinks:
    - https://www.googleapis.com/compute/v1/projects/my-gcp-project-123/regions/europe-west1/subnetworks/gke-nodes
    - https://www.googleapis.com/compute/v1/projects/my-gcp-project-123/regions/europe-west1/subnetworks/app-vms
  natIpNames:
    - prod-euwest1-nat-ip-0
  logFilter: ALL
```

### Using Foreign Key References

Reference other Planton-managed resources instead of hardcoding values:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpRouterNat
metadata:
  name: ref-nat
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpRouterNat.ref-nat
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: my-project
      fieldPath: status.outputs.project_id
  vpcSelfLink:
    valueFrom:
      kind: GcpVpc
      name: my-vpc
      fieldPath: status.outputs.network_self_link
  region: us-central1
  routerName: prod-uscentral1-router
  natName: prod-uscentral1-nat
  logFilter: ERRORS_ONLY
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | Name of the Cloud NAT gateway as created in GCP |
| `routerSelfLink` | `string` | Self-link URL of the Cloud Router created for this NAT |
| `natIpAddresses` | `string[]` | External IP addresses used by this NAT (auto-allocated or static) |

## Related Components

- [GcpVpc](/docs/catalog/gcp/vpc) — provides the VPC network that the Cloud Router attaches to
- [GcpSubnetwork](/docs/catalog/gcp/subnetwork) — subnets that can be scoped for NAT coverage
- [GcpProject](/docs/catalog/gcp/project) — the GCP project where the router and NAT are created
- [GcpGkeCluster](/docs/catalog/gcp/gke-cluster) — private GKE clusters that depend on Cloud NAT for outbound internet access
- [GcpComputeInstance](/docs/catalog/gcp/compute-instance) — private VMs that use Cloud NAT for egress
