# DigitalOcean Kubernetes Node Pool - Terraform Module

This Terraform module provisions and manages node pools for DigitalOcean Kubernetes (DOKS) clusters based on the Planton DigitalOceanKubernetesNodePool specification.

## Overview

The module creates fully-configured DOKS node pools with support for:
- Fixed-size or autoscaling configurations
- Kubernetes labels for pod scheduling
- Kubernetes taints for workload isolation
- DigitalOcean tags for cost attribution
- Production-ready defaults

## Prerequisites

1. **Terraform**: Version 1.5 or higher
2. **DigitalOcean Account**: Active account with API access
3. **DigitalOcean API Token**: Personal access token with read/write permissions
4. **Existing DOKS Cluster**: Pre-existing cluster to add the node pool to

## Installation

### Initialize Terraform

```bash
terraform init
```

This will download the DigitalOcean provider (~2.0).

### Set Required Variables

```bash
export TF_VAR_digitalocean_token="your-do-api-token-here"
```

**Security Note:** Never commit API tokens to version control.

## Usage

### Basic Fixed-Size Pool

```hcl
module "app_workers" {
  source = "./path/to/module"

  metadata = {
    name = "app-workers"
    env  = "production"
  }

  spec = {
    node_pool_name = "app-workers"
    cluster = {
      value = "your-cluster-id"
    }
    size       = "s-2vcpu-4gb"
    node_count = 3
  }

  digitalocean_token = var.digitalocean_token
}
```

### Autoscaling Pool with Labels

```hcl
module "autoscale_workers" {
  source = "./path/to/module"

  metadata = {
    name = "autoscale-workers"
    tags = ["team:platform"]
  }

  spec = {
    node_pool_name = "autoscale-workers"
    cluster = {
      value = "your-cluster-id"
    }
    size       = "s-4vcpu-8gb"
    node_count = 5
    auto_scale = true
    min_nodes  = 3
    max_nodes  = 10
    labels = {
      workload = "application"
      env      = "production"
    }
    tags = ["env:production", "autoscaling:enabled"]
  }

  digitalocean_token = var.digitalocean_token
}
```

### GPU Pool with Taints

```hcl
module "gpu_workers" {
  source = "./path/to/module"

  metadata = {
    name = "gpu-workers"
  }

  spec = {
    node_pool_name = "gpu-workers"
    cluster = {
      value = "your-cluster-id"
    }
    size       = "g-4vcpu-16gb"
    node_count = 2
    auto_scale = true
    min_nodes  = 0
    max_nodes  = 4
    labels = {
      hardware = "gpu"
    }
    taints = [
      {
        key    = "nvidia.com/gpu"
        value  = "true"
        effect = "NoSchedule"
      }
    ]
    tags = ["hardware:gpu"]
  }

  digitalocean_token = var.digitalocean_token
}
```

## Variables

### Required Variables

| Variable | Type | Description |
|----------|------|-------------|
| `metadata` | object | Resource metadata (name, env, tags, etc.) |
| `spec` | object | Node pool specification (see spec variables) |
| `digitalocean_token` | string | DigitalOcean API token (sensitive) |

### Spec Variables

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `node_pool_name` | string | Yes | - | Unique pool name within cluster |
| `cluster.value` | string | Yes | - | Cluster ID |
| `size` | string | Yes | - | Droplet size (e.g., "s-2vcpu-4gb") |
| `node_count` | number | Yes | - | Initial/desired node count |
| `auto_scale` | bool | No | false | Enable autoscaling |
| `min_nodes` | number | No | 0 | Minimum nodes (if autoscaling) |
| `max_nodes` | number | No | 0 | Maximum nodes (if autoscaling) |
| `labels` | map(string) | No | {} | Kubernetes labels |
| `taints` | list(object) | No | [] | Kubernetes taints |
| `tags` | list(string) | No | [] | DigitalOcean tags |

### Taint Object Structure

```hcl
{
  key    = string  # Taint key
  value  = string  # Taint value
  effect = string  # NoSchedule, PreferNoSchedule, or NoExecute
}
```

## Outputs

| Output | Description |
|--------|-------------|
| `node_pool_id` | Node pool UUID |
| `node_ids` | List of node IDs |
| `node_names` | List of node names |
| `actual_node_count` | Current number of nodes |
| `pool_name` | Pool name |
| `pool_size` | Droplet size |

## Deployment Workflow

### 1. Plan

```bash
terraform plan -out=tfplan
```

### 2. Apply

```bash
terraform apply tfplan
```

**Deployment time:** ~2-3 minutes per node pool.

### 3. Verify

```bash
# Check outputs
terraform output node_pool_id

# Verify via kubectl
kubectl get nodes -l planton-resource-name=app-workers
```

## Common Operations

### Update Node Count

```hcl
spec = {
  node_count = 5  # Changed from 3
  # ... other config
}
```

```bash
terraform apply
```

### Enable Autoscaling

```hcl
spec = {
  auto_scale = true
  min_nodes  = 3
  max_nodes  = 10
  # ... other config
}
```

### Add Labels

```hcl
spec = {
  labels = {
    workload = "web"
    tier     = "frontend"
  }
  # ... other config
}
```

### Add Taints

```hcl
spec = {
  taints = [
    {
      key    = "dedicated"
      value  = "backend"
      effect = "NoSchedule"
    }
  ]
  # ... other config
}
```

## State Management

### Remote State (Recommended)

```hcl
terraform {
  backend "s3" {
    bucket = "my-terraform-state"
    key    = "doks/node-pools/app-workers.tfstate"
    region = "us-east-1"
  }
}
```

## Best Practices

1. **Separate Pools by Workload Type**
   - System, application, batch, GPU pools
   - Prevents resource contention

2. **Use Autoscaling for Variable Loads**
   - Set appropriate min/max boundaries
   - Monitor cost implications

3. **Apply Labels for Scheduling**
   - Use labels with nodeSelector or affinity
   - Document label conventions

4. **Use Taints for Isolation**
   - GPU nodes, dedicated pools
   - Prevents unwanted pods

5. **Tag for Cost Attribution**
   - Team, environment, cost-center
   - Enable granular billing analysis

## Troubleshooting

### "Cluster not found"

**Solution:**
```bash
# Verify cluster ID
doctl kubernetes cluster list

# Update spec.cluster.value
```

### "Droplet size not available"

**Solution:**
```bash
# List available sizes
doctl kubernetes options sizes

# Choose valid size for your region
```

### "Invalid taint effect"

**Solution:** Effect must be one of:
- `NoSchedule`
- `PreferNoSchedule`
- `NoExecute`

### Pods Not Scheduling

**Check tolerations:**
```yaml
tolerations:
  - key: nvidia.com/gpu
    operator: "Equal"
    value: "true"
    effect: NoSchedule
```

## Cost Estimation

**Example Costs (Monthly):**

- Small pool: 3 × s-2vcpu-4gb @ $20/node = ~$60
- Medium pool: 5 × s-4vcpu-8gb @ $43/node = ~$215  
- GPU pool: 2 × g-4vcpu-16gb @ ~$150/node = ~$300

Autoscaling reduces costs by scaling down during off-peak hours.

## Cleanup

```bash
# Destroy node pool
terraform destroy

# Warning: This will drain and delete all nodes
```

## Reference

- [DigitalOcean Terraform Provider](https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs)
- [DOKS Node Pools](https://docs.digitalocean.com/products/kubernetes/how-to/add-node-pools/)
- [Kubernetes Taints](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/)
- [Planton Documentation](https://docs.planton.dev/)

---

**Version:** 1.0.0  
**Last Updated:** 2025-11-16

