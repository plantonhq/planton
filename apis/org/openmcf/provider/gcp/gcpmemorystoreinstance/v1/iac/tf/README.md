# GcpMemorystoreInstance Terraform Module

Provisions a Google Cloud Memorystore instance using the new-generation API
(`google_memorystore_instance`). Supports Valkey engine, native sharding,
PSC networking, RDB/AOF persistence, CMEK encryption, and automated backups.

## Requirements

| Name | Version |
|------|---------|
| google | ~> 6.0 |

The `~> 6.0` provider version is required for the `google_memorystore_instance`
resource and the `desired_auto_created_endpoints` field.

## Usage

This module is designed to be driven by OpenMCF manifests. See the component's
`examples.md` for YAML manifests that translate into Terraform variable values.

```hcl
module "memorystore" {
  source = "./path/to/module"

  metadata = {
    name = "my-cache"
  }

  spec = {
    project_id    = { value = "my-gcp-project" }
    instance_name = "my-cache"
    location      = "us-central1"
    shard_count   = 3
    mode          = "CLUSTER"
    node_type     = "HIGHMEM_MEDIUM"
    replica_count = 1

    psc_auto_connections = [{
      network    = { value = "projects/my-project/global/networks/my-vpc" }
      project_id = { value = "my-project" }
    }]
  }
}
```

## Outputs

| Name | Description |
|------|-------------|
| discovery_address | IP address of the instance's discovery endpoint |
| discovery_port | Port of the instance's discovery endpoint |
| instance_uid | Server-generated unique identifier |
| node_size_gb | Memory size per node in GB |
