# GCP Bigtable Instance - Terraform Module

Terraform module that provisions a Cloud Bigtable instance with one or more clusters.

## Resources Created

- `google_bigtable_instance` - The Bigtable instance with inline cluster configuration

## Usage

```hcl
module "bigtable_instance" {
  source = "./path/to/module"

  provider_config = {
    service_account_key_base64 = var.gcp_sa_key
  }

  metadata = {
    name = "my-bigtable"
    org  = "my-org"
    env  = "production"
  }

  spec = {
    project_id    = "my-gcp-project"
    instance_name = "my-bigtable-instance"
    display_name  = "Production Bigtable"

    deletion_protection = true
    force_destroy       = false

    clusters = [
      {
        cluster_id   = "cluster-us-central1-a"
        zone         = "us-central1-a"
        storage_type = "SSD"
        autoscaling_config = {
          min_nodes  = 3
          max_nodes  = 30
          cpu_target = 65
        }
      },
      {
        cluster_id   = "cluster-us-central1-b"
        zone         = "us-central1-b"
        storage_type = "SSD"
        autoscaling_config = {
          min_nodes  = 3
          max_nodes  = 30
          cpu_target = 65
        }
      }
    ]
  }
}
```

## Inputs

| Name | Description | Type | Required |
|------|-------------|------|----------|
| provider_config | GCP provider configuration | object | yes |
| metadata | Resource metadata (name, org, env, id) | object | yes |
| spec | Bigtable instance specification | object | yes |

## Outputs

| Name | Description |
|------|-------------|
| instance_id | Fully qualified instance resource name |
| instance_name | Short instance name |
