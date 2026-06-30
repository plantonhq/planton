# GCP GKE Cluster

Deploys a private GKE cluster control plane on Google Cloud with VPC-native networking, configurable release channels, Workload Identity, and Calico network policy enforcement. The component provisions the cluster itself — node pools and networking resources (VPC, subnets, Cloud NAT) are managed by separate Planton components.

## What Gets Created

When you deploy a GcpGkeCluster resource, Planton provisions:

- **GKE Cluster** — a `google_container_cluster` resource with:
  - Private cluster configuration (private nodes by default, private endpoint disabled)
  - VPC-native IP allocation using secondary ranges for pods and services
  - Release channel for automatic control plane upgrades (default: REGULAR)
  - Workload Identity enabled by default (`PROJECT_ID.svc.id.goog`)
  - Calico network policy enforcement enabled by default
  - Default node pool removed (node pools are managed separately)

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **An existing GCP project** — referenced via `projectId`
- **A VPC network** with a subnetwork that has secondary ranges for pods and services
- **A Cloud NAT** configured on the VPC for private node outbound internet access
- **IAM permissions** to create GKE clusters in the target project

## Quick Start

Create a file `gke.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpGkeCluster
metadata:
  name: my-cluster
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.GcpGkeCluster.my-cluster
spec:
  clusterName: dev-cluster
  projectId:
    value: my-gcp-project-123
  location: us-central1
  networkSelfLink:
    value: projects/my-gcp-project-123/global/networks/my-vpc
  subnetworkSelfLink:
    value: projects/my-gcp-project-123/regions/us-central1/subnetworks/my-subnet
  clusterSecondaryRangeName:
    value: pods
  servicesSecondaryRangeName:
    value: services
  masterIpv4CidrBlock: "172.16.0.16/28"
  routerNatName:
    value: my-nat
```

Deploy:

```shell
planton apply -f gke.yaml
```

This creates a private GKE cluster in `us-central1` with Workload Identity enabled, REGULAR release channel, and Calico network policies.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `clusterName` | `string` | Name of the GKE cluster in GCP. | 1-40 chars, lowercase, letters/numbers/hyphens, must start with a letter |
| `projectId` | `StringValueOrRef` | GCP project ID. Can reference a GcpProject resource via `valueFrom`. | Required |
| `location` | `string` | Region or zone for the cluster (e.g., `us-central1` for regional, `us-central1-a` for zonal). | Required, must match GCP location pattern |
| `networkSelfLink` | `StringValueOrRef` | VPC network self-link. Can reference a GcpVpc resource via `valueFrom`. | Required |
| `subnetworkSelfLink` | `StringValueOrRef` | VPC subnetwork self-link. Can reference a GcpSubnetwork resource via `valueFrom`. | Required |
| `clusterSecondaryRangeName` | `StringValueOrRef` | Name of the secondary IP range on the subnetwork for pod IPs. Can reference a GcpSubnetwork. | Required |
| `servicesSecondaryRangeName` | `StringValueOrRef` | Name of the secondary IP range on the subnetwork for service IPs. Can reference a GcpSubnetwork. | Required |
| `masterIpv4CidrBlock` | `string` | RFC 1918 CIDR block for the control plane private endpoint. | Must be a /28 CIDR block |
| `routerNatName` | `StringValueOrRef` | Cloud NAT name for private node outbound access. Can reference a GcpRouterNat resource. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enablePublicNodes` | `bool` | `false` | When `true`, nodes are created with public IPs. When `false`, nodes are private (recommended). |
| `releaseChannel` | `enum` | `REGULAR` | Kubernetes release channel: `RAPID`, `REGULAR`, `STABLE`, or `NONE`. Controls automatic control plane upgrades. |
| `disableNetworkPolicy` | `bool` | `false` | Disable Calico network policy enforcement. |
| `disableWorkloadIdentity` | `bool` | `false` | Disable Workload Identity (mapping Kubernetes service accounts to GCP service accounts). |

## Examples

### Regional Cluster with Stable Channel

A production cluster using the STABLE release channel for maximum reliability:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpGkeCluster
metadata:
  name: prod-cluster
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpGkeCluster.prod-cluster
spec:
  clusterName: prod-cluster
  projectId:
    value: my-prod-project
  location: us-central1
  networkSelfLink:
    value: projects/my-prod-project/global/networks/prod-vpc
  subnetworkSelfLink:
    value: projects/my-prod-project/regions/us-central1/subnetworks/prod-subnet
  clusterSecondaryRangeName:
    value: pods
  servicesSecondaryRangeName:
    value: services
  masterIpv4CidrBlock: "172.16.0.16/28"
  routerNatName:
    value: prod-nat
  releaseChannel: STABLE
```

### Full-Featured with Foreign Key References

Using Planton resource references for the entire networking stack:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpGkeCluster
metadata:
  name: ref-cluster
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpGkeCluster.ref-cluster
spec:
  clusterName: ref-cluster
  projectId:
    valueFrom:
      kind: GcpProject
      name: my-project
      field: status.outputs.project_id
  location: us-east1
  networkSelfLink:
    valueFrom:
      kind: GcpVpc
      name: my-vpc
      field: status.outputs.self_link
  subnetworkSelfLink:
    valueFrom:
      kind: GcpSubnetwork
      name: my-subnet
      field: status.outputs.self_link
  clusterSecondaryRangeName:
    valueFrom:
      kind: GcpSubnetwork
      name: my-subnet
      field: status.outputs.pods_secondary_range_name
  servicesSecondaryRangeName:
    valueFrom:
      kind: GcpSubnetwork
      name: my-subnet
      field: status.outputs.services_secondary_range_name
  masterIpv4CidrBlock: "172.16.0.32/28"
  routerNatName:
    valueFrom:
      kind: GcpRouterNat
      name: my-nat
      field: metadata.name
  releaseChannel: REGULAR
  enablePublicNodes: false
  disableNetworkPolicy: false
  disableWorkloadIdentity: false
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `endpoint` | `string` | Kubernetes API server endpoint (private IP for private clusters) |
| `cluster_ca_certificate` | `string` | Base64-encoded CA certificate for the cluster's API server |
| `workload_identity_pool` | `string` | Workload Identity pool identifier (e.g., `my-project.svc.id.goog`) — only set when Workload Identity is enabled |

## Related Components

- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project for cluster creation
- [GcpVpc](/docs/catalog/gcp/gcpvpc) — provides the VPC network
- [GcpSubnetwork](/docs/catalog/gcp/gcpsubnetwork) — provides the subnetwork with secondary IP ranges for pods and services
- [GcpRouterNat](/docs/catalog/gcp/gcprouternat) — provides Cloud NAT for private node outbound internet access
