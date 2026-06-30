# Terraform Module - GCP GKE Node Pool

This directory contains the Terraform implementation for the **GcpGkeNodePool** resource, enabling declarative node pool management using HCL.

## Overview

The Terraform module provisions GKE node pools with production-ready defaults:
- **Autoscaling or fixed size**: Scale-to-zero support or predictable capacity
- **Spot VMs**: Cost optimization for fault-tolerant workloads (up to 91% savings)
- **Auto-upgrade and auto-repair**: Enabled by default with customizable settings
- **Custom service accounts**: Least-privilege IAM for security
- **Flexible disk options**: Standard, SSD, or Balanced persistent disks
- **Node labels**: For Kubernetes workload scheduling via `nodeSelector`

## Directory Structure

```
tf/
├── main.tf          # Node pool resource and data sources
├── variables.tf     # Input variable definitions
├── outputs.tf       # Output value definitions
├── locals.tf        # Local variable definitions
├── provider.tf      # Provider configuration
├── backend.tf       # Backend configuration (optional)
└── README.md        # This file
```

## Files Overview

### main.tf

Contains the main resource definitions:

- **`data.google_container_cluster.cluster`**: Looks up parent GKE cluster to fetch location information
- **`google_container_node_pool.node_pool`**: The node pool resource with:
  - Node configuration (machine type, disk, image, service account)
  - Autoscaling or fixed size configuration
  - Management settings (auto-upgrade, auto-repair)
  - Upgrade settings (surge and unavailability controls)
  - Labels and network tags

### variables.tf

Defines input variables:
- `metadata`: Resource metadata (name, id, org, env, labels, tags)
- `spec`: Node pool specification matching `spec.proto`:
  - Cluster references (`cluster_project_id`, `cluster_name`)
  - Machine configuration (`machine_type`, `disk_size_gb`, `disk_type`, `image_type`)
  - Sizing (`node_count` or `autoscaling` - mutually exclusive)
  - Management settings (`management.disable_auto_upgrade`, `management.disable_auto_repair`)
  - Cost optimization (`spot`)
  - Custom service account and node labels

### outputs.tf

Defines output values:
- `node_pool_name`: Name of the created node pool
- `instance_group_urls`: URLs of managed instance groups (one per zone for regional clusters)
- `min_nodes`: Effective minimum size
- `max_nodes`: Effective maximum size
- `current_node_count`: Current number of nodes (managed by autoscaler if enabled)

### locals.tf

Defines local variables:
- `resource_id`: Derived from metadata.id or metadata.name
- `final_gcp_labels`: Merged labels (base + org + env + resource_id)
- `merged_node_labels`: Planton labels + user-specified node labels
- `auto_upgrade_enabled` / `auto_repair_enabled`: Inverted from disable flags
- `network_tag`: GKE network tag (`gke-<cluster-name>`)
- `oauth_scopes`: Minimal OAuth scopes for nodes

### provider.tf

Specifies required providers:
- `google`: ~> 6.0

### backend.tf

Optional backend configuration. By default, Terraform stores state locally. For team collaboration, configure remote backends (GCS, S3, Terraform Cloud).

## Prerequisites

### Tools

- **Terraform** 1.0+ or **OpenTofu** 1.6+ ([install Terraform](https://www.terraform.io/downloads) | [install OpenTofu](https://opentofu.org/docs/intro/install/))
- **gcloud CLI** ([install](https://cloud.google.com/sdk/docs/install))

### Authentication

Authenticate with GCP:

```bash
gcloud auth application-default login
```

This creates credentials at `~/.config/gcloud/application_default_credentials.json` that Terraform uses automatically.

### Required Permissions

The authenticated account needs:
- `roles/container.admin` (GKE cluster and node pool management)
- `roles/compute.viewer` (to lookup cluster details)
- `roles/iam.serviceAccountUser` (if using custom service accounts)

## Standalone Usage

### 1. Create Variable Values File

Create a `terraform.tfvars` file with your node pool configuration:

```hcl
metadata = {
  name = "prod-general"
  org  = "my-org"
  env  = "prod"
}

spec = {
  cluster_project_id = {
    value = "my-gcp-project"
  }
  
  cluster_name = {
    value = "prod-cluster"
  }
  
  machine_type  = "n2-standard-4"
  disk_size_gb  = 100
  disk_type     = "pd-ssd"
  image_type    = "COS_CONTAINERD"
  service_account = "gke-prod-sa@my-gcp-project.iam.gserviceaccount.com"
  
  autoscaling = {
    min_nodes       = 3
    max_nodes       = 10
    location_policy = "BALANCED"
  }
  
  node_labels = {
    "workload-tier" = "general"
    "environment"   = "production"
  }
  
  management = {
    disable_auto_upgrade = false
    disable_auto_repair  = false
  }
  
  spot = false
}
```

### 2. Initialize Terraform

```bash
cd iac/tf
terraform init
```

This downloads the Google provider and initializes the backend.

### 3. Plan Changes

Preview what Terraform will create:

```bash
terraform plan
```

Terraform shows a detailed execution plan with all resources to be created/modified/destroyed.

### 4. Apply Configuration

Deploy the node pool:

```bash
terraform apply
```

Terraform prompts for confirmation before applying changes. Type `yes` to proceed.

### 5. View Outputs

```bash
terraform output
```

Outputs include:
- `node_pool_name`: Name of the created node pool
- `instance_group_urls`: URLs of managed instance groups
- `current_node_count`: Current number of nodes
- `min_nodes`: Effective minimum size
- `max_nodes`: Effective maximum size

### 6. Update Configuration

Modify `terraform.tfvars` and re-run:

```bash
terraform plan   # Preview changes
terraform apply  # Apply updates
```

Terraform computes the diff and applies only necessary changes.

### 7. Destroy Resources

Remove the node pool:

```bash
terraform destroy
```

**Warning:** This terminates all nodes and evicts all pods in the pool. Ensure workloads are drained or migrated first.

## Environment Variables

### Required

- **GCP Project**: Set via `GOOGLE_PROJECT` environment variable or in `provider.tf`

### Optional

- **GOOGLE_APPLICATION_CREDENTIALS**: Path to service account JSON (alternative to `gcloud auth`)
- **GOOGLE_REGION**: Default GCP region (node pool uses cluster location automatically)

Example using service account:

```bash
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json
export GOOGLE_PROJECT=my-gcp-project
terraform apply
```

## Example Configurations

### Minimal Fixed-Size Pool

```hcl
metadata = {
  name = "minimal-pool"
}

spec = {
  cluster_project_id = { value = "my-project" }
  cluster_name       = { value = "my-cluster" }
  node_count         = 3
}
```

### Dev/Test Pool with Spot VMs

```hcl
metadata = {
  name = "dev-pool"
  env  = "dev"
}

spec = {
  cluster_project_id = { value = "dev-project" }
  cluster_name       = { value = "dev-cluster" }
  
  machine_type = "e2-medium"
  disk_size_gb = 50
  spot         = true
  
  autoscaling = {
    min_nodes       = 0    # Scale to zero
    max_nodes       = 5
    location_policy = "ANY"  # Hunt for Spot capacity
  }
  
  node_labels = {
    "spot" = "true"
    "cost-tier" = "low"
  }
}
```

### Production High-Memory Pool

```hcl
metadata = {
  name = "high-memory-pool"
  env  = "prod"
}

spec = {
  cluster_project_id = { value = "prod-project" }
  cluster_name       = { value = "prod-cluster" }
  
  machine_type = "n2-highmem-32"
  disk_size_gb = 200
  disk_type    = "pd-ssd"
  
  node_count = 2  # Fixed size for consistent capacity
  
  node_labels = {
    "workload-type" = "high-memory"
    "cache-tier"    = "primary"
  }
}
```

## Validation

### Variable Validation

The module includes built-in validation:

**Mutually exclusive sizing:**
```hcl
validation {
  condition = (var.spec.node_count != null && var.spec.autoscaling == null) ||
              (var.spec.node_count == null && var.spec.autoscaling != null)
  error_message = "Exactly one of node_count or autoscaling must be specified"
}
```

**Disk type validation:**
```hcl
validation {
  condition     = contains(["pd-standard", "pd-ssd", "pd-balanced"], var.spec.disk_type)
  error_message = "disk_type must be one of: pd-standard, pd-ssd, pd-balanced"
}
```

### Testing Configuration

Validate syntax and configuration:

```bash
terraform validate
```

Format code:

```bash
terraform fmt
```

## Troubleshooting

### "Cluster not found" Error

**Cause:** The parent GKE cluster doesn't exist or the name/project is incorrect.

**Solution:**
1. Verify cluster exists:
   ```bash
   gcloud container clusters list --project=my-gcp-project
   ```
2. Check `cluster_name` and `cluster_project_id` in `terraform.tfvars`
3. Ensure cluster is in the expected location (region or zone)

### "Insufficient Permissions" Error

**Cause:** Authenticated account lacks required IAM permissions.

**Solution:**
1. Grant `roles/container.admin`:
   ```bash
   gcloud projects add-iam-policy-binding my-gcp-project \
     --member="user:your-email@example.com" \
     --role="roles/container.admin"
   ```
2. Wait a few minutes for IAM propagation

### "Quota Exceeded" Error

**Cause:** GCP project quota insufficient for requested resources (CPUs, IPs, disk).

**Solution:**
1. Check quotas:
   ```bash
   gcloud compute project-info describe --project=my-gcp-project
   ```
2. Request quota increase: GCP Console → IAM & Admin → Quotas
3. Use smaller `max_nodes` or less resource-intensive `machine_type`

### "Node Configuration Change Requires Recreation"

**Expected behavior:** Changes to `machine_type`, `disk_type`, `disk_size_gb`, or `image_type` require node recreation.

Terraform output shows:
```
  # google_container_node_pool.node_pool must be replaced
  -/+ resource "google_container_node_pool" "node_pool" {
        machine_type = "n2-standard-2" -> "n2-standard-4" # forces replacement
  }
```

GKE cordons, drains, and replaces nodes automatically. To minimize disruption:
- Configure `PodDisruptionBudgets` for critical workloads
- The module sets `max_surge = 2` and `max_unavailable = 1` for gradual rollout
- Perform updates during maintenance windows

### State Lock Error

**Cause:** Multiple `terraform apply` runs or interrupted previous run.

**Solution:**
```bash
# Force unlock (use with caution)
terraform force-unlock <LOCK_ID>
```

Better: Use remote backend with locking (GCS, S3, Terraform Cloud).

## Integration with Planton CLI

This module is typically invoked via the Planton CLI, not standalone. The CLI:
- Manages variable values from YAML manifests
- Handles state storage and locking
- Provides unified UX across all deployment components

For most users, **use the CLI** instead of running Terraform directly:

```bash
planton apply -f node-pool.yaml
```

## Advanced Configuration

### Remote State Backend

For team collaboration, configure a GCS backend in `backend.tf`:

```hcl
terraform {
  backend "gcs" {
    bucket = "my-terraform-state"
    prefix = "gcp-gke-node-pools/prod-general"
  }
}
```

### Custom Provider Configuration

Override provider settings in `provider.tf`:

```hcl
provider "google" {
  project = "my-gcp-project"
  region  = "us-central1"
  
  # Optional: Use service account
  credentials = file("/path/to/service-account.json")
}
```

### Lifecycle Rules

Prevent accidental destruction:

```hcl
resource "google_container_node_pool" "node_pool" {
  # ... configuration ...
  
  lifecycle {
    prevent_destroy = true  # Terraform will refuse to destroy
  }
}
```

## Related Documentation

- **[Component README](../../README.md)**: User-facing overview and quick start
- **[Examples](../../examples.md)**: Working YAML manifests
- **[Research Document](../../docs/README.md)**: Deep dive into GKE node pools
- **[Pulumi Module](../pulumi/README.md)**: Pulumi alternative

## Support

For issues or questions:
- **Planton Repository**: https://github.com/plantonhq/planton
- **Terraform GCP Provider Docs**: https://registry.terraform.io/providers/hashicorp/google/latest/docs
- **GKE Node Pools Documentation**: https://cloud.google.com/kubernetes-engine/docs/concepts/node-pools

---

**Note**: This module requires an existing `GcpGkeCluster` resource. Node pools cannot exist without a parent cluster.

