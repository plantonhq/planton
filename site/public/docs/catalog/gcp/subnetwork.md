---
title: "Subnetwork"
description: "Subnetwork deployment documentation"
icon: "package"
order: 100
componentName: "gcpsubnetwork"
---

# GCP Subnetwork

Deploys a GCP VPC subnetwork in a specified region with a primary CIDR range, optional secondary IP ranges for alias IPs (commonly used by GKE for Pod and Service CIDRs), and optional Private Google Access. The module also enables the Compute Engine API on the target project automatically.

## What Gets Created

When you deploy a GcpSubnetwork resource, OpenMCF provisions:

- **Compute Engine API enablement** — a `google_project_service` resource that ensures `compute.googleapis.com` is active in the target project
- **Subnetwork** — a `google_compute_subnetwork` resource in the specified region and VPC, configured with the primary CIDR, secondary ranges, and Private Google Access setting

## Prerequisites

- **GCP credentials** configured via environment variables or OpenMCF provider config
- **An existing GCP project** where the subnetwork will be created
- **An existing VPC network** in custom subnet mode (the VPC's self-link is required)
- **A non-overlapping primary CIDR range** that does not conflict with other subnets in the same VPC

## Quick Start

Create a file `subnet.yaml`:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpSubnetwork
metadata:
  name: my-subnet
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.GcpSubnetwork.my-subnet
spec:
  projectId: my-gcp-project
  vpcSelfLink: https://www.googleapis.com/compute/v1/projects/my-gcp-project/global/networks/my-vpc
  region: us-central1
  ipCidrRange: "10.0.0.0/24"
  subnetworkName: my-subnet
```

Deploy:

```shell
openmcf apply -f subnet.yaml
```

This creates a subnetwork named `my-subnet` in `us-central1` with a `/24` primary CIDR range.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `projectId` | `string` | GCP project ID where the subnetwork is created. Supports `valueFrom` referencing a GcpProject resource. | Required |
| `vpcSelfLink` | `string` | Self-link of the parent VPC network. Supports `valueFrom` referencing a GcpVpc resource. | Required |
| `region` | `string` | GCP region for the subnetwork (e.g. `us-central1`). Cannot be changed after creation. | Required; must match `^[a-z]([-a-z0-9]*[a-z0-9])?$` |
| `ipCidrRange` | `string` | Primary IPv4 CIDR range for the subnetwork (e.g. `10.0.0.0/24`). Must be unique within the VPC. | Required; must match IPv4 CIDR format |
| `subnetworkName` | `string` | Name of the subnetwork resource in GCP. | Required; 1-63 chars, lowercase letters/numbers/hyphens, must start with a letter and end with a letter or number |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `secondaryIpRanges` | `object[]` | `[]` | Secondary IP ranges for alias IPs. Each entry requires `rangeName` and `ipCidrRange`. Commonly used for GKE Pod and Service CIDRs. |
| `secondaryIpRanges[].rangeName` | `string` | — | Name for the secondary range. 1-63 chars, lowercase letters/numbers/hyphens. |
| `secondaryIpRanges[].ipCidrRange` | `string` | — | IPv4 CIDR for the secondary range. Must not overlap with other ranges in the VPC. |
| `privateIpGoogleAccess` | `bool` | `false` | When `true`, VMs without external IPs in this subnetwork can reach Google APIs and services over internal networking. |

## Examples

### Subnetwork with Private Google Access

A subnetwork that allows VMs without external IPs to access Google APIs:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpSubnetwork
metadata:
  name: private-subnet
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.GcpSubnetwork.private-subnet
spec:
  projectId: my-gcp-project
  vpcSelfLink: https://www.googleapis.com/compute/v1/projects/my-gcp-project/global/networks/my-vpc
  region: us-east1
  ipCidrRange: "10.10.0.0/20"
  subnetworkName: private-subnet
  privateIpGoogleAccess: true
```

### GKE-Ready Subnetwork with Secondary Ranges

A subnetwork configured with secondary IP ranges for GKE Pod and Service CIDRs:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpSubnetwork
metadata:
  name: gke-subnet
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpSubnetwork.gke-subnet
spec:
  projectId: my-gcp-project
  vpcSelfLink: https://www.googleapis.com/compute/v1/projects/my-gcp-project/global/networks/prod-vpc
  region: us-central1
  ipCidrRange: "10.0.0.0/20"
  subnetworkName: gke-nodes
  privateIpGoogleAccess: true
  secondaryIpRanges:
    - rangeName: gke-pods
      ipCidrRange: "10.4.0.0/14"
    - rangeName: gke-services
      ipCidrRange: "10.8.0.0/20"
```

### Using Foreign Key References

Reference other OpenMCF-managed resources instead of hardcoding IDs:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpSubnetwork
metadata:
  name: ref-subnet
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpSubnetwork.ref-subnet
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: my-project
      field: status.outputs.project_id
  vpcSelfLink:
    valueFrom:
      kind: GcpVpc
      name: prod-vpc
      field: status.outputs.network_self_link
  region: europe-west1
  ipCidrRange: "10.20.0.0/20"
  subnetworkName: ref-subnet
  privateIpGoogleAccess: true
  secondaryIpRanges:
    - rangeName: pods
      ipCidrRange: "10.24.0.0/14"
    - rangeName: services
      ipCidrRange: "10.28.0.0/20"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `subnetworkSelfLink` | `string` | Self-link URL of the created subnetwork, used to reference this subnet from GKE clusters and other resources |
| `region` | `string` | GCP region where the subnetwork was created |
| `ipCidrRange` | `string` | Primary IPv4 CIDR of the subnetwork |
| `secondaryRanges` | `object[]` | List of secondary ranges created, each containing `rangeName` and `ipCidrRange` |

## Related Components

- [GcpVpc](/docs/catalog/gcp/vpc) — provides the parent VPC network for this subnetwork
- [GcpGkeCluster](/docs/catalog/gcp/gke-cluster) — consumes subnetwork and secondary ranges for node, Pod, and Service networking
- [GcpRouterNat](/docs/catalog/gcp/router-nat) — provides NAT gateway for subnetworks without external IPs
- [GcpProject](/docs/catalog/gcp/project) — manages the GCP project that hosts the subnetwork
- [GcpComputeInstance](/docs/catalog/gcp/compute-instance) — VMs deployed into this subnetwork
