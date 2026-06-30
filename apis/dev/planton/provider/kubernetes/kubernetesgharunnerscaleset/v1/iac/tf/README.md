# KubernetesGhaRunnerScaleSet Terraform Module

This Terraform module deploys a GitHub Actions Runner Scale Set on Kubernetes using the official Helm chart.

## Prerequisites

- Kubernetes cluster with `kubectl` access
- [KubernetesGhaRunnerScaleSetController](../../kubernetesgharunnerscalesetcontroller/v1) installed
- GitHub PAT token or GitHub App credentials

## Usage

```hcl
module "gha_runners" {
  source = "./iac/tf"

  metadata = {
    name = "my-runners"
  }

  spec = {
    namespace = {
      value = "gha-runners"
    }
    create_namespace = true

    github = {
      config_url = "https://github.com/myorg/myrepo"
      pat_token = {
        token = var.github_token
      }
    }

    container_mode = {
      type = "DIND"
    }

    scaling = {
      min_runners = 0
      max_runners = 5
    }
  }
}
```

## Inputs

| Name | Description | Type | Required |
|------|-------------|------|----------|
| metadata | Resource metadata (name, id, org, env) | object | yes |
| spec | Runner scale set specification | object | yes |

## Outputs

| Name | Description |
|------|-------------|
| namespace | Namespace where runners are deployed |
| release_name | Helm release name |
| runner_scale_set_name | Name to use in workflow runs-on |
| pvc_names | Names of created PVCs |

## Persistent Volumes

This module supports creating PVCs for caching:

```hcl
persistent_volumes = [
  {
    name          = "npm-cache"
    size          = "20Gi"
    mount_path    = "/home/runner/.npm"
    storage_class = "standard"
  },
  {
    name       = "gradle-cache"
    size       = "50Gi"
    mount_path = "/home/runner/.gradle"
  }
]
```

