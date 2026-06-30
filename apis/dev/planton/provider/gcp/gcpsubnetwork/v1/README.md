# GCP Subnetwork

## Overview

`GcpSubnetwork` is a Planton resource for creating and managing custom-mode subnets in Google Cloud Platform VPC networks. It provides a simple, declarative interface for defining subnet configuration including primary and secondary IP ranges, regional placement, and Private Google Access settings.

**Key Features:**
- ✅ Custom-mode subnet creation in existing VPC networks
- ✅ Support for secondary IP ranges (for GKE pods and services)
- ✅ Private Google Access configuration
- ✅ Multi-region subnet deployment
- ✅ Automated API enablement (compute.googleapis.com)
- ✅ Pulumi and Terraform/OpenTofu support

## Prerequisites

- Existing GCP project
- Existing custom-mode VPC network (see `GcpVpc` resource)
- IAM permissions: `compute.subnetworks.create`, `compute.subnetworks.get`

## API Reference

### GcpSubnetworkSpec

The main specification for defining a GCP subnetwork:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `project_id` | string | Yes | GCP project ID where subnet will be created |
| `vpc_self_link` | string | Yes | Self-link of the parent VPC network |
| `region` | string | Yes | GCP region (e.g., "us-central1") |
| `ip_cidr_range` | string | Yes | Primary IPv4 CIDR (e.g., "10.0.0.0/20") |
| `secondary_ip_ranges` | list | No | Secondary ranges for alias IPs (GKE) |
| `private_ip_google_access` | bool | No | Enable internal-only access to Google APIs (default: false) |

### GcpSubnetworkSecondaryRange

Structure for secondary IP ranges:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `range_name` | string | Yes | Unique name (1-63 chars, lowercase) |
| `ip_cidr_range` | string | Yes | IPv4 CIDR for this range |

### Stack Outputs

After deployment, the following outputs are available:

| Output | Description |
|--------|-------------|
| `subnetwork_self_link` | Self-link URL of the created subnetwork |
| `region` | Region where the subnetwork resides |
| `ip_cidr_range` | Primary IPv4 CIDR of the subnet |
| `secondary_ranges` | List of secondary ranges with names and CIDRs |

## Quick Start

### Basic Subnet

Create a simple subnet in a region:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpSubnetwork
metadata:
  name: app-subnet-us-central1
spec:
  project_id: my-gcp-project
  vpc_self_link: projects/my-gcp-project/global/networks/my-vpc
  region: us-central1
  ip_cidr_range: 10.100.0.0/20
  private_ip_google_access: true
```

### Subnet with Secondary Ranges (GKE)

For GKE clusters, define secondary ranges for pods and services:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpSubnetwork
metadata:
  name: gke-subnet-us-west1
spec:
  project_id: prod-project
  vpc_self_link: projects/prod-project/global/networks/prod-vpc
  region: us-west1
  ip_cidr_range: 10.50.0.0/20
  secondary_ip_ranges:
    - range_name: pods
      ip_cidr_range: 10.50.16.0/18    # 16,384 IPs for pods
    - range_name: services
      ip_cidr_range: 10.50.80.0/24    # 256 IPs for services
  private_ip_google_access: true
```

## Deployment

### Using Pulumi (Default)

```shell
# Preview changes
planton pulumi preview --manifest manifest.yaml

# Deploy
planton pulumi up --manifest manifest.yaml

# Destroy
planton pulumi destroy --manifest manifest.yaml
```

### Using Terraform/OpenTofu

```shell
# Initialize
planton tofu init --manifest manifest.yaml --backend-type gcs \
  --backend-config="bucket=my-state-bucket" \
  --backend-config="prefix=gcp-subnets/my-subnet"

# Plan
planton tofu plan --manifest manifest.yaml

# Apply
planton tofu apply --manifest manifest.yaml --auto-approve

# Destroy
planton tofu destroy --manifest manifest.yaml --auto-approve
```

## Important Considerations

### Subnet Sizing

Choose subnet sizes carefully—**CIDR ranges cannot be changed after creation**:

- **/24** (256 IPs): Small services, VPC connectors (Cloud Run/Functions)
- **/22** (1,024 IPs): Moderate workloads
- **/20** (4,096 IPs): Production GKE clusters (standard size)
- **/16** (65,536 IPs): Very large environments

**GCP reserves 4 IPs per subnet** (network, broadcast, gateway, DNS).

### Secondary Ranges for GKE

If deploying GKE clusters, define secondary ranges **at subnet creation time**:

- **Pods range**: Size based on `max nodes × 110 pods per node`
  - Example: 100 nodes → use /18 or /17 (16k-32k IPs)
- **Services range**: Size based on Kubernetes Services count
  - Example: 200 services → use /24 (256 IPs)

**Cannot add secondary ranges after GKE cluster creation.**

### Private Google Access

Enable `private_ip_google_access: true` when:
- VMs or GKE nodes don't have external IPs (security best practice)
- Workloads need access to GCS, GCR, BigQuery, or other Google APIs
- Running private GKE clusters

## CIDR Planning Best Practices

1. **Plan for Growth**: Allocate 2-3x your initial capacity
2. **Avoid Overlaps**: Coordinate with on-prem networks and other VPCs
3. **Document Allocations**: Maintain an IPAM spreadsheet
4. **Reserve Ranges**: Leave room for future subnets and environments

Example allocation for production:
```
10.0.0.0/16    Production
  10.0.0.0/20  us-central1 app subnet
  10.0.16.0/20 us-east1 app subnet
  10.0.32.0/18 us-central1 GKE (primary + secondaries)
10.1.0.0/16    Staging
10.2.0.0/16    Development
```

## Common Patterns

### Multi-Region Subnets

Deploy subnets across regions for high availability:

```yaml
# US East
---
apiVersion: gcp.planton.dev/v1
kind: GcpSubnetwork
metadata:
  name: app-subnet-us-east1
spec:
  project_id: prod-project
  vpc_self_link: projects/prod-project/global/networks/prod-vpc
  region: us-east1
  ip_cidr_range: 10.0.0.0/20
  private_ip_google_access: true
---
# Europe West
apiVersion: gcp.planton.dev/v1
kind: GcpSubnetwork
metadata:
  name: app-subnet-eu-west1
spec:
  project_id: prod-project
  vpc_self_link: projects/prod-project/global/networks/prod-vpc
  region: europe-west1
  ip_cidr_range: 10.0.16.0/20
  private_ip_google_access: true
```

### Integration with GKE

Reference secondary ranges in GKE cluster configuration:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GkeCluster
metadata:
  name: prod-cluster
spec:
  # ... other fields
  network_config:
    subnetwork: gke-subnet-us-west1
    cluster_secondary_range_name: pods
    services_secondary_range_name: services
```

## Outputs and References

Access outputs for downstream resources:

```yaml
# Reference in another resource
apiVersion: gcp.planton.dev/v1
kind: GkeCluster
metadata:
  name: my-cluster
spec:
  network_config:
    subnetwork_ref:
      kind: GcpSubnetwork
      name: gke-subnet-us-west1
      field_path: status.outputs.subnetwork_self_link
```

## Troubleshooting

### "Range already in use"

**Cause**: CIDR overlaps with existing subnet in the VPC.

**Solution**: Choose a non-overlapping CIDR range.

### "Cannot add secondary ranges after creation"

**Cause**: Attempted to modify secondary ranges after deployment.

**Solution**: Destroy and recreate subnet with desired ranges (requires GKE cluster recreation if attached).

### "API not enabled"

**Cause**: compute.googleapis.com API disabled in project.

**Solution**: The module automatically enables required APIs. Wait 30-60 seconds for API propagation.

## Further Reading

- [Research Documentation](docs/README.md) - Deep dive into GCP networking, CIDR planning, and deployment methods
- [Examples](examples.md) - More practical examples and use cases
- [GCP VPC Documentation](https://cloud.google.com/vpc/docs/vpc) - Official Google Cloud VPC guide
- [GKE Networking](https://cloud.google.com/kubernetes-engine/docs/concepts/alias-ips) - Alias IPs and secondary ranges

## Related Resources

- `GcpVpc` - Create custom-mode VPC networks
- `GcpRouterNat` - NAT gateway for outbound internet access
- `GkeCluster` - Kubernetes clusters on GCP
- `GcpProject` - GCP project management

---

**Need help?** Check the [examples](examples.md) or [research documentation](docs/README.md) for detailed guidance.

