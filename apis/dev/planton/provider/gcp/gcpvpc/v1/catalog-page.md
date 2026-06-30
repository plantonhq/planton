# GCP VPC

Deploys a GCP VPC network in custom subnet mode by default, with configurable dynamic routing and optional Private Services Access for Google managed services like Cloud SQL and Memorystore. The component automatically enables the Compute Engine API on the target project before creating the network.

## What Gets Created

When you deploy a GcpVpc resource, Planton provisions:

- **Compute Engine API enablement** — a `google_project_service` resource that activates `compute.googleapis.com` on the target project
- **VPC Network** — a `google_compute_network` resource with the specified name, subnet mode, and routing mode in the target project
- **Private Services IP Range** (conditional) — a `google_compute_global_address` resource of type `INTERNAL` with purpose `VPC_PEERING`, created only when `privateServicesAccess.enabled` is `true`; reserves a CIDR block for Google managed services
- **Private Services Connection** (conditional) — a `google_service_networking_connection` resource that peers the VPC with Google's `servicenetworking.googleapis.com` network, created only when `privateServicesAccess.enabled` is `true`

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **An existing GCP project** — referenced via `projectId`
- **IAM permissions** to enable APIs and create VPC networks in the target project
- **Service Networking API enabled** (`servicenetworking.googleapis.com`) on the project if using Private Services Access — this can be enabled via a GcpProject resource

## Quick Start

Create a file `vpc.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpVpc
metadata:
  name: my-vpc
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.GcpVpc.my-vpc
spec:
  projectId:
    value: my-gcp-project-123
  networkName: dev-network
```

Deploy:

```shell
planton apply -f vpc.yaml
```

This creates a custom-mode VPC named `dev-network` with regional routing in the specified GCP project.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `projectId` | `StringValueOrRef` | GCP project ID where the VPC is created. Can reference a GcpProject resource via `valueFrom`. | Required |
| `networkName` | `string` | Name of the VPC network in GCP. | 1-63 chars, lowercase letters/numbers/hyphens, must start with a letter and end with a letter or number |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `autoCreateSubnetworks` | `bool` | `false` | When `true`, GCP automatically creates a subnet in every region. When `false` (recommended), subnets are managed separately via GcpSubnetwork resources. |
| `routingMode` | `enum` | `REGIONAL` | Dynamic routing mode for Cloud Routers: `REGIONAL` (routes advertised in one region only) or `GLOBAL` (routes advertised across all regions). Use `GLOBAL` for multi-region or hybrid connectivity. |
| `privateServicesAccess.enabled` | `bool` | `false` | Enables VPC peering with Google's service network, allowing managed services (Cloud SQL, Memorystore, Filestore) to use private IPs from this VPC. |
| `privateServicesAccess.ipRangePrefixLength` | `int32` | `16` | CIDR prefix length for the private services IP allocation. A `/16` reserves 65,536 addresses. Valid range: 8-24. Use a smaller prefix (more IPs) when running many managed service instances. |

## Examples

### Custom-Mode VPC with Regional Routing

A basic VPC for a single-region deployment:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpVpc
metadata:
  name: dev-vpc
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.GcpVpc.dev-vpc
spec:
  projectId:
    value: my-dev-project-123
  networkName: dev-network
```

### Multi-Region VPC with Global Routing

A VPC with global dynamic routing for multi-region workloads or hybrid VPN/Interconnect setups:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpVpc
metadata:
  name: prod-vpc
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpVpc.prod-vpc
spec:
  projectId:
    value: my-prod-project-456
  networkName: prod-network
  routingMode: GLOBAL
```

### VPC with Private Services Access

A VPC configured for private connectivity to Google managed services, using a GcpProject foreign key reference:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpVpc
metadata:
  name: data-vpc
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpVpc.data-vpc
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: my-project
      field: status.outputs.project_id
  networkName: data-network
  routingMode: GLOBAL
  privateServicesAccess:
    enabled: true
    ipRangePrefixLength: 20
```

This reserves a `/20` block (4,096 IPs) for Google managed services and creates the VPC peering connection. Cloud SQL, Memorystore, and Filestore instances in this project can then be assigned private IPs from the allocated range.

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `networkSelfLink` | `string` | Full self-link URL of the created VPC network (e.g., `projects/my-project/global/networks/my-vpc`) |
| `privateServicesIpRangeName` | `string` | Name of the allocated IP range for private services — only set when `privateServicesAccess.enabled` is `true` |
| `privateServicesIpRangeCidr` | `string` | CIDR of the allocated IP range (e.g., `10.100.0.0/16`) — only set when `privateServicesAccess.enabled` is `true` |

## Related Components

- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project and can enable the Service Networking API required for Private Services Access
- [GcpSubnetwork](/docs/catalog/gcp/gcpsubnetwork) — creates subnets within this VPC with primary and secondary IP ranges
- [GcpRouterNat](/docs/catalog/gcp/gcprouternat) — provides Cloud NAT for private workload outbound internet access
- [GcpGkeCluster](/docs/catalog/gcp/gcpgkecluster) — deploys a GKE cluster into this VPC
- [GcpCloudSql](/docs/catalog/gcp/gcpcloudsql) — deploys Cloud SQL instances that can use Private Services Access for private IP connectivity
