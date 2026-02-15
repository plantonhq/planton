# GCP Cloud Composer Environment - Terraform Module

Terraform module that provisions a Cloud Composer environment — a managed Apache Airflow service for authoring, scheduling, and monitoring data pipelines.

## Provider Requirements

- `google` provider `~> 6.0`

## Resources Created

- `google_composer_environment` — The Composer environment with all configuration including networking, software config, workloads, CMEK, maintenance windows, recovery, and access control

## Usage

```hcl
module "composer_environment" {
  source = "./path/to/module"

  provider_config = {
    service_account_key_base64 = var.gcp_sa_key
  }

  metadata = {
    name = "my-composer"
    org  = "my-org"
    env  = "production"
  }

  spec = {
    project_id      = "my-gcp-project"
    region          = "us-central1"
    environment_name = "my-composer-env"
    environment_size = "ENVIRONMENT_SIZE_MEDIUM"

    node_config = {
      network = "projects/my-gcp-project/global/networks/prod-vpc"
      subnetwork = "projects/my-gcp-project/regions/us-central1/subnetworks/prod-subnet"
    }

    software_config = {
      image_version = "composer-2.9.7-airflow-2.9.3"
      pypi_packages = {
        numpy    = ">=1.21.0"
        pandas   = ">=1.3.0"
        requests = ""
      }
    }

    workloads_config = {
      scheduler = {
        cpu        = 2.0
        memory_gb  = 7.5
        storage_gb = 5.0
        count      = 1
      }
      worker = {
        cpu        = 4.0
        memory_gb  = 15.0
        storage_gb = 10.0
        min_count  = 2
        max_count  = 10
      }
    }

    kms_key_name = "projects/my-gcp-project/locations/us-central1/keyRings/composer-kr/cryptoKeys/composer-key"
  }
}
```

## Inputs

| Name | Description | Type | Required |
|------|-------------|------|----------|
| provider_config | GCP provider configuration | object | yes |
| metadata | Resource metadata (name, org, env, id) | object | yes |
| spec | Composer environment specification | object | yes |

## Outputs

| Name | Description |
|------|-------------|
| environment_id | Fully qualified environment resource name (`projects/{project}/locations/{region}/environments/{name}`) |
| environment_name | Short environment name |
