# KubernetesJob Terraform Module

This Terraform module deploys a KubernetesJob to a Kubernetes cluster.

## Overview

The module creates the following Kubernetes resources:

1. **Namespace** (optional) - Created if `create_namespace: true`
2. **ServiceAccount** - For pod identity and RBAC
3. **ConfigMaps** - From `spec.config_maps`
4. **Secret** - For environment secrets with direct values
5. **Image Pull Secret** (optional) - If Docker credentials are provided
6. **Job** - The main batch workload

## Usage

### Basic Example

```hcl
module "kubernetes_job" {
  source = "./path/to/module"

  metadata = {
    name = "data-migration"
  }

  spec = {
    target_cluster = {
      cluster_name = "my-cluster"
    }
    namespace        = "batch-jobs"
    create_namespace = true
    image = {
      repo = "myregistry/migration-runner"
      tag  = "v1.0.0"
    }
    resources = {
      limits = {
        cpu    = "1000m"
        memory = "2Gi"
      }
      requests = {
        cpu    = "250m"
        memory = "512Mi"
      }
    }
    backoff_limit              = 3
    active_deadline_seconds    = 3600
    ttl_seconds_after_finished = 86400
    command = ["python", "/app/migrate.py"]
  }
}
```

### With Environment Variables

```hcl
module "kubernetes_job" {
  source = "./path/to/module"

  metadata = {
    name = "etl-job"
  }

  spec = {
    target_cluster = {
      cluster_name = "my-cluster"
    }
    namespace        = "data-processing"
    create_namespace = true
    image = {
      repo = "myregistry/etl-runner"
      tag  = "v2.0.0"
    }
    resources = {
      limits = {
        cpu    = "2000m"
        memory = "4Gi"
      }
      requests = {
        cpu    = "500m"
        memory = "1Gi"
      }
    }
    env = {
      variables = {
        INPUT_PATH = {
          value = "/data/input"
        }
        OUTPUT_PATH = {
          value = "/data/output"
        }
      }
      secrets = {
        DATABASE_PASSWORD = {
          secret_ref = {
            name = "db-credentials"
            key  = "password"
          }
        }
      }
    }
  }
}
```

### Parallel Job

```hcl
module "kubernetes_job" {
  source = "./path/to/module"

  metadata = {
    name = "parallel-processor"
  }

  spec = {
    target_cluster = {
      cluster_name = "my-cluster"
    }
    namespace        = "batch"
    create_namespace = true
    image = {
      repo = "myregistry/processor"
      tag  = "v1.0.0"
    }
    resources = {
      limits = {
        cpu    = "1000m"
        memory = "1Gi"
      }
      requests = {
        cpu    = "250m"
        memory = "256Mi"
      }
    }
    parallelism = 5
    completions = 20
    backoff_limit = 3
  }
}
```

## Inputs

| Name | Description | Type | Required |
|------|-------------|------|----------|
| metadata | Resource metadata including name, org, env | object | yes |
| spec | Job specification | object | yes |
| docker_config_json | Docker credentials for private registries | string | no |

## Outputs

| Name | Description |
|------|-------------|
| namespace | The Kubernetes namespace |
| job_name | The name of the Job |
| service_account_name | The service account name |
| resource_id | The unique resource ID |

## Requirements

| Name | Version |
|------|---------|
| terraform | >= 1.0 |
| kubernetes | ~> 2.35 |

## Resources Created

- `kubernetes_namespace` (conditional)
- `kubernetes_service_account`
- `kubernetes_secret` (for env secrets and image pull)
- `kubernetes_config_map` (for each config_map entry)
- `kubernetes_job`

## Notes

- Jobs run to completion and then stop
- Use `ttl_seconds_after_finished` for automatic cleanup
- Set `active_deadline_seconds` to prevent runaway jobs
- Use `backoff_limit` to control retry behavior
