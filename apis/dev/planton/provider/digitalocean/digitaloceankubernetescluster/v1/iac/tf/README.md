# DigitalOcean Kubernetes Cluster - Terraform Module

This Terraform module provisions and manages DigitalOcean Kubernetes (DOKS) clusters based on the Planton DigitalOceanKubernetesCluster specification.

## Overview

The module creates a fully-configured DOKS cluster with support for:
- High availability control plane
- Node pool autoscaling
- VPC isolation
- Control plane firewall
- Container registry integration
- Scheduled maintenance windows
- Automatic security patch upgrades

## Prerequisites

1. **Terraform**: Version 1.5 or higher
2. **DigitalOcean Account**: Active account with API access
3. **DigitalOcean API Token**: Personal access token with read/write permissions
4. **VPC**: Pre-existing VPC in the target region

## Installation

### Initialize Terraform

```bash
terraform init
```

This will download the DigitalOcean provider (~2.0).

### Set Required Variables

The module requires a DigitalOcean API token for authentication:

```bash
export TF_VAR_digitalocean_token="your-do-api-token-here"
```

**Security Note:** Never commit API tokens to version control. Use environment variables, Terraform Cloud, or a secrets manager.

## Usage

### Basic Example

```hcl
module "doks_cluster" {
  source = "./path/to/module"

  metadata = {
    name = "my-cluster"
    env  = "dev"
  }

  spec = {
    cluster_name       = "my-cluster"
    region             = "nyc1"
    kubernetes_version = "1.29"
    vpc = {
      value = "your-vpc-uuid"
    }
    default_node_pool = {
      size       = "s-2vcpu-4gb"
      node_count = 3
    }
  }

  digitalocean_token = var.digitalocean_token
}
```

### Production Example with All Features

```hcl
module "production_cluster" {
  source = "./path/to/module"

  metadata = {
    name = "prod-cluster"
    env  = "production"
    tags = ["team:platform", "criticality:high"]
  }

  spec = {
    cluster_name       = "prod-cluster"
    region             = "sfo3"
    kubernetes_version = "1.29"
    
    vpc = {
      value = "your-production-vpc-uuid"
    }
    
    # High availability control plane
    highly_available = true
    
    # Auto-upgrade settings
    auto_upgrade          = true
    disable_surge_upgrade = false
    
    # Maintenance window (Sunday 2 AM UTC)
    maintenance_window = "sunday=02:00"
    
    # Enable DOCR integration
    registry_integration = true
    
    # Restrict API access to specific IPs
    control_plane_firewall_allowed_ips = [
      "203.0.113.10/32",  # Office IP
      "198.51.100.0/24"   # CI/CD subnet
    ]
    
    # Tags for organization
    tags = ["env:production", "compliance:required"]
    
    # Autoscaling node pool
    default_node_pool = {
      size       = "s-4vcpu-8gb"
      node_count = 5
      auto_scale = true
      min_nodes  = 5
      max_nodes  = 10
    }
  }

  digitalocean_token = var.digitalocean_token
}
```

## Variables

### Required Variables

| Variable | Type | Description |
|----------|------|-------------|
| `metadata` | object | Resource metadata (name, env, tags, etc.) |
| `spec` | object | Cluster specification (see spec variables below) |
| `digitalocean_token` | string | DigitalOcean API token (sensitive) |

### Spec Variables

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `cluster_name` | string | Yes | - | Unique cluster name in DigitalOcean account |
| `region` | string | Yes | - | DigitalOcean region (e.g., "nyc1", "sfo3") |
| `kubernetes_version` | string | Yes | - | Kubernetes version (e.g., "1.29") |
| `vpc.value` | string | Yes | - | VPC UUID for cluster networking |
| `highly_available` | bool | No | false | Enable HA control plane ($40/month) |
| `auto_upgrade` | bool | No | false | Enable automatic patch upgrades |
| `disable_surge_upgrade` | bool | No | false | Disable surge upgrades (not recommended) |
| `maintenance_window` | string | No | "" | Maintenance window (e.g., "sunday=02:00") |
| `registry_integration` | bool | No | false | Enable DOCR integration |
| `control_plane_firewall_allowed_ips` | list(string) | No | [] | Allowed IPs for API access (CIDR notation) |
| `tags` | list(string) | No | [] | Tags for resource organization |
| `default_node_pool.size` | string | Yes | - | Droplet size (e.g., "s-2vcpu-4gb") |
| `default_node_pool.node_count` | number | Yes | - | Initial node count |
| `default_node_pool.auto_scale` | bool | No | false | Enable autoscaling |
| `default_node_pool.min_nodes` | number | No | 0 | Minimum nodes (if autoscaling) |
| `default_node_pool.max_nodes` | number | No | 0 | Maximum nodes (if autoscaling) |

## Outputs

| Output | Description | Sensitive |
|--------|-------------|-----------|
| `cluster_id` | Cluster UUID | No |
| `kubeconfig` | Base64-encoded kubeconfig | Yes |
| `api_server_endpoint` | Kubernetes API server URL | No |
| `cluster_urn` | DigitalOcean URN | No |
| `cluster_status` | Current cluster status | No |
| `created_at` | Creation timestamp | No |
| `updated_at` | Last update timestamp | No |
| `node_pool_id` | Default node pool ID | No |
| `ipv4_address` | Control plane public IP | No |

## Accessing Outputs

```bash
# Get cluster ID
terraform output cluster_id

# Get kubeconfig (sensitive output)
terraform output -raw kubeconfig > kubeconfig.yaml
export KUBECONFIG=kubeconfig.yaml

# Get API endpoint
terraform output api_server_endpoint
```

## Deployment Workflow

### 1. Plan

```bash
terraform plan -out=tfplan
```

Review the plan to verify:
- Cluster configuration matches expectations
- VPC UUID is correct
- Node pool sizing is appropriate
- Firewall rules are correctly configured

### 2. Apply

```bash
terraform apply tfplan
```

**Deployment time:** 3-5 minutes for cluster provisioning.

### 3. Verify

```bash
# Export kubeconfig
terraform output -raw kubeconfig > ~/.kube/doks-config
export KUBECONFIG=~/.kube/doks-config

# Verify cluster connectivity
kubectl get nodes
kubectl cluster-info
```

## Common Operations

### Update Kubernetes Version

```hcl
spec = {
  kubernetes_version = "1.30"  # Update to newer version
  # ... other config
}
```

```bash
terraform plan
terraform apply
```

**Note:** The module ignores version changes to prevent drift on auto-upgrades. Remove from `ignore_changes` if manual version upgrades are needed.

### Scale Node Pool

```hcl
spec = {
  default_node_pool = {
    node_count = 5  # Changed from 3
    # ... other config
  }
}
```

```bash
terraform apply
```

### Enable Autoscaling

```hcl
spec = {
  default_node_pool = {
    auto_scale = true
    min_nodes  = 3
    max_nodes  = 10
    # ... other config
  }
}
```

### Update Control Plane Firewall

```hcl
spec = {
  control_plane_firewall_allowed_ips = [
    "203.0.113.10/32",
    "198.51.100.50/32"  # Added new IP
  ]
  # ... other config
}
```

```bash
terraform apply
```

## State Management

### Remote State (Recommended)

For team collaboration and safety, use remote state:

**S3 Backend:**

```hcl
terraform {
  backend "s3" {
    bucket = "my-terraform-state"
    key    = "doks/prod-cluster.tfstate"
    region = "us-east-1"
  }
}
```

**Terraform Cloud:**

```hcl
terraform {
  backend "remote" {
    organization = "my-org"
    workspaces {
      name = "doks-prod"
    }
  }
}
```

### State Locking

Enable state locking to prevent concurrent modifications:
- S3: Use DynamoDB table for locking
- Terraform Cloud: Built-in locking

## Cost Estimation

### Example Costs (Monthly)

**Development Cluster:**
- Control plane: Free (non-HA)
- 2x s-1vcpu-2gb nodes: ~$24
- **Total:** ~$24/month

**Staging Cluster:**
- Control plane: Free (non-HA)
- 3x s-2vcpu-4gb nodes: ~$60
- **Total:** ~$60/month

**Production Cluster:**
- Control plane: $40 (HA, may be waived)
- 5x s-4vcpu-8gb nodes: ~$215
- Load balancers: ~$10 each
- **Total:** ~$255-295/month

Use the [DigitalOcean Pricing Calculator](https://www.digitalocean.com/pricing/calculator) for accurate estimates.

## Security Best Practices

1. **API Token Security**
   - Never commit tokens to version control
   - Use environment variables or secrets managers
   - Rotate tokens regularly

2. **Control Plane Firewall**
   - Always restrict API access in production
   - Use specific CIDR blocks, avoid 0.0.0.0/0
   - Update firewall rules when IPs change

3. **VPC Isolation**
   - Use dedicated VPCs per environment
   - Enable VPC peering for controlled cross-cluster communication

4. **RBAC**
   - Implement least-privilege access controls
   - Use separate kubeconfigs for different teams/roles

5. **Network Policies**
   - Deploy NetworkPolicy resources for pod-to-pod security
   - DOKS includes Cilium CNI with NetworkPolicy support

## Troubleshooting

### Common Issues

#### 1. "VPC not found" Error

**Symptom:** Terraform fails with VPC UUID not found

**Solution:**
```bash
# List available VPCs
doctl vpcs list

# Use correct UUID in spec.vpc.value
```

#### 2. "Kubernetes version not supported"

**Symptom:** Invalid Kubernetes version error

**Solution:**
```bash
# List available versions
doctl kubernetes options versions

# Update spec.kubernetes_version to supported version
```

#### 3. Cannot Access Cluster API

**Symptom:** kubectl commands timeout

**Possible Causes:**
- Control plane firewall blocks your IP
- Kubeconfig not exported correctly

**Solution:**
```bash
# Check your public IP
curl ifconfig.me

# Add IP to control_plane_firewall_allowed_ips
# Re-export kubeconfig
terraform output -raw kubeconfig > ~/.kube/config
```

#### 4. Autoscaling Not Working

**Symptom:** Nodes don't scale despite load

**Solution:**
- Verify `auto_scale = true`
- Ensure `min_nodes` and `max_nodes` are set
- Check pod resource requests are defined

```bash
# Check cluster autoscaler logs
kubectl logs -n kube-system -l app=cluster-autoscaler
```

## Module Maintenance

### Updating Provider Version

```hcl
terraform {
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "~> 2.1"  # Update version
    }
  }
}
```

```bash
terraform init -upgrade
```

### Destroying Clusters

**Warning:** This permanently deletes the cluster and all workloads.

```bash
# Plan destruction
terraform plan -destroy

# Destroy (requires confirmation)
terraform destroy

# Auto-approve (use with caution)
terraform destroy -auto-approve
```

## Reference

- [DigitalOcean Terraform Provider Documentation](https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs)
- [DOKS Documentation](https://docs.digitalocean.com/products/kubernetes/)
- [Planton Documentation](https://docs.planton.dev/)

## Support

For issues or questions:
- Planton: [GitHub Issues](https://github.com/plantonhq/planton/issues)
- DigitalOcean Support: [Support Portal](https://www.digitalocean.com/support)

---

**Version:** 1.0.0  
**Last Updated:** 2025-11-16

