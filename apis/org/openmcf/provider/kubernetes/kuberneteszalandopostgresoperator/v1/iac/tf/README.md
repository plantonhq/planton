# Zalando Postgres Operator - Terraform Module

This Terraform module deploys the [Zalando Postgres Operator](https://github.com/zalando/postgres-operator) on Kubernetes clusters, providing automated PostgreSQL database management with optional Cloudflare R2 backup support.

## Overview

The Terraform module provides infrastructure-as-code deployment of the Zalando Postgres Operator with:
- Helm chart-based operator deployment
- Configurable CPU and memory resources
- Optional Cloudflare R2 backup configuration
- Automatic label propagation to managed databases

## Prerequisites

| Requirement | Purpose |
|------------|---------|
| **Terraform** | Version 1.0+ |
| **Kubernetes Cluster** | Target for operator deployment |
| **kubectl** | Verification and debugging |
| **Helm Provider** | Helm chart deployment |
| **Cloudflare R2** (optional) | Backup storage |

## Quick Start

### Using OpenMCF CLI

```bash
# Create manifest
cat > postgres-operator.yaml <<EOF
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesZalandoPostgresOperator
metadata:
  name: postgres-operator
spec:
  target_cluster:
    cluster_name: my-gke-cluster
  namespace:
    value: postgres-operator
  create_namespace: true
  container:
    resources:
      requests:
        cpu: 50m
        memory: 100Mi
      limits:
        cpu: 1000m
        memory: 1Gi
EOF

# Deploy with Terraform backend
openmcf terraform apply --manifest postgres-operator.yaml
```

### Standalone Terraform Usage

```bash
cd iac/tf

# Initialize Terraform
terraform init

# Review plan
terraform plan

# Apply configuration
terraform apply
```

## Module Structure

```
iac/tf/
├── provider.tf     # Kubernetes and Helm provider configuration
├── variables.tf    # Input variable definitions
├── locals.tf       # Local value computations
├── main.tf         # Resource definitions
├── outputs.tf      # Output value definitions
└── README.md       # This file
```

## Input Variables

### Metadata

```hcl
variable "metadata" {
  type = object({
    name    = string
    id      = optional(string)
    org     = optional(string)
    env     = optional(string)
    labels  = optional(map(string))
    tags    = optional(list(string))
    version = optional(object({ id = string, message = string }))
  })
}
```

### Spec

```hcl
variable "spec" {
  type = object({
    namespace        = string
    create_namespace = bool
    container = object({
      resources = object({
        limits = object({
          cpu    = string
          memory = string
        })
        requests = object({
          cpu    = string
          memory = string
        })
      })
    })
    backup_config = optional(object({
      bucket        = string
      object_prefix = optional(string)
      schedule      = string
      credentials = object({
        cloudflare_account_id = string
        access_key_id         = string
        secret_access_key     = string
      })
      enable_wal_g_backup        = optional(bool, true)
      enable_wal_g_restore       = optional(bool, true)
      enable_clone_wal_g_restore = optional(bool, true)
    }))
  })
}
```

## Example Configuration

### Basic Deployment (No Backups)

```hcl
module "postgres_operator" {
  source = "./iac/tf"

  metadata = {
    name = "postgres-operator"
    org  = "my-company"
    env  = "development"
  }

  spec = {
    namespace        = "postgres-operator"
    create_namespace = true
    container = {
      resources = {
        requests = {
          cpu    = "50m"
          memory = "100Mi"
        }
        limits = {
          cpu    = "1000m"
          memory = "1Gi"
        }
      }
    }
  }
}
```

### Production with Backups

```hcl
module "postgres_operator" {
  source = "./iac/tf"

  metadata = {
    name = "postgres-operator"
    id   = "pgop-prod-001"
    org  = "acme-corp"
    env  = "production"
  }

  spec = {
    namespace        = "postgres-operator"
    create_namespace = true
    container = {
      resources = {
        requests = {
          cpu    = "100m"
          memory = "256Mi"
        }
        limits = {
          cpu    = "2000m"
          memory = "2Gi"
        }
      }
    }
    backup_config = {
      bucket        = "postgres-backups-prod"
      object_prefix = "prod/pg"
      schedule      = "0 2 * * *"
      credentials = {
        cloudflare_account_id = var.cloudflare_account_id
        access_key_id         = var.r2_access_key_id
        secret_access_key     = var.r2_secret_access_key
      }
      enable_wal_g_backup        = true
      enable_wal_g_restore       = true
      enable_clone_wal_g_restore = true
    }
  }
}
```

## Outputs

| Output | Description |
|--------|-------------|
| `namespace` | Operator namespace (`postgres-operator`) |
| `service` | Operator service name |
| `port_forward_command` | kubectl command for local access |
| `kube_endpoint` | Internal cluster endpoint |
| `ingress_endpoint` | Public endpoint (not applicable) |

## Resources Created

### Always Created

1. **Helm Release**: `postgres-operator` (from Zalando Helm chart)

### Conditionally Created

2. **Kubernetes Namespace**: Created only when `create_namespace: true`

3. **Kubernetes Secret**: `r2-postgres-backup-credentials` (only when `backup_config` provided)
   - Contains R2 access key ID and secret access key
4. **Kubernetes ConfigMap**: `postgres-pod-backup-config`
   - Contains WAL-G environment variables
   - Referenced by operator for all PostgreSQL databases

## Configuration Details

### Namespace Management

Control whether the namespace should be created or use an existing one:

```hcl
spec = {
  namespace        = "postgres-operator"
  create_namespace = true  # Creates the namespace
}
```

**Using Existing Namespace:**

```hcl
spec = {
  namespace        = "existing-postgres-namespace"
  create_namespace = false  # Uses existing namespace
}
```

**Important**: When `create_namespace = false`, the namespace must already exist in the cluster.

### Resource Limits

The operator container resources are fully configurable:

```hcl
spec = {
  namespace        = "postgres-operator"
  create_namespace = true
  container = {
    resources = {
      requests = {
        cpu    = "50m"      # Minimum guaranteed
        memory = "100Mi"
      }
      limits = {
        cpu    = "1000m"    # Maximum allowed
        memory = "1Gi"
      }
    }
  }
}
```

### Backup Configuration

When `backup_config` is provided:

1. **Secret Creation**: Stores R2 credentials
2. **ConfigMap Creation**: Stores WAL-G environment variables
3. **Operator Configuration**: Helm chart configured with `pod_environment_configmap`

All PostgreSQL databases created by the operator automatically inherit the backup configuration.

### Helm Chart

- **Chart**: `postgres-operator/postgres-operator`
- **Version**: `1.12.2`
- **Repository**: https://opensource.zalando.com/postgres-operator/charts/postgres-operator

### Label Inheritance

The operator is configured to propagate these labels to all PostgreSQL databases:

1. `resource`
2. `organization`
3. `environment`
4. `resource_kind`
5. `resource_id`

## Terraform Commands

### Initialize

```bash
terraform init
```

### Plan

```bash
terraform plan \
  -var-file="vars/production.tfvars"
```

### Apply

```bash
terraform apply \
  -var-file="vars/production.tfvars"
```

### Destroy

```bash
terraform destroy \
  -var-file="vars/production.tfvars"
```

### Format

```bash
terraform fmt -recursive
```

### Validate

```bash
terraform validate
```

## Verification

### Check Operator Status

```bash
# Verify namespace
terraform output namespace

# Check operator deployment
kubectl get deployment -n $(terraform output -raw namespace)

# Check operator pods
kubectl get pods -n $(terraform output -raw namespace)

# View operator logs
kubectl logs -n $(terraform output -raw namespace) deployment/postgres-operator -f
```

### Verify Backup Configuration

```bash
# Check if Secret exists
kubectl get secret -n postgres-operator r2-postgres-backup-credentials

# Check if ConfigMap exists
kubectl get configmap -n postgres-operator postgres-pod-backup-config

# View ConfigMap contents
kubectl get configmap -n postgres-operator postgres-pod-backup-config -o yaml
```

### Verify Helm Release

```bash
# List Helm releases
helm list -n postgres-operator

# Get Helm values
helm get values postgres-operator -n postgres-operator

# Get Helm manifest
helm get manifest postgres-operator -n postgres-operator
```

## Troubleshooting

### Operator Not Starting

```bash
# Check Terraform state
terraform show

# Check events
kubectl get events -n postgres-operator --sort-by='.lastTimestamp'

# Check pod status
kubectl describe pod -n postgres-operator -l app.kubernetes.io/name=postgres-operator
```

### Backup ConfigMap Not Created

```bash
# Verify backup_config is set
terraform show | grep backup_config

# Check conditional resource creation
terraform state list | grep backup
```

### Helm Release Failed

```bash
# Check Helm status
helm status postgres-operator -n postgres-operator

# Check Helm history
helm history postgres-operator -n postgres-operator

# Rollback if needed
helm rollback postgres-operator -n postgres-operator
```

## Best Practices

### Use Variable Files

Create environment-specific `.tfvars` files:

**vars/development.tfvars**
```hcl
metadata = {
  name = "postgres-operator"
  env  = "development"
}

spec = {
  container = {
    resources = {
      requests = { cpu = "25m", memory = "50Mi" }
      limits   = { cpu = "500m", memory = "512Mi" }
    }
  }
}
```

**vars/production.tfvars**
```hcl
metadata = {
  name = "postgres-operator"
  env  = "production"
}

spec = {
  container = {
    resources = {
      requests = { cpu = "100m", memory = "256Mi" }
      limits   = { cpu = "2000m", memory = "2Gi" }
    }
  }
  backup_config = {
    bucket   = "postgres-backups-prod"
    schedule = "0 2 * * *"
    credentials = {
      cloudflare_account_id = var.cloudflare_account_id
      access_key_id         = var.r2_access_key_id
      secret_access_key     = var.r2_secret_access_key
    }
  }
}
```

### Use Remote State

Configure S3 or Terraform Cloud for state storage:

```hcl
terraform {
  backend "s3" {
    bucket = "my-terraform-state"
    key    = "postgres-operator/terraform.tfstate"
    region = "us-east-1"
  }
}
```

### Secure Credentials

Use environment variables or external secret management:

```bash
# Set environment variables
export TF_VAR_r2_access_key_id="your-key-id"
export TF_VAR_r2_secret_access_key="your-secret-key"

# Or use external secret providers
# - AWS Secrets Manager
# - HashiCorp Vault
# - Azure Key Vault
```

## Limitations

1. **Single Operator**: Designed for one operator instance per cluster
2. **R2 Only**: Backup currently only supports Cloudflare R2
3. **Helm Chart Version**: Module uses pinned version `1.12.2`
4. **Namespace Existence**: When `create_namespace: false`, namespace must already exist (no validation at plan time)

## Migration from Pulumi

If migrating from Pulumi to Terraform:

1. Export Pulumi state: `pulumi stack export > state.json`
2. Extract resource values
3. Import existing resources:
   ```bash
   terraform import kubernetes_namespace_v1.postgres_operator postgres-operator
   terraform import helm_release.postgres_operator postgres-operator/postgres-operator
   ```
4. Run `terraform plan` to verify no changes

## References

- [Spec Definition](../../spec.proto)
- [Stack Outputs](../../stack_outputs.proto)
- [Examples](../../examples.md)
- [Zalando Operator Docs](https://postgres-operator.readthedocs.io/)
- [Helm Chart](https://github.com/zalando/postgres-operator/tree/master/charts/postgres-operator)
- [Terraform Kubernetes Provider](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs)
- [Terraform Helm Provider](https://registry.terraform.io/providers/hashicorp/helm/latest/docs)

## Support

For issues or questions:
- Review [component README](../../README.md)
- Check [troubleshooting section](#troubleshooting)
- Consult [Pulumi overview](../pulumi/overview.md) for design decisions
- File an issue in the OpenMCF repository

