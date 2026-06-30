# Terraform Implementation - GCP GKE Cluster

This directory contains the Terraform implementation for the **GcpGkeCluster** resource.

## Overview

The Terraform implementation provides a declarative way to deploy production-ready GKE control planes using HCL. It leverages the official Google Terraform provider to create private, VPC-native clusters with Workload Identity, network policies, and auto-upgrades.

## Directory Structure

```
tf/
├── main.tf          # GKE cluster resource
├── variables.tf     # Input variable definitions
├── outputs.tf       # Output value definitions
├── locals.tf        # Local variable definitions
├── provider.tf      # Provider configuration
├── backend.tf       # Backend configuration (optional)
└── README.md        # This file
```

## Files Overview

### main.tf

Contains the main GKE cluster resource definition:

- **`google_container_cluster.cluster`**: GKE control plane with:
  - Private cluster configuration (private nodes, private endpoint settings)
  - VPC-native networking (IP allocation policy with secondary ranges)
  - Workload Identity configuration (conditional, based on disable flag)
  - Network policy configuration (Calico enforcement)
  - Release channel for auto-upgrades
  - Resource labels

### variables.tf

Defines input variables:
- `metadata`: Resource metadata (name, id, org, env)
- `spec`: Cluster specification (project, location, networking, security)

### outputs.tf

Defines output values:
- `endpoint`: Cluster API server IP address
- `cluster_ca_certificate`: CA certificate for cluster verification (sensitive)
- `workload_identity_pool`: Workload Identity pool (if enabled)

### locals.tf

Defines local variables:
- `gcp_labels`: Labels for GCP resources (from metadata)
- `release_channel_map`: Maps proto enum to GKE channel strings
- `release_channel`: Resolved release channel string

### provider.tf

Specifies required providers:
- `google`: v6.19.0

### backend.tf

Backend configuration for remote state storage (optional).

## Prerequisites

- Terraform 1.0+
- GCP account with appropriate permissions
- Existing VPC network with:
  - Subnet with primary IP range for nodes
  - Secondary IP range for pods
  - Secondary IP range for services
- Cloud NAT for private node egress

## Required GCP Permissions

The service account needs:

- `roles/container.admin` (to create GKE clusters)
- `roles/compute.networkUser` (to use VPC subnets)
- `roles/iam.serviceAccountUser` (to use service accounts)

## Usage

### Direct Terraform Usage

1. **Initialize Terraform**:
   ```bash
   terraform init
   ```

2. **Create tfvars file**:
   Create `terraform.tfvars`:
   ```hcl
   metadata = {
     name = "prod-cluster"
     id   = "gke-abc123"
     org  = "acme-corp"
     env = {
       id = "production"
     }
   }

   spec = {
     project_id = {
       value = "my-project-12345"
     }
     location = "us-central1"
     subnetwork_self_link = {
       value = "https://www.googleapis.com/compute/v1/projects/my-project/regions/us-central1/subnetworks/my-subnet"
     }
     cluster_secondary_range_name = {
       value = "pods"
     }
     services_secondary_range_name = {
       value = "services"
     }
     master_ipv4_cidr_block = "172.16.0.0/28"
     enable_public_nodes    = false
     release_channel        = 2  # REGULAR
     disable_network_policy = false
     disable_workload_identity = false
     router_nat_name = {
       value = "my-nat"
     }
   }
   ```

3. **Plan changes**:
   ```bash
   terraform plan
   ```

4. **Apply changes**:
   ```bash
   terraform apply
   ```

5. **View outputs**:
   ```bash
   terraform output endpoint
   terraform output workload_identity_pool
   ```

6. **Destroy resources**:
   ```bash
   terraform destroy
   ```

### Using Planton CLI (Recommended)

```bash
planton terraform apply --manifest gcpgkecluster.yaml --stack org/project/stack
```

## Input Variables

### metadata

```hcl
metadata = {
  name = "prod-cluster"    # Cluster name
  id   = "gke-abc123"      # Resource ID
  org  = "acme-corp"       # Organization
  env = {
    id = "production"      # Environment
  }
}
```

### spec

```hcl
spec = {
  # GCP project where cluster is created
  project_id = {
    value = "my-project-12345"
  }

  # Regional or zonal location
  location = "us-central1"  # Regional for HA, or "us-central1-a" for zonal

  # VPC subnet (must exist with secondary ranges)
  subnetwork_self_link = {
    value = "https://www.googleapis.com/compute/v1/projects/my-project/regions/us-central1/subnetworks/my-subnet"
  }

  # Secondary IP range for pods
  cluster_secondary_range_name = {
    value = "pods"
  }

  # Secondary IP range for services
  services_secondary_range_name = {
    value = "services"
  }

  # Control plane private endpoint CIDR (must be /28)
  master_ipv4_cidr_block = "172.16.0.0/28"

  # Whether nodes have public IPs (default: false)
  enable_public_nodes = false

  # Release channel (0=unspecified, 1=RAPID, 2=REGULAR, 3=STABLE, 4=NONE)
  release_channel = 2  # REGULAR (default)

  # Disable network policies (default: false = enabled)
  disable_network_policy = false

  # Disable Workload Identity (default: false = enabled)
  disable_workload_identity = false

  # Cloud NAT name (for private node egress)
  router_nat_name = {
    value = "my-nat"
  }
}
```

## Release Channels

| Value | Enum | Channel | Description |
|-------|------|---------|-------------|
| 0 | Unspecified | REGULAR | Default to REGULAR |
| 1 | RAPID | RAPID | Bleeding-edge, new versions quickly |
| 2 | REGULAR | REGULAR | Production recommended |
| 3 | STABLE | STABLE | Conservative, battle-tested versions |
| 4 | NONE | UNSPECIFIED | Manual upgrades (not recommended) |

## Outputs

After applying, these outputs are available:

```bash
terraform output endpoint               # API server IP
terraform output workload_identity_pool # PROJECT_ID.svc.id.goog
```

The `cluster_ca_certificate` output is sensitive and not displayed by default:

```bash
terraform output -raw cluster_ca_certificate
```

## Examples

### Minimal Production Cluster

```hcl
metadata = {
  name = "prod-cluster"
  id   = "gke-001"
  org  = "acme"
  env  = { id = "prod" }
}

spec = {
  project_id                     = { value = "acme-prod-12345" }
  location                       = "us-central1"  # Regional
  subnetwork_self_link           = { value = "https://..." }
  cluster_secondary_range_name   = { value = "pods" }
  services_secondary_range_name  = { value = "services" }
  master_ipv4_cidr_block         = "172.16.0.0/28"
  enable_public_nodes            = false
  release_channel                = 2  # REGULAR
  disable_network_policy         = false
  disable_workload_identity      = false
  router_nat_name                = { value = "prod-nat" }
}
```

### Development Cluster (Cost-Optimized)

```hcl
metadata = {
  name = "dev-cluster"
  id   = "gke-dev-001"
  org  = "acme"
  env  = { id = "dev" }
}

spec = {
  project_id                     = { value = "acme-dev-67890" }
  location                       = "us-central1-a"  # Zonal (cheaper)
  subnetwork_self_link           = { value = "https://..." }
  cluster_secondary_range_name   = { value = "dev-pods" }
  services_secondary_range_name  = { value = "dev-services" }
  master_ipv4_cidr_block         = "172.16.0.16/28"
  enable_public_nodes            = false
  release_channel                = 1  # RAPID (bleeding-edge)
  disable_network_policy         = true  # Simplify dev
  disable_workload_identity      = true  # Not needed for dev
  router_nat_name                = { value = "dev-nat" }
}
```

### High-Security Cluster (Stable Channel)

```hcl
metadata = {
  name = "secure-prod"
  id   = "gke-secure-001"
  org  = "finance-corp"
  env  = { id = "prod" }
}

spec = {
  project_id                     = { value = "finance-prod-99999" }
  location                       = "us-east1"
  subnetwork_self_link           = { value = "https://..." }
  cluster_secondary_range_name   = { value = "secure-pods" }
  services_secondary_range_name  = { value = "secure-services" }
  master_ipv4_cidr_block         = "172.16.0.32/28"
  enable_public_nodes            = false
  release_channel                = 3  # STABLE (conservative)
  disable_network_policy         = false  # Maximum security
  disable_workload_identity      = false  # Required for compliance
  router_nat_name                = { value = "secure-nat" }
}
```

## State Management

Terraform state can be stored:

- **Locally**: Default, state in `terraform.tfstate`
- **Remote**: Use backend configuration (GCS, S3, etc.)

Example GCS backend (edit `backend.tf`):

```hcl
terraform {
  backend "gcs" {
    bucket = "my-terraform-state"
    prefix = "gke-clusters/prod"
  }
}
```

## Verification

### Using gcloud

1. **List clusters**:
   ```bash
   gcloud container clusters list --project=my-project
   ```

2. **Get cluster details**:
   ```bash
   gcloud container clusters describe prod-cluster \
     --region=us-central1 \
     --project=my-project
   ```

3. **Configure kubectl**:
   ```bash
   gcloud container clusters get-credentials prod-cluster \
     --region=us-central1 \
     --project=my-project
   ```

4. **Verify cluster access**:
   ```bash
   kubectl cluster-info
   kubectl get nodes
   ```

### Using Terraform

```bash
# Show all outputs
terraform output

# Show specific output
terraform output endpoint

# Show all resources
terraform state list

# Show specific resource
terraform state show google_container_cluster.cluster
```

## Troubleshooting

### Error: "Invalid master_ipv4_cidr_block"

- Must be exactly `/28` CIDR (16 IPs)
- Must be RFC1918 private range (10.0.0.0/8, 172.16.0.0/12, or 192.168.0.0/16)
- Cannot overlap with VPC primary range, pod range, service range, or peered VPC ranges

**Fix**: Choose a non-overlapping /28 range, e.g., `172.16.0.0/28`

### Error: "Subnetwork does not exist"

- Verify the subnetwork exists in the correct project and region
- Check the `subnetwork_self_link` format (must be full GCP resource URL)
- Ensure service account has `roles/compute.networkUser` on the subnet

### Error: "Secondary range not found"

- Verify the subnet has secondary IP ranges defined
- Check the range names match exactly (case-sensitive)
- Use `gcloud compute networks subnets describe` to list ranges

### Error: "Workload Identity pool not created"

- Check that `disable_workload_identity = false`
- Verify the project ID is correct
- Ensure Workload Identity is enabled on the project:
  ```bash
  gcloud container clusters update prod-cluster \
    --region=us-central1 \
    --workload-pool=PROJECT_ID.svc.id.goog
  ```

### Permission Denied

Ensure service account has required roles:

```bash
gcloud projects add-iam-policy-binding my-project \
  --member="serviceAccount:terraform@project.iam.gserviceaccount.com" \
  --role="roles/container.admin"

gcloud projects add-iam-policy-binding my-project \
  --member="serviceAccount:terraform@project.iam.gserviceaccount.com" \
  --role="roles/compute.networkUser"
```

## Best Practices

1. **Use Remote State**: Store state in GCS for team collaboration and locking
2. **Separate Environments**: Use Terraform workspaces or separate directories for dev/staging/prod
3. **Regional Clusters for Production**: Zonal clusters lack control plane HA
4. **Plan Master CIDR Allocation**: Use a dedicated /16 subnet (e.g., `172.16.0.0/16`) for all cluster master CIDRs
5. **Enable Workload Identity**: Avoid shared node-level service account keys
6. **Enable Network Policies**: Implement microsegmentation from day one
7. **Use Release Channels**: Prefer REGULAR for production (auto-upgrades with balanced schedule)
8. **Version Control Everything**: Keep Terraform files in Git, tfvars files in gitignored or encrypted

## Integration with Node Pools

After creating a cluster, provision node pools using the `GcpGkeNodePool` resource:

1. The cluster is created with `remove_default_node_pool = true`
2. No compute nodes exist yet (only control plane)
3. Create `GcpGkeNodePool` resources referencing this cluster
4. Node pools have independent lifecycles (can be updated without touching control plane)

## Integration with Planton

This Terraform module integrates with Planton CLI, which:

- Converts YAML manifests to tfvars
- Manages Terraform state remotely
- Handles GCP credentials securely
- Provides consistent multi-cloud interface
- Orchestrates dependency ordering (VPC → Subnet → NAT → Cluster)

## Next Steps

After creating a GKE cluster:

1. **Provision Node Pools**: Use `GcpGkeNodePool` resources to add compute nodes
2. **Configure kubectl**: Run `gcloud container clusters get-credentials`
3. **Set Up Workload Identity**: Bind Kubernetes Service Accounts to Google Service Accounts
4. **Deploy Applications**: Apply Kubernetes manifests
5. **Implement Network Policies**: Define microsegmentation rules

For comprehensive design rationale and production best practices, see the [research documentation](../../docs/README.md).

For real-world configuration examples, see [examples.md](../../examples.md).

